//go:build cgo && !no_rust
// +build cgo,!no_rust

package scanner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/ericfisherdev/goclean/internal/types"
)

// TestRustSynParser_BasicFunctionality tests basic CGO parsing functionality
func TestRustSynParser_BasicFunctionality(t *testing.T) {
	parser, err := NewRustSynParser(true)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	// Test simple Rust code
	rustCode := `
fn hello_world() -> String {
    "Hello, World!".to_string()
}

pub struct TestStruct {
    pub field1: i32,
    pub field2: String,
}

impl TestStruct {
    pub fn new(value: i32) -> Self {
        Self {
            field1: value,
            field2: format!("Value: {}", value),
        }
    }
}
`

	astInfo, err := parser.ParseRustFile([]byte(rustCode), "test.rs")
	if err != nil {
		t.Fatalf("Failed to parse Rust code: %v", err)
	}

	// Verify parsed AST
	if astInfo == nil {
		t.Fatal("AST info is nil")
	}

	if len(astInfo.Functions) != 2 { // hello_world + new
		t.Errorf("Expected 2 functions, got %d", len(astInfo.Functions))
	}

	if len(astInfo.Structs) != 1 {
		t.Errorf("Expected 1 struct, got %d", len(astInfo.Structs))
	}

	if len(astInfo.Impls) != 1 {
		t.Errorf("Expected 1 impl block, got %d", len(astInfo.Impls))
	}

	// Verify function details
	helloWorldFunc := findFunctionByName(astInfo.Functions, "hello_world")
	if helloWorldFunc == nil {
		t.Error("hello_world function not found")
	} else {
		if helloWorldFunc.ReturnType != "String" {
			t.Errorf("Expected return type 'String', got '%s'", helloWorldFunc.ReturnType)
		}
		if helloWorldFunc.IsPublic {
			t.Error("hello_world should not be public")
		}
	}

	// Verify struct details
	if astInfo.Structs[0].Name != "TestStruct" {
		t.Errorf("Expected struct name 'TestStruct', got '%s'", astInfo.Structs[0].Name)
	}
	if !astInfo.Structs[0].IsPublic {
		t.Error("TestStruct should be public")
	}
}

// TestRustSynParser_ConfigurationOptions tests different parsing configurations
func TestRustSynParser_ConfigurationOptions(t *testing.T) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	rustCode := `
/// This is a documented function
pub fn documented_function() -> i32 {
    42
}

fn private_function() -> i32 {
    24
}

use std::collections::HashMap;

macro_rules! simple_macro {
    () => {
        println!("Hello from macro!");
    };
}
`

	// Test with all options enabled
	fullConfig := &RustSynConfig{
		IncludeDocs:       true,
		IncludePositions:  true,
		ParseMacros:       true,
		IncludePrivate:    true,
		MaxComplexityCalc: 100,
		IncludeGenerics:   true,
	}

	astInfo, err := parser.ParseRustFileWithConfig([]byte(rustCode), "test.rs", fullConfig)
	if err != nil {
		t.Fatalf("Failed to parse with full config: %v", err)
	}

	if len(astInfo.Functions) != 2 {
		t.Errorf("Expected 2 functions with full config, got %d", len(astInfo.Functions))
	}

	// Test with private functions excluded
	publicOnlyConfig := &RustSynConfig{
		IncludeDocs:       true,
		IncludePositions:  true,
		ParseMacros:       true,
		IncludePrivate:    false,
		MaxComplexityCalc: 100,
		IncludeGenerics:   true,
	}

	astInfo, err = parser.ParseRustFileWithConfig([]byte(rustCode), "test.rs", publicOnlyConfig)
	if err != nil {
		t.Fatalf("Failed to parse with public-only config: %v", err)
	}

	// Should only have the public function
	publicFunctions := 0
	for _, fn := range astInfo.Functions {
		if fn.IsPublic {
			publicFunctions++
		}
	}

	if publicFunctions == 0 {
		t.Error("Expected at least one public function")
	}
}

