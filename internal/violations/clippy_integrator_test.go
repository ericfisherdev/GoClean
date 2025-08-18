package violations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
)

func TestClippyIntegrator_NewClippyIntegrator(t *testing.T) {
	integrator := NewClippyIntegrator(nil)
	if integrator == nil {
		t.Fatal("Expected ClippyIntegrator to be created")
	}
	
	if integrator.Name() != "ClippyIntegrator" {
		t.Errorf("Expected name 'ClippyIntegrator', got '%s'", integrator.Name())
	}
	
	if integrator.Description() == "" {
		t.Error("Expected non-empty description")
	}
}

func TestClippyIntegrator_DetectNonRustFile(t *testing.T) {
	integrator := NewClippyIntegrator(DefaultDetectorConfig())
	
	fileInfo := &models.FileInfo{
		Path:     "test.go",
		Language: "go",
		Lines:    10,
	}
	
	violations := integrator.Detect(fileInfo, nil)
	if len(violations) != 0 {
		t.Errorf("Expected no violations for non-Rust file, got %d", len(violations))
	}
}

func TestClippyIntegrator_FindCargoProjectRoot(t *testing.T) {
	integrator := NewClippyIntegrator(DefaultDetectorConfig())
	
	// Create a temporary directory structure
	tempDir := t.TempDir()
	
	// Create nested directory structure
	subDir := filepath.Join(tempDir, "src", "submodule")
	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	
	// Create Cargo.toml in the root
	cargoPath := filepath.Join(tempDir, "Cargo.toml")
	err = os.WriteFile(cargoPath, []byte(`[package]
name = "test_project"
version = "0.1.0"
edition = "2021"
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create Cargo.toml: %v", err)
	}
	
	// Test finding from subdirectory
	testFile := filepath.Join(subDir, "test.rs")
	projectRoot := integrator.findCargoProjectRoot(testFile)
	
	if projectRoot != tempDir {
		t.Errorf("Expected project root '%s', got '%s'", tempDir, projectRoot)
	}
}

func TestClippyIntegrator_FindCargoProjectRootNoCargoToml(t *testing.T) {
	integrator := NewClippyIntegrator(DefaultDetectorConfig())
	
	// Create a temporary directory without Cargo.toml
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.rs")
	
	projectRoot := integrator.findCargoProjectRoot(testFile)
	
	if projectRoot != "" {
		t.Errorf("Expected empty project root for directory without Cargo.toml, got '%s'", projectRoot)
	}
}

func TestClippyIntegrator_ParseClippyOutput(t *testing.T) {
	integrator := NewClippyIntegrator(DefaultDetectorConfig())
	
	// Mock clippy JSON output
	clippyJSON := `{"reason":"compiler-message","package_id":"test 0.1.0 (path+file:///tmp/test)","target":{"kind":["bin"],"crate_types":["bin"],"name":"test","src_path":"/tmp/test/src/main.rs","edition":"2021","required-features":[]},"message":{"message":"this function has too many arguments (5/4)","code":{"code":"clippy::too_many_arguments","explanation":""},"level":"warning","spans":[{"file_name":"src/main.rs","byte_start":0,"byte_end":50,"line_start":1,"line_end":1,"column_start":1,"column_end":51,"is_primary":true,"text":[{"text":"fn test_function(a: i32, b: i32, c: i32, d: i32, e: i32) {","highlight_start":1,"highlight_end":51}],"label":"","suggested_replacement":null,"suggestion_applicability":null,"expansion":null}],"children":[{"message":"help: consider passing some arguments as a single struct","code":null,"level":"help","spans":[],"children":[],"rendered":null}],"rendered":"warning: this function has too many arguments (5/4)\n --> src/main.rs:1:1\n  |\n1 | fn test_function(a: i32, b: i32, c: i32, d: i32, e: i32) {\n  | ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^\n  |\n  = help: consider passing some arguments as a single struct\n"}}`
	
	violations := integrator.parseClippyOutput([]byte(clippyJSON), "src/main.rs")
	
	if len(violations) != 1 {
		t.Fatalf("Expected 1 violation, got %d", len(violations))
	}
	
	violation := violations[0]
	
	// Check basic violation properties
	if violation.Type != models.ViolationTypeCyclomaticComplexity {
		t.Errorf("Expected violation type %s, got %s", models.ViolationTypeCyclomaticComplexity, violation.Type)
	}
	
	if violation.Line != 1 {
		t.Errorf("Expected line 1, got %d", violation.Line)
	}
	
	if violation.Column != 1 {
		t.Errorf("Expected column 1, got %d", violation.Column)
	}
	
	if !strings.Contains(violation.Message, "Detected by rust-clippy") {
		t.Errorf("Expected message to contain 'Detected by rust-clippy', got '%s'", violation.Message)
	}
	
	if violation.Rule != "clippy::too_many_arguments" {
		t.Errorf("Expected rule 'clippy::too_many_arguments', got '%s'", violation.Rule)
	}
	
	if !strings.Contains(violation.Description, "rust-clippy") {
		t.Errorf("Expected description to contain 'rust-clippy', got '%s'", violation.Description)
	}
}

func TestClippyIntegrator_ParseClippyOutputMultipleMessages(t *testing.T) {
	integrator := NewClippyIntegrator(DefaultDetectorConfig())
	
	// Mock multiple clippy messages
	clippyOutput := `{"reason":"build-finished","success":false}
{"reason":"compiler-message","package_id":"test 0.1.0","target":{"kind":["bin"],"crate_types":["bin"],"name":"test","src_path":"src/main.rs","edition":"2021","required-features":[]},"message":{"message":"missing documentation for a function","code":{"code":"clippy::missing_docs_in_private_items","explanation":""},"level":"warning","spans":[{"file_name":"src/main.rs","byte_start":0,"byte_end":25,"line_start":1,"line_end":1,"column_start":1,"column_end":26,"is_primary":true,"text":[{"text":"fn undocumented_function() {","highlight_start":1,"highlight_end":26}],"label":"","suggested_replacement":null,"suggestion_applicability":null,"expansion":null}],"children":[],"rendered":"warning: missing documentation for a function"}}
{"reason":"compiler-message","package_id":"test 0.1.0","target":{"kind":["bin"],"crate_types":["bin"],"name":"test","src_path":"src/main.rs","edition":"2021","required-features":[]},"message":{"message":"this loop could be written as a for loop","code":{"code":"clippy::while_let_loop","explanation":""},"level":"warning","spans":[{"file_name":"src/main.rs","byte_start":50,"byte_end":100,"line_start":3,"line_end":5,"column_start":5,"column_end":6,"is_primary":true,"text":[{"text":"    while let Some(x) = iter.next() {","highlight_start":5,"highlight_end":40}],"label":"","suggested_replacement":null,"suggestion_applicability":null,"expansion":null}],"children":[{"message":"help: consider using for instead","code":null,"level":"help","spans":[],"children":[],"rendered":null}],"rendered":"warning: this loop could be written as a for loop"}}`
	
	violations := integrator.parseClippyOutput([]byte(clippyOutput), "src/main.rs")
	
	if len(violations) != 2 {
		t.Fatalf("Expected 2 violations, got %d", len(violations))
	}
	
	// Check first violation (missing docs)
	docViolation := violations[0]
	if docViolation.Type != models.ViolationTypeMissingDocumentation {
		t.Errorf("Expected first violation type %s, got %s", models.ViolationTypeMissingDocumentation, docViolation.Type)
	}
	
	// Check second violation (style)
	styleViolation := violations[1]
	if styleViolation.Type != models.ViolationTypeStructure {
		t.Errorf("Expected second violation type %s, got %s", models.ViolationTypeStructure, styleViolation.Type)
	}
}

func TestClippyIntegrator_MapClippySeverity(t *testing.T) {
	integrator := NewClippyIntegrator(DefaultDetectorConfig())
	
	tests := []struct {
		level    string
		lintCode string
		expected models.Severity
	}{
		{"error", "any_lint", models.SeverityHigh},
		{"warning", "correctness", models.SeverityHigh},
		{"warning", "suspicious", models.SeverityMedium},
		{"warning", "complexity", models.SeverityMedium},
		{"warning", "perf", models.SeverityMedium},
		{"warning", "style", models.SeverityLow},
		{"warning", "unknown", models.SeverityMedium},
		{"info", "any", models.SeverityLow},
	}
	
	for _, tt := range tests {
		t.Run(tt.level+"_"+tt.lintCode, func(t *testing.T) {
			severity := integrator.mapClippySeverity(tt.level, tt.lintCode)
			if severity != tt.expected {
				t.Errorf("Expected severity %v for level '%s' and lint '%s', got %v", 
					tt.expected, tt.level, tt.lintCode, severity)
			}
		})
	}
}

func TestClippyIntegrator_MapClippyLintToViolationType(t *testing.T) {
	integrator := NewClippyIntegrator(DefaultDetectorConfig())
	
	tests := []struct {
		lintCode string
		expected models.ViolationType
	}{
		{"cognitive_complexity", models.ViolationTypeCyclomaticComplexity},
		{"too_many_arguments", models.ViolationTypeCyclomaticComplexity},
		{"missing_docs", models.ViolationTypeMissingDocumentation},
		{"wrong_self_convention", models.ViolationTypeNaming},
		{"todo", models.ViolationTypeTodo},
		{"unimplemented", models.ViolationTypeTodo},
		{"panic", models.ViolationTypeTodo},
		{"magic_number", models.ViolationTypeMagicNumber},
		{"approx_constant", models.ViolationTypeMagicNumber},
		{"duplicate", models.ViolationTypeDuplication},
		{"similar", models.ViolationTypeDuplication},
		{"unknown_lint", models.ViolationTypeStructure},
	}
	
	for _, tt := range tests {
		t.Run(tt.lintCode, func(t *testing.T) {
			violationType := integrator.mapClippyLintToViolationType(tt.lintCode)
			if violationType != tt.expected {
				t.Errorf("Expected violation type %v for lint '%s', got %v", 
					tt.expected, tt.lintCode, violationType)
			}
		})
	}
}

func TestClippyIntegrator_GenerateClippySuggestion(t *testing.T) {
	integrator := NewClippyIntegrator(DefaultDetectorConfig())
	
	tests := []struct {
		name        string
		diagnostic  ClippyDiagnostic
		expectedContains string
	}{
		{
			name: "Diagnostic with help message",
			diagnostic: ClippyDiagnostic{
				Message: "this function has too many arguments",
				Children: []ClippyDiagnostic{
					{Message: "help: consider passing some arguments as a single struct"},
				},
			},
			expectedContains: "consider passing some arguments as a single struct",
		},
		{
			name: "Diagnostic with explanation",
			diagnostic: ClippyDiagnostic{
				Message: "magic number detected",
				Code: &ClippyCode{
					Explanation: "Magic numbers make code harder to maintain",
				},
			},
			expectedContains: "Magic numbers make code harder to maintain",
		},
		{
			name: "Diagnostic without help or explanation",
			diagnostic: ClippyDiagnostic{
				Message: "some lint message",
			},
			expectedContains: "Consider addressing this clippy lint",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestion := integrator.generateClippySuggestion(tt.diagnostic)
			if !strings.Contains(suggestion, tt.expectedContains) {
				t.Errorf("Expected suggestion to contain '%s', got '%s'", 
					tt.expectedContains, suggestion)
			}
		})
	}
}

func TestClippyIntegrator_ExtractCodeSnippet(t *testing.T) {
	integrator := NewClippyIntegrator(DefaultDetectorConfig())
	
	tests := []struct {
		name     string
		span     ClippySpan
		expected string
	}{
		{
			name: "Span with text",
			span: ClippySpan{
				Text: []ClippyText{
					{Text: "  fn test_function() {  "},
				},
			},
			expected: "fn test_function() {",
		},
		{
			name: "Span without text",
			span: ClippySpan{
				Text: []ClippyText{},
			},
			expected: "",
		},
		{
			name: "Span with multiple text entries",
			span: ClippySpan{
				Text: []ClippyText{
					{Text: "let x = 42;"},
					{Text: "let y = 24;"},
				},
			},
			expected: "let x = 42;",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := integrator.extractCodeSnippet(&tt.span)
			if result != tt.expected {
				t.Errorf("Expected code snippet '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestClippyIntegrator_ConvertDiagnosticToViolation(t *testing.T) {
	integrator := NewClippyIntegrator(DefaultDetectorConfig())
	
	diagnostic := ClippyDiagnostic{
		Message: "this function has too many arguments (5/4)",
		Code: &ClippyCode{
			Code: "clippy::too_many_arguments",
		},
		Level: "warning",
		Spans: []ClippySpan{
			{
				FileName:    "src/main.rs",
				LineStart:   10,
				ColumnStart: 5,
				IsPrimary:   true,
				Text: []ClippyText{
					{Text: "fn test_function(a: i32, b: i32, c: i32, d: i32, e: i32) {"},
				},
			},
		},
	}
	
	violation := integrator.convertDiagnosticToViolation(diagnostic, "src/main.rs")
	
	if violation == nil {
		t.Fatal("Expected violation to be created")
	}
	
	if violation.Type != models.ViolationTypeCyclomaticComplexity {
		t.Errorf("Expected violation type %s, got %s", models.ViolationTypeCyclomaticComplexity, violation.Type)
	}
	
	if violation.Line != 10 {
		t.Errorf("Expected line 10, got %d", violation.Line)
	}
	
	if violation.Column != 5 {
		t.Errorf("Expected column 5, got %d", violation.Column)
	}
	
	if !strings.Contains(violation.Message, "Detected by rust-clippy") {
		t.Errorf("Expected message to contain 'Detected by rust-clippy', got '%s'", violation.Message)
	}
	
	if !strings.Contains(violation.Description, "rust-clippy") {
		t.Errorf("Expected description to contain 'rust-clippy', got '%s'", violation.Description)
	}
}

func TestClippyIntegrator_ConvertDiagnosticToViolationNonPrimarySpan(t *testing.T) {
	integrator := NewClippyIntegrator(DefaultDetectorConfig())
	
	diagnostic := ClippyDiagnostic{
		Message: "test message",
		Level:   "warning",
		Spans: []ClippySpan{
			{
				FileName:  "src/main.rs",
				IsPrimary: false, // Non-primary span
			},
		},
	}
	
	violation := integrator.convertDiagnosticToViolation(diagnostic, "src/main.rs")
	
	if violation != nil {
		t.Error("Expected no violation for diagnostic without primary span")
	}
}

func TestClippyIntegrator_ConvertDiagnosticToViolationWrongFile(t *testing.T) {
	integrator := NewClippyIntegrator(DefaultDetectorConfig())
	
	diagnostic := ClippyDiagnostic{
		Message: "test message",
		Level:   "warning",
		Spans: []ClippySpan{
			{
				FileName:  "src/other.rs", // Different file
				IsPrimary: true,
			},
		},
	}
	
	violation := integrator.convertDiagnosticToViolation(diagnostic, "src/main.rs")
	
	if violation != nil {
		t.Error("Expected no violation for diagnostic from different file")
	}
}

func TestClippyIntegrator_ConvertDiagnosticToViolationInfoLevel(t *testing.T) {
	integrator := NewClippyIntegrator(DefaultDetectorConfig())
	
	diagnostic := ClippyDiagnostic{
		Message: "test message",
		Level:   "info", // Info level should be skipped
		Spans: []ClippySpan{
			{
				FileName:  "src/main.rs",
				IsPrimary: true,
			},
		},
	}
	
	violation := integrator.convertDiagnosticToViolation(diagnostic, "src/main.rs")
	
	if violation != nil {
		t.Error("Expected no violation for info level diagnostic")
	}
}

func TestClippyIntegrator_ConvertDiagnosticToViolationNilCode(t *testing.T) {
	integrator := NewClippyIntegrator(DefaultDetectorConfig())
	
	diagnostic := ClippyDiagnostic{
		Message: "test message",
		Code:    nil, // Nil code field
		Level:   "warning",
		Spans: []ClippySpan{
			{
				FileName:    "src/main.rs",
				LineStart:   10,
				ColumnStart: 5,
				IsPrimary:   true,
			},
		},
	}
	
	violation := integrator.convertDiagnosticToViolation(diagnostic, "src/main.rs")
	
	if violation != nil {
		t.Fatal("Expected no violation to be created for nil Code")
	}
}

// TestClippyIntegrator_Integration provides an integration test for clippy
// Note: This test requires cargo and rust to be installed and will be skipped if not available
func TestClippyIntegrator_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	integrator := NewClippyIntegrator(DefaultDetectorConfig())
	
	// Check if clippy is available
	if !integrator.isClippyAvailable() {
		t.Skip("Clippy not available, skipping integration test")
	}
	
	// Create a temporary Rust project
	tempDir := t.TempDir()
	
	// Create Cargo.toml
	cargoToml := `[package]
name = "test_project"
version = "0.1.0"
edition = "2021"

[dependencies]
`
	err := os.WriteFile(filepath.Join(tempDir, "Cargo.toml"), []byte(cargoToml), 0644)
	if err != nil {
		t.Fatalf("Failed to create Cargo.toml: %v", err)
	}
	
	// Create src directory
	srcDir := filepath.Join(tempDir, "src")
	err = os.MkdirAll(srcDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create src directory: %v", err)
	}
	
	// Create main.rs with clippy violations
	mainRs := `fn main() {
    println!("Hello, world!");
}

// This function has intentional clippy violations
fn problematic_function(a: i32, b: i32, c: i32, d: i32, e: i32) -> i32 {
    let magic = 42; // Magic number
    a + b + c + d + e + magic
}
`
	err = os.WriteFile(filepath.Join(srcDir, "main.rs"), []byte(mainRs), 0644)
	if err != nil {
		t.Fatalf("Failed to create main.rs: %v", err)
	}
	
	// Test the clippy integration
	fileInfo := &models.FileInfo{
		Path:     filepath.Join(srcDir, "main.rs"),
		Language: "rust",
		Lines:    8,
	}
	
	violations := integrator.Detect(fileInfo, nil)
	
	// Should detect some violations (exact count may vary by clippy version)
	if len(violations) == 0 {
		t.Log("No violations detected - this may be expected if clippy configuration differs")
	} else {
		t.Logf("Detected %d clippy violations", len(violations))
		for i, v := range violations {
			t.Logf("  Violation %d: %s - %s (line %d)", i+1, v.Type, v.Message, v.Line)
		}
	}
	
	// Verify all violations have proper attribution
	for _, v := range violations {
		if !strings.Contains(v.Description, "rust-clippy") {
			t.Errorf("Expected description to contain 'rust-clippy', got '%s'", v.Description)
		}
		if !strings.Contains(v.Message, "Detected by rust-clippy") {
			t.Errorf("Expected message to contain 'Detected by rust-clippy', got '%s'", v.Message)
		}
	}
}