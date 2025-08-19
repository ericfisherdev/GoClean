package violations

import (
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// TestRustDocumentationDetector_Name tests the detector name
func TestRustDocumentationDetector_Name(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustDocumentationDetector(config)

	expectedName := "Rust Documentation Quality"
	if detector.Name() != expectedName {
		t.Errorf("Expected name '%s', got '%s'", expectedName, detector.Name())
	}
}

// TestRustDocumentationDetector_Description tests the detector description
func TestRustDocumentationDetector_Description(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustDocumentationDetector(config)

	description := detector.Description()
	if description == "" {
		t.Error("Expected non-empty description")
	}

	// Check that description mentions key components
	expectedKeywords := []string{"missing", "documentation", "Rust"}
	for _, keyword := range expectedKeywords {
		if !containsString(description, keyword) {
			t.Errorf("Description should contain '%s': %s", keyword, description)
		}
	}
}

// TestRustDocumentationDetector_DetectWithNilASTInfo tests detection with nil AST info
func TestRustDocumentationDetector_DetectWithNilASTInfo(t *testing.T) {
	config := DefaultDetectorConfig()
	config.RequireCommentsForPublic = true
	detector := NewRustDocumentationDetector(config)

	fileInfo := &models.FileInfo{
		Path:     "test.rs",
		Language: "rust",
	}

	violations := detector.Detect(fileInfo, nil)
	if len(violations) != 0 {
		t.Errorf("Expected 0 violations for nil AST info, got %d", len(violations))
	}
}

// TestRustDocumentationDetector_DetectWithInvalidASTInfo tests detection with invalid AST info
func TestRustDocumentationDetector_DetectWithInvalidASTInfo(t *testing.T) {
	config := DefaultDetectorConfig()
	config.RequireCommentsForPublic = true
	detector := NewRustDocumentationDetector(config)

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

// TestRustDocumentationDetector_DetectPublicFunctionWithoutDocs tests missing function documentation
func TestRustDocumentationDetector_DetectPublicFunctionWithoutDocs(t *testing.T) {
	config := DefaultDetectorConfig()
	config.RequireCommentsForPublic = true
	detector := NewRustDocumentationDetector(config)

	fileInfo := &models.FileInfo{
		Path:     "test.rs",
		Language: "rust",
	}

	rustAstInfo := &types.RustASTInfo{
		FilePath: "test.rs",
		Functions: []*types.RustFunctionInfo{
			{
				Name:           "public_function",
				StartLine:      10,
				StartColumn:    5,
				IsPublic:       true,
				HasDocComments: false,
				Visibility:     "pub",
			},
		},
	}

	violations := detector.Detect(fileInfo, rustAstInfo)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
		return
	}

	violation := violations[0]
	if violation.Type != models.ViolationTypeMissingDocumentation {
		t.Errorf("Expected violation type %s, got %s", models.ViolationTypeMissingDocumentation, violation.Type)
	}

	if violation.Severity != models.SeverityMedium {
		t.Errorf("Expected severity %s, got %s", models.SeverityMedium, violation.Severity)
	}

	if violation.Line != 10 {
		t.Errorf("Expected line 10, got %d", violation.Line)
	}

	if violation.Rule != "rust-missing-function-documentation" {
		t.Errorf("Expected rule 'rust-missing-function-documentation', got '%s'", violation.Rule)
	}
}

// TestRustDocumentationDetector_DetectUnsafeFunctionWithoutDocs tests missing unsafe function documentation
func TestRustDocumentationDetector_DetectUnsafeFunctionWithoutDocs(t *testing.T) {
	config := DefaultDetectorConfig()
	config.RequireCommentsForPublic = true
	detector := NewRustDocumentationDetector(config)

	fileInfo := &models.FileInfo{
		Path:     "test.rs",
		Language: "rust",
	}

	rustAstInfo := &types.RustASTInfo{
		FilePath: "test.rs",
		Functions: []*types.RustFunctionInfo{
			{
				Name:           "unsafe_function",
				StartLine:      15,
				StartColumn:    5,
				IsPublic:       true,
				IsUnsafe:       true,
				HasDocComments: false,
				Visibility:     "pub",
			},
		},
	}

	violations := detector.Detect(fileInfo, rustAstInfo)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
		return
	}

	violation := violations[0]
	if violation.Severity != models.SeverityHigh {
		t.Errorf("Expected high severity for unsafe function, got %s", violation.Severity)
	}

	expectedSuggestion := "Add doc comments (///) describing what function 'unsafe_function' does, its parameters, and return value. Include a # Safety section explaining the safety requirements"
	if violation.Suggestion != expectedSuggestion {
		t.Errorf("Expected suggestion about safety section, got: %s", violation.Suggestion)
	}
}

