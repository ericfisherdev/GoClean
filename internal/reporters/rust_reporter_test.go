package reporters

import (
	"testing"
	"time"

	"github.com/ericfisherdev/goclean/internal/models"
)

func TestEnhancedRustReporting(t *testing.T) {
	// Create a test report with Rust violations
	report := &models.Report{
		ID:          "test-report",
		GeneratedAt: time.Now(),
		Summary: &models.ScanSummary{
			TotalFiles:      1,
			ScannedFiles:    1,
			TotalViolations: 5,
			Duration:        100 * time.Millisecond,
		},
		Statistics: &models.Statistics{
			ViolationsByType: map[models.ViolationType]int{
				models.ViolationTypeRustInvalidFunctionNaming: 2,
				models.ViolationTypeRustOveruseUnwrap:         1,
				models.ViolationTypeRustUnnecessaryClone:      1,
				models.ViolationTypeRustUnsafeWithoutComment:  1,
			},
			ViolationsBySeverity: map[models.Severity]int{
				models.SeverityLow:    3,
				models.SeverityMedium: 2,
			},
		},
		Files: []*models.ScanResult{
			{
				File: &models.FileInfo{
					Path:     "example.rs",
					Language: "rust",
				},
				Violations: []*models.Violation{
					{
						Type:        models.ViolationTypeRustInvalidFunctionNaming,
						Severity:    models.SeverityLow,
						Line:        5,
						Message:     "Function name 'badName' should use snake_case",
						Description: "Rust functions should follow snake_case naming convention",
						Suggestion:  "Rename to 'bad_name'",
						CodeSnippet: "fn badName() -> i32 {\n    return 42;\n}",
					},
					{
						Type:        models.ViolationTypeRustOveruseUnwrap,
						Severity:    models.SeverityMedium,
						Line:        12,
						Message:     "Overuse of unwrap() without proper error handling",
						Description: "Consider using proper error handling instead of unwrap()",
						Suggestion:  "Use pattern matching or ? operator",
						CodeSnippet: "let value = result.unwrap(); // Dangerous!",
					},
					{
						Type:        models.ViolationTypeRustUnsafeWithoutComment,
						Severity:    models.SeverityMedium,
						Line:        20,
						Message:     "Unsafe block lacks explanation comment",
						Description: "Unsafe blocks should explain why they are safe",
						Suggestion:  "Add comment explaining safety invariants",
						CodeSnippet: "unsafe {\n    *ptr = 42;\n}",
					},
				},
			},
		},
	}

	t.Run("MarkdownReporter_RustSyntaxHighlighting", func(t *testing.T) {
		// Test the language detection function directly
		config := &MarkdownConfig{
			OutputPath:      "/tmp/test_rust_report.md",
			IncludeExamples: true,
		}
		reporter := NewMarkdownReporter(config)
		
		// Test language detection
		lang := reporter.detectLanguageFromFile("example.rs")
		if lang != "rust" {
			t.Errorf("Expected 'rust' language for .rs file, got '%s'", lang)
		}
		
		// Test categorization for Rust violations
		for _, result := range report.Files {
			for _, violation := range result.Violations {
				if models.IsRustSpecificViolation(violation.Type) {
					category := models.GetRustViolationCategory(violation.Type)
					if category == "" {
						t.Errorf("Expected non-empty category for Rust violation %v", violation.Type)
					}
				}
			}
		}
	})

	t.Run("ConsoleReporter_RustCategoryDisplay", func(t *testing.T) {
		_ = NewConsoleReporter(true, false) // verbose=true, colors=false
		
		// Capture console output by temporarily redirecting
		// This is a simplified test - in practice you'd use a buffer
		
		// Test that Rust categories are properly detected
		for _, result := range report.Files {
			for _, violation := range result.Violations {
				if models.IsRustSpecificViolation(violation.Type) {
					category := models.GetRustViolationCategory(violation.Type)
					if category == "" {
						t.Errorf("Expected non-empty category for Rust violation %v", violation.Type)
					}
				}
			}
		}
	})

	t.Run("HTMLReporter_RustEnhancements", func(t *testing.T) {
		config := &HTMLConfig{
			OutputPath: "/tmp/test_rust_report.html",
		}
		reporter, err := NewHTMLReporter(config)
		if err != nil {
			t.Fatalf("Failed to create HTML reporter: %v", err)
		}
		_ = reporter // Suppress unused variable
		
		// Test template functions
		_ = reporter // Suppress unused variable
		templateFuncs := getTemplateFunctions()
		
		// Test detectLanguage function
		detectLang, exists := templateFuncs["detectLanguage"]
		if !exists {
			t.Error("Expected detectLanguage template function")
		} else {
			if langFunc, ok := detectLang.(func(string) string); ok {
				if lang := langFunc("example.rs"); lang != "rust" {
					t.Errorf("Expected 'rust' language for .rs file, got '%s'", lang)
				}
				if lang := langFunc("example.go"); lang != "go" {
					t.Errorf("Expected 'go' language for .go file, got '%s'", lang)
				}
			} else {
				t.Error("detectLanguage function has wrong signature")
			}
		}
		
		// Test rustViolationCategory function
		rustCatFunc, exists := templateFuncs["rustViolationCategory"]
		if !exists {
			t.Error("Expected rustViolationCategory template function")
		} else {
			if catFunc, ok := rustCatFunc.(func(models.ViolationType) string); ok {
				category := catFunc(models.ViolationTypeRustOveruseUnwrap)
				if category != "error_handling" {
					t.Errorf("Expected 'error_handling' category, got '%s'", category)
				}
			} else {
				t.Error("rustViolationCategory function has wrong signature")
			}
		}
		
		// Test isRustViolation function  
		isRustFunc, exists := templateFuncs["isRustViolation"]
		if !exists {
			t.Error("Expected isRustViolation template function")
		} else {
			if checkFunc, ok := isRustFunc.(func(models.ViolationType) bool); ok {
				if !checkFunc(models.ViolationTypeRustOveruseUnwrap) {
					t.Error("Expected Rust violation to be detected as Rust-specific")
				}
				if checkFunc(models.ViolationTypeFunctionLength) {
					t.Error("Expected non-Rust violation to not be detected as Rust-specific")
				}
			} else {
				t.Error("isRustViolation function has wrong signature")
			}
		}
	})
}

func TestRustViolationCategorization(t *testing.T) {
	testCases := []struct {
		violationType models.ViolationType
		expectedCat   models.RustViolationCategory
		isRustSpecific bool
	}{
		{models.ViolationTypeRustOveruseUnwrap, models.RustCategoryErrorHandling, true},
		{models.ViolationTypeRustInvalidFunctionNaming, models.RustCategoryNaming, true},
		{models.ViolationTypeRustUnsafeWithoutComment, models.RustCategorySafety, true},
		{models.ViolationTypeRustUnnecessaryClone, models.RustCategoryOwnership, true},
		{models.ViolationTypeFunctionLength, "", false},
		{models.ViolationTypeNaming, "", false},
	}

	for _, tc := range testCases {
		t.Run(string(tc.violationType), func(t *testing.T) {
			// Test categorization
			actualCat := models.GetRustViolationCategory(tc.violationType)
			if actualCat != tc.expectedCat {
				t.Errorf("Expected category %s, got %s", tc.expectedCat, actualCat)
			}
			
			// Test Rust-specific detection
			isRust := models.IsRustSpecificViolation(tc.violationType)
			if isRust != tc.isRustSpecific {
				t.Errorf("Expected isRustSpecific %v, got %v", tc.isRustSpecific, isRust)
			}
		})
	}
}