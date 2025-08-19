package violations

import (
	"strings"
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

func TestRustTodoTrackerDetector_NewDetector(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustTodoTrackerDetector(config)
	
	if detector == nil {
		t.Fatal("Expected detector to be created")
	}
	
	if detector.Name() != "Rust Technical Debt Tracker" {
		t.Errorf("Expected name 'Rust Technical Debt Tracker', got '%s'", detector.Name())
	}
	
	if detector.Description() == "" {
		t.Error("Expected non-empty description")
	}
}

func TestRustTodoTrackerDetector_Detect_NoASTInfo(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustTodoTrackerDetector(config)
	
	fileInfo := &models.FileInfo{
		Path:     "/test/example.rs",
		Language: "Rust",
	}
	
	violations := detector.Detect(fileInfo, nil)
	if len(violations) != 0 {
		t.Errorf("Expected no violations when astInfo is nil, got %d", len(violations))
	}
}

func TestRustTodoTrackerDetector_Detect_WrongASTType(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustTodoTrackerDetector(config)
	
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

func TestRustTodoTrackerDetector_Detect_ValidRustAST(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustTodoTrackerDetector(config)
	
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

func TestRustTodoTrackerDetector_CheckForMarkers(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustTodoTrackerDetector(config)
	
	tests := []struct {
		name        string
		text        string
		expectViolation bool
		expectedMarker  string
		expectedSeverity models.Severity
	}{
		{
			name:        "TODO marker",
			text:        "TODO: Implement user authentication",
			expectViolation: true,
			expectedMarker:  "TODO",
			expectedSeverity: models.SeverityLow,
		},
		{
			name:        "FIXME marker",
			text:        "FIXME: This validation is broken",
			expectViolation: true,
			expectedMarker:  "FIXME",
			expectedSeverity: models.SeverityHigh,
		},
		{
			name:        "HACK marker",
			text:        "HACK: Temporary workaround for database",
			expectViolation: true,
			expectedMarker:  "HACK",
			expectedSeverity: models.SeverityMedium,
		},
		{
			name:        "BUG marker",
			text:        "BUG: Memory leak in this function",
			expectViolation: true,
			expectedMarker:  "BUG",
			expectedSeverity: models.SeverityHigh,
		},
		{
			name:        "OPTIMIZE marker",
			text:        "OPTIMIZE: This loop can be faster",
			expectViolation: true,
			expectedMarker:  "OPTIMIZE",
			expectedSeverity: models.SeverityLow,
		},
		{
			name:        "UNIMPLEMENTED marker",
			text:        "UNIMPLEMENTED: Feature not ready",
			expectViolation: true,
			expectedMarker:  "UNIMPLEMENTED",
			expectedSeverity: models.SeverityMedium,
		},
		{
			name:        "UNREACHABLE marker",
			text:        "UNREACHABLE: This should never happen",
			expectViolation: true,
			expectedMarker:  "UNREACHABLE",
			expectedSeverity: models.SeverityMedium,
		},
		{
			name:        "Regular comment",
			text:        "This is a regular comment",
			expectViolation: false,
		},
		{
			name:        "Empty text",
			text:        "",
			expectViolation: false,
		},
		{
			name:        "Case insensitive TODO",
			text:        "todo: implement this",
			expectViolation: true,
			expectedMarker:  "TODO",
			expectedSeverity: models.SeverityLow,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violation := detector.checkForMarkers(tt.text, 10, "/test/file.rs")
			
			if tt.expectViolation && violation == nil {
				t.Error("Expected violation to be created, got nil")
			}
			
			if !tt.expectViolation && violation != nil {
				t.Errorf("Expected no violation, got: %+v", violation)
			}
			
			if violation != nil {
				if violation.Type != models.ViolationTypeTodo {
					t.Errorf("Expected violation type %v, got %v", models.ViolationTypeTodo, violation.Type)
				}
				
				if violation.Severity != tt.expectedSeverity {
					t.Errorf("Expected severity %v, got %v", tt.expectedSeverity, violation.Severity)
				}
				
				if violation.Line != 10 {
					t.Errorf("Expected line 10, got %d", violation.Line)
				}
				
				if violation.File != "/test/file.rs" {
					t.Errorf("Expected file '/test/file.rs', got '%s'", violation.File)
				}
				
				// Check that the message contains the expected marker
				if tt.expectedMarker != "" && !strings.Contains(violation.Message, tt.expectedMarker) {
					t.Errorf("Expected message to contain '%s', got '%s'", tt.expectedMarker, violation.Message)
				}
			}
		})
	}
}

func TestRustTodoTrackerDetector_CheckForRustSpecificMarkers(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustTodoTrackerDetector(config)
	
	tests := []struct {
		name        string
		line        string
		expectViolation bool
		expectedMarker  string
		expectedSeverity models.Severity
	}{
		{
			name:        "panic! macro",
			line:        `panic!("Something went wrong");`,
			expectViolation: true,
			expectedMarker:  "PANIC",
			expectedSeverity: models.SeverityHigh,
		},
		{
			name:        "unimplemented! macro",
			line:        `unimplemented!("Feature not ready");`,
			expectViolation: true,
			expectedMarker:  "UNIMPLEMENTED",
			expectedSeverity: models.SeverityMedium,
		},
		{
			name:        "unreachable! macro",
			line:        `unreachable!("This should never happen");`,
			expectViolation: true,
			expectedMarker:  "UNREACHABLE",
			expectedSeverity: models.SeverityMedium,
		},
		{
			name:        "todo! macro",
			line:        `todo!("Implement this function");`,
			expectViolation: true,
			expectedMarker:  "TODO",
			expectedSeverity: models.SeverityLow,
		},
		{
			name:        "panic! without message",
			line:        `panic!();`,
			expectViolation: true,
			expectedMarker:  "PANIC",
			expectedSeverity: models.SeverityHigh,
		},
		{
			name:        "todo! without message",
			line:        `todo!();`,
			expectViolation: true,
			expectedMarker:  "TODO",
			expectedSeverity: models.SeverityLow,
		},
		{
			name:        "Regular function call",
			line:        `println!("Hello, world!");`,
			expectViolation: false,
		},
		{
			name:        "Regular code",
			line:        `let x = 42;`,
			expectViolation: false,
		},
		{
			name:        "Function definition",
			line:        `fn test_function() -> i32 {`,
			expectViolation: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violation := detector.checkForRustSpecificMarkers(tt.line, 15, "/test/code.rs")
			
			if tt.expectViolation && violation == nil {
				t.Error("Expected violation to be created, got nil")
			}
			
			if !tt.expectViolation && violation != nil {
				t.Errorf("Expected no violation, got: %+v", violation)
			}
			
			if violation != nil {
				if violation.Type != models.ViolationTypeTodo {
					t.Errorf("Expected violation type %v, got %v", models.ViolationTypeTodo, violation.Type)
				}
				
				if violation.Severity != tt.expectedSeverity {
					t.Errorf("Expected severity %v, got %v", tt.expectedSeverity, violation.Severity)
				}
				
				if violation.Line != 15 {
					t.Errorf("Expected line 15, got %d", violation.Line)
				}
				
				if violation.File != "/test/code.rs" {
					t.Errorf("Expected file '/test/code.rs', got '%s'", violation.File)
				}
				
				// Check that the message contains the expected marker
				if tt.expectedMarker != "" && !strings.Contains(violation.Message, tt.expectedMarker) {
					t.Errorf("Expected message to contain '%s', got '%s'", tt.expectedMarker, violation.Message)
				}
			}
		})
	}
}

func TestRustTodoTrackerDetector_ClassifyRustMarkerSeverity(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustTodoTrackerDetector(config)
	
	tests := []struct {
		marker   string
		expected models.Severity
	}{
		{"TODO", models.SeverityLow},
		{"OPTIMIZE", models.SeverityLow},
		{"REFACTOR", models.SeverityLow},
		{"HACK", models.SeverityMedium},
		{"XXX", models.SeverityMedium},
		{"UNWRAP", models.SeverityMedium},
		{"UNIMPLEMENTED", models.SeverityMedium},
		{"UNREACHABLE", models.SeverityMedium},
		{"BUG", models.SeverityHigh},
		{"FIXME", models.SeverityHigh},
		{"PANIC", models.SeverityHigh},
		{"UNKNOWN", models.SeverityInfo},
	}
	
	for _, tt := range tests {
		t.Run(tt.marker, func(t *testing.T) {
			result := detector.classifyRustMarkerSeverity(tt.marker)
			if result != tt.expected {
				t.Errorf("classifyRustMarkerSeverity(%s) = %v, want %v", tt.marker, result, tt.expected)
			}
		})
	}
}

func TestRustTodoTrackerDetector_GetRustMarkerDescription(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustTodoTrackerDetector(config)
	
	tests := []struct {
		marker   string
		expected string
	}{
		{"TODO", "Pending task or feature implementation"},
		{"FIXME", "Known issue that needs fixing"},
		{"HACK", "Temporary workaround that should be improved"},
		{"PANIC", "Code that intentionally panics - may need proper error handling"},
		{"UNIMPLEMENTED", "Placeholder for unimplemented functionality"},
		{"UNREACHABLE", "Code marked as unreachable - verify logic correctness"},
		{"UNKNOWN", "Technical debt marker"},
	}
	
	for _, tt := range tests {
		t.Run(tt.marker, func(t *testing.T) {
			result := detector.getRustMarkerDescription(tt.marker)
			if result != tt.expected {
				t.Errorf("getRustMarkerDescription(%s) = %q, want %q", tt.marker, result, tt.expected)
			}
		})
	}
}

func TestRustTodoTrackerDetector_GetRustMarkerSuggestion(t *testing.T) {
	config := DefaultDetectorConfig()
	detector := NewRustTodoTrackerDetector(config)
	
	tests := []struct {
		marker   string
		contains string // Check if suggestion contains this string
	}{
		{"TODO", "Schedule and complete"},
		{"FIXME", "Prioritize fixing"},
		{"HACK", "Replace this workaround"},
		{"PANIC", "Replace panic! with proper error handling"},
		{"UNWRAP", "Replace unwrap() with proper error handling"},
		{"UNIMPLEMENTED", "Implement the missing functionality"},
		{"UNREACHABLE", "Verify that this code is truly unreachable"},
		{"UNKNOWN", "Address this technical debt"},
	}
	
	for _, tt := range tests {
		t.Run(tt.marker, func(t *testing.T) {
			result := detector.getRustMarkerSuggestion(tt.marker)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("getRustMarkerSuggestion(%s) = %q, expected to contain %q", tt.marker, result, tt.contains)
			}
		})
	}
}

