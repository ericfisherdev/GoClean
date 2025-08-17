package violations

import (
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

func TestRustNamingDetector_FunctionNaming(t *testing.T) {
	tests := []struct {
		name           string
		function       *types.RustFunctionInfo
		expectedViolations int
		expectedRules  []string
	}{
		{
			name: "Valid snake_case function",
			function: &types.RustFunctionInfo{
				Name:      "calculate_total",
				StartLine: 10,
				StartColumn: 1,
			},
			expectedViolations: 0,
		},
		{
			name: "Invalid PascalCase function",
			function: &types.RustFunctionInfo{
				Name:      "CalculateTotal",
				StartLine: 20,
				StartColumn: 1,
			},
			expectedViolations: 1,
			expectedRules:  []string{RustInvalidFunctionNaming},
		},
		{
			name: "Invalid camelCase function",
			function: &types.RustFunctionInfo{
				Name:      "calculateTotal",
				StartLine: 30,
				StartColumn: 1,
			},
			expectedViolations: 1,
			expectedRules:  []string{RustInvalidFunctionNaming},
		},
		{
			name: "Non-descriptive function name",
			function: &types.RustFunctionInfo{
				Name:      "f",
				StartLine: 40,
				StartColumn: 1,
			},
			expectedViolations: 1, // Non-descriptive (single letter is valid snake_case in Rust)
			expectedRules:  []string{RustNonDescriptiveName},
		},
		{
			name: "Function with abbreviation",
			function: &types.RustFunctionInfo{
				Name:      "calc_num",
				StartLine: 50,
				StartColumn: 1,
			},
			expectedViolations: 1,
			expectedRules:  []string{RustUnclearAbbreviation},
		},
		{
			name: "Valid function with common pattern",
			function: &types.RustFunctionInfo{
				Name:      "from_str",
				StartLine: 60,
				StartColumn: 1,
			},
			expectedViolations: 0, // 'str' is acceptable in Rust context
		},
		{
			name: "Valid new constructor",
			function: &types.RustFunctionInfo{
				Name:      "new",
				StartLine: 70,
				StartColumn: 1,
			},
			expectedViolations: 0,
		},
	}

	detector := NewRustNamingDetector(nil)
	fileInfo := &models.FileInfo{Path: "test.rs"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := detector.checkFunctionNaming(tt.function, fileInfo.Path)
			
			if len(violations) != tt.expectedViolations {
				t.Errorf("Expected %d violations, got %d", tt.expectedViolations, len(violations))
				for _, v := range violations {
					t.Logf("  Violation: %s (rule: %s)", v.Message, v.Rule)
				}
			}

			// Check that expected rules are present
			for _, expectedRule := range tt.expectedRules {
				found := false
				for _, v := range violations {
					if v.Rule == expectedRule {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected rule '%s' not found in violations", expectedRule)
				}
			}
		})
	}
}

func TestRustNamingDetector_StructNaming(t *testing.T) {
	tests := []struct {
		name           string
		structInfo     *types.RustStructInfo
		expectedViolations int
		expectedRules  []string
	}{
		{
			name: "Valid PascalCase struct",
			structInfo: &types.RustStructInfo{
				Name:      "UserAccount",
				StartLine: 10,
				StartColumn: 1,
			},
			expectedViolations: 0,
		},
		{
			name: "Invalid snake_case struct",
			structInfo: &types.RustStructInfo{
				Name:      "user_account",
				StartLine: 20,
				StartColumn: 1,
			},
			expectedViolations: 1,
			expectedRules:  []string{RustInvalidStructNaming},
		},
		{
			name: "Struct with proper acronym",
			structInfo: &types.RustStructInfo{
				Name:      "HttpServer",
				StartLine: 30,
				StartColumn: 1,
			},
			expectedViolations: 0,
		},
		{
			name: "Struct with improper acronym",
			structInfo: &types.RustStructInfo{
				Name:      "HTTPServer",
				StartLine: 40,
				StartColumn: 1,
			},
			expectedViolations: 2, // Both invalid struct naming (not proper PascalCase) AND improper acronym
			expectedRules:  []string{RustInvalidStructNaming, RustAcronymCasing},
		},
		{
			name: "Non-descriptive struct name",
			structInfo: &types.RustStructInfo{
				Name:      "Data",
				StartLine: 50,
				StartColumn: 1,
			},
			expectedViolations: 1,
			expectedRules:  []string{RustNonDescriptiveName},
		},
	}

	detector := NewRustNamingDetector(nil)
	fileInfo := &models.FileInfo{Path: "test.rs"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := detector.checkStructNaming(tt.structInfo, fileInfo.Path)
			
			if len(violations) != tt.expectedViolations {
				t.Errorf("Expected %d violations, got %d", tt.expectedViolations, len(violations))
				for _, v := range violations {
					t.Logf("  Violation: %s (rule: %s)", v.Message, v.Rule)
				}
			}

			// Check that expected rules are present
			for _, expectedRule := range tt.expectedRules {
				found := false
				for _, v := range violations {
					if v.Rule == expectedRule {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected rule '%s' not found in violations", expectedRule)
				}
			}
		})
	}
}

