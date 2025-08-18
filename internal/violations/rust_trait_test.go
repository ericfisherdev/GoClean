package violations

import (
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

func TestRustTraitDetector(t *testing.T) {
	config := &DetectorConfig{
		RustConfig: &RustDetectorConfig{
			MaxTraitComplexity:    10,
			MaxTraitLines:        30,
			MaxTraitMethods:      5,
			MaxAssociatedTypes:   3,
			MaxComplexTraitParams: 1,
		},
	}
	detector := NewRustTraitDetector(config)

	tests := []struct {
		name           string
		astInfo        *types.RustASTInfo
		expectedCount  int
		expectedTypes  []models.ViolationType
	}{
		{
			name: "valid traits",
			astInfo: &types.RustASTInfo{
				FilePath: "test.rs",
				Traits: []*types.RustTraitInfo{
					{
						Name:           "ValidTrait",
						StartLine:      1,
						EndLine:        10,
						StartColumn:    1,
						EndColumn:      2,
						MethodCount:    3,
						IsPublic:       true,
						HasDocComments: true,
					},
				},
			},
			expectedCount: 0,
			expectedTypes: []models.ViolationType{},
		},
		{
			name: "overly complex trait",
			astInfo: &types.RustASTInfo{
				FilePath: "test.rs",
				Traits: []*types.RustTraitInfo{
					{
						Name:           "VeryComplexTraitWithManyMethods",
						StartLine:      1,
						EndLine:        100,
						StartColumn:    1,
						EndColumn:      2,
						MethodCount:    15,
						IsPublic:       true,
						HasDocComments: false,
					},
				},
			},
			expectedCount: 4, // complexity, size (lines), size (methods), associated types
			expectedTypes: []models.ViolationType{
				models.ViolationTypeRustOverlyComplexTrait,
			},
		},
		{
			name: "invalid trait naming",
			astInfo: &types.RustASTInfo{
				FilePath: "test.rs",
				Traits: []*types.RustTraitInfo{
					{
						Name:           "invalid_trait_name",
						StartLine:      1,
						EndLine:        5,
						StartColumn:    1,
						EndColumn:      2,
						MethodCount:    2,
						IsPublic:       true,
						HasDocComments: true,
					},
					{
						Name:           "T", // Non-descriptive name
						StartLine:      10,
						EndLine:        15,
						StartColumn:    1,
						EndColumn:      2,
						MethodCount:    1,
						IsPublic:       true,
						HasDocComments: true,
					},
				},
			},
			expectedCount: 2, // Two naming violations (snake_case and non-descriptive)
			expectedTypes: []models.ViolationType{
				models.ViolationTypeRustInvalidTraitNaming,
			},
		},
		{
			name: "complex trait bounds in implementation",
			astInfo: &types.RustASTInfo{
				FilePath: "test.rs",
				Impls: []*types.RustImplInfo{
					{
						StartLine:   20,
						EndLine:     50,
						StartColumn: 1,
						EndColumn:   2,
						TargetType:  "MyStruct<T, U>",
						TraitName:   "ComplexTrait<Clone + Send + Sync + Display>",
						MethodCount: 3,
					},
				},
			},
			expectedCount: 1, // Complex trait bounds
			expectedTypes: []models.ViolationType{
				models.ViolationTypeRustTraitBoundComplexity,
			},
		},
		{
			name: "functions with complex trait bounds",
			astInfo: &types.RustASTInfo{
				FilePath: "test.rs",
				Functions: []*types.RustFunctionInfo{
					{
						Name:        "complex_function",
						StartLine:   30,
						EndLine:     35,
						StartColumn: 1,
						EndColumn:   2,
						ReturnType:  "impl Iterator<Item = Result<T, E>> + Send + Sync",
						Parameters: []types.RustParameterInfo{
							{
								Name: "param1",
								Type: "impl Clone + Send + Sync",
							},
							{
								Name: "param2", 
								Type: "impl Iterator<Item = T> + ExactSizeIterator + Send",
							},
						},
					},
				},
			},
			expectedCount: 2, // Return type complexity + too many complex params
			expectedTypes: []models.ViolationType{
				models.ViolationTypeRustTraitBoundComplexity,
			},
		},
		{
			name: "large implementation",
			astInfo: &types.RustASTInfo{
				FilePath: "test.rs",
				Impls: []*types.RustImplInfo{
					{
						StartLine:   40,
						EndLine:     200,
						StartColumn: 1,
						EndColumn:   2,
						TargetType:  "MyLargeStruct",
						TraitName:   "SomeTrait",
						MethodCount: 20, // Too many methods
					},
				},
			},
			expectedCount: 1, // Large implementation
			expectedTypes: []models.ViolationType{
				models.ViolationTypeRustOverlyComplexTrait,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := detector.Detect(tt.astInfo)
			
			if len(violations) != tt.expectedCount {
				t.Errorf("Expected %d violations, got %d", tt.expectedCount, len(violations))
				for i, v := range violations {
					t.Logf("Violation %d: %s - %s", i, v.Type, v.Message)
				}
				return
			}

			// Verify violation types
			for _, expectedType := range tt.expectedTypes {
				found := false
				for _, violation := range violations {
					if violation.Type == expectedType {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected violation type %s not found", expectedType)
				}
			}
		})
	}
}

func TestRustTraitDetector_ComplexityCalculation(t *testing.T) {
	detector := NewRustTraitDetector(nil)

	tests := []struct {
		name     string
		trait    *types.RustTraitInfo
		expected int
	}{
		{
			name: "simple trait",
			trait: &types.RustTraitInfo{
				Name:        "SimpleTrait",
				StartLine:   1,
				EndLine:     10,
				MethodCount: 2,
			},
			expected: 6, // 2 methods * 2 + estimated associated types (1*2)
		},
		{
			name: "complex trait with many methods",
			trait: &types.RustTraitInfo{
				Name:        "ComplexTrait",
				StartLine:   1,
				EndLine:     100,
				MethodCount: 10,
			},
			expected: 44, // 10*2 + (100-1+1-20)/5 + estimated associated types = 20 + 16 + 8 = 44
		},
		{
			name: "trait with multiple responsibility indicators",
			trait: &types.RustTraitInfo{
				Name:        "DataProcessorAndManager",
				StartLine:   1,
				EndLine:     30,
				MethodCount: 5,
			},
			expected: 21, // 5*2 + 5 (multiple responsibilities) + 6 (estimated associated types) = 21
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			complexity := detector.calculateTraitComplexity(tt.trait)
			if complexity != tt.expected {
				t.Errorf("Expected complexity %d, got %d", tt.expected, complexity)
			}
		})
	}
}

func TestRustTraitDetector_NamingValidation(t *testing.T) {
	detector := NewRustTraitDetector(nil)

	tests := []struct {
		name     string
		traitName string
		valid    bool
	}{
		{"valid PascalCase", "MyTrait", true},
		{"valid single word", "Display", true},
		{"valid with numbers", "Trait2D", true},
		{"invalid snake_case", "my_trait", false},
		{"invalid kebab-case", "my-trait", false},
		{"invalid starting with lowercase", "myTrait", false},
		{"invalid with spaces", "My Trait", false},
		{"invalid with underscores", "My_Trait", false},
		{"empty name", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.isValidRustTraitName(tt.traitName)
			if result != tt.valid {
				t.Errorf("Expected %s to be valid=%v, got %v", tt.traitName, tt.valid, result)
			}
		})
	}
}

