//go:build cgo && !no_rust
// +build cgo,!no_rust

package scanner

import (
	"runtime"
	"sync"
	"testing"
	"time"
	"fmt"
	"strings"
)

// MemoryLeakDetector helps detect memory leaks in CGO operations
type MemoryLeakDetector struct {
	initialHeapAlloc  uint64
	initialHeapSys    uint64
	initialNumGC      uint32
	initialMallocs    uint64
	initialFrees      uint64
	samples           []MemorySample
	mutex             sync.Mutex
}

// MemorySample represents a memory usage sample at a point in time
type MemorySample struct {
	Timestamp    time.Time
	HeapAlloc    uint64
	HeapSys      uint64
	NumGC        uint32
	Mallocs      uint64
	Frees        uint64
	CGOCalls     int64
	ParserMetrics map[string]interface{}
}

// NewMemoryLeakDetector creates a new memory leak detector
func NewMemoryLeakDetector() *MemoryLeakDetector {
	detector := &MemoryLeakDetector{
		samples: make([]MemorySample, 0),
	}
	
	// Record initial state
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	
	detector.initialHeapAlloc = m.Alloc
	detector.initialHeapSys = m.HeapSys
	detector.initialNumGC = m.NumGC
	detector.initialMallocs = m.Mallocs
	detector.initialFrees = m.Frees
	
	return detector
}

// TakeSample records current memory state
func (d *MemoryLeakDetector) TakeSample(parser *RustSynParser) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	var parserMetrics map[string]interface{}
	if parser != nil {
		parserMetrics = parser.GetMemorySafetyMetrics()
	}
	
	sample := MemorySample{
		Timestamp:     time.Now(),
		HeapAlloc:     m.Alloc,
		HeapSys:       m.HeapSys,
		NumGC:         m.NumGC,
		Mallocs:       m.Mallocs,
		Frees:         m.Frees,
		CGOCalls:      runtime.NumCgoCall(),
		ParserMetrics: parserMetrics,
	}
	
	d.samples = append(d.samples, sample)
}

// AnalyzeLeaks analyzes collected samples for potential memory leaks
func (d *MemoryLeakDetector) AnalyzeLeaks(t *testing.T) LeakAnalysisResult {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	if len(d.samples) < 2 {
		return LeakAnalysisResult{
			HasLeak: false,
			Message: "Insufficient samples for analysis",
		}
	}
	
	// Analyze memory trends
	firstSample := d.samples[0]
	lastSample := d.samples[len(d.samples)-1]
	
	heapGrowth := int64(lastSample.HeapAlloc) - int64(firstSample.HeapAlloc)
	mallocGrowth := int64(lastSample.Mallocs) - int64(firstSample.Mallocs)
	freeGrowth := int64(lastSample.Frees) - int64(firstSample.Frees)
	
	// Calculate allocation rate
	duration := lastSample.Timestamp.Sub(firstSample.Timestamp)
	allocRate := float64(mallocGrowth) / duration.Seconds()
	freeRate := float64(freeGrowth) / duration.Seconds()
	
	result := LeakAnalysisResult{
		HeapGrowth:     heapGrowth,
		MallocGrowth:   mallocGrowth,
		FreeGrowth:     freeGrowth,
		AllocRate:      allocRate,
		FreeRate:       freeRate,
		SampleCount:    len(d.samples),
		Duration:       duration,
	}
	
	// Check for memory leaks
	const maxAcceptableGrowth = 5 * 1024 * 1024 // 5MB
	const maxAllocFreeImbalance = 0.1            // 10% imbalance
	
	if heapGrowth > maxAcceptableGrowth {
		result.HasLeak = true
		result.Message = fmt.Sprintf("Excessive heap growth: %d bytes", heapGrowth)
	} else if mallocGrowth > 0 && freeGrowth > 0 {
		imbalance := float64(mallocGrowth-freeGrowth) / float64(mallocGrowth)
		if imbalance > maxAllocFreeImbalance {
			result.HasLeak = true
			result.Message = fmt.Sprintf("Allocation/free imbalance: %.2f%%", imbalance*100)
		}
	}
	
	// Analyze parser-specific metrics
	if lastSample.ParserMetrics != nil {
		if allocatedStrings, ok := lastSample.ParserMetrics["allocated_c_strings"].(int64); ok && allocatedStrings > 0 {
			result.HasLeak = true
			result.Message = fmt.Sprintf("C string leak detected: %d allocated strings", allocatedStrings)
		}
	}
	
	if !result.HasLeak {
		result.Message = "No significant memory leaks detected"
	}
	
	// Log analysis results
	t.Logf("Memory Leak Analysis Results:")
	t.Logf("  Duration: %v", duration)
	t.Logf("  Heap growth: %d bytes", heapGrowth)
	t.Logf("  Malloc growth: %d", mallocGrowth)
	t.Logf("  Free growth: %d", freeGrowth)
	t.Logf("  Alloc rate: %.2f/sec", allocRate)
	t.Logf("  Free rate: %.2f/sec", freeRate)
	t.Logf("  Has leak: %v", result.HasLeak)
	t.Logf("  Message: %s", result.Message)
	
	return result
}