// TestRustDocumentationDetector_DetectPublicStructWithoutDocs tests missing struct documentation
func TestRustDocumentationDetector_DetectPublicStructWithoutDocs(t *testing.T) {
	config := DefaultDetectorConfig()
	config.RequireCommentsForPublic = true
	detector := NewRustDocumentationDetector(config)

	fileInfo := &models.FileInfo{
		Path:     "test.rs",
		Language: "rust",
	}

	rustAstInfo := &types.RustASTInfo{
		FilePath: "test.rs",
		Structs: []*types.RustStructInfo{
			{
				Name:           "PublicStruct",
				StartLine:      20,
				StartColumn:    5,
				IsPublic:       true,
				HasDocComments: false,
				Visibility:     "pub",
			},
		},
	}

	violations := detector.Detect(fileInfo, rustAstInfo)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
		return
	}

	violation := violations[0]
	if violation.Type != models.ViolationTypeMissingDocumentation {
		t.Errorf("Expected violation type %s, got %s", models.ViolationTypeMissingDocumentation, violation.Type)
	}

	if violation.Rule != "rust-missing-struct-documentation" {
		t.Errorf("Expected rule 'rust-missing-struct-documentation', got '%s'", violation.Rule)
	}
}

// TestRustDocumentationDetector_DetectPublicEnumWithoutDocs tests missing enum documentation
func TestRustDocumentationDetector_DetectPublicEnumWithoutDocs(t *testing.T) {
	config := DefaultDetectorConfig()
	config.RequireCommentsForPublic = true
	detector := NewRustDocumentationDetector(config)

	fileInfo := &models.FileInfo{
		Path:     "test.rs",
		Language: "rust",
	}

	rustAstInfo := &types.RustASTInfo{
		FilePath: "test.rs",
		Enums: []*types.RustEnumInfo{
			{
				Name:           "PublicEnum",
				StartLine:      25,
				StartColumn:    5,
				IsPublic:       true,
				HasDocComments: false,
				Visibility:     "pub",
			},
		},
	}

	violations := detector.Detect(fileInfo, rustAstInfo)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
		return
	}

	violation := violations[0]
	if violation.Rule != "rust-missing-enum-documentation" {
		t.Errorf("Expected rule 'rust-missing-enum-documentation', got '%s'", violation.Rule)
	}
}

// TestRustDocumentationDetector_DetectPublicTraitWithoutDocs tests missing trait documentation
func TestRustDocumentationDetector_DetectPublicTraitWithoutDocs(t *testing.T) {
	config := DefaultDetectorConfig()
	config.RequireCommentsForPublic = true
	detector := NewRustDocumentationDetector(config)

	fileInfo := &models.FileInfo{
		Path:     "test.rs",
		Language: "rust",
	}

	rustAstInfo := &types.RustASTInfo{
		FilePath: "test.rs",
		Traits: []*types.RustTraitInfo{
			{
				Name:           "PublicTrait",
				StartLine:      30,
				StartColumn:    5,
				IsPublic:       true,
				HasDocComments: false,
				Visibility:     "pub",
			},
		},
	}

	violations := detector.Detect(fileInfo, rustAstInfo)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
		return
	}

	violation := violations[0]
	if violation.Rule != "rust-missing-trait-documentation" {
		t.Errorf("Expected rule 'rust-missing-trait-documentation', got '%s'", violation.Rule)
	}
}

