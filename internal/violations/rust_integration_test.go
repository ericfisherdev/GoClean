package violations

import (
	"io/ioutil"
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

func TestRustFunctionDetector_Integration(t *testing.T) {
	// Create a test Rust file content
	rustContent := `// Test Rust file
pub fn too_many_params(a: i32, b: String, c: bool, d: Vec<u8>, e: &str) -> String {
    println!("Line 1");
    println!("Line 2");
    println!("Line 3");
    println!("Line 4");
    println!("Line 5");
    println!("Line 6");
    println!("Line 7");
    println!("Line 8");
    println!("Line 9");
    println!("Line 10");
    println!("Line 11");
    println!("Line 12");
    println!("Line 13");
    println!("Line 14");
    println!("Line 15");
    println!("Line 16");
    println!("Line 17");
    println!("Line 18");
    println!("Line 19");
    println!("Line 20");
    println!("Line 21");
    println!("Line 22");
    println!("Line 23");
    println!("Line 24");
    println!("Line 25");
    println!("Line 26");
    println!("Line 27");
    println!("Line 28");
    println!("Line 29");
    println!("Line 30");
    "result".to_string()
}

unsafe fn dangerous_function() -> *mut u8 {
    std::ptr::null_mut()
}

pub fn undocumented_public() {
    println!("This is public but has no docs");
}`

	// Write to temporary file
	tmpFile, err := ioutil.TempFile("", "test_rust_*.rs")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer tmpFile.Close()

	_, err = tmpFile.WriteString(rustContent)
	if err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}

	filePath := tmpFile.Name()

	// Create a simple RustASTInfo for testing (simulating parsed data)
	rustASTInfo := &types.RustASTInfo{
		FilePath: filePath,
		Functions: []*types.RustFunctionInfo{
			{
				Name:           "too_many_params",
				StartLine:      2,
				EndLine:        32,
				StartColumn:    1,
				EndColumn:      1,
				LineCount:      31, // Long function
				Complexity:     3,
				Parameters: []types.RustParameterInfo{
					{Name: "a", Type: "i32"},
					{Name: "b", Type: "String"},
					{Name: "c", Type: "bool"},
					{Name: "d", Type: "Vec<u8>"},
					{Name: "e", Type: "&str"},
				},
				IsPublic:       true,
				HasDocComments: false,
				Visibility:     "pub",
			},
			{
				Name:           "dangerous_function",
				StartLine:      34,
				EndLine:        36,
				StartColumn:    1,
				EndColumn:      1,
				LineCount:      3,
				Complexity:     1,
				Parameters:     []types.RustParameterInfo{},
				IsUnsafe:       true,
				HasDocComments: false,
			},
			{
				Name:           "undocumented_public",
				StartLine:      38,
				EndLine:        40,
				StartColumn:    1,
				EndColumn:      1,
				LineCount:      3,
				Complexity:     1,
				Parameters:     []types.RustParameterInfo{},
				IsPublic:       true,
				HasDocComments: false,
				Visibility:     "pub",
			},
		},
	}

	// Create detector config
	config := &DetectorConfig{
		MaxFunctionLines:         25,
		MaxCyclomaticComplexity: 10,
		MaxParameters:           3,
		RequireCommentsForPublic: true,
	}

	// Create file info
	fileInfo := &models.FileInfo{
		Path:     filePath,
		Language: "rust",
		Lines:    40,
	}

	// Test RustFunctionDetector directly
	detector := NewRustFunctionDetector(config)
	violations := detector.Detect(fileInfo, rustASTInfo)

	// Verify violations were found
	if len(violations) == 0 {
		t.Error("Expected violations to be found, got none")
	}

	// Count violation types
	violationCounts := make(map[models.ViolationType]int)
	for _, v := range violations {
		violationCounts[v.Type]++
	}

	// Verify expected violations
	expectedViolations := map[models.ViolationType]int{
		models.ViolationTypeFunctionLength:         1, // too_many_params is too long
		models.ViolationTypeParameterCount:        1, // too_many_params has too many params
		models.ViolationTypeMissingDocumentation:  3, // all three functions missing docs
	}

	for expectedType, expectedCount := range expectedViolations {
		if count, found := violationCounts[expectedType]; !found {
			t.Errorf("Expected violation type %s not found", expectedType)
		} else if count != expectedCount {
			t.Errorf("Expected %d violations of type %s, got %d", expectedCount, expectedType, count)
		}
	}

	// Verify all violations have proper fields
	for _, v := range violations {
		if v.Message == "" {
			t.Error("Violation missing message")
		}
		if v.File != filePath {
			t.Errorf("Expected file %s, got %s", filePath, v.File)
		}
		if v.Line <= 0 {
			t.Error("Violation missing valid line number")
		}
		if v.Rule == "" {
			t.Error("Violation missing rule")
		}
		if v.Suggestion == "" {
			t.Error("Violation missing suggestion")
		}
	}

	t.Logf("Successfully detected %d violations in Rust file", len(violations))
	for _, v := range violations {
		t.Logf("  - %s: %s (line %d)", v.Type, v.Message, v.Line)
	}
}

func TestRustFunctionDetector_WithDetectorRegistry(t *testing.T) {
	// Test integration with DetectorRegistry
	config := &DetectorConfig{
		MaxFunctionLines:         10,
		MaxCyclomaticComplexity: 5,
		MaxParameters:           2,
		RequireCommentsForPublic: true,
	}

	// Create detector registry and register Rust function detector
	registry := NewDetectorRegistry()
	registry.RegisterDetector(NewRustFunctionDetector(config))

	// Create file info and Rust AST info
	rustASTInfo := &types.RustASTInfo{
		FilePath: "test.rs",
		Functions: []*types.RustFunctionInfo{
			{
				Name:           "problematic_function",
				StartLine:      1,
				EndLine:        15,
				StartColumn:    1,
				EndColumn:      1,
				LineCount:      15, // Too long
				Complexity:     8,  // Too complex
				Parameters: []types.RustParameterInfo{
					{Name: "a", Type: "i32"},
					{Name: "b", Type: "String"},
					{Name: "c", Type: "bool"}, // Too many params
				},
				IsPublic:       true,
				HasDocComments: false, // Missing docs
				Visibility:     "pub",
			},
		},
	}

	fileInfo := &models.FileInfo{
		Path:     "test.rs",
		Language: "rust",
		Lines:    15,
	}

	// Run violation detection through registry
	violations := registry.DetectAll(fileInfo, rustASTInfo)

	// Verify violations were detected
	if len(violations) == 0 {
		t.Error("Expected violations to be detected by DetectorRegistry")
	}

	// Should have multiple violations for the problematic function
	violationTypes := make(map[models.ViolationType]bool)
	for _, v := range violations {
		violationTypes[v.Type] = true
	}

	expectedTypes := []models.ViolationType{
		models.ViolationTypeFunctionLength,
		models.ViolationTypeCyclomaticComplexity,
		models.ViolationTypeParameterCount,
		models.ViolationTypeMissingDocumentation,
	}

	for _, expectedType := range expectedTypes {
		if !violationTypes[expectedType] {
			t.Errorf("Expected violation type %s not found in results", expectedType)
		}
	}

	t.Logf("DetectorRegistry integration test successful: %d violations detected", len(violations))
}