// TestRustSynParser_ErrorHandling tests error handling in CGO calls
func TestRustSynParser_ErrorHandling(t *testing.T) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	// Test with invalid Rust code
	invalidCode := `
fn broken_function( {
    this is not valid rust
}
`

	_, err = parser.ParseRustFile([]byte(invalidCode), "broken.rs")
	if err == nil {
		t.Error("Expected error for invalid Rust code")
	}

	// Test with empty file
	_, err = parser.ParseRustFile([]byte(""), "empty.rs")
	if err != nil {
		t.Errorf("Empty file should not cause error: %v", err)
	}

	// Test with very large file (should be rejected)
	largeCode := make([]byte, 100*1024*1024) // 100MB
	for i := range largeCode {
		largeCode[i] = 'a'
	}

	_, err = parser.ParseRustFile(largeCode, "large.rs")
	if err == nil {
		t.Error("Expected error for excessively large file")
	}
}

// TestRustSynParser_MemoryManagement tests memory allocation and deallocation
func TestRustSynParser_MemoryManagement(t *testing.T) {
	parser, err := NewRustSynParser(true)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	rustCode := `
fn test_function() {
    println!("Testing memory management");
}
`

	// Parse multiple times and check memory tracking
	initialMetrics := parser.GetMemorySafetyMetrics()
	
	for i := 0; i < 10; i++ {
		_, err := parser.ParseRustFile([]byte(rustCode), fmt.Sprintf("test_%d.rs", i))
		if err != nil {
			t.Fatalf("Parsing failed at iteration %d: %v", i, err)
		}
	}

	finalMetrics := parser.GetMemorySafetyMetrics()
	
	// Check that no C strings are leaked
	allocatedStrings := finalMetrics["allocated_c_strings"].(int64)
	if allocatedStrings > 0 {
		t.Errorf("Memory leak detected: %d allocated C strings", allocatedStrings)
	}

	// Check that no parse calls are active
	activeCalls := finalMetrics["active_parse_calls"].(int64)
	if activeCalls > 0 {
		t.Errorf("Active parse calls not cleaned up: %d", activeCalls)
	}

	// Validate overall memory state
	if err := parser.ValidateMemoryState(); err != nil {
		t.Errorf("Memory state validation failed: %v", err)
	}
}

// TestRustSynParser_ConcurrentAccess tests thread safety of CGO calls
func TestRustSynParser_ConcurrentAccess(t *testing.T) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	rustCode := `
fn concurrent_test(value: i32) -> i32 {
    value * 2
}

struct ConcurrentStruct {
    field: String,
}
`

	const numGoroutines = 20
	const parsesPerGoroutine = 5

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*parsesPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for j := 0; j < parsesPerGoroutine; j++ {
				fileName := fmt.Sprintf("concurrent_%d_%d.rs", goroutineID, j)
				_, err := parser.ParseRustFile([]byte(rustCode), fileName)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d, parse %d: %w", goroutineID, j, err)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("Concurrent parsing error: %v", err)
	}

	// Verify final state
	metrics := parser.GetMemorySafetyMetrics()
	if err := parser.ValidateMemoryState(); err != nil {
		t.Errorf("Memory state validation failed after concurrent access: %v", err)
	}

	t.Logf("Final memory metrics after concurrent test: %+v", metrics)
}