// TestRustDocumentationDetector_DetectPublicModuleWithoutDocs tests missing module documentation
func TestRustDocumentationDetector_DetectPublicModuleWithoutDocs(t *testing.T) {
	config := DefaultDetectorConfig()
	config.RequireCommentsForPublic = true
	detector := NewRustDocumentationDetector(config)

	fileInfo := &models.FileInfo{
		Path:     "test.rs",
		Language: "rust",
	}

	rustAstInfo := &types.RustASTInfo{
		FilePath: "test.rs",
		Modules: []*types.RustModuleInfo{
			{
				Name:           "public_module",
				StartLine:      35,
				StartColumn:    5,
				IsPublic:       true,
				HasDocComments: false,
				Visibility:     "pub",
			},
		},
	}

	violations := detector.Detect(fileInfo, rustAstInfo)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
		return
	}

	violation := violations[0]
	if violation.Severity != models.SeverityLow {
		t.Errorf("Expected low severity for module, got %s", violation.Severity)
	}

	if violation.Rule != "rust-missing-module-documentation" {
		t.Errorf("Expected rule 'rust-missing-module-documentation', got '%s'", violation.Rule)
	}
}

// TestRustDocumentationDetector_DetectPublicConstantWithoutDocs tests missing constant documentation
func TestRustDocumentationDetector_DetectPublicConstantWithoutDocs(t *testing.T) {
	config := DefaultDetectorConfig()
	config.RequireCommentsForPublic = true
	detector := NewRustDocumentationDetector(config)

	fileInfo := &models.FileInfo{
		Path:     "test.rs",
		Language: "rust",
	}

	rustAstInfo := &types.RustASTInfo{
		FilePath: "test.rs",
		Constants: []*types.RustConstantInfo{
			{
				Name:           "PUBLIC_CONSTANT",
				Line:           40,
				Column:         5,
				IsPublic:       true,
				HasDocComments: false,
				Visibility:     "pub",
			},
		},
	}

	violations := detector.Detect(fileInfo, rustAstInfo)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
		return
	}

	violation := violations[0]
	if violation.Severity != models.SeverityLow {
		t.Errorf("Expected low severity for constant, got %s", violation.Severity)
	}

	if violation.Rule != "rust-missing-constant-documentation" {
		t.Errorf("Expected rule 'rust-missing-constant-documentation', got '%s'", violation.Rule)
	}
}

// TestRustDocumentationDetector_DetectPublicMacroWithoutDocs tests missing macro documentation
func TestRustDocumentationDetector_DetectPublicMacroWithoutDocs(t *testing.T) {
	config := DefaultDetectorConfig()
	config.RequireCommentsForPublic = true
	detector := NewRustDocumentationDetector(config)

	fileInfo := &models.FileInfo{
		Path:     "test.rs",
		Language: "rust",
	}

	rustAstInfo := &types.RustASTInfo{
		FilePath: "test.rs",
		Macros: []*types.RustMacroInfo{
			{
				Name:           "public_macro",
				StartLine:      45,
				StartColumn:    5,
				IsPublic:       true,
				HasDocComments: false,
				MacroType:      "macro_rules!",
			},
		},
	}

	violations := detector.Detect(fileInfo, rustAstInfo)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
		return
	}

	violation := violations[0]
	if violation.Severity != models.SeverityMedium {
		t.Errorf("Expected medium severity for macro, got %s", violation.Severity)
	}

	if violation.Rule != "rust-missing-macro-documentation" {
		t.Errorf("Expected rule 'rust-missing-macro-documentation', got '%s'", violation.Rule)
	}

	expectedSuggestion := "Add doc comments (///) describing what macro_rules! macro 'public_macro' does, its syntax, and usage examples"
	if violation.Suggestion != expectedSuggestion {
		t.Errorf("Expected macro-specific suggestion, got: %s", violation.Suggestion)
	}
}

// TestRustDocumentationDetector_IgnorePrivateItems tests that private items are ignored
func TestRustDocumentationDetector_IgnorePrivateItems(t *testing.T) {
	config := DefaultDetectorConfig()
	config.RequireCommentsForPublic = true
	detector := NewRustDocumentationDetector(config)

	fileInfo := &models.FileInfo{
		Path:     "test.rs",
		Language: "rust",
	}

	rustAstInfo := &types.RustASTInfo{
		FilePath: "test.rs",
		Functions: []*types.RustFunctionInfo{
			{
				Name:           "private_function",
				StartLine:      10,
				StartColumn:    5,
				IsPublic:       false,
				HasDocComments: false,
				Visibility:     "private",
			},
		},
		Structs: []*types.RustStructInfo{
			{
				Name:           "PrivateStruct",
				StartLine:      20,
				StartColumn:    5,
				IsPublic:       false,
				HasDocComments: false,
				Visibility:     "private",
			},
		},
	}

	violations := detector.Detect(fileInfo, rustAstInfo)

	if len(violations) != 0 {
		t.Errorf("Expected 0 violations for private items, got %d", len(violations))
	}
}

