package violations

import (
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

func TestRustCommentedCodeDetector_NewDetector(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustCommentedCodeDetector(config)
	
	if detector == nil {
		t.Fatal("Expected detector to be created")
	}
	
	if detector.Name() != "Rust Commented Code Detector" {
		t.Errorf("Expected name 'Rust Commented Code Detector', got '%s'", detector.Name())
	}
	
	if detector.Description() == "" {
		t.Error("Expected non-empty description")
	}
}

func TestRustCommentedCodeDetector_Detect_NoASTInfo(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustCommentedCodeDetector(config)
	
	fileInfo := &models.FileInfo{
		Path:     "/test/example.rs",
		Language: "Rust",
	}
	
	violations := detector.Detect(fileInfo, nil)
	if len(violations) != 0 {
		t.Errorf("Expected no violations when astInfo is nil, got %d", len(violations))
	}
}

func TestRustCommentedCodeDetector_Detect_WrongASTType(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustCommentedCodeDetector(config)
	
	fileInfo := &models.FileInfo{
		Path:     "/test/example.rs",
		Language: "Rust",
	}
	
	// Pass Go AST info instead of Rust AST info
	goAstInfo := &types.GoASTInfo{}
	
	violations := detector.Detect(fileInfo, goAstInfo)
	if len(violations) != 0 {
		t.Errorf("Expected no violations with wrong AST type, got %d", len(violations))
	}
}

func TestRustCommentedCodeDetector_Detect_ValidRustAST(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustCommentedCodeDetector(config)
	
	fileInfo := &models.FileInfo{
		Path:     "/test/example.rs",
		Language: "Rust",
	}
	
	rustAstInfo := &types.RustASTInfo{
		FilePath: "/test/example.rs",
		Functions: []*types.RustFunctionInfo{
			{
				Name:      "test_fn",
				StartLine: 1,
				EndLine:   5,
			},
		},
	}
	
	// Since we can't easily mock file reading in the current implementation,
	// this test mainly verifies the basic flow works
	violations := detector.Detect(fileInfo, rustAstInfo)
	
	// We expect no violations since file reading returns empty content
	if len(violations) != 0 {
		t.Errorf("Expected no violations with empty content, got %d", len(violations))
	}
}

func TestRustCommentedCodeDetector_LooksLikeRustCode(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustCommentedCodeDetector(config)
	
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{
			name:     "Rust function",
			text:     "fn test_function() { println!(\"Hello\"); }",
			expected: true,
		},
		{
			name:     "Rust struct",
			text:     "struct MyStruct { field: i32 }",
			expected: true,
		},
		{
			name:     "Rust if statement",
			text:     "if x > 0 { return true; }",
			expected: true,
		},
		{
			name:     "Rust let binding",
			text:     "let mut x = 42;",
			expected: true,
		},
		{
			name:     "Rust match expression",
			text:     "match value { Some(x) => x, None => 0 }",
			expected: true,
		},
		{
			name:     "Rust macro call",
			text:     "println!(\"Debug: {:?}\", value);",
			expected: true,
		},
		{
			name:     "Simple documentation",
			text:     "This is a simple documentation comment.",
			expected: false,
		},
		{
			name:     "Short text",
			text:     "Short",
			expected: false,
		},
		{
			name:     "Empty text",
			text:     "",
			expected: false,
		},
		{
			name:     "Regular sentence",
			text:     "This is just a regular English sentence explaining something.",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.looksLikeRustCode(tt.text)
			if result != tt.expected {
				t.Errorf("looksLikeRustCode(%q) = %v, want %v", tt.text, result, tt.expected)
			}
		})
	}
}

func TestRustCommentedCodeDetector_IsDocumentation(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustCommentedCodeDetector(config)
	
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{
			name:     "TODO marker",
			text:     "TODO: Implement this feature",
			expected: true,
		},
		{
			name:     "FIXME marker",
			text:     "FIXME: This is broken",
			expected: true,
		},
		{
			name:     "Copyright notice",
			text:     "Copyright 2023 Company Name",
			expected: true,
		},
		{
			name:     "Rust doc comment marker",
			text:     "/// This is a doc comment",
			expected: true,
		},
		{
			name:     "Rust inner doc comment",
			text:     "//! This is an inner doc comment",
			expected: true,
		},
		{
			name:     "Example tag",
			text:     "Example: fn test() { }",
			expected: true,
		},
		{
			name:     "Proper sentence",
			text:     "This function calculates the sum of two numbers.",
			expected: true,
		},
		{
			name:     "Code-like text",
			text:     "fn test() { return 42; }",
			expected: false,
		},
		{
			name:     "Variable assignment",
			text:     "let x = 10",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.isDocumentation(tt.text)
			if result != tt.expected {
				t.Errorf("isDocumentation(%q) = %v, want %v", tt.text, result, tt.expected)
			}
		})
	}
}

func TestRustCommentedCodeDetector_CreateViolation(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustCommentedCodeDetector(config)
	
	tests := []struct {
		name        string
		commentText string
		expectViolation bool
	}{
		{
			name:        "Clear Rust code",
			commentText: "fn test() { let x = 42; println!(\"{}\", x); }",
			expectViolation: true,
		},
		{
			name:        "Documentation comment",
			commentText: "This function performs a calculation.",
			expectViolation: false,
		},
		{
			name:        "Single Rust keyword",
			commentText: "fn",
			expectViolation: false,
		},
		{
			name:        "Multiple Rust patterns",
			commentText: "let mut x = vec![1, 2, 3]; for item in x { println!(\"{}\", item); }",
			expectViolation: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violation := detector.createViolation(tt.commentText, 10, "/test/file.rs")
			
			if tt.expectViolation && violation == nil {
				t.Error("Expected violation to be created, got nil")
			}
			
			if !tt.expectViolation && violation != nil {
				t.Errorf("Expected no violation, got: %+v", violation)
			}
			
			if violation != nil {
				if violation.Type != models.ViolationTypeCommentedCode {
					t.Errorf("Expected violation type %v, got %v", models.ViolationTypeCommentedCode, violation.Type)
				}
				
				if violation.Severity != models.SeverityLow {
					t.Errorf("Expected severity %v, got %v", models.SeverityLow, violation.Severity)
				}
				
				if violation.Line != 10 {
					t.Errorf("Expected line 10, got %d", violation.Line)
				}
				
				if violation.File != "/test/file.rs" {
					t.Errorf("Expected file '/test/file.rs', got '%s'", violation.File)
				}
			}
		})
	}
}