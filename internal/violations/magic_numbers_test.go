package violations
import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)
func TestMagicNumberDetector_Detect_NoMagicNumbers(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewMagicNumberDetector(config)
	source := `
package main
func example() {
	x := 0
	y := 1 
	z := 10
	array := make([]int, 100)
	timeout := 60 * 24 // Common time values
}`
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}
	fileInfo := &models.FileInfo{
		Path:     "/test/example.go",
		Language: "Go",
	}
	astInfo := &types.GoASTInfo{
		AST:     astFile,
		FileSet: fset,
	}
	violations := detector.Detect(fileInfo, astInfo)
	// Should not detect violations for acceptable numbers
	if len(violations) != 0 {
		t.Errorf("Expected no violations, got %d", len(violations))
	}
}
func TestMagicNumberDetector_Detect_WithMagicNumbers(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewMagicNumberDetector(config)
	source := `
package main
func example() {
	x := 3600  // Magic number - should be a constant
	buffer := make([]byte, 9999) // Magic number (changed from 8192 which is acceptable)
	discount := price * 0.15     // Magic number
	y := 73            // Magic number
}`
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}
	fileInfo := &models.FileInfo{
		Path:     "/test/example.go",
		Language: "Go",
	}
	astInfo := &types.GoASTInfo{
		AST:     astFile,
		FileSet: fset,
	}
	violations := detector.Detect(fileInfo, astInfo)
	// Should detect magic numbers
	if len(violations) < 3 {
		t.Errorf("Expected at least 3 violations for magic numbers, got %d", len(violations))
		for i, v := range violations {
			t.Logf("Violation %d: Line %d, Message: %s, Code: %s", i+1, v.Line, v.Message, v.CodeSnippet)
		}
	}
	// Check violation types
	for _, violation := range violations {
		if violation.Type != models.ViolationTypeMagicNumber {
			t.Errorf("Expected magic number violation type, got %s", violation.Type)
		}
		if violation.Severity != models.SeverityLow {
			t.Errorf("Expected low severity, got %s", violation.Severity)
		}
	}
}
func TestMagicNumberDetector_AcceptableValues(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewMagicNumberDetector(config)
	testCases := []struct {
		name  string
		value int
		acceptable bool
	}{
		{"Zero", 0, true},
		{"One", 1, true},
		{"Two", 2, true},
		{"Ten", 10, true},
		{"Hour minutes", 60, true},
		{"Day hours", 24, true},
		{"Week days", 7, true},
		{"Year days", 365, true},
		{"Kilobyte", 1024, true},
		{"Small array size", 5, true},
		{"Random large number", 12345, false},
		{"Magic retry count", 73, false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := detector.isAcceptableInt(tc.value)
			if result != tc.acceptable {
				t.Errorf("Value %d: expected acceptable=%t, got %t", tc.value, tc.acceptable, result)
			}
		})
	}
}
func TestMagicNumberDetector_AcceptableFloats(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewMagicNumberDetector(config)
	testCases := []struct {
		name  string
		value float64
		acceptable bool
	}{
		{"Zero", 0.0, true},
		{"One", 1.0, true},
		{"Half", 0.5, true},
		{"Quarter", 0.25, true},
		{"Pi approximation", 3.14, true},
		{"Pi precise", 3.14159, true},
		{"Pi very precise", 3.141592653589793, true},
		{"E approximation", 2.71828, true},
		{"E precise", 2.718281828459045, true},
		{"Pi close approximation", 3.1416, true},  // Should be accepted as Pi approximation
		{"E close approximation", 2.718, true},   // Should be accepted as e approximation
		{"Random decimal", 2.7845, false},
		{"Magic percentage", 0.123, false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := detector.isAcceptableFloat(tc.value)
			if result != tc.acceptable {
				t.Errorf("Value %f: expected acceptable=%t, got %t", tc.value, tc.acceptable, result)
			}
		})
	}
}

// Test const declarations should NOT be flagged
func TestMagicNumberDetector_ConstDeclarations(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewMagicNumberDetector(config)
	
	source := `
package main

const (
    YouthSportsWeight   = 25.0
    AdultFitnessWeight  = 20.0
    MaxConnections     = 100
    DefaultTimeout     = 30
)

var (
    GlobalVar = 42  // This should be flagged
)
`
	
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}
	
	fileInfo := &models.FileInfo{
		Path:     "/test/example.go",
		Language: "Go",
	}
	astInfo := &types.GoASTInfo{
		AST:     astFile,
		FileSet: fset,
	}
	
	violations := detector.Detect(fileInfo, astInfo)
	
	// Should only detect the global var, not the constants
	expectedViolations := 1
	if len(violations) != expectedViolations {
		t.Errorf("Expected %d violations (only GlobalVar), got %d", expectedViolations, len(violations))
		for i, v := range violations {
			t.Logf("Violation %d: Line %d, Message: %s, Code: %s", i+1, v.Line, v.Message, v.CodeSnippet)
		}
	}
	
	// Verify the violation is for GlobalVar = 42 (should be around line 12 due to whitespace)
	if len(violations) > 0 {
		violation := violations[0]
		if violation.CodeSnippet != "42" {
			t.Errorf("Expected violation for '42', got '%s' on line %d", violation.CodeSnippet, violation.Line)
		}
	}
}

