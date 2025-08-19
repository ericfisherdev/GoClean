//go:build cgo && !no_rust
// +build cgo,!no_rust

package scanner

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestMemorySafetyTracking tests the memory safety tracking functionality
func TestMemorySafetyTracking(t *testing.T) {
	parser, err := NewRustSynParser(true)
	if err != nil {
		t.Skipf("Skipping test: Rust parser not available: %v", err)
	}
	defer parser.Cleanup()

	// Test basic allocation tracking
	cstr1 := parser.allocateCString("test string 1")
	cstr2 := parser.allocateCString("test string 2")

	metrics := parser.GetMemorySafetyMetrics()
	allocatedStrings := metrics["allocated_c_strings"].(int64)
	if allocatedStrings != 2 {
		t.Errorf("Expected 2 allocated strings, got %d", allocatedStrings)
	}

	parser.freeCString(cstr1)
	parser.freeCString(cstr2)

	metrics = parser.GetMemorySafetyMetrics()
	allocatedStrings = metrics["allocated_c_strings"].(int64)
	if allocatedStrings != 0 {
		t.Errorf("Expected 0 allocated strings after cleanup, got %d", allocatedStrings)
	}
}

// TestConcurrentParsing tests memory safety under concurrent operations
func TestConcurrentParsing(t *testing.T) {
	parser, err := NewRustSynParser(true)
	if err != nil {
		t.Skipf("Skipping test: Rust parser not available: %v", err)
	}
	defer parser.Cleanup()

	rustCode := `
fn hello_world() -> String {
    "Hello, World!".to_string()
}

struct TestStruct {
    field1: i32,
    field2: String,
}

impl TestStruct {
    fn new(value: i32) -> Self {
        Self {
            field1: value,
            field2: format!("Value: {}", value),
        }
    }
}
`

	const numGoroutines = 10
	const operationsPerGoroutine = 5

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*operationsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				_, err := parser.ParseRustFile([]byte(rustCode), "test.rs")
				if err != nil {
					errors <- err
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent parsing error: %v", err)
	}

	// Verify memory safety metrics
	metrics := parser.GetMemorySafetyMetrics()
	allocatedStrings := metrics["allocated_c_strings"].(int64)
	activeCalls := metrics["active_parse_calls"].(int64)

	if allocatedStrings > 0 {
		t.Errorf("Memory leak detected: %d allocated C strings after concurrent operations", allocatedStrings)
	}

	if activeCalls > 0 {
		t.Errorf("Active parse calls not cleaned up: %d", activeCalls)
	}

	// Validate memory state
	if err := parser.ValidateMemoryState(); err != nil {
		t.Errorf("Memory state validation failed: %v", err)
	}
}

// TestErrorRecovery tests error tracking and recovery mechanisms
func TestErrorRecovery(t *testing.T) {
	parser, err := NewRustSynParser(true)
	if err != nil {
		t.Skipf("Skipping test: Rust parser not available: %v", err)
	}
	defer parser.Cleanup()

	// Test with invalid Rust code
	invalidCode := "this is not valid rust code {{{"

	for i := 0; i < 5; i++ {
		_, err := parser.ParseRustFile([]byte(invalidCode), "invalid.rs")
		if err == nil {
			t.Error("Expected parsing error for invalid code")
		}
	}

	metrics := parser.GetMemorySafetyMetrics()
	consecutiveErrors := metrics["consecutive_errors"].(int64)
	if consecutiveErrors != 5 {
		t.Errorf("Expected 5 consecutive errors, got %d", consecutiveErrors)
	}

	// Test recovery with valid code
	validCode := `fn test() { println!("Hello"); }`
	_, err = parser.ParseRustFile([]byte(validCode), "valid.rs")
	if err != nil {
		t.Errorf("Valid code parsing failed: %v", err)
	}

	metrics = parser.GetMemorySafetyMetrics()
	consecutiveErrors = metrics["consecutive_errors"].(int64)
	if consecutiveErrors != 0 {
		t.Errorf("Expected consecutive errors to reset after successful parse, got %d", consecutiveErrors)
	}
}