// LeakAnalysisResult contains the results of memory leak analysis
type LeakAnalysisResult struct {
	HasLeak      bool
	Message      string
	HeapGrowth   int64
	MallocGrowth int64
	FreeGrowth   int64
	AllocRate    float64
	FreeRate     float64
	SampleCount  int
	Duration     time.Duration
}

// TestRustSynParser_MemoryLeakDetection performs comprehensive memory leak detection
func TestRustSynParser_MemoryLeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak detection in short mode")
	}

	parser, err := NewRustSynParser(true)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	detector := NewMemoryLeakDetector()
	detector.TakeSample(parser)

	// Test various scenarios that might cause leaks
	scenarios := []struct {
		name string
		test func(*testing.T, *RustSynParser, *MemoryLeakDetector)
	}{
		{"RepeatedParsing", testRepeatedParsing},
		{"ConcurrentParsing", testConcurrentParsingLeaks},
		{"ErrorConditions", testErrorConditionLeaks},
		{"LargeFiles", testLargeFileLeaks},
		{"UtilityFunctions", testUtilityFunctionLeaks},
		{"ConfigurationChanges", testConfigurationLeaks},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Take sample before test
			detector.TakeSample(parser)
			
			// Run the test scenario
			scenario.test(t, parser, detector)
			
			// Force cleanup and GC
			parser.ForceCleanup()
			runtime.GC()
			runtime.GC() // Run twice to ensure cleanup
			
			// Take sample after test
			detector.TakeSample(parser)
			
			// Check for leaks specific to this scenario
			if err := parser.ValidateMemoryState(); err != nil {
				t.Errorf("Memory state validation failed after %s: %v", scenario.name, err)
			}
		})
	}

	// Final analysis
	time.Sleep(100 * time.Millisecond) // Allow time for any delayed cleanup
	detector.TakeSample(parser)
	
	result := detector.AnalyzeLeaks(t)
	if result.HasLeak {
		t.Errorf("Memory leak detected: %s", result.Message)
	}
}

// testRepeatedParsing tests for leaks in repeated parsing operations
func testRepeatedParsing(t *testing.T, parser *RustSynParser, detector *MemoryLeakDetector) {
	rustCode := `
fn repeated_test_function(value: i32) -> String {
    format!("Repeated test: {}", value)
}

struct RepeatedTestStruct {
    field: i32,
}

impl RepeatedTestStruct {
    fn new(value: i32) -> Self {
        Self { field: value }
    }
}
`

	const iterations = 200
	for i := 0; i < iterations; i++ {
		_, err := parser.ParseRustFile([]byte(rustCode), fmt.Sprintf("repeated_%d.rs", i))
		if err != nil {
			t.Fatalf("Repeated parsing failed at iteration %d: %v", i, err)
		}

		// Take sample every 50 iterations
		if i%50 == 49 {
			detector.TakeSample(parser)
		}
	}
}

