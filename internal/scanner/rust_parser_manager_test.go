package scanner

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestRustParserManager_Initialization tests parser manager initialization
func TestRustParserManager_Initialization(t *testing.T) {
	manager := NewRustParserManager(true)
	defer manager.Cleanup()

	// Verify manager is initialized
	status := manager.GetStatus()
	if status["parser_type"] == "unknown" {
		t.Error("Parser manager should have initialized to a known parser type")
	}

	// Check capabilities
	capabilities := manager.GetCapabilities()
	if capabilities == nil {
		t.Error("Capabilities should not be nil")
	}

	if capabilities.ParserType == "" {
		t.Error("Parser type should be set")
	}

	if capabilities.InitializedAt.IsZero() {
		t.Error("Initialization time should be set")
	}

	t.Logf("Initialized with parser type: %s", capabilities.ParserType)
	t.Logf("Accuracy level: %s", capabilities.AccuracyLevel)
	t.Logf("Performance level: %s", capabilities.PerformanceLevel)
}

// TestRustParserManager_FallbackMessage tests fallback messaging
func TestRustParserManager_FallbackMessage(t *testing.T) {
	manager := NewRustParserManager(false)
	defer manager.Cleanup()

	message := manager.GetFallbackMessage()
	if message == "" {
		t.Error("Fallback message should not be empty")
	}

	// Message should contain useful information
	if !strings.Contains(message, "Rust parsing") {
		t.Error("Message should mention Rust parsing")
	}

	t.Logf("Fallback message: %s", message)
}

// TestRustParserManager_ParseRustFile tests file parsing with fallback
func TestRustParserManager_ParseRustFile(t *testing.T) {
	manager := NewRustParserManager(true)
	defer manager.Cleanup()

	rustCode := `
fn hello_world() -> String {
    "Hello, World!".to_string()
}

pub struct TestStruct {
    pub field: i32,
}

impl TestStruct {
    pub fn new(value: i32) -> Self {
        Self { field: value }
    }
}
`

	astInfo, err := manager.ParseRustFile([]byte(rustCode), "test.rs")
	
	// Check the result based on available parser
	capabilities := manager.GetCapabilities()
	
	switch capabilities.ParserType {
	case "syn-crate":
		// Syn parser should succeed with good results
		if err != nil {
			t.Errorf("Syn parser should succeed: %v", err)
		}
		if astInfo == nil {
			t.Error("AST info should not be nil with syn parser")
		}
		if len(astInfo.Functions) < 2 { // hello_world + new
			t.Errorf("Expected at least 2 functions, got %d", len(astInfo.Functions))
		}

	case "regex-fallback":
		// Regex parser might succeed with basic results
		if err != nil {
			t.Logf("Regex parser error (expected): %v", err)
		} else if astInfo != nil {
			t.Logf("Regex parser found %d functions, %d structs", 
				len(astInfo.Functions), len(astInfo.Structs))
		}

	case "no-op-fallback":
		// Fallback mode should return an error
		if err == nil {
			t.Error("Fallback mode should return an error")
		}
		if astInfo != nil {
			t.Error("AST info should be nil in fallback mode")
		}

	default:
		t.Errorf("Unknown parser type: %s", capabilities.ParserType)
	}

	// Check statistics
	status := manager.GetStatus()
	attemptCount := status["attempt_count"].(int64)
	if attemptCount == 0 {
		t.Error("Attempt count should be incremented")
	}

	t.Logf("Parse attempt completed with parser: %s", capabilities.ParserType)
}

// TestRustParserManager_ValidateSyntax tests syntax validation
func TestRustParserManager_ValidateSyntax(t *testing.T) {
	manager := NewRustParserManager(false)
	defer manager.Cleanup()

	// Test valid Rust code
	validCode := `fn main() { println!("Hello, world!"); }`
	isValid, err := manager.ValidateSyntax([]byte(validCode))

	capabilities := manager.GetCapabilities()
	
	if capabilities.HasSyntaxValidation {
		// Parser supports syntax validation
		if err != nil {
			t.Errorf("Syntax validation should not error on valid code: %v", err)
		}
		if !isValid {
			t.Error("Valid code should be marked as valid")
		}

		// Test invalid Rust code
		invalidCode := `fn main( { println!("Hello, world!"); }`
		isValid, err = manager.ValidateSyntax([]byte(invalidCode))
		if err != nil {
			t.Errorf("Syntax validation should not error on invalid code: %v", err)
		}
		if isValid {
			t.Error("Invalid code should be marked as invalid")
		}
	} else {
		// Parser doesn't support syntax validation
		if capabilities.ParserType == "no-op-fallback" && err == nil {
			t.Error("Fallback mode should return error for syntax validation")
		}
		t.Logf("Syntax validation not supported by %s", capabilities.ParserType)
	}
}

