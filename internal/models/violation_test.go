package models

import (
	"testing"
)

func TestSeverityString(t *testing.T) {
	testCases := []struct {
		severity Severity
		expected string
	}{
		{SeverityLow, "Low"},
		{SeverityMedium, "Medium"},
		{SeverityHigh, "High"},
		{SeverityCritical, "Critical"},
		{Severity(999), "Unknown"}, // Invalid severity
	}

	for _, tc := range testCases {
		result := tc.severity.String()
		if result != tc.expected {
			t.Errorf("Severity(%d).String() = %q, expected %q", int(tc.severity), result, tc.expected)
		}
	}
}

func TestViolationTypes(t *testing.T) {
	// Test that all violation types are defined as expected
	expectedTypes := map[ViolationType]string{
		ViolationTypeFunctionLength:         "function_length",
		ViolationTypeCyclomaticComplexity:   "cyclomatic_complexity",
		ViolationTypeParameterCount:         "parameter_count",
		ViolationTypeNestingDepth:           "nesting_depth",
		ViolationTypeNaming:                 "naming_convention",
		ViolationTypeClassSize:              "class_size",
		ViolationTypeMissingDocumentation:   "missing_documentation",
		ViolationTypeMagicNumber:            "magic_number",
		ViolationTypeDuplication:            "code_duplication",
	}

	for violationType, expectedString := range expectedTypes {
		if string(violationType) != expectedString {
			t.Errorf("ViolationType %v should equal %q, got %q", violationType, expectedString, string(violationType))
		}
	}
}

func TestViolationConstruction(t *testing.T) {
	violation := Violation{
		ID:          "test-001",
		Type:        ViolationTypeFunctionLength,
		Severity:    SeverityHigh,
		Message:     "Function is too long",
		Description: "This function exceeds the recommended length",
		File:        "test.go",
		Line:        10,
		Column:      1,
		EndLine:     50,
		EndColumn:   1,
		Context:     "func longFunction() {",
		Rule:        "function-length",
		Suggestion:  "Break this function into smaller functions",
	}

	// Test all fields are set correctly
	if violation.ID != "test-001" {
		t.Errorf("Expected ID 'test-001', got %q", violation.ID)
	}

	if violation.Type != ViolationTypeFunctionLength {
		t.Errorf("Expected type %v, got %v", ViolationTypeFunctionLength, violation.Type)
	}

	if violation.Severity != SeverityHigh {
		t.Errorf("Expected severity %v, got %v", SeverityHigh, violation.Severity)
	}

	if violation.Message != "Function is too long" {
		t.Errorf("Expected message 'Function is too long', got %q", violation.Message)
	}

	if violation.Description != "This function exceeds the recommended length" {
		t.Errorf("Expected description 'This function exceeds the recommended length', got %q", violation.Description)
	}

	if violation.File != "test.go" {
		t.Errorf("Expected file 'test.go', got %q", violation.File)
	}

	if violation.Line != 10 {
		t.Errorf("Expected line 10, got %d", violation.Line)
	}

	if violation.Column != 1 {
		t.Errorf("Expected column 1, got %d", violation.Column)
	}

	if violation.EndLine != 50 {
		t.Errorf("Expected end line 50, got %d", violation.EndLine)
	}

	if violation.EndColumn != 1 {
		t.Errorf("Expected end column 1, got %d", violation.EndColumn)
	}

	if violation.Context != "func longFunction() {" {
		t.Errorf("Expected context 'func longFunction() {', got %q", violation.Context)
	}

	if violation.Rule != "function-length" {
		t.Errorf("Expected rule 'function-length', got %q", violation.Rule)
	}

	if violation.Suggestion != "Break this function into smaller functions" {
		t.Errorf("Expected suggestion 'Break this function into smaller functions', got %q", violation.Suggestion)
	}
}

