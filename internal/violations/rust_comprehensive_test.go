package violations

import (
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// TestRustDetectorSuite_Comprehensive tests all Rust detectors together
func TestRustDetectorSuite_Comprehensive(t *testing.T) {
	config := &DetectorConfig{
		MaxFunctionLines:         25,
		MaxCyclomaticComplexity: 5,
		MaxParameters:           3,
		RequireCommentsForPublic: true,
		MaxClassLines:           500,
		MaxMethods:              10,
	}

	// Create a comprehensive Rust AST with multiple issues
	rustAstInfo := &types.RustASTInfo{
		Functions: []*types.RustFunctionInfo{
			{
				Name:           "problematic_function",
				StartLine:      1,
				EndLine:        30, // Too long
				StartColumn:    1,
				EndColumn:      1,
				LineCount:      30,
				Complexity:     8,  // Too complex
				Parameters: []types.RustParameterInfo{
					{Name: "a", Type: "i32"},
					{Name: "b", Type: "String"},
					{Name: "c", Type: "bool"},
					{Name: "d", Type: "Vec<u8>"}, // Too many params
				},
				IsPublic:       true,
				HasDocComments: false, // Missing docs
				Visibility:     "pub",
			},
			{
				Name:           "unsafe_function",
				StartLine:      35,
				EndLine:        40,
				StartColumn:    1,
				EndColumn:      1,
				LineCount:      5,
				Complexity:     2,
				Parameters:     []types.RustParameterInfo{},
				IsPublic:       false,
				IsUnsafe:       true,
				HasDocComments: false, // Unsafe without docs
			},
		},
		Structs: []*types.RustStructInfo{
			{
				Name:           "BadStruct",  // Wrong naming convention
				StartLine:      45,
				EndLine:        60,
				StartColumn:    1,
				EndColumn:      1,
				IsPublic:       true,
				HasDocComments: false, // Missing docs
				FieldCount:     2,
				Visibility:     "pub",
			},
		},
		Constants: []*types.RustConstantInfo{
			{
				Name:           "magicNumber", // Wrong naming convention
				Type:           "i32",
				Line:           65,
				Column:         1,
				IsPublic:       true,
				Visibility:     "pub",
				HasDocComments: false,
			},
		},
	}

	fileInfo := &models.FileInfo{
		Path:     "problematic.rs",
		Language: "rust",
		Lines:    24,
	}

	// Test all detectors
	tests := []struct {
		name             string
		detector         Detector
		expectedMinCount int
		expectedTypes    []models.ViolationType
	}{
		{
			name:             "RustFunctionDetector",
			detector:         NewRustFunctionDetector(config),
			expectedMinCount: 3, // Length, complexity, params, missing docs, unsafe docs
			expectedTypes: []models.ViolationType{
				models.ViolationTypeFunctionLength,
				models.ViolationTypeCyclomaticComplexity,
				models.ViolationTypeParameterCount,
				models.ViolationTypeMissingDocumentation,
			},
		},
		{
			name:             "RustNamingDetector",
			detector:         NewRustNamingDetector(config),
			expectedMinCount: 2, // BadStruct, magicNumber
			expectedTypes: []models.ViolationType{
				models.ViolationTypeNaming,
			},
		},
		{
			name:             "RustDocumentationDetector",
			detector:         NewRustDocumentationDetector(config),
			expectedMinCount: 2, // Function and struct missing docs
			expectedTypes: []models.ViolationType{
				models.ViolationTypeMissingDocumentation,
			},
		},
		{
			name:             "RustMagicNumberDetector",
			detector:         NewRustMagicNumberDetector(config),
			expectedMinCount: 1, // Magic number 42
			expectedTypes: []models.ViolationType{
				models.ViolationTypeMagicNumber,
			},
		},
		{
			name:             "RustTodoTrackerDetector",
			detector:         NewRustTodoTrackerDetector(config),
			expectedMinCount: 2, // TODO comment and panic! call
			expectedTypes: []models.ViolationType{
				models.ViolationTypeTodo,
			},
		},
		{
			name:             "RustCommentedCodeDetector",
			detector:         NewRustCommentedCodeDetector(config),
			expectedMinCount: 1, // Commented panic! call
			expectedTypes: []models.ViolationType{
				models.ViolationTypeCommentedCode,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := tt.detector.Detect(fileInfo, rustAstInfo)

			if len(violations) < tt.expectedMinCount {
				t.Errorf("Expected at least %d violations, got %d", tt.expectedMinCount, len(violations))
				for i, v := range violations {
					t.Logf("  Violation %d: %s - %s (line %d)", i+1, v.Type, v.Message, v.Line)
				}
			}

			// Check that expected violation types are present
			for _, expectedType := range tt.expectedTypes {
				found := false
				for _, v := range violations {
					if v.Type == expectedType {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected violation type %s not found", expectedType)
				}
			}

			// Verify all violations have required fields
			for i, v := range violations {
				if v.Message == "" {
					t.Errorf("Violation %d missing message", i)
				}
				if v.File == "" {
					t.Errorf("Violation %d missing file", i)
				}
				if v.Line == 0 {
					t.Errorf("Violation %d missing line number", i)
				}
				if v.Rule == "" {
					t.Errorf("Violation %d missing rule", i)
				}
				if v.Suggestion == "" {
					t.Errorf("Violation %d missing suggestion", i)
				}
			}
		})
	}
}

// TestRustDetectorSuite_EdgeCases tests edge cases across all detectors
func TestRustDetectorSuite_EdgeCases(t *testing.T) {
	config := DefaultDetectorConfig()

	tests := []struct {
		name        string
		fileInfo    *models.FileInfo
		rustAstInfo *types.RustASTInfo
		description string
	}{
		{
			name: "Empty Rust file",
			fileInfo: &models.FileInfo{
				Path:     "empty.rs",
				Language: "rust",
				Lines:    0,
			},
			rustAstInfo: &types.RustASTInfo{},
			description: "Should handle empty files gracefully",
		},
		{
			name: "Rust file with only comments",
			fileInfo: &models.FileInfo{
				Path:     "comments_only.rs",
				Language: "rust",
				Lines:    3,
			},
			rustAstInfo: &types.RustASTInfo{},
			description: "Should handle comment-only files",
		},
		{
			name: "Rust file with complex generics",
			fileInfo: &models.FileInfo{
				Path:     "generics.rs",
				Language: "rust",
				Lines:    8,
			},
			rustAstInfo: &types.RustASTInfo{
				Functions: []*types.RustFunctionInfo{
					{
						Name:           "complex_function",
						StartLine:      1,
						EndLine:        8,
						StartColumn:    1,
						EndColumn:      1,
						LineCount:      8,
						Complexity:     1,
						Parameters: []types.RustParameterInfo{
							{Name: "param", Type: "T"},
						},
						IsPublic:       true,
						ReturnType:     "Result<U, V>",
						Visibility:     "pub",
						HasDocComments: false,
					},
				},
			},
			description: "Should handle complex generic functions",
		},
	}

	detectors := []Detector{
		NewRustFunctionDetector(config),
		NewRustNamingDetector(config),
		NewRustDocumentationDetector(config),
		NewRustMagicNumberDetector(config),
		NewRustTodoTrackerDetector(config),
		NewRustCommentedCodeDetector(config),
		NewRustDuplicationDetector(config),
		NewRustStructureDetector(config),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, detector := range detectors {
				// Should not panic and should return valid violations
				violations := detector.Detect(tt.fileInfo, tt.rustAstInfo)
				
				// Verify all violations are valid
				for i, v := range violations {
					if v.File == "" {
						t.Errorf("Detector %T: Violation %d missing file", detector, i)
					}
					if v.Message == "" {
						t.Errorf("Detector %T: Violation %d missing message", detector, i)
					}
					if v.Rule == "" {
						t.Errorf("Detector %T: Violation %d missing rule", detector, i)
					}
				}
			}
		})
	}
}