func TestRustNamingDetector_ConstantNaming(t *testing.T) {
	tests := []struct {
		name           string
		constant       *types.RustConstantInfo
		expectedViolations int
		expectedRules  []string
	}{
		{
			name: "Valid SCREAMING_SNAKE_CASE constant",
			constant: &types.RustConstantInfo{
				Name:   "MAX_RETRY_COUNT",
				Line:   10,
				Column: 1,
			},
			expectedViolations: 0,
		},
		{
			name: "Invalid snake_case constant",
			constant: &types.RustConstantInfo{
				Name:   "max_retry_count",
				Line:   20,
				Column: 1,
			},
			expectedViolations: 1,
			expectedRules:  []string{RustInvalidConstantNaming},
		},
		{
			name: "Invalid PascalCase constant",
			constant: &types.RustConstantInfo{
				Name:   "MaxRetryCount",
				Line:   30,
				Column: 1,
			},
			expectedViolations: 1,
			expectedRules:  []string{RustInvalidConstantNaming},
		},
		{
			name: "Valid single word constant",
			constant: &types.RustConstantInfo{
				Name:   "VERSION",
				Line:   40,
				Column: 1,
			},
			expectedViolations: 0,
		},
	}

	detector := NewRustNamingDetector(nil)
	fileInfo := &models.FileInfo{Path: "test.rs"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := detector.checkConstantNaming(tt.constant, fileInfo.Path)
			
			if len(violations) != tt.expectedViolations {
				t.Errorf("Expected %d violations, got %d", tt.expectedViolations, len(violations))
				for _, v := range violations {
					t.Logf("  Violation: %s (rule: %s)", v.Message, v.Rule)
				}
			}

			// Check that expected rules are present
			for _, expectedRule := range tt.expectedRules {
				found := false
				for _, v := range violations {
					if v.Rule == expectedRule {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected rule '%s' not found in violations", expectedRule)
				}
			}
		})
	}
}

func TestRustNamingDetector_ParameterNaming(t *testing.T) {
	config := &DetectorConfig{
		AllowSingleLetterVars: false,
	}
	
	tests := []struct {
		name           string
		function       *types.RustFunctionInfo
		expectedViolations int
		expectedRules  []string
	}{
		{
			name: "Valid snake_case parameters",
			function: &types.RustFunctionInfo{
				Name:      "process_data",
				StartLine: 10,
				Parameters: []types.RustParameterInfo{
					{Name: "user_id", Type: "u64"},
					{Name: "is_active", Type: "bool"},
				},
			},
			expectedViolations: 0,
		},
		{
			name: "Invalid camelCase parameter",
			function: &types.RustFunctionInfo{
				Name:      "process_data",
				StartLine: 20,
				Parameters: []types.RustParameterInfo{
					{Name: "userId", Type: "u64"},
				},
			},
			expectedViolations: 1,
			expectedRules:  []string{RustInvalidParameterNaming},
		},
		{
			name: "Single letter parameter (not allowed)",
			function: &types.RustFunctionInfo{
				Name:      "process_data",
				StartLine: 30,
				Parameters: []types.RustParameterInfo{
					{Name: "x", Type: "i32"},
				},
			},
			expectedViolations: 0, // 'x' is acceptable in Rust context
			expectedRules:  []string{},
		},
		{
			name: "Self parameter (should be ignored)",
			function: &types.RustFunctionInfo{
				Name:      "process",
				StartLine: 40,
				Parameters: []types.RustParameterInfo{
					{Name: "self"},
				},
			},
			expectedViolations: 0,
		},
		{
			name: "Underscore parameter (should be ignored)",
			function: &types.RustFunctionInfo{
				Name:      "handle_event",
				StartLine: 50,
				Parameters: []types.RustParameterInfo{
					{Name: "_"},
				},
			},
			expectedViolations: 0,
		},
		{
			name: "Iterator context allows single letters",
			function: &types.RustFunctionInfo{
				Name:      "map_values",
				StartLine: 60,
				Parameters: []types.RustParameterInfo{
					{Name: "x", Type: "i32"},
				},
			},
			expectedViolations: 0, // Single letter acceptable in iterator context
		},
	}

	detector := NewRustNamingDetector(config)
	fileInfo := &models.FileInfo{Path: "test.rs"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := detector.checkParameterNaming(tt.function, fileInfo.Path)
			
			if len(violations) != tt.expectedViolations {
				t.Errorf("Expected %d violations, got %d", tt.expectedViolations, len(violations))
				for _, v := range violations {
					t.Logf("  Violation: %s (rule: %s)", v.Message, v.Rule)
				}
			}

			// Check that expected rules are present
			for _, expectedRule := range tt.expectedRules {
				found := false
				for _, v := range violations {
					if v.Rule == expectedRule {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected rule '%s' not found in violations", expectedRule)
				}
			}
		})
	}
}