func TestRustTraitDetector_NonDescriptiveNames(t *testing.T) {
	detector := NewRustTraitDetector(nil)

	tests := []struct {
		name      string
		traitName string
		nonDesc   bool
	}{
		{"descriptive name", "Serializable", false},
		{"acceptable short name", "Eq", false},
		{"non-descriptive generic", "T", true},
		{"non-descriptive trait", "Trait", true},
		{"non-descriptive manager", "Manager", true},
		{"non-descriptive helper", "Helper", true},
		{"too short", "X", true},
		{"acceptable UI", "UI", false},
		{"acceptable API", "API", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.isNonDescriptiveName(tt.traitName)
			if result != tt.nonDesc {
				t.Errorf("Expected %s to be non-descriptive=%v, got %v", tt.traitName, tt.nonDesc, result)
			}
		})
	}
}

func TestRustTraitDetector_TraitBoundComplexity(t *testing.T) {
	detector := NewRustTraitDetector(nil)

	tests := []struct {
		name        string
		typeStr     string
		complex     bool
	}{
		{"simple type", "String", false},
		{"simple generic", "Vec<T>", false},
		{"complex bounds", "impl Iterator<Item = T> + Send + Sync", true},
		{"very complex", "where T: Clone + Send + Sync + Display", true},
		{"long type name", "VeryLongTypeNameThatIndicatesComplexityInTheTypeSystemAndRequiresManyCharactersToReachTheThreshold<T>", true},
		{"dyn trait", "dyn Display + Send + Sync", true},
		{"empty type", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.hasComplexTraitBoundsInType(tt.typeStr)
			if result != tt.complex {
				t.Errorf("Expected %s to be complex=%v, got %v", tt.typeStr, tt.complex, result)
			}
		})
	}
}