// TestMemoryPooling tests the memory pooling functionality
func TestMemoryPooling(t *testing.T) {
	pool := NewMemoryPool()

	// Test string slice pooling
	slice1 := pool.GetStringSlice()
	slice1 = append(slice1, "test1", "test2", "test3")
	pool.PutStringSlice(slice1)

	slice2 := pool.GetStringSlice()
	if len(slice2) != 0 {
		t.Errorf("Expected empty slice from pool, got length %d", len(slice2))
	}
	if cap(slice2) < 3 {
		t.Errorf("Expected capacity >= 3 from reused slice, got %d", cap(slice2))
	}

	// Test buffer pooling
	buffer1 := pool.GetBuffer()
	buffer1 = append(buffer1, []byte("test data")...)
	pool.PutBuffer(buffer1)

	buffer2 := pool.GetBuffer()
	if len(buffer2) != 0 {
		t.Errorf("Expected empty buffer from pool, got length %d", len(buffer2))
	}
	if cap(buffer2) < 9 {
		t.Errorf("Expected capacity >= 9 from reused buffer, got %d", cap(buffer2))
	}
}

// TestPerformanceProfiler tests the performance profiling functionality
func TestPerformanceProfiler(t *testing.T) {
	profiler := NewPerformanceProfiler()

	// Create test parsing function
	testCode := []byte(`
fn fibonacci(n: u32) -> u32 {
    match n {
        0 => 0,
        1 => 1,
        _ => fibonacci(n - 1) + fibonacci(n - 2),
    }
}
`)

	parseFunc := func(content []byte) error {
		// Simulate parsing work
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	// Run multiple benchmarks
	for i := 0; i < 5; i++ {
		result := profiler.BenchmarkParsingOperation("test_parse", testCode, parseFunc)
		if !result.Success {
			t.Errorf("Benchmark %d failed: %s", i, result.Error)
		}
		if result.Duration < 10*time.Millisecond {
			t.Errorf("Benchmark %d duration too short: %v", i, result.Duration)
		}
	}

	// Test benchmark summary
	summary := profiler.GetBenchmarkSummary()
	benchmarkCount := summary["benchmark_count"].(int)
	if benchmarkCount != 5 {
		t.Errorf("Expected 5 benchmarks, got %d", benchmarkCount)
	}

	successRate := summary["success_rate_percent"].(float64)
	if successRate != 100.0 {
		t.Errorf("Expected 100%% success rate, got %.2f%%", successRate)
	}

	// Test recent benchmarks retrieval
	recent := profiler.GetRecentBenchmarks(3)
	if len(recent) != 3 {
		t.Errorf("Expected 3 recent benchmarks, got %d", len(recent))
	}

	// Test clearing benchmarks
	profiler.ClearBenchmarks()
	summary = profiler.GetBenchmarkSummary()
	benchmarkCount = summary["benchmark_count"].(int)
	if benchmarkCount != 0 {
		t.Errorf("Expected 0 benchmarks after clearing, got %d", benchmarkCount)
	}
}

// TestMemoryLeakDetection tests memory leak detection under stress
func TestMemoryLeakDetection(t *testing.T) {
	parser, err := NewRustSynParser(true)
	if err != nil {
		t.Skipf("Skipping test: Rust parser not available: %v", err)
	}
	defer parser.Cleanup()

	// Record initial memory
	var initialMemStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&initialMemStats)

	// Perform many parsing operations
	rustCode := `
fn test_function(param1: i32, param2: String) -> Result<String, Error> {
    if param1 > 0 {
        Ok(format!("Value: {}, Message: {}", param1, param2))
    } else {
        Err(Error::new("Invalid parameter"))
    }
}

struct DataProcessor {
    buffer: Vec<u8>,
    capacity: usize,
}

impl DataProcessor {
    fn new(capacity: usize) -> Self {
        Self {
            buffer: Vec::with_capacity(capacity),
            capacity,
        }
    }
    
    fn process(&mut self, data: &[u8]) -> Result<Vec<u8>, ProcessError> {
        if data.len() > self.capacity {
            return Err(ProcessError::TooLarge);
        }
        self.buffer.extend_from_slice(data);
        Ok(self.buffer.clone())
    }
}
`

	const iterations = 100
	for i := 0; i < iterations; i++ {
		_, err := parser.ParseRustFile([]byte(rustCode), "stress_test.rs")
		if err != nil {
			t.Fatalf("Parsing failed at iteration %d: %v", i, err)
		}

		// Force cleanup every 25 iterations
		if i%25 == 24 {
			parser.ForceCleanup()
		}
	}

	// Final validation
	if err := parser.ValidateMemoryState(); err != nil {
		t.Errorf("Memory state validation failed after stress test: %v", err)
	}

	// Check that memory usage is reasonable
	var finalMemStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&finalMemStats)

	memoryIncrease := finalMemStats.Alloc - initialMemStats.Alloc
	const maxAcceptableIncrease = 50 * 1024 * 1024 // 50MB

	if memoryIncrease > maxAcceptableIncrease {
		t.Errorf("Memory usage increased by %d bytes (%.2f MB), which exceeds the acceptable limit of %d bytes",
			memoryIncrease, float64(memoryIncrease)/(1024*1024), maxAcceptableIncrease)
	}

	metrics := parser.GetMemorySafetyMetrics()
	t.Logf("Final memory safety metrics: %+v", metrics)
}

