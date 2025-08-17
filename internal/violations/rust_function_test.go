package violations

import (
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

func TestRustFunctionDetector_Detect(t *testing.T) {
	tests := []struct {
		name           string
		config         *DetectorConfig
		fileInfo       *models.FileInfo
		rustAstInfo    *types.RustASTInfo
		expectedCount  int
		expectedTypes  []models.ViolationType
	}{
		{
			name:   "No violations for simple function",
			config: DefaultDetectorConfig(),
			fileInfo: &models.FileInfo{
				Path:     "test.rs",
				Language: "rust",
			},
			rustAstInfo: &types.RustASTInfo{
				Functions: []*types.RustFunctionInfo{
					{
						Name:           "simple_function",
						StartLine:      1,
						EndLine:        5,
						StartColumn:    1,
						EndColumn:      1,
						LineCount:      5,
						Complexity:     2,
						Parameters:     []types.RustParameterInfo{},
						IsPublic:       false,
						HasDocComments: false,
					},
				},
			},
			expectedCount: 0,
			expectedTypes: []models.ViolationType{},
		},
		{
			name: "Function too long",
			config: &DetectorConfig{
				MaxFunctionLines:         20,
				MaxCyclomaticComplexity: 10,
				MaxParameters:           4,
			},
			fileInfo: &models.FileInfo{
				Path:     "test.rs",
				Language: "rust",
			},
			rustAstInfo: &types.RustASTInfo{
				Functions: []*types.RustFunctionInfo{
					{
						Name:        "long_function",
						StartLine:   1,
						EndLine:     50,
						StartColumn: 1,
						EndColumn:   1,
						LineCount:   50,
						Complexity:  5,
						Parameters:  []types.RustParameterInfo{},
						IsPublic:    false,
					},
				},
			},
			expectedCount: 1,
			expectedTypes: []models.ViolationType{models.ViolationTypeFunctionLength},
		},
		{
			name: "High cyclomatic complexity",
			config: &DetectorConfig{
				MaxFunctionLines:         100,
				MaxCyclomaticComplexity: 5,
				MaxParameters:           4,
			},
			fileInfo: &models.FileInfo{
				Path:     "test.rs",
				Language: "rust",
			},
			rustAstInfo: &types.RustASTInfo{
				Functions: []*types.RustFunctionInfo{
					{
						Name:        "complex_function",
						StartLine:   1,
						EndLine:     30,
						StartColumn: 1,
						EndColumn:   1,
						LineCount:   30,
						Complexity:  15,
						Parameters:  []types.RustParameterInfo{},
						IsPublic:    false,
					},
				},
			},
			expectedCount: 1,
			expectedTypes: []models.ViolationType{models.ViolationTypeCyclomaticComplexity},
		},
		{
			name: "Too many parameters",
			config: &DetectorConfig{
				MaxFunctionLines:         100,
				MaxCyclomaticComplexity: 10,
				MaxParameters:           3,
			},
			fileInfo: &models.FileInfo{
				Path:     "test.rs",
				Language: "rust",
			},
			rustAstInfo: &types.RustASTInfo{
				Functions: []*types.RustFunctionInfo{
					{
						Name:        "many_params",
						StartLine:   1,
						EndLine:     10,
						StartColumn: 1,
						EndColumn:   1,
						LineCount:   10,
						Complexity:  2,
						Parameters: []types.RustParameterInfo{
							{Name: "a", Type: "i32"},
							{Name: "b", Type: "String"},
							{Name: "c", Type: "bool"},
							{Name: "d", Type: "Vec<u8>"},
							{Name: "e", Type: "&str"},
						},
						IsPublic: false,
					},
				},
			},
			expectedCount: 1,
			expectedTypes: []models.ViolationType{models.ViolationTypeParameterCount},
		},
		{
			name: "Missing documentation on public function",
			config: &DetectorConfig{
				MaxFunctionLines:         100,
				MaxCyclomaticComplexity: 10,
				MaxParameters:           4,
				RequireCommentsForPublic: true,
			},
			fileInfo: &models.FileInfo{
				Path:     "test.rs",
				Language: "rust",
			},
			rustAstInfo: &types.RustASTInfo{
				Functions: []*types.RustFunctionInfo{
					{
						Name:           "public_function",
						StartLine:      1,
						EndLine:        5,
						StartColumn:    1,
						EndColumn:      1,
						LineCount:      5,
						Complexity:     2,
						Parameters:     []types.RustParameterInfo{},
						IsPublic:       true,
						HasDocComments: false,
						Visibility:     "pub",
					},
				},
			},
			expectedCount: 1,
			expectedTypes: []models.ViolationType{models.ViolationTypeMissingDocumentation},
		},
		{
			name: "Unsafe function without documentation",
			config: &DetectorConfig{
				MaxFunctionLines:         100,
				MaxCyclomaticComplexity: 10,
				MaxParameters:           4,
				RequireCommentsForPublic: false,
			},
			fileInfo: &models.FileInfo{
				Path:     "test.rs",
				Language: "rust",
			},
			rustAstInfo: &types.RustASTInfo{
				Functions: []*types.RustFunctionInfo{
					{
						Name:           "unsafe_function",
						StartLine:      1,
						EndLine:        5,
						StartColumn:    1,
						EndColumn:      1,
						LineCount:      5,
						Complexity:     2,
						Parameters:     []types.RustParameterInfo{},
						IsPublic:       false,
						IsUnsafe:       true,
						HasDocComments: false,
					},
				},
			},
			expectedCount: 1,
			expectedTypes: []models.ViolationType{models.ViolationTypeMissingDocumentation},
		},
		{
			name: "Complex async function",
			config: &DetectorConfig{
				MaxFunctionLines:         20,
				MaxCyclomaticComplexity: 10,
				MaxParameters:           4,
			},
			fileInfo: &models.FileInfo{
				Path:     "test.rs",
				Language: "rust",
			},
			rustAstInfo: &types.RustASTInfo{
				Functions: []*types.RustFunctionInfo{
					{
						Name:        "async_handler",
						StartLine:   1,
						EndLine:     15,
						StartColumn: 1,
						EndColumn:   1,
						LineCount:   15, // Over half of max (20/2 = 10)
						Complexity:  3,
						Parameters:  []types.RustParameterInfo{},
						IsPublic:    false,
						IsAsync:     true,
					},
				},
			},
			expectedCount: 1,
			expectedTypes: []models.ViolationType{models.ViolationTypeFunctionLength},
		},
		{
			name: "Multiple violations",
			config: &DetectorConfig{
				MaxFunctionLines:         10,
				MaxCyclomaticComplexity: 3,
				MaxParameters:           2,
				RequireCommentsForPublic: true,
			},
			fileInfo: &models.FileInfo{
				Path:     "test.rs",
				Language: "rust",
			},
			rustAstInfo: &types.RustASTInfo{
				Functions: []*types.RustFunctionInfo{
					{
						Name:        "problematic_function",
						StartLine:   1,
						EndLine:     25,
						StartColumn: 1,
						EndColumn:   1,
						LineCount:   25,
						Complexity:  8,
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
			},
			expectedCount: 4, // Length, complexity, params, missing doc
			expectedTypes: []models.ViolationType{
				models.ViolationTypeFunctionLength,
				models.ViolationTypeCyclomaticComplexity,
				models.ViolationTypeParameterCount,
				models.ViolationTypeMissingDocumentation,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewRustFunctionDetector(tt.config)
			violations := detector.Detect(tt.fileInfo, tt.rustAstInfo)

			if len(violations) != tt.expectedCount {
				t.Errorf("Expected %d violations, got %d", tt.expectedCount, len(violations))
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

			// Verify violations have required fields
			for _, v := range violations {
				if v.Message == "" {
					t.Error("Violation missing message")
				}
				if v.File == "" {
					t.Error("Violation missing file")
				}
				if v.Line == 0 {
					t.Error("Violation missing line number")
				}
				if v.Rule == "" {
					t.Error("Violation missing rule")
				}
				if v.Suggestion == "" {
					t.Error("Violation missing suggestion")
				}
			}
		})
	}
}

func TestRustFunctionDetector_GenerateSignature(t *testing.T) {
	detector := NewRustFunctionDetector(nil)

	tests := []struct {
		name     string
		function *types.RustFunctionInfo
		expected string
	}{
		{
			name: "Simple private function",
			function: &types.RustFunctionInfo{
				Name:       "simple",
				Visibility: "private",
			},
			expected: "fn simple()",
		},
		{
			name: "Public function with parameters",
			function: &types.RustFunctionInfo{
				Name:       "process",
				Visibility: "pub",
				IsPublic:   true,
				Parameters: []types.RustParameterInfo{
					{Name: "data", Type: "String", IsMutable: false, IsRef: false},
					{Name: "count", Type: "usize", IsMutable: false, IsRef: false},
				},
			},
			expected: "pub fn process(data: String, count: usize)",
		},
		{
			name: "Async unsafe function with return type",
			function: &types.RustFunctionInfo{
				Name:       "fetch_data",
				Visibility: "pub(crate)",
				IsPublic:   true,
				IsAsync:    true,
				IsUnsafe:   true,
				ReturnType: "Result<Vec<u8>, Error>",
				Parameters: []types.RustParameterInfo{
					{Name: "url", Type: "str", IsRef: true},
				},
			},
			expected: "pub(crate) async unsafe fn fetch_data(url: &str) -> Result<Vec<u8>, Error>",
		},
		{
			name: "Const function",
			function: &types.RustFunctionInfo{
				Name:       "calculate",
				Visibility: "pub",
				IsPublic:   true,
				IsConst:    true,
				ReturnType: "u32",
				Parameters: []types.RustParameterInfo{
					{Name: "x", Type: "u32"},
				},
			},
			expected: "pub const fn calculate(x: u32) -> u32",
		},
		{
			name: "Function with mutable reference parameter",
			function: &types.RustFunctionInfo{
				Name:       "modify",
				Visibility: "private",
				Parameters: []types.RustParameterInfo{
					{Name: "data", Type: "Vec<i32>", IsMutable: true, IsRef: true},
				},
			},
			expected: "fn modify(mut data: &Vec<i32>)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.generateRustFunctionSignature(tt.function)
			if result != tt.expected {
				t.Errorf("Expected signature '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestRustFunctionDetector_Severity(t *testing.T) {
	config := &DetectorConfig{
		MaxFunctionLines:        20,
		MaxCyclomaticComplexity: 5,
		MaxParameters:          3,
	}
	detector := NewRustFunctionDetector(config)

	// Test function length severity
	if severity := detector.getSeverityForFunctionLength(15); severity != models.SeverityLow {
		t.Errorf("Expected Low severity for 15 lines, got %v", severity)
	}
	if severity := detector.getSeverityForFunctionLength(35); severity != models.SeverityMedium {
		t.Errorf("Expected Medium severity for 35 lines, got %v", severity)
	}
	if severity := detector.getSeverityForFunctionLength(50); severity != models.SeverityHigh {
		t.Errorf("Expected High severity for 50 lines, got %v", severity)
	}

	// Test complexity severity
	if severity := detector.getSeverityForComplexity(4); severity != models.SeverityLow {
		t.Errorf("Expected Low severity for complexity 4, got %v", severity)
	}
	if severity := detector.getSeverityForComplexity(8); severity != models.SeverityMedium {
		t.Errorf("Expected Medium severity for complexity 8, got %v", severity)
	}
	if severity := detector.getSeverityForComplexity(12); severity != models.SeverityHigh {
		t.Errorf("Expected High severity for complexity 12, got %v", severity)
	}

	// Test parameter count severity
	if severity := detector.getSeverityForParameterCount(4); severity != models.SeverityLow {
		t.Errorf("Expected Low severity for 4 parameters, got %v", severity)
	}
	if severity := detector.getSeverityForParameterCount(5); severity != models.SeverityMedium {
		t.Errorf("Expected Medium severity for 5 parameters, got %v", severity)
	}
	if severity := detector.getSeverityForParameterCount(8); severity != models.SeverityHigh {
		t.Errorf("Expected High severity for 8 parameters, got %v", severity)
	}
}