// +build integration

package reporters

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ericfisherdev/goclean/internal/config"
	"github.com/ericfisherdev/goclean/internal/models"
)

// TestRustHTMLReportGeneration tests HTML report generation for Rust violations
func TestRustHTMLReportGeneration(t *testing.T) {
	// Create test violations
	violations := []models.Violation{
		{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityMedium,
			Rule:        "RUST_INVALID_FUNCTION_NAMING",
			Message:     "Function 'getUserData' should use snake_case naming",
			File:        "src/lib.rs",
			Line:        10,
			Column:      1,
			Suggestion:  "Rename to 'get_user_data'",
			CodeSnippet: "fn getUserData() -> String {",
		},
		{
			Type:        "RUST_OVERUSE_UNWRAP",
			Severity:    models.SeverityHigh,
			Rule:        "RUST_OVERUSE_UNWRAP",
			Message:     "Use of unwrap() may cause panic - consider using ? operator",
			File:        "src/main.rs",
			Line:        25,
			Column:      15,
			Suggestion:  "Replace .unwrap() with ? operator for proper error propagation",
			CodeSnippet: "let file = File::open(\"config.txt\").unwrap();",
		},
		{
			Type:        models.ViolationTypeMagicNumber,
			Severity:    models.SeverityLow,
			Rule:        "RUST_MAGIC_NUMBER",
			Message:     "Magic number 42 should be extracted to a named constant",
			File:        "src/utils.rs",
			Line:        8,
			Column:      20,
			Suggestion:  "Extract to a const like 'const DEFAULT_VALUE: i32 = 42;'",
			CodeSnippet: "let result = value * 42;",
		},
		{
			Type:        models.ViolationTypeMissingDocumentation,
			Severity:    models.SeverityLow,
			Rule:        "RUST_MISSING_DOCUMENTATION",
			Message:     "Missing documentation for public function 'process_data'",
			File:        "src/lib.rs",
			Line:        50,
			Column:      1,
			Suggestion:  "Add documentation comment: /// Process the input data",
			CodeSnippet: "pub fn process_data(input: &str) -> Result<String, Error> {",
		},
	}

	// Create statistics
	stats := models.Statistics{
		TotalViolations:  len(violations),
		FilesScanned:     3,
		FilesByLanguage:  map[string]int{"rust": 3},
		ViolationsByType: map[models.ViolationType]int{
			models.ViolationTypeNaming:              1,
			"RUST_OVERUSE_UNWRAP":                   1,
			models.ViolationTypeMagicNumber:         1,
			models.ViolationTypeMissingDocumentation: 1,
		},
		ViolationsBySeverity: map[models.Severity]int{
			models.SeverityLow:    2,
			models.SeverityMedium: 1,
			models.SeverityHigh:   1,
		},
	}

	// Test HTML report generation
	t.Run("HTML Report", func(t *testing.T) {
		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "rust-report.html")
		
		htmlReporter, err := NewHTMLReporter(&HTMLConfig{
			OutputPath: outputPath,
			Title:      "Rust Code Analysis Report",
		})
		if err != nil {
			t.Fatalf("Failed to create HTML reporter: %v", err)
		}

		rpt := buildReport(violations, stats)
		err = htmlReporter.Generate(rpt)
		if err != nil {
			t.Fatalf("Failed to generate HTML report: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("HTML report file was not created")
		}

		// Read and verify content
		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read HTML report: %v", err)
		}

		htmlContent := string(content)

		// Verify HTML content with stable elements
		expectedContent := []string{
			"GoClean Code Analysis Report",
			"Function 'getUserData' should use snake_case naming",
			"Use of unwrap() may cause panic - consider using ? operator",
			"Magic number 42 should be extracted to a named constant",
			"Missing documentation for public function 'process_data'",
			"src/lib.rs",
			"src/main.rs",
			"src/utils.rs",
		}

		for _, expected := range expectedContent {
			if !strings.Contains(htmlContent, expected) {
				t.Errorf("HTML report missing expected content: %s", expected)
			}
		}
	})

	// Test Markdown report generation
	t.Run("Markdown Report", func(t *testing.T) {
		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "rust-violations.md")
		
		mdReporter := NewMarkdownReporter(&MarkdownConfig{
			OutputPath: outputPath,
		})

		rpt := buildReport(violations, stats)
		err := mdReporter.Generate(rpt)
		if err != nil {
			t.Fatalf("Failed to generate Markdown report: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("Markdown report file was not created")
		}

		// Read and verify content
		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read Markdown report: %v", err)
		}

		mdContent := string(content)

		// Verify Rust-specific content
		expectedPatterns := []string{
			"## Violations by File",
			"### src/lib.rs",
			"### src/main.rs",
			"### src/utils.rs",
			"RUST_INVALID_FUNCTION_NAMING",
			"RUST_OVERUSE_UNWRAP",
			"```rust",
			"snake_case",
			"unwrap()",
		}

		for _, pattern := range expectedPatterns {
			if !strings.Contains(mdContent, pattern) {
				t.Errorf("Markdown report missing expected pattern: %s", pattern)
			}
		}

		// Verify code snippets are properly formatted
		if !strings.Contains(mdContent, "```rust\nfn getUserData()") {
			t.Error("Rust code snippets not properly formatted in Markdown")
		}
	})

	// Test Console report generation
	t.Run("Console Report", func(t *testing.T) {
		var buf bytes.Buffer
		
 		consoleReporter := NewConsoleReporter(&config.ConsoleConfig{
 			Colored: false, // Disable colors for testing
 			Output:  &buf,
 			Verbose: false,
 		})

		rpt := buildReport(violations, stats)
		err := consoleReporter.Generate(rpt)
		if err != nil {
			t.Fatalf("Failed to generate console report: %v", err)
		}

		output := buf.String()

		// Verify console output contains Rust violations
		expectedOutput := []string{
			"src/lib.rs",
			"src/main.rs",
			"src/utils.rs",
			"RUST_INVALID_FUNCTION_NAMING",
			"RUST_OVERUSE_UNWRAP",
			"snake_case",
			"unwrap()",
		}

		for _, expected := range expectedOutput {
			if !strings.Contains(output, expected) {
				t.Errorf("Console output missing expected content: %s", expected)
			}
		}
	})

	// Test JSON report generation
	t.Run("JSON Report", func(t *testing.T) {
		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "rust-violations.json")
		
		jsonReporter := NewJSONReporter(&config.JSONConfig{
			Path: outputPath,
		})

		rpt := buildReport(violations, stats)
		err := jsonReporter.Generate(rpt)
		if err != nil {
			t.Fatalf("Failed to generate JSON report: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("JSON report file was not created")
		}

		// Read and verify content
		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read JSON report: %v", err)
		}

		jsonContent := string(content)

		// Verify JSON contains Rust-specific fields
		expectedJSON := []string{
			`"rule":"RUST_INVALID_FUNCTION_NAMING"`,
			`"rule":"RUST_OVERUSE_UNWRAP"`,
			`"file":"src/lib.rs"`,
			`"file":"src/main.rs"`,
			`"file":"src/utils.rs"`,
			`"message":"Function 'getUserData' should use snake_case naming"`,
			`"code_snippet":"fn getUserData() -> String {"`,
		}

		for _, expected := range expectedJSON {
			if !strings.Contains(jsonContent, expected) {
				t.Errorf("JSON report missing expected content: %s", expected)
			}
		}
	})
}