// TestRustSynParser_ComplexRustFeatures tests parsing of complex Rust language features
func TestRustSynParser_ComplexRustFeatures(t *testing.T) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	complexRustCode := `
use std::collections::HashMap;
use std::sync::{Arc, Mutex};

/// A generic trait with associated types
pub trait ProcessorTrait<T> {
    type Output;
    type Error;
    
    fn process(&self, input: T) -> Result<Self::Output, Self::Error>;
}

/// Generic struct with lifetime parameters
pub struct DataProcessor<'a, T> 
where 
    T: Clone + Send + Sync,
{
    name: &'a str,
    data: Vec<T>,
    processor: Box<dyn Fn(&T) -> T + Send + Sync>,
}

impl<'a, T> DataProcessor<'a, T> 
where 
    T: Clone + Send + Sync,
{
    /// Create a new data processor
    pub fn new(name: &'a str) -> Self {
        Self {
            name,
            data: Vec::new(),
            processor: Box::new(|x| x.clone()),
        }
    }
    
    /// Async function with complex error handling
    pub async fn process_async(&mut self, items: Vec<T>) -> Result<Vec<T>, ProcessingError> {
        let mut results = Vec::with_capacity(items.len());
        
        for item in items {
            match tokio::time::timeout(
                std::time::Duration::from_secs(1),
                self.process_single(item.clone()),
            ).await {
                Ok(Ok(result)) => results.push(result),
                Ok(Err(e)) => return Err(ProcessingError::ProcessingFailed(e.into())),
                Err(_) => return Err(ProcessingError::Timeout),
            }
        }
        
        Ok(results)
    }
    
    async fn process_single(&self, item: T) -> Result<T, Box<dyn std::error::Error>> {
        tokio::time::sleep(std::time::Duration::from_millis(10)).await;
        Ok((self.processor)(&item))
    }
}

/// Complex enum with associated data
#[derive(Debug, Clone)]
pub enum ProcessingError {
    InvalidInput(String),
    ProcessingFailed(Box<dyn std::error::Error + Send + Sync>),
    Timeout,
    NetworkError { code: u16, message: String },
}

/// Macro for creating processors
macro_rules! create_processor {
    ($name:ident, $input_type:ty, $output_type:ty) => {
        pub struct $name;
        
        impl ProcessorTrait<$input_type> for $name {
            type Output = $output_type;
            type Error = ProcessingError;
            
            fn process(&self, input: $input_type) -> Result<Self::Output, Self::Error> {
                // Implementation would go here
                Ok(Default::default())
            }
        }
    };
}

create_processor!(StringProcessor, String, String);
create_processor!(NumberProcessor, i32, f64);

/// Const function with complex calculations
const fn calculate_capacity(size: usize) -> usize {
    if size < 10 {
        size * 2
    } else {
        size + 10
    }
}

/// Static with complex initialization
static GLOBAL_PROCESSOR: once_cell::sync::Lazy<Arc<Mutex<HashMap<String, Box<dyn ProcessorTrait<String, Output = String, Error = ProcessingError> + Send + Sync>>>>> = 
    once_cell::sync::Lazy::new(|| Arc::new(Mutex::new(HashMap::new())));
`

	astInfo, err := parser.ParseRustFile([]byte(complexRustCode), "complex.rs")
	if err != nil {
		t.Fatalf("Failed to parse complex Rust code: %v", err)
	}

	// Verify parsing of complex features
	if len(astInfo.Traits) == 0 {
		t.Error("Expected trait to be parsed")
	}

	if len(astInfo.Structs) == 0 {
		t.Error("Expected struct to be parsed")
	}

	if len(astInfo.Enums) == 0 {
		t.Error("Expected enum to be parsed")
	}

	if len(astInfo.Macros) == 0 {
		t.Error("Expected macro to be parsed")
	}

	// Check for complex function signatures
	asyncFunc := findFunctionByName(astInfo.Functions, "process_async")
	if asyncFunc == nil {
		t.Error("process_async function not found")
	} else {
		if !asyncFunc.IsAsync {
			t.Error("process_async should be marked as async")
		}
	}

	// Verify enum parsing
	if len(astInfo.Enums) > 0 {
		processingError := astInfo.Enums[0]
		if processingError.Name != "ProcessingError" {
			t.Errorf("Expected enum name 'ProcessingError', got '%s'", processingError.Name)
		}
		if processingError.VariantCount == 0 {
			t.Error("ProcessingError enum should have variants")
		}
	}
}