func TestRustTraitDetector_MultipleResponsibilities(t *testing.T) {
	detector := NewRustTraitDetector(nil)

	tests := []struct {
		name      string
		traitName string
		multiple  bool
	}{
		{"single responsibility", "Serializable", false},
		{"clear single purpose", "Iterator", false},
		{"multiple with And", "ReadAndWrite", true},
		{"multiple with Manager", "DataManager", true},
		{"multiple with Processor", "MessageProcessor", true},
		{"multiple capital letters", "HTTPClientManagerAndProcessor", true},
		{"acceptable multiple caps", "XMLParser", true}, // XMLParser has 4 capital letters, which exceeds the threshold of 3
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.hasMultipleResponsibilities(tt.traitName)
			if result != tt.multiple {
				t.Errorf("Expected %s to have multiple responsibilities=%v, got %v", tt.traitName, tt.multiple, result)
			}
		})
	}
}

func TestRustTraitDetector_ConfigIntegration(t *testing.T) {
	// Test with custom configuration
	config := &DetectorConfig{
		RustConfig: &RustDetectorConfig{
			MaxTraitComplexity:    5,  // Very low threshold
			MaxTraitLines:        10, // Very low threshold
			MaxTraitMethods:      2,  // Very low threshold
			MaxAssociatedTypes:   1,  // Very low threshold
			MaxComplexTraitParams: 0, // No complex params allowed
		},
	}
	
	detector := NewRustTraitDetector(config)

	astInfo := &types.RustASTInfo{
		FilePath: "test.rs",
		Traits: []*types.RustTraitInfo{
			{
				Name:        "ModeratelyComplexTrait",
				StartLine:   1,
				EndLine:     20, // Above threshold
				MethodCount: 4,  // Above threshold
				IsPublic:    true,
			},
		},
	}

	violations := detector.Detect(astInfo)
	
	// Should detect multiple violations due to low thresholds
	if len(violations) < 2 {
		t.Errorf("Expected at least 2 violations with strict config, got %d", len(violations))
	}
	
	// Verify violations are of expected types
	complexityFound := false
	for _, violation := range violations {
		if violation.Type == models.ViolationTypeRustOverlyComplexTrait {
			complexityFound = true
		}
	}
	
	if !complexityFound {
		t.Error("Expected complexity violation not found")
	}
}

func TestRustTraitDetector_EmptyInput(t *testing.T) {
	detector := NewRustTraitDetector(nil)

	// Test with nil AST info
	violations := detector.Detect(nil)
	if len(violations) != 0 {
		t.Errorf("Expected 0 violations for nil input, got %d", len(violations))
	}

	// Test with empty AST info
	emptyAST := &types.RustASTInfo{
		FilePath: "empty.rs",
		Traits:   []*types.RustTraitInfo{},
		Impls:    []*types.RustImplInfo{},
		Functions: []*types.RustFunctionInfo{},
	}
	
	violations = detector.Detect(emptyAST)
	if len(violations) != 0 {
		t.Errorf("Expected 0 violations for empty AST, got %d", len(violations))
	}
}

func TestRustTraitDetector_DetectorInterface(t *testing.T) {
	detector := NewRustTraitDetector(nil)

	// Test detector implements required interface methods
	name := detector.Name()
	if name != "RustTraitDetector" {
		t.Errorf("Expected detector name 'RustTraitDetector', got '%s'", name)
	}

	description := detector.Description()
	if description == "" {
		t.Error("Expected non-empty description")
	}
	
	if len(description) < 20 {
		t.Errorf("Expected meaningful description, got '%s'", description)
	}
}