// TestRustDocumentationDetector_IgnoreDocumentedItems tests that documented items don't generate violations
func TestRustDocumentationDetector_IgnoreDocumentedItems(t *testing.T) {
	config := DefaultDetectorConfig()
	config.RequireCommentsForPublic = true
	detector := NewRustDocumentationDetector(config)

	fileInfo := &models.FileInfo{
		Path:     "test.rs",
		Language: "rust",
	}

	rustAstInfo := &types.RustASTInfo{
		FilePath: "test.rs",
		Functions: []*types.RustFunctionInfo{
			{
				Name:           "documented_function",
				StartLine:      10,
				StartColumn:    5,
				IsPublic:       true,
				HasDocComments: true,
				Visibility:     "pub",
			},
		},
		Structs: []*types.RustStructInfo{
			{
				Name:           "DocumentedStruct",
				StartLine:      20,
				StartColumn:    5,
				IsPublic:       true,
				HasDocComments: true,
				Visibility:     "pub",
			},
		},
	}

	violations := detector.Detect(fileInfo, rustAstInfo)

	if len(violations) != 0 {
		t.Errorf("Expected 0 violations for documented items, got %d", len(violations))
	}
}

// TestRustDocumentationDetector_RequireCommentsDisabled tests behavior when RequireCommentsForPublic is disabled
func TestRustDocumentationDetector_RequireCommentsDisabled(t *testing.T) {
	config := DefaultDetectorConfig()
	config.RequireCommentsForPublic = false
	detector := NewRustDocumentationDetector(config)

	fileInfo := &models.FileInfo{
		Path:     "test.rs",
		Language: "rust",
	}

	rustAstInfo := &types.RustASTInfo{
		FilePath: "test.rs",
		Functions: []*types.RustFunctionInfo{
			{
				Name:           "undocumented_function",
				StartLine:      10,
				StartColumn:    5,
				IsPublic:       true,
				HasDocComments: false,
				Visibility:     "pub",
			},
		},
	}

	violations := detector.Detect(fileInfo, rustAstInfo)

	if len(violations) != 0 {
		t.Errorf("Expected 0 violations when RequireCommentsForPublic is disabled, got %d", len(violations))
	}
}

// TestRustDocumentationDetector_MultipleViolations tests detection of multiple violations
func TestRustDocumentationDetector_MultipleViolations(t *testing.T) {
	config := DefaultDetectorConfig()
	config.RequireCommentsForPublic = true
	detector := NewRustDocumentationDetector(config)

	fileInfo := &models.FileInfo{
		Path:     "test.rs",
		Language: "rust",
	}

	rustAstInfo := &types.RustASTInfo{
		FilePath: "test.rs",
		Functions: []*types.RustFunctionInfo{
			{
				Name:           "undocumented_function",
				StartLine:      10,
				StartColumn:    5,
				IsPublic:       true,
				HasDocComments: false,
				Visibility:     "pub",
			},
		},
		Structs: []*types.RustStructInfo{
			{
				Name:           "UndocumentedStruct",
				StartLine:      20,
				StartColumn:    5,
				IsPublic:       true,
				HasDocComments: false,
				Visibility:     "pub",
			},
		},
		Enums: []*types.RustEnumInfo{
			{
				Name:           "UndocumentedEnum",
				StartLine:      30,
				StartColumn:    5,
				IsPublic:       true,
				HasDocComments: false,
				Visibility:     "pub",
			},
		},
	}

	violations := detector.Detect(fileInfo, rustAstInfo)

	expectedViolations := 3
	if len(violations) != expectedViolations {
		t.Errorf("Expected %d violations, got %d", expectedViolations, len(violations))
	}

	// Check that we have one of each type
	rules := make(map[string]int)
	for _, violation := range violations {
		rules[violation.Rule]++
	}

	expectedRules := []string{
		"rust-missing-function-documentation",
		"rust-missing-struct-documentation",
		"rust-missing-enum-documentation",
	}

	for _, rule := range expectedRules {
		if rules[rule] != 1 {
			t.Errorf("Expected 1 violation for rule '%s', got %d", rule, rules[rule])
		}
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr || 
		     findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}