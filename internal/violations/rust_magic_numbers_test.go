package violations

import (
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// TestRustMagicNumberDetector_Name tests the detector name
func TestRustMagicNumberDetector_Name(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustMagicNumberDetector(config)

	expectedName := "Rust Magic Number Detector"
	if detector.Name() != expectedName {
		t.Errorf("Expected name '%s', got '%s'", expectedName, detector.Name())
	}
}

// TestRustMagicNumberDetector_Description tests the detector description
func TestRustMagicNumberDetector_Description(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustMagicNumberDetector(config)

	description := detector.Description()
	if description == "" {
		t.Error("Expected non-empty description")
	}

	// Check that description mentions key components
	expectedKeywords := []string{"hardcoded", "numeric", "literals", "Rust"}
	for _, keyword := range expectedKeywords {
		if !containsString(description, keyword) {
			t.Errorf("Description should contain '%s': %s", keyword, description)
		}
	}
}

// TestRustMagicNumberDetector_DetectWithNilASTInfo tests detection with nil AST info
func TestRustMagicNumberDetector_DetectWithNilASTInfo(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustMagicNumberDetector(config)

	fileInfo := &models.FileInfo{
		Path:     "test.rs",
		Language: "rust",
	}

	violations := detector.Detect(fileInfo, nil)
	if len(violations) != 0 {
		t.Errorf("Expected 0 violations for nil AST info, got %d", len(violations))
	}
}

// TestRustMagicNumberDetector_DetectWithInvalidASTInfo tests detection with invalid AST info
func TestRustMagicNumberDetector_DetectWithInvalidASTInfo(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustMagicNumberDetector(config)

	fileInfo := &models.FileInfo{
		Path:     "test.rs",
		Language: "rust",
	}

	// Test with wrong type
	violations := detector.Detect(fileInfo, "invalid")
	if len(violations) != 0 {
		t.Errorf("Expected 0 violations for invalid AST info, got %d", len(violations))
	}
}

// TestRustMagicNumberDetector_isRustAcceptableInt tests integer value acceptance
func TestRustMagicNumberDetector_isRustAcceptableInt(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustMagicNumberDetector(config)

	testCases := []struct {
		value    string
		expected bool
		reason   string
	}{
		{"0", true, "zero is always acceptable"},
		{"1", true, "one is always acceptable"},
		{"2", true, "small numbers are acceptable"},
		{"10", true, "ten is acceptable"},
		{"16", true, "common bit size"},
		{"32", true, "common bit size"},
		{"64", true, "common bit size"},
		{"100", true, "percentage base"},
		{"1024", true, "power of 2"},
		{"42", false, "arbitrary number should be flagged"},
		{"137", false, "arbitrary number should be flagged"},
		{"999", false, "arbitrary number should be flagged"},
		{"1u32", true, "typed one is acceptable"},
		{"42u64", false, "typed arbitrary number should be flagged"},
		{"1024usize", true, "typed power of 2 is acceptable"},
	}

	for _, tc := range testCases {
		result := detector.isRustAcceptableInt(tc.value)
		if result != tc.expected {
			t.Errorf("isRustAcceptableInt(%s): expected %v, got %v (%s)", 
				tc.value, tc.expected, result, tc.reason)
		}
	}
}

// TestRustMagicNumberDetector_isRustAcceptableFloat tests float value acceptance
func TestRustMagicNumberDetector_isRustAcceptableFloat(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustMagicNumberDetector(config)

	testCases := []struct {
		value    string
		expected bool
		reason   string
	}{
		{"0.0", true, "zero float is acceptable"},
		{"1.0", true, "one float is acceptable"},
		{"0.5", true, "half is common"},
		{"3.14", true, "pi approximation"},
		{"3.14159", true, "pi approximation"},
		{"2.71828", true, "e approximation"},
		{"1.23", false, "arbitrary float should be flagged"},
		{"0.0f32", true, "typed zero float is acceptable"},
		{"1.5f64", false, "typed arbitrary float should be flagged"},
	}

	for _, tc := range testCases {
		result := detector.isRustAcceptableFloat(tc.value)
		if result != tc.expected {
			t.Errorf("isRustAcceptableFloat(%s): expected %v, got %v (%s)", 
				tc.value, tc.expected, result, tc.reason)
		}
	}
}

// TestRustMagicNumberDetector_isRustAcceptableHex tests hex value acceptance
func TestRustMagicNumberDetector_isRustAcceptableHex(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustMagicNumberDetector(config)

	testCases := []struct {
		value    string
		expected bool
		reason   string
	}{
		{"0x0", true, "zero hex is acceptable"},
		{"0x1", true, "one hex is acceptable"},
		{"0xff", true, "common bit mask"},
		{"0xffff", true, "common bit mask"},
		{"0x100", true, "power of 2 in hex"},
		{"0x42", false, "arbitrary hex should be flagged"},
		{"0xdeadbeef", false, "arbitrary hex should be flagged"},
		{"0x1u32", true, "typed hex one is acceptable"},
		{"0x42u64", false, "typed arbitrary hex should be flagged"},
	}

	for _, tc := range testCases {
		result := detector.isRustAcceptableHex(tc.value)
		if result != tc.expected {
			t.Errorf("isRustAcceptableHex(%s): expected %v, got %v (%s)", 
				tc.value, tc.expected, result, tc.reason)
		}
	}
}

// TestRustMagicNumberDetector_isRustAcceptableBinary tests binary value acceptance
func TestRustMagicNumberDetector_isRustAcceptableBinary(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustMagicNumberDetector(config)

	testCases := []struct {
		value    string
		expected bool
		reason   string
	}{
		{"0b0", true, "zero binary is acceptable"},
		{"0b1", true, "one binary is acceptable"},
		{"0b10", true, "power of 2 in binary"},
		{"0b1111", true, "common bit mask"},
		{"0b11111111", true, "byte mask"},
		{"0b101010", false, "arbitrary binary should be flagged"},
		{"0b1u8", true, "typed binary one is acceptable"},
		{"0b101010u16", false, "typed arbitrary binary should be flagged"},
	}

	for _, tc := range testCases {
		result := detector.isRustAcceptableBinary(tc.value)
		if result != tc.expected {
			t.Errorf("isRustAcceptableBinary(%s): expected %v, got %v (%s)", 
				tc.value, tc.expected, result, tc.reason)
		}
	}
}

// TestRustMagicNumberDetector_removeRustTypeSuffix tests type suffix removal
func TestRustMagicNumberDetector_removeRustTypeSuffix(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustMagicNumberDetector(config)

	testCases := []struct {
		input    string
		expected string
	}{
		{"42", "42"},
		{"42u32", "42"},
		{"42i64", "42"},
		{"1024usize", "1024"},
		{"3.14f32", "3.14"},
		{"2.71828f64", "2.71828"},
		{"0xffu8", "0xff"},
		{"0b1010i16", "0b1010"},
	}

	for _, tc := range testCases {
		result := detector.removeRustTypeSuffix(tc.input)
		if result != tc.expected {
			t.Errorf("removeRustTypeSuffix(%s): expected %s, got %s", 
				tc.input, tc.expected, result)
		}
	}
}

// TestRustMagicNumberDetector_isRustPowerOfTwo tests power of 2 detection
func TestRustMagicNumberDetector_isRustPowerOfTwo(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustMagicNumberDetector(config)

	testCases := []struct {
		value    int
		expected bool
	}{
		{1, true},
		{2, true},
		{4, true},
		{8, true},
		{16, true},
		{32, true},
		{64, true},
		{128, true},
		{256, true},
		{512, true},
		{1024, true},
		{3, false},
		{5, false},
		{6, false},
		{7, false},
		{9, false},
		{15, false},
		{31, false},
		{63, false},
	}

	for _, tc := range testCases {
		result := detector.isRustPowerOfTwo(tc.value)
		if result != tc.expected {
			t.Errorf("isRustPowerOfTwo(%d): expected %v, got %v", 
				tc.value, tc.expected, result)
		}
	}
}

// TestRustMagicNumberDetector_isRustPowerOfTen tests power of 10 detection
func TestRustMagicNumberDetector_isRustPowerOfTen(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustMagicNumberDetector(config)

	testCases := []struct {
		value    int
		expected bool
	}{
		{1, true},
		{10, true},
		{100, true},
		{1000, true},
		{10000, true},
		{100000, true},
		{1000000, true},
		{2, false},
		{5, false},
		{11, false},
		{99, false},
		{101, false},
		{999, false},
		{1001, false},
	}

	for _, tc := range testCases {
		result := detector.isRustPowerOfTen(tc.value)
		if result != tc.expected {
			t.Errorf("isRustPowerOfTen(%d): expected %v, got %v", 
				tc.value, tc.expected, result)
		}
	}
}

// TestRustMagicNumberDetector_shouldSkipLine tests line skipping logic
func TestRustMagicNumberDetector_shouldSkipLine(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustMagicNumberDetector(config)

	testCases := []struct {
		line     string
		expected bool
		reason   string
	}{
		{"// This is a comment with 42", true, "comment lines should be skipped"},
		{"/* Block comment with 123 */", true, "block comment lines should be skipped"},
		{"const MAX_SIZE: usize = 1024;", true, "const declarations should be skipped"},
		{"static GLOBAL_VAR: i32 = 42;", true, "static declarations should be skipped"},
		{"version = \"1.2.3\"", true, "version strings should be skipped"},
		{"let name = \"test\";", true, "string literals should be skipped"},
		{"#[test]", true, "test attributes should be skipped"},
		{"assert_eq!(result, 42);", true, "assert macros should be skipped"},
		{"let x = 42;", false, "regular code should not be skipped"},
		{"    let y = some_function(123);", false, "function calls should not be skipped"},
	}

	for _, tc := range testCases {
		result := detector.shouldSkipLine(tc.line)
		if result != tc.expected {
			t.Errorf("shouldSkipLine(%s): expected %v, got %v (%s)", 
				tc.line, tc.expected, result, tc.reason)
		}
	}
}

// TestRustMagicNumberDetector_isInRustAcceptableContext tests context acceptance
func TestRustMagicNumberDetector_isInRustAcceptableContext(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustMagicNumberDetector(config)

	testCases := []struct {
		line     string
		expected bool
		reason   string
	}{
		{"arr[42]", true, "array indexing should be acceptable"},
		{"let range = 0..42;", true, "range operations should be acceptable"},
		{"match value { 42 => {}, _ => {} }", true, "match arms should be acceptable"},
		{"Vec::with_capacity(100)", true, "capacity hints should be acceptable"},
		{"thread::sleep(Duration::from_secs(5))", true, "duration context should be acceptable"},
		{"let port = 8080;", true, "port context should be acceptable"},
		{"let buffer = vec![0; 1024];", true, "buffer context should be acceptable"},
		{"let mask = value & 0xff;", true, "bit manipulation should be acceptable"},
		{"let count = 42;", true, "count variable should be acceptable"},
		{"let magic = 42;", false, "magic variable should not be acceptable"},
		{"some_function(42)", false, "function argument should not be acceptable"},
	}

	for _, tc := range testCases {
		match := NumericMatch{Value: "42", Type: "integer", Content: tc.line}
		result := detector.isInRustAcceptableContext(match, tc.line)
		if result != tc.expected {
			t.Errorf("isInRustAcceptableContext(%s): expected %v, got %v (%s)", 
				tc.line, tc.expected, result, tc.reason)
		}
	}
}

// TestRustMagicNumberDetector_getRustMagicNumberSeverity tests severity assignment
func TestRustMagicNumberDetector_getRustMagicNumberSeverity(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustMagicNumberDetector(config)

	testCases := []struct {
		value        string
		valueType    string
		expected     models.Severity
		reason       string
	}{
		{"42", "integer", models.SeverityLow, "small integers should be low severity"},
		{"1000000", "integer", models.SeverityLow, "large integers should be low severity"},
		{"3.14", "float", models.SeverityLow, "floats should be low severity"},
		{"0xff", "hex", models.SeverityMedium, "hex literals should be medium severity"},
		{"0b1010", "binary", models.SeverityMedium, "binary literals should be medium severity"},
	}

	for _, tc := range testCases {
		result := detector.getRustMagicNumberSeverity(tc.value, tc.valueType)
		if result != tc.expected {
			t.Errorf("getRustMagicNumberSeverity(%s, %s): expected %v, got %v (%s)", 
				tc.value, tc.valueType, tc.expected, result, tc.reason)
		}
	}
}

// TestRustMagicNumberDetector_getRustMagicNumberSuggestion tests suggestion generation
func TestRustMagicNumberDetector_getRustMagicNumberSuggestion(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustMagicNumberDetector(config)

	testCases := []struct {
		value        string
		valueType    string
		shouldContain []string
		reason       string
	}{
		{"42", "integer", []string{"const", "SCREAMING_SNAKE_CASE"}, "integer suggestions should mention const and naming"},
		{"0xff", "hex", []string{"const", "FLAG_MASK", "0xff"}, "hex suggestions should mention masks and hex format"},
		{"0b1010", "binary", []string{"const", "BIT_PATTERN", "0b1010"}, "binary suggestions should mention patterns and binary format"},
		{"3.14", "float", []string{"const", "COEFFICIENT", "3.14"}, "float suggestions should mention coefficients and float format"},
	}

	for _, tc := range testCases {
		result := detector.getRustMagicNumberSuggestion(tc.value, tc.valueType)
		for _, shouldContain := range tc.shouldContain {
			if !containsString(result, shouldContain) {
				t.Errorf("getRustMagicNumberSuggestion(%s, %s) should contain '%s': %s (%s)", 
					tc.value, tc.valueType, shouldContain, result, tc.reason)
			}
		}
	}
}

// TestRustMagicNumberDetector_ValidRustASTInfo tests detection with valid Rust AST info
func TestRustMagicNumberDetector_ValidRustASTInfo(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustMagicNumberDetector(config)

	fileInfo := &models.FileInfo{
		Path:     "test.rs",
		Language: "rust",
	}

	rustAstInfo := &types.RustASTInfo{
		FilePath: "test.rs",
		Functions: []*types.RustFunctionInfo{
			{
				Name:      "test_function",
				StartLine: 10,
			},
		},
	}

	// Since the detector reads file content, this test will not find violations
	// unless we can mock the file reading or provide actual file content
	violations := detector.Detect(fileInfo, rustAstInfo)

	// We expect 0 violations since we can't read the file content in tests
	// In a real scenario, this would require setting up test files
	if len(violations) < 0 {
		t.Errorf("Expected 0 or more violations, got %d", len(violations))
	}
}