// TestRustParserManager_ErrorHandling tests error handling and recovery
func TestRustParserManager_ErrorHandling(t *testing.T) {
	manager := NewRustParserManager(false)
	defer manager.Cleanup()

	// Test with completely invalid content
	invalidCodes := []string{
		"", // empty
		"not rust code at all",
		"fn broken_function(",
		"struct Incomplete {",
		strings.Repeat("x", 10000), // very long invalid content
	}

	for i, code := range invalidCodes {
		t.Run(fmt.Sprintf("InvalidCode_%d", i), func(t *testing.T) {
			_, err := manager.ParseRustFile([]byte(code), fmt.Sprintf("invalid_%d.rs", i))
			
			capabilities := manager.GetCapabilities()
			
			if capabilities.ParserType == "no-op-fallback" {
				// Fallback should always error
				if err == nil {
					t.Error("Fallback mode should return error for any input")
				}
			} else {
				// Other parsers might handle some invalid input gracefully
				t.Logf("Parser %s result for invalid code %d: error=%v", 
					capabilities.ParserType, i, err)
			}
		})
	}

	// Check error history
	status := manager.GetStatus()
	errors := status["recent_errors"].([]string)
	if len(errors) == 0 && status["parser_type"] != "no-op-fallback" {
		t.Logf("No errors recorded (which is fine if parser handles invalid input gracefully)")
	} else {
		t.Logf("Recorded %d recent errors", len(errors))
	}
}

// TestRustParserManager_Performance tests performance monitoring
func TestRustParserManager_Performance(t *testing.T) {
	manager := NewRustParserManager(false)
	defer manager.Cleanup()

	// Perform multiple operations to test performance tracking
	testCode := `fn perf_test() { println!("Performance test"); }`
	
	const iterations = 10
	startTime := time.Now()

	for i := 0; i < iterations; i++ {
		_, err := manager.ParseRustFile([]byte(testCode), fmt.Sprintf("perf_%d.rs", i))
		
		// Don't fail the test if parsing fails (depends on available parser)
		if err != nil {
			t.Logf("Parsing iteration %d failed: %v", i, err)
		}
	}

	duration := time.Since(startTime)
	
	// Check statistics
	status := manager.GetStatus()
	attemptCount := status["attempt_count"].(int64)
	successCount := status["success_count"].(int64)
	successRate := status["success_rate"].(float64)

	t.Logf("Performance test results:")
	t.Logf("  Duration: %v", duration)
	t.Logf("  Attempts: %d", attemptCount)
	t.Logf("  Successes: %d", successCount)
	t.Logf("  Success rate: %.2f%%", successRate)
	t.Logf("  Avg time per operation: %v", duration/iterations)

	if attemptCount < iterations {
		t.Errorf("Expected at least %d attempts, got %d", iterations, attemptCount)
	}
}

// TestRustParserManager_GlobalInstance tests the global manager instance
func TestRustParserManager_GlobalInstance(t *testing.T) {
	// Get global instance
	manager1 := GetGlobalParserManager(false)
	manager2 := GetGlobalParserManager(true) // Different verbose setting shouldn't matter

	// Should be the same instance
	if manager1 != manager2 {
		t.Error("Global parser manager instances should be the same")
	}

	// Test functionality
	capabilities := manager1.GetCapabilities()
	if capabilities == nil {
		t.Error("Global manager should have capabilities")
	}

	// Test cleanup
	defer CleanupGlobalParserManager()

	// Test basic operation
	testCode := `fn global_test() {}`
	_, err := manager1.ParseRustFile([]byte(testCode), "global_test.rs")
	
	// Don't fail if parsing fails (depends on available parser)
	if err != nil {
		t.Logf("Global manager parsing failed (may be expected): %v", err)
	}
}

// TestRustParserManager_StatusReporting tests detailed status reporting
func TestRustParserManager_StatusReporting(t *testing.T) {
	manager := NewRustParserManager(true)
	defer manager.Cleanup()

	// Get initial status
	status := manager.GetStatus()
	
	// Verify all expected fields are present
	expectedFields := []string{
		"parser_type", "fallback_reason", "initialized_at", "uptime_seconds",
		"attempt_count", "success_count", "success_rate", "has_syn_parser",
		"has_regex_parser", "recent_errors", "cgo_enabled", "rust_available",
	}

	for _, field := range expectedFields {
		if _, exists := status[field]; !exists {
			t.Errorf("Status should include field: %s", field)
		}
	}

	// Check specific values
	if status["cgo_enabled"] != CGOEnabled {
		t.Errorf("CGO enabled status mismatch: expected %v, got %v", 
			CGOEnabled, status["cgo_enabled"])
	}

	// Check uptime
	uptime := status["uptime_seconds"].(float64)
	if uptime <= 0 {
		t.Error("Uptime should be positive")
	}

	t.Logf("Manager status:")
	for key, value := range status {
		t.Logf("  %s: %v", key, value)
	}
}

// BenchmarkRustParserManager_ParsePerformance benchmarks parsing performance
func BenchmarkRustParserManager_ParsePerformance(b *testing.B) {
	manager := NewRustParserManager(false)
	defer manager.Cleanup()

	rustCode := []byte(`
fn benchmark_function(x: i32, y: i32) -> i32 {
    x + y
}

struct BenchmarkStruct {
    field1: i32,
    field2: String,
}

impl BenchmarkStruct {
    fn new() -> Self {
        Self {
            field1: 42,
            field2: "benchmark".to_string(),
        }
    }
}
`)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := manager.ParseRustFile(rustCode, "benchmark.rs")
		if err != nil {
			// Don't fail benchmark if parsing fails (depends on available parser)
			continue
		}
	}

	b.StopTimer()

	// Report final statistics
	status := manager.GetStatus()
	capabilities := manager.GetCapabilities()
	
	b.Logf("Parser type: %s", capabilities.ParserType)
	b.Logf("Success rate: %.2f%%", status["success_rate"].(float64))
}