// TestMixedLanguageReportGeneration tests report generation for mixed Go/Rust projects
func TestMixedLanguageReportGeneration(t *testing.T) {
	// Create mixed violations
	violations := []models.Violation{
		// Go violations
		{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityMedium,
			Rule:        "GO_INVALID_FUNCTION_NAMING",
			Message:     "Function 'get_data' should use camelCase naming",
			File:        "main.go",
			Line:        15,
			Column:      1,
			CodeSnippet: "func get_data() string {",
		},
		{
			Type:        models.ViolationTypeFunctionLength,
			Severity:    models.SeverityHigh,
			Rule:        "GO_FUNCTION_TOO_LONG",
			Message:     "Function 'processData' is 150 lines long (max: 50)",
			File:        "processor.go",
			Line:        25,
			Column:      1,
			CodeSnippet: "func processData(input []byte) error {",
		},
		// Rust violations
		{
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityMedium,
			Rule:        "RUST_INVALID_STRUCT_NAMING",
			Message:     "Struct 'user_data' should use PascalCase naming",
			File:        "src/models.rs",
			Line:        8,
			Column:      1,
			CodeSnippet: "struct user_data {",
		},
		{
			Type:        "RUST_UNNECESSARY_CLONE",
			Severity:    models.SeverityMedium,
			Rule:        "RUST_UNNECESSARY_CLONE",
			Message:     "Unnecessary clone detected - value is not used after clone",
			File:        "src/lib.rs",
			Line:        42,
			Column:      10,
			CodeSnippet: "let result = data.clone();",
		},
	}

	// Create statistics for mixed project
	stats := models.Statistics{
		TotalViolations:  len(violations),
		FilesScanned:     4,
		FilesByLanguage:  map[string]int{"go": 2, "rust": 2},
		ViolationsByType: map[models.ViolationType]int{
			models.ViolationTypeNaming:        2,
			models.ViolationTypeFunctionLength: 1,
			"RUST_UNNECESSARY_CLONE":          1,
		},
		ViolationsBySeverity: map[models.Severity]int{
			models.SeverityMedium: 3,
			models.SeverityHigh:   1,
		},
	}

	// Test HTML report for mixed project
	t.Run("Mixed HTML Report", func(t *testing.T) {
		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "mixed-report.html")
		
		htmlReporter, err := NewHTMLReporter(&config.HTMLConfig{
			Path: outputPath,
		})
		if err != nil {
			t.Fatalf("Failed to create HTML reporter: %v", err)
		}

		rpt := buildReport(violations, stats)
		err := htmlReporter.Generate(rpt)
		if err != nil {
			t.Fatalf("Failed to generate mixed HTML report: %v", err)
		}

		// Read and verify content
		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read mixed HTML report: %v", err)
		}

		htmlContent := string(content)

		// Verify both languages are represented
		goPatterns := []string{
			"main.go",
			"processor.go",
			"GO_INVALID_FUNCTION_NAMING",
			"camelCase",
		}

		rustPatterns := []string{
			"src/models.rs",
			"src/lib.rs",
			"RUST_INVALID_STRUCT_NAMING",
			"PascalCase",
			"RUST_UNNECESSARY_CLONE",
		}

		for _, pattern := range goPatterns {
			if !strings.Contains(htmlContent, pattern) {
				t.Errorf("Mixed report missing Go content: %s", pattern)
			}
		}

		for _, pattern := range rustPatterns {
			if !strings.Contains(htmlContent, pattern) {
				t.Errorf("Mixed report missing Rust content: %s", pattern)
			}
		}

		// Verify language statistics
		if !strings.Contains(htmlContent, "go") || !strings.Contains(htmlContent, "rust") {
			t.Error("Mixed report missing language statistics")
		}
	})

	// Test Markdown report for mixed project
	t.Run("Mixed Markdown Report", func(t *testing.T) {
		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "mixed-violations.md")
		
		mdReporter := NewMarkdownReporter(&MarkdownConfig{
			OutputPath: outputPath,
		})

		rpt := buildReport(violations, stats)
		err := mdReporter.Generate(rpt)
		if err != nil {
			t.Fatalf("Failed to generate mixed Markdown report: %v", err)
		}

		// Read and verify content
		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read mixed Markdown report: %v", err)
		}

		mdContent := string(content)

		// Verify proper code block syntax highlighting
		if !strings.Contains(mdContent, "```go") {
			t.Error("Go code blocks not properly highlighted")
		}
		if !strings.Contains(mdContent, "```rust") {
			t.Error("Rust code blocks not properly highlighted")
		}

		// Verify statistics section
		if !strings.Contains(mdContent, "Files by Language") {
			t.Error("Language statistics missing from Markdown report")
		}
	})
}

// buildReport creates a properly structured models.Report from violations and statistics
func buildReport(vs []models.Violation, stats models.Statistics) *models.Report {
	byFile := map[string][]*models.Violation{}
	for i := range vs {
		v := &vs[i]
		byFile[v.File] = append(byFile[v.File], v)
	}
	var files []*models.ScanResult
	for filePath, vlist := range byFile {
		files = append(files, &models.ScanResult{
			File: &models.FileInfo{
				Path:      filePath,
				Name:      filePath, // Simplified for tests
				Extension: ".rs",    // Assumed for Rust tests
				Language:  "rust",
				Scanned:   true,
			},
			Violations: vlist,
		})
	}
	return &models.Report{
		ID:          "test-report",
		GeneratedAt: time.Now(),
		Summary: &models.ScanSummary{
			TotalFiles:      stats.FilesScanned,
			TotalViolations: stats.TotalViolations,
		},
		Statistics: &stats,
		Files:      files,
	}
}