// TestLargeFileHandling tests memory safety with large files
func TestLargeFileHandling(t *testing.T) {
	parser, err := NewRustSynParser(true)
	if err != nil {
		t.Skipf("Skipping test: Rust parser not available: %v", err)
	}
	defer parser.Cleanup()

	// Create a large Rust file (simulate)
	largeContent := make([]byte, 10*1024*1024) // 10MB
	for i := range largeContent {
		largeContent[i] = 'a'
	}

	// Try to parse very large content (should be rejected)
	_, err = parser.ParseRustFile(largeContent, "large_file.rs")
	if err == nil {
		t.Error("Expected error for file exceeding size limit")
	}

	// Verify no memory leak after large file rejection
	metrics := parser.GetMemorySafetyMetrics()
	allocatedStrings := metrics["allocated_c_strings"].(int64)
	if allocatedStrings > 0 {
		t.Errorf("Memory leak detected after large file rejection: %d allocated strings", allocatedStrings)
	}
}

// BenchmarkParsingPerformance benchmarks the enhanced parsing performance
func BenchmarkParsingPerformance(b *testing.B) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		b.Skipf("Skipping benchmark: Rust parser not available: %v", err)
	}
	defer parser.Cleanup()

	rustCode := []byte(`
pub struct ComplexStruct<T, U> 
where 
    T: Clone + Send + Sync,
    U: std::fmt::Display,
{
    data: Vec<T>,
    metadata: HashMap<String, U>,
    processor: Box<dyn Fn(&T) -> Result<U, Box<dyn std::error::Error>> + Send + Sync>,
}

impl<T, U> ComplexStruct<T, U> 
where 
    T: Clone + Send + Sync,
    U: std::fmt::Display,
{
    pub fn new() -> Self {
        Self {
            data: Vec::new(),
            metadata: HashMap::new(),
            processor: Box::new(|_| Err("Not implemented".into())),
        }
    }
    
    pub async fn process_data(&mut self, items: Vec<T>) -> Result<Vec<U>, ProcessingError> {
        let mut results = Vec::with_capacity(items.len());
        
        for item in items {
            match (self.processor)(&item) {
                Ok(result) => {
                    results.push(result);
                    self.data.push(item.clone());
                }
                Err(e) => return Err(ProcessingError::ProcessingFailed(e)),
            }
        }
        
        Ok(results)
    }
    
    pub fn get_statistics(&self) -> ProcessingStatistics {
        ProcessingStatistics {
            total_items: self.data.len(),
            metadata_entries: self.metadata.len(),
            memory_usage: std::mem::size_of_val(&self.data) + 
                         std::mem::size_of_val(&self.metadata),
        }
    }
}
`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseRustFile(rustCode, "benchmark.rs")
		if err != nil {
			b.Fatalf("Parsing failed: %v", err)
		}
	}
}