// Test descriptive variable names should NOT be flagged
func TestMagicNumberDetector_DescriptiveVariableNames(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewMagicNumberDetector(config)
	
	source := `
package main

func example() {
    connectionTimeout := 45
    maxRetries := 5
    bufferSize := 1024
    maxConnections := 100
    defaultWeight := 25.0
    
    // These should be flagged
    x := 42
    y := 999
}
`
	
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}
	
	fileInfo := &models.FileInfo{
		Path:     "/test/example.go",
		Language: "Go",
	}
	astInfo := &types.GoASTInfo{
		AST:     astFile,
		FileSet: fset,
	}
	
	violations := detector.Detect(fileInfo, astInfo)
	
	// Should only detect x and y, not the descriptive variables
	expectedViolations := 2
	if len(violations) != expectedViolations {
		t.Errorf("Expected %d violations (only x and y), got %d", expectedViolations, len(violations))
		for i, v := range violations {
			t.Logf("Violation %d: Line %d, Message: %s, Code: %s", i+1, v.Line, v.Message, v.CodeSnippet)
		}
	}
}

// Test powers of 10 should NOT be flagged  
func TestMagicNumberDetector_PowersOfTen(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewMagicNumberDetector(config)
	
	testCases := []struct {
		name       string
		value      int
		acceptable bool
	}{
		{"Ten", 10, true},
		{"Hundred", 100, true},
		{"Thousand", 1000, true},
		{"Ten thousand", 10000, true},
		{"Hundred thousand", 100000, true},
		{"Million", 1000000, true},
		{"Not power of 10", 150, false},
		{"Not power of 10", 999, false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := detector.isAcceptableInt(tc.value)
			if result != tc.acceptable {
				t.Errorf("Value %d: expected acceptable=%t, got %t", tc.value, tc.acceptable, result)
			}
		})
	}
}

// Test mathematical constants in variable assignments
func TestMagicNumberDetector_MathematicalConstants(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewMagicNumberDetector(config)
	
	source := `
package main

func example() {
    pi := 3.14159
    e := 2.71828
    piApprox := 3.14
    eApprox := 2.718
    
    // These should be flagged
    randomFloat := 2.7845
    magicValue := 15.75
}
`
	
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}
	
	fileInfo := &models.FileInfo{
		Path:     "/test/example.go",
		Language: "Go",
	}
	astInfo := &types.GoASTInfo{
		AST:     astFile,
		FileSet: fset,
	}
	
	violations := detector.Detect(fileInfo, astInfo)
	
	// Should only detect randomFloat and magicValue, not the mathematical constants
	expectedViolations := 2
	if len(violations) != expectedViolations {
		t.Errorf("Expected %d violations (only randomFloat and magicValue), got %d", expectedViolations, len(violations))
		for i, v := range violations {
			t.Logf("Violation %d: Line %d, Message: %s, Code: %s", i+1, v.Line, v.Message, v.CodeSnippet)
		}
	}
}

// Test specific price constants should NOT be flagged
func TestMagicNumberDetector_PriceConstants(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewMagicNumberDetector(config)
	
	source := `
package main

const (
    YouthSportsMinPrice   = 75.0
    YouthSportsMaxPrice   = 250.0
    AdultFitnessMinPrice  = 50.0
    AdultFitnessMaxPrice  = 180.0
    SeniorsMinPrice       = 25.0
    SeniorsMaxPrice       = 120.0
    AquaticsMinPrice      = 40.0
    AquaticsMaxPrice      = 200.0
    ArtsCraftsMinPrice    = 35.0
    ArtsCraftsMaxPrice    = 150.0
    SpecialEventsMinPrice = 15.0
    SpecialEventsMaxPrice = 75.0
    CampsMinPrice         = 150.0
    CampsMaxPrice         = 400.0
)

func processOrder() {
    // This should be flagged as it's not in a const block
    x := 8.5
}
`
	
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}
	
	fileInfo := &models.FileInfo{
		Path:     "/test/example.go",
		Language: "Go",
	}
	astInfo := &types.GoASTInfo{
		AST:     astFile,
		FileSet: fset,
	}
	
	violations := detector.Detect(fileInfo, astInfo)
	
	// Should only detect the x assignment, not any of the constants
	expectedViolations := 1
	if len(violations) != expectedViolations {
		t.Errorf("Expected %d violations (only x), got %d", expectedViolations, len(violations))
		for i, v := range violations {
			t.Logf("Violation %d: Line %d, Message: %s, Code: %s", i+1, v.Line, v.Message, v.CodeSnippet)
		}
	}
	
	// Verify the violation is for x = 8.5
	if len(violations) > 0 {
		violation := violations[0]
		if violation.CodeSnippet != "8.5" {
			t.Errorf("Expected violation for '8.5', got '%s' on line %d", violation.CodeSnippet, violation.Line)
		}
	}
	
	// Verify that none of the price constants were flagged
	priceValues := []string{"75.0", "250.0", "50.0", "180.0", "25.0", "120.0", "40.0", "200.0", "35.0", "150.0", "15.0", "75.0", "150.0", "400.0"}
	for _, violation := range violations {
		for _, priceValue := range priceValues {
			if violation.CodeSnippet == priceValue {
				t.Errorf("Price constant %s should not be flagged as magic number", priceValue)
			}
		}
	}
}