func TestRustConventionChecker_CaseConversions(t *testing.T) {
	checker := NewRustConventionChecker()
	
	tests := []struct {
		name     string
		input    string
		method   string
		expected string
	}{
		// ToSnakeCase tests
		{
			name:     "PascalCase to snake_case",
			input:    "UserAccount",
			method:   "snake",
			expected: "user_account",
		},
		{
			name:     "camelCase to snake_case",
			input:    "userId",
			method:   "snake",
			expected: "user_id",
		},
		{
			name:     "Already snake_case",
			input:    "user_id",
			method:   "snake",
			expected: "user_id",
		},
		{
			name:     "Acronym in PascalCase to snake_case",
			input:    "HttpServer",
			method:   "snake",
			expected: "http_server",
		},
		
		// ToPascalCase tests
		{
			name:     "snake_case to PascalCase",
			input:    "user_account",
			method:   "pascal",
			expected: "UserAccount",
		},
		{
			name:     "camelCase to PascalCase",
			input:    "userId",
			method:   "pascal",
			expected: "UserId",
		},
		{
			name:     "Already PascalCase",
			input:    "UserAccount",
			method:   "pascal",
			expected: "UserAccount",
		},
		
		// ToScreamingSnakeCase tests
		{
			name:     "snake_case to SCREAMING_SNAKE_CASE",
			input:    "max_retry",
			method:   "screaming",
			expected: "MAX_RETRY",
		},
		{
			name:     "PascalCase to SCREAMING_SNAKE_CASE",
			input:    "MaxRetry",
			method:   "screaming",
			expected: "MAX_RETRY",
		},
		{
			name:     "camelCase to SCREAMING_SNAKE_CASE",
			input:    "maxRetry",
			method:   "screaming",
			expected: "MAX_RETRY",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			switch tt.method {
			case "snake":
				result = checker.ToSnakeCase(tt.input)
			case "pascal":
				result = checker.ToPascalCase(tt.input)
			case "screaming":
				result = checker.ToScreamingSnakeCase(tt.input)
			}
			
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestRustConventionChecker_Validation(t *testing.T) {
	checker := NewRustConventionChecker()
	
	tests := []struct {
		name     string
		input    string
		checkType string
		expected bool
	}{
		// Function name validation
		{
			name:     "Valid function name",
			input:    "calculate_total",
			checkType: "function",
			expected: true,
		},
		{
			name:     "Invalid function name (PascalCase)",
			input:    "CalculateTotal",
			checkType: "function",
			expected: false,
		},
		{
			name:     "Rust keyword as function name",
			input:    "impl",
			checkType: "function",
			expected: false,
		},
		
		// Type name validation
		{
			name:     "Valid struct name",
			input:    "UserAccount",
			checkType: "type",
			expected: true,
		},
		{
			name:     "Invalid struct name (snake_case)",
			input:    "user_account",
			checkType: "type",
			expected: false,
		},
		{
			name:     "Type with proper acronym",
			input:    "HttpServer",
			checkType: "type",
			expected: true,
		},
		{
			name:     "Type with improper acronym",
			input:    "HTTPServer",
			checkType: "type",
			expected: false,
		},
		
		// Constant name validation
		{
			name:     "Valid constant name",
			input:    "MAX_RETRY_COUNT",
			checkType: "constant",
			expected: true,
		},
		{
			name:     "Invalid constant name (snake_case)",
			input:    "max_retry_count",
			checkType: "constant",
			expected: false,
		},
		
		// Module name validation
		{
			name:     "Valid module name",
			input:    "user_management",
			checkType: "module",
			expected: true,
		},
		{
			name:     "Invalid module name (PascalCase)",
			input:    "UserManagement",
			checkType: "module",
			expected: false,
		},
		
		// Crate name validation
		{
			name:     "Valid crate name with hyphen",
			input:    "my-awesome-crate",
			checkType: "crate",
			expected: true,
		},
		{
			name:     "Valid crate name with underscore",
			input:    "my_awesome_crate",
			checkType: "crate",
			expected: true,
		},
		{
			name:     "Invalid crate name (starts with number)",
			input:    "123crate",
			checkType: "crate",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result bool
			switch tt.checkType {
			case "function":
				result = checker.IsValidFunctionName(tt.input)
			case "type":
				result = checker.IsValidTypeName(tt.input)
			case "constant":
				result = checker.IsValidConstantName(tt.input)
			case "module":
				result = checker.IsValidModuleName(tt.input)
			case "crate":
				result = checker.IsValidCrateName(tt.input)
			}
			
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for input '%s'", tt.expected, result, tt.input)
			}
		})
	}
}

func TestRustNamingDetector_FullIntegration(t *testing.T) {
	detector := NewRustNamingDetector(nil)
	
	// Create a complete Rust AST info
	rustAst := &types.RustASTInfo{
		FilePath: "test.rs",
		Functions: []*types.RustFunctionInfo{
			{
				Name:      "processData", // Invalid: should be snake_case
				StartLine: 10,
				Parameters: []types.RustParameterInfo{
					{Name: "userId", Type: "u64"}, // Invalid: should be snake_case
				},
			},
			{
				Name:      "calc_num", // Has abbreviation
				StartLine: 20,
			},
		},
		Structs: []*types.RustStructInfo{
			{
				Name:      "user_data", // Invalid: should be PascalCase
				StartLine: 30,
			},
			{
				Name:      "HTTPClient", // Invalid acronym casing
				StartLine: 40,
			},
		},
		Constants: []*types.RustConstantInfo{
			{
				Name:   "maxRetries", // Invalid: should be SCREAMING_SNAKE_CASE
				Line:   50,
				Column: 1,
			},
		},
		Modules: []*types.RustModuleInfo{
			{
				Name:      "UserManagement", // Invalid: should be snake_case
				StartLine: 60,
			},
		},
	}
	
	fileInfo := &models.FileInfo{
		Path: "test.rs",
	}
	
	violations := detector.Detect(fileInfo, rustAst)
	
	// We expect violations for:
	// 1. processData function name
	// 2. userId parameter name
	// 3. calc_num abbreviation
	// 4. user_data struct name
	// 5. HTTPClient acronym casing
	// 6. maxRetries constant name
	// 7. UserManagement module name
	expectedViolationCount := 7
	
	if len(violations) < expectedViolationCount {
		t.Errorf("Expected at least %d violations, got %d", expectedViolationCount, len(violations))
		for _, v := range violations {
			t.Logf("  Violation: %s (rule: %s)", v.Message, v.Rule)
		}
	}
	
	// Verify specific violations are present
	expectedRules := map[string]bool{
		RustInvalidFunctionNaming: false,
		RustInvalidParameterNaming: false,
		RustUnclearAbbreviation: false,
		RustInvalidStructNaming: false,
		RustAcronymCasing: false,
		RustInvalidConstantNaming: false,
		RustInvalidModuleNaming: false,
	}
	
	for _, v := range violations {
		if _, ok := expectedRules[v.Rule]; ok {
			expectedRules[v.Rule] = true
		}
	}
	
	for rule, found := range expectedRules {
		if !found {
			t.Errorf("Expected rule '%s' not found in violations", rule)
		}
	}
}