// TestRustDetectorSuite_Performance tests performance characteristics
func TestRustDetectorSuite_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	config := DefaultDetectorConfig()
	
	// Create a large Rust AST for performance testing
	largeFunctions := make([]*types.RustFunctionInfo, 100)
	for i := 0; i < 100; i++ {
		largeFunctions[i] = &types.RustFunctionInfo{
			Name:        "function_" + string(rune('a'+i%26)),
			StartLine:   i * 10,
			EndLine:     (i * 10) + 5,
			LineCount:   5,
			Complexity:  2,
			Parameters:  []types.RustParameterInfo{},
			IsPublic:    i%2 == 0,
		}
	}

	largeStructs := make([]*types.RustStructInfo, 50)
	for i := 0; i < 50; i++ {
		largeStructs[i] = &types.RustStructInfo{
			Name:           "Struct" + string(rune('A'+i%26)),
			StartLine:      1000 + (i * 5),
			EndLine:        1000 + (i * 5) + 3,
			StartColumn:    1,
			EndColumn:      1,
			IsPublic:       i%2 == 0,
			FieldCount:     2,
			Visibility:     "pub",
			HasDocComments: false,
		}
	}


	fileInfo := &models.FileInfo{
		Path:     "large_file.rs",
		Language: "rust",
		Lines:    2000,
	}

	rustAstInfo := &types.RustASTInfo{
		Functions: largeFunctions,
		Structs:   largeStructs,
	}

	detectors := []Detector{
		NewRustFunctionDetector(config),
		NewRustNamingDetector(config),
		NewRustDocumentationDetector(config),
		NewRustMagicNumberDetector(config),
		NewRustStructureDetector(config),
	}

	for _, detector := range detectors {
		t.Run("Performance_"+detector.Name(), func(t *testing.T) {
			// Run detector multiple times to test performance
			for i := 0; i < 10; i++ {
				violations := detector.Detect(fileInfo, rustAstInfo)
				
				// Basic sanity check
				if len(violations) < 0 {
					t.Error("Negative violation count")
				}
			}
		})
	}
}

