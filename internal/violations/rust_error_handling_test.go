package violations

import (
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

func TestRustErrorHandlingDetector_Detect(t *testing.T) {
	detector := NewRustErrorHandlingDetector(nil)

	tests := []struct {
		name         string
		code         string
		expectedViolations int
		expectedTypes []models.ViolationType
	}{
		{
			name: "overuse of unwrap",
			code: `fn main() {
    let result = Some(42);
    let value = result.unwrap();
    println!("{}", value);
}`,
			expectedViolations: 1,
			expectedTypes: []models.ViolationType{models.ViolationTypeRustOveruseUnwrap},
		},
		{
			name: "missing error propagation",
			code: `fn process_file() -> Result<String, std::io::Error> {
    match std::fs::read_to_string("file.txt") {
        Ok(content) => Ok(content),
        Err(e) => return Err(e),
    }
}`,
			expectedViolations: 1,
			expectedTypes: []models.ViolationType{models.ViolationTypeRustMissingErrorPropagation},
		},
		{
			name: "panic prone code - array indexing",
			code: `fn main() {
    let arr = [1, 2, 3];
    let value = arr[10];
    println!("{}", value);
}`,
			expectedViolations: 1,
			expectedTypes: []models.ViolationType{models.ViolationTypeRustPanicProneCode},
		},
		{
			name: "panic prone code - division",
			code: `fn divide(a: i32, b: i32) -> i32 {
    a / b
}`,
			expectedViolations: 1,
			expectedTypes: []models.ViolationType{models.ViolationTypeRustPanicProneCode},
		},
		{
			name: "improper expect usage",
			code: `fn main() {
    let result = Some(42);
    let value = result.expect("error");
    println!("{}", value);
}`,
			expectedViolations: 1,
			expectedTypes: []models.ViolationType{models.ViolationTypeRustImproperExpect},
		},
		{
			name: "unhandled result",
			code: `fn main() {
    std::fs::read_to_string("file.txt");
}`,
			expectedViolations: 1,
			expectedTypes: []models.ViolationType{models.ViolationTypeRustUnhandledResult},
		},
		{
			name: "good error handling",
			code: `fn process_file() -> Result<String, Box<dyn std::error::Error>> {
    let content = std::fs::read_to_string("file.txt")?;
    Ok(content.trim().to_string())
}`,
			expectedViolations: 0,
			expectedTypes: []models.ViolationType{},
		},
		{
			name: "multiple error handling violations",
			code: `fn bad_function() -> Result<i32, Box<dyn std::error::Error>> {
    let result = Some(42);
    let value = result.unwrap(); // unnecessary unwrap
    let arr = [1, 2, 3];
    let item = arr[value]; // panic-prone indexing
    match std::fs::read_to_string("file.txt") {
        Ok(content) => Ok(content.len() as i32),
        Err(e) => return Err(Box::new(e)), // could use ?
    }
}`,
			expectedViolations: 3,
			expectedTypes: []models.ViolationType{
				models.ViolationTypeRustOveruseUnwrap,
				models.ViolationTypeRustPanicProneCode,
				models.ViolationTypeRustMissingErrorPropagation,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file with the test code
			filePath := createTempFileWithContent(t, tt.code, ".rs")
			defer removeTempFile(t, filePath)

			fileInfo := &models.FileInfo{
				Path:     filePath,
				Language: "rust",
			}

			rustAstInfo := &types.RustASTInfo{
				FilePath: filePath,
			}

			violations := detector.Detect(fileInfo, rustAstInfo)

			if len(violations) != tt.expectedViolations {
				t.Errorf("Expected %d violations, got %d", tt.expectedViolations, len(violations))
				for i, v := range violations {
					t.Errorf("Violation %d: Type=%s, Message=%s", i+1, v.Type, v.Message)
				}
			}

			// Check violation types
			if len(violations) > 0 && len(tt.expectedTypes) > 0 {
				foundTypes := make(map[models.ViolationType]bool)
				for _, v := range violations {
					foundTypes[v.Type] = true
				}

				for _, expectedType := range tt.expectedTypes {
					if !foundTypes[expectedType] {
						t.Errorf("Expected violation type %s not found", expectedType)
					}
				}
			}
		})
	}
}

func TestRustErrorHandlingDetector_Name(t *testing.T) {
	detector := NewRustErrorHandlingDetector(nil)
	expected := "Rust Error Handling Analysis"
	if detector.Name() != expected {
		t.Errorf("Expected name %s, got %s", expected, detector.Name())
	}
}

func TestRustErrorHandlingDetector_Description(t *testing.T) {
	detector := NewRustErrorHandlingDetector(nil)
	description := detector.Description()
	if description == "" {
		t.Error("Description should not be empty")
	}
	
	// Check that description mentions key concepts
	expectedKeywords := []string{"error", "handling", "unwrap", "expect"}
	for _, keyword := range expectedKeywords {
		if !containsIgnoreCase(description, keyword) {
			t.Errorf("Description should mention '%s'", keyword)
		}
	}
}

func TestRustErrorHandlingDetector_DetectUnwrapOveruse(t *testing.T) {
	detector := NewRustErrorHandlingDetector(nil)

	testCases := []struct {
		name     string
		code     string
		expected int
	}{
		{
			name:     "unwrap in loop",
			code:     `for item in vec { item.unwrap(); }`,
			expected: 1,
		},
		{
			name:     "unwrap in main function",
			code:     `fn main() { result.unwrap(); }`,
			expected: 0, // main function unwraps are usually okay
		},
		{
			name:     "unwrap_or usage",
			code:     `let value = result.unwrap_or(42);`,
			expected: 0, // unwrap_or is safer
		},
		{
			name:     "multiple unwraps",
			code:     `result1.unwrap(); result2.unwrap(); result3.unwrap(); result4.unwrap(); result5.unwrap();`,
			expected: 1, // Overall count violation (individual unwraps are not flagged as problematic by default)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filePath := createTempFileWithContent(t, tc.code, ".rs")
			defer removeTempFile(t, filePath)

			fileInfo := &models.FileInfo{
				Path:     filePath,
				Language: "rust",
			}

			rustAstInfo := &types.RustASTInfo{
				FilePath: filePath,
			}

			violations := detector.Detect(fileInfo, rustAstInfo)
			
			unwrapViolations := 0
			for _, v := range violations {
				if v.Type == models.ViolationTypeRustOveruseUnwrap {
					unwrapViolations++
				}
			}

			if unwrapViolations != tc.expected {
				t.Errorf("Expected %d unwrap violations, got %d", tc.expected, unwrapViolations)
			}
		})
	}
}

