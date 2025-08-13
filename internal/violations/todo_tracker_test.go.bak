package violations
import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)
func TestTodoTrackerDetector_Detect_NoMarkers(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewTodoTrackerDetector(config)
	source := `
package main
// This is a regular comment
// Another normal comment describing the function
func example() {
	// Simple inline comment
	return 42
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
	if len(violations) != 0 {
		t.Errorf("Expected no violations, got %d", len(violations))
	}
}
func TestTodoTrackerDetector_Detect_WithMarkers(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewTodoTrackerDetector(config)
	source := `
package main
// TODO: Implement user authentication
func login() {
	// FIXME: This validation is incomplete
	return false
}
// HACK: Temporary workaround for database connection
func connectDB() {
	// XXX: This is dangerous and needs review
	return nil
}
// BUG: Function returns wrong value sometimes
func calculate() {
	// OPTIMIZE: This could be more efficient
	return 0
}
// REFACTOR: Extract this logic into separate functions
func processData() {
	return
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
	// Should detect all markers
	expectedCount := 7 // TODO, FIXME, HACK, XXX, BUG, OPTIMIZE, REFACTOR
	if len(violations) != expectedCount {
		t.Errorf("Expected %d violations, got %d", expectedCount, len(violations))
	}
	// Check violation types and severities
	markerSeverities := map[string]models.Severity{
		"TODO":     models.SeverityLow,
		"FIXME":    models.SeverityHigh,
		"HACK":     models.SeverityMedium,
		"XXX":      models.SeverityMedium,
		"BUG":      models.SeverityHigh,
		"OPTIMIZE": models.SeverityLow,
		"REFACTOR": models.SeverityLow,
	}
	markerCounts := make(map[string]int)
	for _, violation := range violations {
		if violation.Type != models.ViolationTypeTodo {
			t.Errorf("Expected todo violation type, got %s", violation.Type)
		}
		// Extract marker from message
		for marker, expectedSeverity := range markerSeverities {
			if len(violation.Message) > len(marker) && violation.Message[:len(marker)] == marker {
				markerCounts[marker]++
				if violation.Severity != expectedSeverity {
					t.Errorf("Marker %s: expected severity %s, got %s", marker, expectedSeverity, violation.Severity)
				}
				break
			}
		}
	}
	// Check all markers were found
	for marker := range markerSeverities {
		if markerCounts[marker] == 0 {
			t.Errorf("Marker %s was not detected", marker)
		}
	}
}
func TestTodoTrackerDetector_SeverityClassification(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewTodoTrackerDetector(config)
	testCases := []struct {
		marker           string
		expectedSeverity models.Severity
	}{
		{"TODO", models.SeverityLow},
		{"FIXME", models.SeverityHigh},
		{"HACK", models.SeverityMedium},
		{"XXX", models.SeverityMedium},
		{"BUG", models.SeverityHigh},
		{"OPTIMIZE", models.SeverityLow},
		{"REFACTOR", models.SeverityLow},
		{"UNKNOWN", models.SeverityInfo},
	}
	for _, tc := range testCases {
		t.Run(tc.marker, func(t *testing.T) {
			severity := detector.classifyMarkerSeverity(tc.marker)
			if severity != tc.expectedSeverity {
				t.Errorf("Marker %s: expected severity %s, got %s", tc.marker, tc.expectedSeverity, severity)
			}
		})
	}
}
func TestTodoTrackerDetector_CaseInsensitive(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewTodoTrackerDetector(config)
	source := `
package main
// todo: lowercase marker
// Todo: mixed case marker
// TODO: uppercase marker
func example() {
	return
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
	// Should detect all variations
	expectedCount := 3
	if len(violations) != expectedCount {
		t.Errorf("Expected %d violations, got %d", expectedCount, len(violations))
	}
}