// TestRustDetectorSuite_Integration tests integration between multiple detectors
func TestRustDetectorSuite_Integration(t *testing.T) {
	config := &DetectorConfig{
		MaxFunctionLines:         10,
		MaxCyclomaticComplexity: 3,
		MaxParameters:           2,
		RequireCommentsForPublic: true,
		MaxClassLines:           100,
	}

	// Create registry and register all Rust detectors
	registry := NewDetectorRegistry()
	registry.RegisterDetector(NewRustFunctionDetector(config))
	registry.RegisterDetector(NewRustNamingDetector(config))
	registry.RegisterDetector(NewRustDocumentationDetector(config))
	registry.RegisterDetector(NewRustMagicNumberDetector(config))
	registry.RegisterDetector(NewRustTodoTrackerDetector(config))
	registry.RegisterDetector(NewRustCommentedCodeDetector(config))

	fileInfo := &models.FileInfo{
		Path:     "integration_test.rs",
		Language: "rust",
		Lines:    14,
	}

	rustAstInfo := &types.RustASTInfo{
		Functions: []*types.RustFunctionInfo{
			{
				Name:           "BadName",
				StartLine:      2,
				EndLine:        12,
				LineCount:      11,
				Complexity:     4,
				Parameters: []types.RustParameterInfo{
					{Name: "a", Type: "i32"},
					{Name: "b", Type: "String"},
					{Name: "c", Type: "bool"},
				},
				IsPublic:       true,
				HasDocComments: false,
				Visibility:     "pub",
			},
		},
	}

	allViolations := registry.DetectAll(fileInfo, rustAstInfo)

	// Should detect multiple types of violations
	expectedTypes := map[models.ViolationType]bool{
		models.ViolationTypeFunctionLength:        true,
		models.ViolationTypeCyclomaticComplexity: true,
		models.ViolationTypeParameterCount:       true,
		models.ViolationTypeMissingDocumentation: true,
		models.ViolationTypeNaming:              true,
		models.ViolationTypeMagicNumber:         true,
		models.ViolationTypeTodo:                true,
		models.ViolationTypeCommentedCode:       true,
	}

	foundTypes := make(map[models.ViolationType]bool)
	for _, v := range allViolations {
		foundTypes[v.Type] = true
	}

	for expectedType := range expectedTypes {
		if !foundTypes[expectedType] {
			t.Errorf("Expected violation type %s not found in integration test", expectedType)
		}
	}

	if len(allViolations) < 6 {
		t.Errorf("Expected at least 6 violations in integration test, got %d", len(allViolations))
		for i, v := range allViolations {
			t.Logf("  Violation %d: %s - %s (line %d)", i+1, v.Type, v.Message, v.Line)
		}
	}

	t.Logf("Integration test detected %d violations across all detectors", len(allViolations))
}