// testConcurrentParsingLeaks tests for leaks in concurrent operations
func testConcurrentParsingLeaks(t *testing.T, parser *RustSynParser, detector *MemoryLeakDetector) {
	rustCode := `
fn concurrent_test(id: usize) -> String {
    format!("Concurrent test {}", id)
}
`

	const numGoroutines = 20
	const parsesPerGoroutine = 10

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for j := 0; j < parsesPerGoroutine; j++ {
				fileName := fmt.Sprintf("concurrent_%d_%d.rs", goroutineID, j)
				_, err := parser.ParseRustFile([]byte(rustCode), fileName)
				if err != nil {
					t.Errorf("Concurrent parsing failed: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()
	detector.TakeSample(parser)
}

// testErrorConditionLeaks tests for leaks when error conditions occur
func testErrorConditionLeaks(t *testing.T, parser *RustSynParser, detector *MemoryLeakDetector) {
	invalidCodes := []string{
		"fn broken_function(",
		"struct Incomplete {",
		"impl NonExistent {",
		"use std::invalid::module;",
		"macro_rules! broken {",
		"fn ( invalid syntax ) {",
		"struct { no_name }",
		"fn fn fn fn",
	}

	for i, code := range invalidCodes {
		_, err := parser.ParseRustFile([]byte(code), fmt.Sprintf("error_%d.rs", i))
		// We expect these to fail, but shouldn't leak memory
		if err == nil {
			t.Errorf("Expected error for invalid code %d, but parsing succeeded", i)
		}
	}

	detector.TakeSample(parser)
}

// testLargeFileLeaks tests for leaks when processing large files
func testLargeFileLeaks(t *testing.T, parser *RustSynParser, detector *MemoryLeakDetector) {
	// Generate a large Rust file
	var codeBuilder strings.Builder
	codeBuilder.WriteString("use std::collections::HashMap;\n\n")

	for i := 0; i < 500; i++ {
		codeBuilder.WriteString(fmt.Sprintf(`
pub struct LargeStruct%d {
    field1: i32,
    field2: String,
    field3: Vec<f64>,
}

impl LargeStruct%d {
    pub fn new() -> Self {
        Self {
            field1: %d,
            field2: "large_test_%d".to_string(),
            field3: vec![1.0, 2.0, 3.0],
        }
    }
}
`, i, i, i, i))
	}

	largeCode := codeBuilder.String()
	
	// Parse the large file multiple times
	for i := 0; i < 5; i++ {
		_, err := parser.ParseRustFile([]byte(largeCode), fmt.Sprintf("large_%d.rs", i))
		if err != nil {
			t.Errorf("Large file parsing failed: %v", err)
		}
		
		detector.TakeSample(parser)
	}
}

// testUtilityFunctionLeaks tests utility functions for memory leaks
func testUtilityFunctionLeaks(t *testing.T, parser *RustSynParser, detector *MemoryLeakDetector) {
	testCode := `fn utility_test() { println!("Testing utilities"); }`

	// Test various utility functions
	for i := 0; i < 50; i++ {
		// Test syntax validation
		_, err := parser.ValidateSyntax([]byte(testCode))
		if err != nil {
			t.Errorf("Syntax validation failed: %v", err)
		}

		// Test expression parsing
		_, err = parser.ParseExpression("42 + 24")
		if err != nil {
			t.Errorf("Expression parsing failed: %v", err)
		}

		// Test can parse check
		_, err = parser.CanParse([]byte(testCode))
		if err != nil {
			t.Errorf("CanParse check failed: %v", err)
		}

		// Test parse stats
		_, err = parser.GetParseStats([]byte(testCode))
		if err != nil {
			t.Errorf("GetParseStats failed: %v", err)
		}

		if i%10 == 9 {
			detector.TakeSample(parser)
		}
	}
}

// testConfigurationLeaks tests different configuration options for leaks
func testConfigurationLeaks(t *testing.T, parser *RustSynParser, detector *MemoryLeakDetector) {
	testCode := `
/// Documented function
pub fn config_test(param: i32) -> String {
    format!("Config test: {}", param)
}

struct ConfigStruct {
    field: i32,
}
`

	configs := []*RustSynConfig{
		{IncludeDocs: true, IncludePositions: true, ParseMacros: true, IncludePrivate: true, MaxComplexityCalc: 100, IncludeGenerics: true},
		{IncludeDocs: false, IncludePositions: false, ParseMacros: false, IncludePrivate: false, MaxComplexityCalc: 50, IncludeGenerics: false},
		{IncludeDocs: true, IncludePositions: false, ParseMacros: true, IncludePrivate: false, MaxComplexityCalc: 200, IncludeGenerics: true},
		{IncludeDocs: false, IncludePositions: true, ParseMacros: false, IncludePrivate: true, MaxComplexityCalc: 25, IncludeGenerics: false},
	}

	for i, config := range configs {
		for j := 0; j < 20; j++ {
			_, err := parser.ParseRustFileWithConfig([]byte(testCode), fmt.Sprintf("config_%d_%d.rs", i, j), config)
			if err != nil {
				t.Errorf("Config parsing failed: %v", err)
			}
		}
		
		detector.TakeSample(parser)
	}
}

// TestRustSynParser_PerformanceProfiler tests the performance profiler functionality  
func TestRustSynParser_PerformanceProfiler(t *testing.T) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	profiler := NewPerformanceProfiler()

	testCodes := map[string][]byte{
		"simple": []byte(`fn simple() { println!("hello"); }`),
		"medium": []byte(`
pub struct Medium {
    field1: i32,
    field2: String,
}

impl Medium {
    pub fn new() -> Self {
        Self {
            field1: 42,
            field2: "test".to_string(),
        }
    }
}
`),
		"complex": []byte(generateComplexRustCodeForProfiling()),
	}

	// Benchmark each code type
	for name, code := range testCodes {
		t.Run(fmt.Sprintf("Profile_%s", name), func(t *testing.T) {
			parseFunc := func(content []byte) error {
				_, err := parser.ParseRustFile(content, fmt.Sprintf("%s.rs", name))
				return err
			}

			// Run multiple benchmarks
			for i := 0; i < 10; i++ {
				result := profiler.BenchmarkParsingOperation(
					fmt.Sprintf("%s_benchmark_%d", name, i),
					code,
					parseFunc,
				)

				if !result.Success {
					t.Errorf("Benchmark failed: %s", result.Error)
				}

				if result.Duration <= 0 {
					t.Errorf("Invalid benchmark duration: %v", result.Duration)
				}

				t.Logf("Benchmark %s_%d: %v, Memory: %d bytes, Success: %v",
					name, i, result.Duration, result.MemoryUsed, result.Success)
			}
		})
	}

	// Test benchmark summary
	summary := profiler.GetBenchmarkSummary()
	t.Logf("Benchmark Summary: %+v", summary)

	benchmarkCount := summary["benchmark_count"].(int)
	if benchmarkCount != 30 { // 3 types * 10 benchmarks each
		t.Errorf("Expected 30 benchmarks, got %d", benchmarkCount)
	}

	successRate := summary["success_rate_percent"].(float64)
	if successRate != 100.0 {
		t.Errorf("Expected 100%% success rate, got %.2f%%", successRate)
	}

	// Test recent benchmarks retrieval
	recent := profiler.GetRecentBenchmarks(5)
	if len(recent) != 5 {
		t.Errorf("Expected 5 recent benchmarks, got %d", len(recent))
	}

	// Test clearing benchmarks
	profiler.ClearBenchmarks()
	summary = profiler.GetBenchmarkSummary()
	benchmarkCount = summary["benchmark_count"].(int)
	if benchmarkCount != 0 {
		t.Errorf("Expected 0 benchmarks after clearing, got %d", benchmarkCount)
	}
}

// generateComplexRustCodeForProfiling creates complex code for profiling tests
func generateComplexRustCodeForProfiling() string {
	return `
use std::collections::{HashMap, BTreeMap};
use std::sync::{Arc, Mutex};

pub trait ComplexTrait<T, U> {
    type Output;
    fn complex_method(&self, input: T) -> Result<Self::Output, U>;
}

pub struct ComplexStruct<T> 
where 
    T: Clone + Send + Sync,
{
    data: HashMap<String, T>,
    metadata: BTreeMap<u64, String>,
    processor: Arc<Mutex<Box<dyn Fn(&T) -> T + Send + Sync>>>,
}

impl<T> ComplexStruct<T> 
where 
    T: Clone + Send + Sync,
{
    pub fn new() -> Self {
        Self {
            data: HashMap::new(),
            metadata: BTreeMap::new(),
            processor: Arc::new(Mutex::new(Box::new(|x| x.clone()))),
        }
    }
    
    pub async fn process_async(&mut self, key: String, value: T) -> Result<T, Box<dyn std::error::Error>> {
        let processor = self.processor.lock().unwrap();
        let result = processor(&value);
        self.data.insert(key, result.clone());
        Ok(result)
    }
}

#[derive(Debug, Clone)]
pub enum ComplexEnum {
    Variant1 { field1: i32, field2: String },
    Variant2(Vec<f64>),
    Variant3,
}

macro_rules! complex_macro {
    ($name:ident, $type:ty) => {
        pub struct $name {
            value: $type,
        }
        
        impl $name {
            pub fn new(value: $type) -> Self {
                Self { value }
            }
        }
    };
}

complex_macro!(IntWrapper, i32);
complex_macro!(StringWrapper, String);
`
}

// BenchmarkRustSynParser_MemoryOperationsDetailed provides detailed memory operation benchmarks
func BenchmarkRustSynParser_MemoryOperationsDetailed(b *testing.B) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		b.Skipf("Skipping benchmark: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	b.Run("StringAllocationOverhead", func(b *testing.B) {
		testStrings := []string{
			"short",
			"medium length string for testing allocation overhead",
			strings.Repeat("very long string for testing allocation overhead with repeated content", 100),
		}
		
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			for _, str := range testStrings {
				cstr := parser.allocateCString(str)
				parser.freeCString(cstr)
			}
		}
	})

	b.Run("MemoryTrackingOverhead", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			_ = parser.GetMemorySafetyMetrics()
		}
	})

	b.Run("MemoryValidationOverhead", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			_ = parser.ValidateMemoryState()
		}
	})

	b.Run("ForceCleanupOverhead", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			parser.ForceCleanup()
		}
	})
}

