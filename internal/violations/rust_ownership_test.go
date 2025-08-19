package violations

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

func TestRustOwnershipDetector_Detect(t *testing.T) {
	detector := NewRustOwnershipDetector(nil)

	tests := []struct {
		name         string
		code         string
		expectedViolations int
		expectedTypes []models.ViolationType
	}{
		{
			name: "unnecessary clone in simple assignment",
			code: `fn main() {
    let x = String::from("hello");
    let y = x.clone();
    println!("{}", y);
}`,
			expectedViolations: 1,
			expectedTypes: []models.ViolationType{models.ViolationTypeRustUnnecessaryClone},
		},
		{
			name: "unnecessary clone in function call",
			code: `fn process_string(s: &str) {
    println!("{}", s);
}

fn main() {
    let text = String::from("hello");
    process_string(text.clone());
}`,
			expectedViolations: 1,
			expectedTypes: []models.ViolationType{models.ViolationTypeRustUnnecessaryClone},
		},
		{
			name: "multiple dereferences",
			code: `fn main() {
    let x = 42;
    let y = &x;
    let z = &y;
    println!("{}", **z);
}`,
			expectedViolations: 1,
			expectedTypes: []models.ViolationType{models.ViolationTypeRustInefficientBorrowing},
		},
		{
			name: "reference-dereference chain",
			code: `fn main() {
    let x = 42;
    let y = &*x;
    println!("{}", y);
}`,
			expectedViolations: 1,
			expectedTypes: []models.ViolationType{models.ViolationTypeRustInefficientBorrowing},
		},
		{
			name: "complex lifetime annotations",
			code: `fn complex_function<'a, 'b, 'c, 'd>(x: &'a str, y: &'b str, z: &'c str, w: &'d str) -> &'a str {
    x
}`,
			expectedViolations: 2, // One for too many lifetimes, one for complex pattern
			expectedTypes: []models.ViolationType{models.ViolationTypeRustComplexLifetime},
		},
		{
			name: "unnecessary move in simple closure",
			code: `fn main() {
    let x = 42;
    let closure = move || { x };
    println!("{}", closure());
}`,
			expectedViolations: 1,
			expectedTypes: []models.ViolationType{models.ViolationTypeRustMoveSemanticsViolation},
		},
		{
			name: "borrow checker bypass with unsafe",
			code: `fn main() {
    let mut x = 42;
    let y = &x;
    unsafe {
        let z = &mut x;
        *z = 43;
    }
    println!("{}", y);
}`,
			expectedViolations: 1,
			expectedTypes: []models.ViolationType{models.ViolationTypeRustBorrowCheckerBypass},
		},
		{
			name: "transmute abuse",
			code: `use std::mem;

fn main() {
    let x: u32 = 42;
    let y: f32 = unsafe { mem::transmute(x) };
    println!("{}", y);
}`,
			expectedViolations: 2, // One for unsafe block, one for transmute
			expectedTypes: []models.ViolationType{models.ViolationTypeRustBorrowCheckerBypass},
		},
		{
			name: "no violations in good code",
			code: `fn main() {
    let x = String::from("hello");
    let y = &x;
    println!("{}", y);
    
    let numbers = vec![1, 2, 3];
    for num in &numbers {
        println!("{}", num);
    }
}`,
			expectedViolations: 0,
			expectedTypes: []models.ViolationType{},
		},
		{
			name: "multiple violations",
			code: `fn main() {
    let x = String::from("hello");
    let y = x.clone(); // unnecessary clone
    let z = &*y;       // inefficient borrowing
    println!("{}", z);
}`,
			expectedViolations: 2,
			expectedTypes: []models.ViolationType{
				models.ViolationTypeRustUnnecessaryClone,
				models.ViolationTypeRustInefficientBorrowing,
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

func TestRustOwnershipDetector_Name(t *testing.T) {
	detector := NewRustOwnershipDetector(nil)
	expected := "Rust Ownership Analysis"
	if detector.Name() != expected {
		t.Errorf("Expected name %s, got %s", expected, detector.Name())
	}
}

func TestRustOwnershipDetector_Description(t *testing.T) {
	detector := NewRustOwnershipDetector(nil)
	description := detector.Description()
	if description == "" {
		t.Error("Description should not be empty")
	}
	
	// Check that description mentions key concepts
	expectedKeywords := []string{"ownership", "borrowing", "clone", "lifetime"}
	for _, keyword := range expectedKeywords {
		if !containsIgnoreCase(description, keyword) {
			t.Errorf("Description should mention '%s'", keyword)
		}
	}
}

func TestRustOwnershipDetector_DetectUnnecessaryClones(t *testing.T) {
	detector := NewRustOwnershipDetector(nil)

	testCases := []struct {
		name     string
		code     string
		expected int
	}{
		{
			name:     "clone in iterator",
			code:     `vec.iter().map(|x| x.clone()).collect()`,
			expected: 1,
		},
		{
			name:     "clone for return",
			code:     `return value.clone();`,
			expected: 1,
		},
		{
			name:     "necessary clone for thread",
			code:     `thread::spawn(move || { value.clone() })`,
			expected: 0,
		},
		{
			name:     "clone with immediate reference",
			code:     `let x = value.clone()&`,
			expected: 1,
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
			
			cloneViolations := 0
			for _, v := range violations {
				if v.Type == models.ViolationTypeRustUnnecessaryClone {
					cloneViolations++
				}
			}

			if cloneViolations != tc.expected {
				t.Errorf("Expected %d clone violations, got %d", tc.expected, cloneViolations)
			}
		})
	}
}

func TestRustOwnershipDetector_DetectComplexLifetimes(t *testing.T) {
	detector := NewRustOwnershipDetector(nil)

	complexLifetimeCode := `
fn very_complex<'a, 'b, 'c, 'd, 'e>(
    x: &'a str, 
    y: &'b str, 
    z: &'c str, 
    w: &'d str, 
    v: &'e str
) -> &'a str {
    x
}`

	filePath := createTempFileWithContent(t, complexLifetimeCode, ".rs")
	defer removeTempFile(t, filePath)

	fileInfo := &models.FileInfo{
		Path:     filePath,
		Language: "rust",
	}

	rustAstInfo := &types.RustASTInfo{
		FilePath: filePath,
	}

	violations := detector.Detect(fileInfo, rustAstInfo)
	
	found := false
	for _, v := range violations {
		if v.Type == models.ViolationTypeRustComplexLifetime {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find complex lifetime violation")
	}
}

func TestRustOwnershipDetector_DetectBorrowCheckerBypass(t *testing.T) {
	detector := NewRustOwnershipDetector(nil)

	unsafeCode := `
unsafe {
    let x = &mut data;
    let y = &data;
    *x = 42;
}`

	filePath := createTempFileWithContent(t, unsafeCode, ".rs")
	defer removeTempFile(t, filePath)

	fileInfo := &models.FileInfo{
		Path:     filePath,
		Language: "rust",
	}

	rustAstInfo := &types.RustASTInfo{
		FilePath: filePath,
	}

	violations := detector.Detect(fileInfo, rustAstInfo)
	
	found := false
	for _, v := range violations {
		if v.Type == models.ViolationTypeRustBorrowCheckerBypass {
			found = true
			if v.Severity != models.SeverityMedium && v.Severity != models.SeverityHigh {
				t.Errorf("Expected high or medium severity for borrow checker bypass, got %s", v.Severity)
			}
			break
		}
	}

	if !found {
		t.Error("Expected to find borrow checker bypass violation")
	}
}

func TestRustOwnershipDetector_NilAstInfo(t *testing.T) {
	detector := NewRustOwnershipDetector(nil)

	fileInfo := &models.FileInfo{
		Path:     "test.rs",
		Language: "rust",
	}

	violations := detector.Detect(fileInfo, nil)
	if len(violations) != 0 {
		t.Errorf("Expected 0 violations for nil AST info, got %d", len(violations))
	}
}

func TestRustOwnershipDetector_InvalidAstInfo(t *testing.T) {
	detector := NewRustOwnershipDetector(nil)

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

// Helper functions for testing

func createTempFileWithContent(t *testing.T, content, extension string) string {
	tmpFile, err := ioutil.TempFile("", "*"+extension)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	return tmpFile.Name()
}

func removeTempFile(t *testing.T, filePath string) {
	if err := os.Remove(filePath); err != nil {
		t.Errorf("Failed to remove temp file %s: %v", filePath, err)
	}
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}