// TestRustSynParser_UtilityFunctions tests utility functions of the parser
func TestRustSynParser_UtilityFunctions(t *testing.T) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	// Test syntax validation
	validCode := `fn main() { println!("Hello, world!"); }`
	isValid, err := parser.ValidateSyntax([]byte(validCode))
	if err != nil {
		t.Errorf("Syntax validation failed: %v", err)
	}
	if !isValid {
		t.Error("Valid code was marked as invalid")
	}

	invalidCode := `fn main( { println!("Hello, world!"); }`
	isValid, err = parser.ValidateSyntax([]byte(invalidCode))
	if err != nil {
		t.Errorf("Syntax validation failed: %v", err)
	}
	if isValid {
		t.Error("Invalid code was marked as valid")
	}

	// Test expression parsing
	expr, err := parser.ParseExpression("42 + 24")
	if err != nil {
		t.Errorf("Expression parsing failed: %v", err)
	}
	if expr == nil {
		t.Error("Expression result is nil")
	}

	// Test capability query
	capabilities, err := parser.GetCapabilities()
	if err != nil {
		t.Errorf("Failed to get capabilities: %v", err)
	}
	if capabilities == nil {
		t.Error("Capabilities is nil")
	}

	// Test version info
	version, err := parser.GetVersion()
	if err != nil {
		t.Errorf("Failed to get version: %v", err)
	}
	if version == nil {
		t.Error("Version info is nil")
	}

	// Test parsing statistics
	stats, err := parser.GetParseStats([]byte(validCode))
	if err != nil {
		t.Errorf("Failed to get parse stats: %v", err)
	}
	if stats == nil {
		t.Error("Parse stats is nil")
	}

	// Test can parse check
	confidence, err := parser.CanParse([]byte(validCode))
	if err != nil {
		t.Errorf("CanParse check failed: %v", err)
	}
	if confidence <= 0 {
		t.Errorf("Expected positive confidence for valid Rust code, got %d", confidence)
	}
}

// TestRustSynParser_GlobalInstance tests the global parser instance
func TestRustSynParser_GlobalInstance(t *testing.T) {
	// Get global instance
	parser1, err := GetGlobalSynParser()
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}

	// Get it again - should be the same instance
	parser2, err := GetGlobalSynParser()
	if err != nil {
		t.Fatalf("Failed to get global parser second time: %v", err)
	}

	// Should be the same instance
	if parser1 != parser2 {
		t.Error("Global parser instances are different")
	}

	// Test that it's initialized
	if !parser1.IsInitialized() {
		t.Error("Global parser is not initialized")
	}

	// Test cleanup
	defer CleanupGlobalSynParser()

	// Test parsing with global instance
	rustCode := `fn test() { println!("Global parser test"); }`
	_, err = parser1.ParseRustFile([]byte(rustCode), "global_test.rs")
	if err != nil {
		t.Errorf("Global parser failed to parse: %v", err)
	}
}

// TestRustSynParser_JSONSerialization tests JSON parsing and serialization
func TestRustSynParser_JSONSerialization(t *testing.T) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	rustCode := `
pub fn example_function(param: i32) -> String {
    format!("Parameter: {}", param)
}

pub struct ExampleStruct {
    pub field1: i32,
    pub field2: String,
}
`

	astInfo, err := parser.ParseRustFile([]byte(rustCode), "json_test.rs")
	if err != nil {
		t.Fatalf("Failed to parse Rust code: %v", err)
	}

	// Test JSON serialization roundtrip
	jsonData, err := json.Marshal(astInfo)
	if err != nil {
		t.Fatalf("Failed to marshal AST to JSON: %v", err)
	}

	var restoredAST types.RustASTInfo
	err = json.Unmarshal(jsonData, &restoredAST)
	if err != nil {
		t.Fatalf("Failed to unmarshal AST from JSON: %v", err)
	}

	// Verify the roundtrip preserved data
	if len(restoredAST.Functions) != len(astInfo.Functions) {
		t.Errorf("Function count mismatch after JSON roundtrip: %d vs %d", 
			len(restoredAST.Functions), len(astInfo.Functions))
	}

	if len(restoredAST.Structs) != len(astInfo.Structs) {
		t.Errorf("Struct count mismatch after JSON roundtrip: %d vs %d", 
			len(restoredAST.Structs), len(astInfo.Structs))
	}

	// Verify function details are preserved
	if len(astInfo.Functions) > 0 && len(restoredAST.Functions) > 0 {
		original := astInfo.Functions[0]
		restored := restoredAST.Functions[0]

		if original.Name != restored.Name {
			t.Errorf("Function name mismatch: %s vs %s", original.Name, restored.Name)
		}

		if original.IsPublic != restored.IsPublic {
			t.Errorf("Function visibility mismatch: %v vs %v", original.IsPublic, restored.IsPublic)
		}
	}
}