// TestRustSynParser_StressTestWithLeakDetection combines stress testing with leak detection
func TestRustSynParser_StressTestWithLeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	parser, err := NewRustSynParser(true)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	detector := NewMemoryLeakDetector()
	detector.TakeSample(parser)

	// Stress test parameters
	const (
		duration           = 30 * time.Second
		samplingInterval   = 2 * time.Second
		cleanupInterval    = 10 * time.Second
		maxConcurrentOps   = 10
	)

	stopChan := make(chan bool)
	var wg sync.WaitGroup

	// Start sampling routine
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(samplingInterval)
		defer ticker.Stop()
		
		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				detector.TakeSample(parser)
			}
		}
	}()

	// Start cleanup routine
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		
		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				parser.ForceCleanup()
				runtime.GC()
			}
		}
	}()

	// Start stress testing goroutines
	semaphore := make(chan bool, maxConcurrentOps)
	for i := 0; i < maxConcurrentOps; i++ {
		semaphore <- true
	}

	startTime := time.Now()
	operationCount := 0

	for time.Since(startTime) < duration {
		<-semaphore
		wg.Add(1)
		
		go func(opID int) {
			defer wg.Done()
			defer func() { semaphore <- true }()
			
			// Perform various operations
			operations := []func(){
				func() {
					code := fmt.Sprintf(`fn stress_test_%d() { println!("Stress test %d"); }`, opID, opID)
					parser.ParseRustFile([]byte(code), fmt.Sprintf("stress_%d.rs", opID))
				},
				func() {
					parser.ValidateSyntax([]byte(`fn valid() {}`))
				},
				func() {
					parser.ParseExpression("42 + 24")
				},
				func() {
					parser.GetParseStats([]byte(`fn stats_test() {}`))
				},
			}
			
			// Execute random operation
			op := operations[opID%len(operations)]
			op()
		}(operationCount)
		
		operationCount++
		time.Sleep(10 * time.Millisecond) // Small delay between operations
	}

	// Stop all routines
	close(stopChan)
	wg.Wait()

	// Final analysis
	detector.TakeSample(parser)
	result := detector.AnalyzeLeaks(t)

	t.Logf("Stress test completed:")
	t.Logf("  Duration: %v", duration)
	t.Logf("  Operations: %d", operationCount)
	t.Logf("  Rate: %.2f ops/sec", float64(operationCount)/duration.Seconds())

	if result.HasLeak {
		t.Errorf("Memory leak detected during stress test: %s", result.Message)
	}

	// Validate final parser state
	if err := parser.ValidateMemoryState(); err != nil {
		t.Errorf("Parser memory state validation failed: %v", err)
	}
}