func TestRustErrorHandlingDetector_DetectPanicProneCode(t *testing.T) {
	detector := NewRustErrorHandlingDetector(nil)

	panicProneCode := `
fn dangerous_function() {
    let arr = [1, 2, 3];
    let idx = arr[10]; // out of bounds
    let result = 42 / 0; // division by zero
    panic!("This will panic"); // explicit panic
}`

	filePath := createTempFileWithContent(t, panicProneCode, ".rs")
	defer removeTempFile(t, filePath)

	fileInfo := &models.FileInfo{
		Path:     filePath,
		Language: "rust",
	}

	rustAstInfo := &types.RustASTInfo{
		FilePath: filePath,
	}

	violations := detector.Detect(fileInfo, rustAstInfo)
	
	panicViolations := 0
	for _, v := range violations {
		if v.Type == models.ViolationTypeRustPanicProneCode {
			panicViolations++
		}
	}

	if panicViolations < 2 { // Should find at least array indexing and panic!
		t.Errorf("Expected at least 2 panic-prone violations, got %d", panicViolations)
	}
}

func TestRustErrorHandlingDetector_DetectImproperExpect(t *testing.T) {
	detector := NewRustErrorHandlingDetector(nil)

	improperExpectCode := `
fn test_expects() {
    result1.expect("error"); // too short
    result2.expect("failed"); // non-descriptive
    result3.expect("This is a proper descriptive error message");
}`

	filePath := createTempFileWithContent(t, improperExpectCode, ".rs")
	defer removeTempFile(t, filePath)

	fileInfo := &models.FileInfo{
		Path:     filePath,
		Language: "rust",
	}

	rustAstInfo := &types.RustASTInfo{
		FilePath: filePath,
	}

	violations := detector.Detect(fileInfo, rustAstInfo)
	
	expectViolations := 0
	for _, v := range violations {
		if v.Type == models.ViolationTypeRustImproperExpect {
			expectViolations++
		}
	}

	if expectViolations != 2 { // Should find 2 improper expects
		t.Errorf("Expected 2 improper expect violations, got %d", expectViolations)
	}
}

func TestRustErrorHandlingDetector_NilAstInfo(t *testing.T) {
	detector := NewRustErrorHandlingDetector(nil)

	fileInfo := &models.FileInfo{
		Path:     "test.rs",
		Language: "rust",
	}

	violations := detector.Detect(fileInfo, nil)
	if len(violations) != 0 {
		t.Errorf("Expected 0 violations for nil AST info, got %d", len(violations))
	}
}

func TestRustErrorHandlingDetector_InvalidAstInfo(t *testing.T) {
	detector := NewRustErrorHandlingDetector(nil)

	fileInfo := &models.FileInfo{
		Path:     "test.rs",
		Language: "rust",
	}

	// Pass invalid AST info (string instead of RustASTInfo)
	violations := detector.Detect(fileInfo, "invalid")
	if len(violations) != 0 {
		t.Errorf("Expected 0 violations for invalid AST info, got %d", len(violations))
	}
}

// Helper functions are defined in rust_ownership_test.go
// func createTempFileWithContent(t *testing.T, content, extension string) string
// func removeTempFile(t *testing.T, filePath string)
// func containsIgnoreCase(s, substr string) bool