// BenchmarkRustSynParser_ParseSmallFile benchmarks parsing small files
func BenchmarkRustSynParser_ParseSmallFile(b *testing.B) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		b.Skipf("Skipping benchmark: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	rustCode := []byte(`
fn main() {
    println!("Hello, world!");
}

struct Simple {
    field: i32,
}
`)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := parser.ParseRustFile(rustCode, "small.rs")
		if err != nil {
			b.Fatalf("Parsing failed: %v", err)
		}
	}
}

// BenchmarkRustSynParser_ParseLargeFile benchmarks parsing larger files
func BenchmarkRustSynParser_ParseLargeFile(b *testing.B) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		b.Skipf("Skipping benchmark: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	// Generate larger Rust code
	var codeBuilder strings.Builder
	codeBuilder.WriteString("use std::collections::HashMap;\n\n")

	for i := 0; i < 100; i++ {
		codeBuilder.WriteString(fmt.Sprintf(`
pub struct Struct%d {
    pub field1: i32,
    pub field2: String,
    pub field3: Vec<f64>,
}

impl Struct%d {
    pub fn new() -> Self {
        Self {
            field1: %d,
            field2: "test_%d".to_string(),
            field3: vec![1.0, 2.0, 3.0],
        }
    }
    
    pub fn process(&self, input: i32) -> Result<String, Box<dyn std::error::Error>> {
        if input > 0 {
            Ok(format!("Processed: {}", input * self.field1))
        } else {
            Err("Invalid input".into())
        }
    }
}
`, i, i, i, i))
	}

	rustCode := []byte(codeBuilder.String())

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := parser.ParseRustFile(rustCode, "large.rs")
		if err != nil {
			b.Fatalf("Parsing failed: %v", err)
		}
	}
}

// BenchmarkRustSynParser_MemoryOperations benchmarks memory-related operations
func BenchmarkRustSynParser_MemoryOperations(b *testing.B) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		b.Skipf("Skipping benchmark: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	b.Run("CStringAllocation", func(b *testing.B) {
		testString := "This is a test string for C allocation benchmarking"
		
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			cstr := parser.allocateCString(testString)
			parser.freeCString(cstr)
		}
	})

	b.Run("MetricsRetrieval", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			_ = parser.GetMemorySafetyMetrics()
		}
	})

	b.Run("MemoryValidation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			_ = parser.ValidateMemoryState()
		}
	})
}

// Helper functions

func findFunctionByName(functions []*types.RustFunctionInfo, name string) *types.RustFunctionInfo {
	for _, fn := range functions {
		if fn.Name == name {
			return fn
		}
	}
	return nil
}

// TestRustSynParser_StressTest performs stress testing to detect potential issues
func TestRustSynParser_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	parser, err := NewRustSynParser(false)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	rustCode := `
fn stress_test_function(param1: i32, param2: String) -> Result<String, Box<dyn std::error::Error>> {
    if param1 > 0 {
        Ok(format!("Value: {}, Message: {}", param1, param2))
    } else {
        Err("Invalid parameter".into())
    }
}

struct StressTestStruct {
    data: Vec<u8>,
    metadata: std::collections::HashMap<String, String>,
}

impl StressTestStruct {
    fn new() -> Self {
        Self {
            data: Vec::new(),
            metadata: std::collections::HashMap::new(),
        }
    }
}
`

	const iterations = 1000
	
	var memBefore runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	for i := 0; i < iterations; i++ {
		_, err := parser.ParseRustFile([]byte(rustCode), fmt.Sprintf("stress_%d.rs", i))
		if err != nil {
			t.Fatalf("Stress test failed at iteration %d: %v", i, err)
		}

		// Force cleanup every 100 iterations
		if i%100 == 99 {
			parser.ForceCleanup()
			runtime.GC()
		}
	}

	// Final validation
	if err := parser.ValidateMemoryState(); err != nil {
		t.Errorf("Memory state validation failed after stress test: %v", err)
	}

	var memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memAfter)

	memIncrease := memAfter.Alloc - memBefore.Alloc
	if memIncrease > 10*1024*1024 { // 10MB threshold
		t.Errorf("Memory increase too large after stress test: %d bytes", memIncrease)
	}

	metrics := parser.GetMemorySafetyMetrics()
	t.Logf("Final stress test metrics: %+v", metrics)
	t.Logf("Memory increase: %d bytes", memIncrease)
}

// TestRustSynParser_EdgeCases tests various edge cases
func TestRustSynParser_EdgeCases(t *testing.T) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	testCases := []struct {
		name     string
		code     string
		shouldFail bool
	}{
		{
			name: "EmptyFile",
			code: "",
			shouldFail: false,
		},
		{
			name: "OnlyComments",
			code: "// This is just a comment\n/* Block comment */",
			shouldFail: false,
		},
		{
			name: "OnlyUseStatements",
			code: "use std::collections::HashMap;\nuse std::sync::Arc;",
			shouldFail: false,
		},
		{
			name: "ComplexGenericFunction",
			code: `fn complex<T, U, V>(x: T, y: U) -> V where T: Clone, U: Send + Sync, V: Default { V::default() }`,
			shouldFail: false,
		},
		{
			name: "AsyncFunction",
			code: `async fn async_function() -> Result<(), Box<dyn std::error::Error>> { Ok(()) }`,
			shouldFail: false,
		},
		{
			name: "UnsafeFunction", 
			code: `unsafe fn unsafe_function(ptr: *mut u8) { *ptr = 42; }`,
			shouldFail: false,
		},
		{
			name: "ConstFunction",
			code: `const fn const_function(x: usize) -> usize { x * 2 }`,
			shouldFail: false,
		},
		{
			name: "IncompleteFunction",
			code: `fn incomplete_function(`,
			shouldFail: true,
		},
		{
			name: "MalformedStruct",
			code: `struct Broken {`,
			shouldFail: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parser.ParseRustFile([]byte(tc.code), fmt.Sprintf("%s.rs", tc.name))
			
			if tc.shouldFail && err == nil {
				t.Errorf("Expected parsing to fail for case %s", tc.name)
			} else if !tc.shouldFail && err != nil {
				t.Errorf("Expected parsing to succeed for case %s, got error: %v", tc.name, err)
			}
		})
	}
}

// TestRustSynParser_UnsafeOperations tests handling of unsafe code blocks
func TestRustSynParser_UnsafeOperations(t *testing.T) {
	parser, err := NewRustSynParser(false)
	if err != nil {
		t.Skipf("Skipping test: Rust syn parser not available: %v", err)
	}
	defer parser.Cleanup()

	// Test that parser doesn't crash on unsafe operations
	unsafeCode := `
use std::ptr;

unsafe fn dangerous_function(ptr: *mut u8, len: usize) -> Vec<u8> {
    if ptr.is_null() {
        return Vec::new();
    }
    
    let slice = std::slice::from_raw_parts(ptr, len);
    slice.to_vec()
}

fn safe_wrapper() -> Vec<u8> {
    let mut data = vec![1, 2, 3, 4, 5];
    unsafe {
        dangerous_function(data.as_mut_ptr(), data.len())
    }
}

struct UnsafeStruct {
    data: *mut u8,
    len: usize,
}

impl UnsafeStruct {
    unsafe fn new(size: usize) -> Self {
        let layout = std::alloc::Layout::from_size_align(size, 1).unwrap();
        let ptr = std::alloc::alloc(layout);
        Self { data: ptr, len: size }
    }
}
`

	astInfo, err := parser.ParseRustFile([]byte(unsafeCode), "unsafe.rs")
	if err != nil {
		t.Fatalf("Failed to parse unsafe code: %v", err)
	}

	// Verify unsafe functions are marked correctly
	unsafeFunc := findFunctionByName(astInfo.Functions, "dangerous_function")
	if unsafeFunc == nil {
		t.Error("dangerous_function not found")
	} else if !unsafeFunc.IsUnsafe {
		t.Error("dangerous_function should be marked as unsafe")
	}

	safeFunc := findFunctionByName(astInfo.Functions, "safe_wrapper")
	if safeFunc == nil {
		t.Error("safe_wrapper not found")
	} else if safeFunc.IsUnsafe {
		t.Error("safe_wrapper should not be marked as unsafe")
	}
}