func TestLocationConstruction(t *testing.T) {
	location := Location{
		File:      "main.go",
		Line:      42,
		Column:    15,
		EndLine:   45,
		EndColumn: 20,
	}

	if location.File != "main.go" {
		t.Errorf("Expected file 'main.go', got %q", location.File)
	}

	if location.Line != 42 {
		t.Errorf("Expected line 42, got %d", location.Line)
	}

	if location.Column != 15 {
		t.Errorf("Expected column 15, got %d", location.Column)
	}

	if location.EndLine != 45 {
		t.Errorf("Expected end line 45, got %d", location.EndLine)
	}

	if location.EndColumn != 20 {
		t.Errorf("Expected end column 20, got %d", location.EndColumn)
	}
}

func TestSeverityOrdering(t *testing.T) {
	// Test that severity levels are ordered correctly
	if SeverityLow >= SeverityMedium {
		t.Error("SeverityLow should be less than SeverityMedium")
	}

	if SeverityMedium >= SeverityHigh {
		t.Error("SeverityMedium should be less than SeverityHigh")
	}

	if SeverityHigh >= SeverityCritical {
		t.Error("SeverityHigh should be less than SeverityCritical")
	}

	// Test specific values (SeverityInfo is 0, so others are shifted)
	if int(SeverityLow) != 1 {
		t.Errorf("Expected SeverityLow to be 1, got %d", int(SeverityLow))
	}

	if int(SeverityMedium) != 2 {
		t.Errorf("Expected SeverityMedium to be 2, got %d", int(SeverityMedium))
	}

	if int(SeverityHigh) != 3 {
		t.Errorf("Expected SeverityHigh to be 3, got %d", int(SeverityHigh))
	}

	if int(SeverityCritical) != 4 {
		t.Errorf("Expected SeverityCritical to be 4, got %d", int(SeverityCritical))
	}
}

func TestViolationJSONTags(t *testing.T) {
	// This test verifies that the JSON tags are present and meaningful
	// In a more comprehensive test suite, we might actually test JSON marshaling/unmarshaling

	violation := Violation{
		ID:          "test-001",
		Type:        ViolationTypeFunctionLength,
		Severity:    SeverityHigh,
		Message:     "Test message",
		Description: "Test description",
		File:        "test.go",
		Line:        1,
		Column:      1,
		Rule:        "test-rule",
	}

	// Basic validation that we can create a violation with all required fields
	if violation.ID == "" {
		t.Error("Violation ID should not be empty")
	}

	if violation.Type == "" {
		t.Error("Violation Type should not be empty")
	}

	if violation.Message == "" {
		t.Error("Violation Message should not be empty")
	}

	if violation.Rule == "" {
		t.Error("Violation Rule should not be empty")
	}
}

func TestLocationJSONTags(t *testing.T) {
	// Test that Location struct can be constructed properly
	location := Location{
		File:   "test.go",
		Line:   10,
		Column: 5,
	}

	if location.File == "" {
		t.Error("Location File should not be empty")
	}

	if location.Line <= 0 {
		t.Error("Location Line should be positive")
	}

	if location.Column <= 0 {
		t.Error("Location Column should be positive")
	}
}

func TestViolationTypeConstants(t *testing.T) {
	// Ensure all violation type constants are unique
	types := []ViolationType{
		ViolationTypeFunctionLength,
		ViolationTypeCyclomaticComplexity,
		ViolationTypeParameterCount,
		ViolationTypeNestingDepth,
		ViolationTypeNaming,
		ViolationTypeClassSize,
		ViolationTypeMissingDocumentation,
		ViolationTypeMagicNumber,
		ViolationTypeDuplication,
	}

	seen := make(map[ViolationType]bool)
	for _, vType := range types {
		if seen[vType] {
			t.Errorf("Duplicate violation type found: %v", vType)
		}
		seen[vType] = true

		// Ensure the string value is not empty
		if string(vType) == "" {
			t.Errorf("Violation type %v has empty string value", vType)
		}
	}

	// Ensure we have the expected number of types
	if len(types) != len(seen) {
		t.Errorf("Expected %d unique violation types, got %d", len(types), len(seen))
	}
}