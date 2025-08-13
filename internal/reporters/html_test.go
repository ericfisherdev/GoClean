package reporters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ericfisherdev/goclean/internal/models"
)

func TestNewHTMLReporter(t *testing.T) {
	config := &HTMLConfig{
		OutputPath:      "test.html",
		AutoRefresh:     true,
		RefreshInterval: 30,
		Theme:           "auto",
	}

	reporter, err := NewHTMLReporter(config)
	if err != nil {
		t.Fatalf("Failed to create HTML reporter: %v", err)
	}

	if reporter.config != config {
		t.Error("Reporter config not set correctly")
	}

	if reporter.template == nil {
		t.Error("Template not loaded")
	}
}

func TestHTMLReporter_Generate(t *testing.T) {
	// Create temporary directory for test output
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test-report.html")

	config := &HTMLConfig{
		OutputPath:      outputPath,
		AutoRefresh:     true,
		RefreshInterval: 10,
		Theme:           "light",
	}

	reporter, err := NewHTMLReporter(config)
	if err != nil {
		t.Fatalf("Failed to create HTML reporter: %v", err)
	}

	// Create test report data
	report := createTestReport()

	// Generate report
	err = reporter.Generate(report)
	if err != nil {
		t.Fatalf("Failed to generate HTML report: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("HTML report file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated HTML file: %v", err)
	}

	htmlContent := string(content)

	// Verify basic HTML structure and key elements
	basicElements := []string{
		"<!DOCTYPE html>",
		"<html",
		"<head>",
		"<body>",
		"GoClean",
		`<meta http-equiv="refresh" content="10">`, // Auto-refresh meta tag
	}

	for _, element := range basicElements {
		if !strings.Contains(htmlContent, element) {
			t.Errorf("Generated HTML missing basic element: %s", element)
		}
	}

}

func TestHTMLReporter_GenerateWithProgress(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "progress-report.html")

	config := &HTMLConfig{
		OutputPath:      outputPath,
		AutoRefresh:     false,
		RefreshInterval: 0,
		Theme:           "dark",
	}

	reporter, err := NewHTMLReporter(config)
	if err != nil {
		t.Fatalf("Failed to create HTML reporter: %v", err)
	}

	report := createTestReport()

	// Track progress messages
	var progressMessages []string
	progressFn := func(message string) {
		progressMessages = append(progressMessages, message)
	}

	err = reporter.GenerateWithProgress(report, progressFn)
	if err != nil {
		t.Fatalf("Failed to generate HTML report with progress: %v", err)
	}

	// Verify progress callback was called
	if len(progressMessages) == 0 {
		t.Error("Progress callback was not called")
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("HTML report file was not created")
	}
}

func TestHTMLReporter_NoAutoRefresh(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "no-refresh-report.html")

	config := &HTMLConfig{
		OutputPath:      outputPath,
		AutoRefresh:     false,
		RefreshInterval: 0,
		Theme:           "auto",
	}

	reporter, err := NewHTMLReporter(config)
	if err != nil {
		t.Fatalf("Failed to create HTML reporter: %v", err)
	}

	report := createTestReport()

	err = reporter.Generate(report)
	if err != nil {
		t.Fatalf("Failed to generate HTML report: %v", err)
	}

	// Read content and verify no auto-refresh meta tag
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated HTML file: %v", err)
	}

	htmlContent := string(content)
	if strings.Contains(htmlContent, `<meta http-equiv="refresh"`) {
		t.Error("HTML should not contain auto-refresh meta tag when disabled")
	}
}

func createTestReport() *models.Report {
	// Create test violations
	violations := []*models.Violation{
		{
			ID:          "test-1",
			Type:        models.ViolationTypeFunctionLength,
			Severity:    models.SeverityCritical,
			Message:     "Test Violation",
			Description: "This is a test violation",
			File:        "test.go",
			Line:        10,
			Column:      1,
			Rule:        "function-length",
			Suggestion:  "Break down this function",
			CodeSnippet: "func testFunction() {\n    // test code\n}",
		},
		{
			ID:          "test-2",
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityMedium,
			Message:     "Poor naming convention",
			Description: "Variable name is too short",
			File:        "test.go",
			Line:        5,
			Column:      5,
			Rule:        "naming-convention",
			Suggestion:  "Use descriptive variable names",
			CodeSnippet: "var x int",
		},
	}

	// Create test file info
	fileInfo := &models.FileInfo{
		Path:         "test.go",
		Name:         "test.go",
		Extension:    ".go",
		Size:         1024,
		Lines:        50,
		ModifiedTime: time.Now(),
		Language:     "Go",
		Scanned:      true,
	}

	// Create test metrics
	metrics := &models.FileMetrics{
		TotalLines:      50,
		CodeLines:       40,
		CommentLines:    5,
		BlankLines:      5,
		FunctionCount:   3,
		ClassCount:      0,
		ComplexityScore: 8,
	}

	// Create test scan result
	scanResult := &models.ScanResult{
		File:       fileInfo,
		Violations: violations,
		Metrics:    metrics,
	}

	// Create test summary
	summary := &models.ScanSummary{
		TotalFiles:      1,
		ScannedFiles:    1,
		SkippedFiles:    0,
		TotalViolations: 2,
		ViolationsByType: map[string]int{
			string(models.ViolationTypeFunctionLength): 1,
			string(models.ViolationTypeNaming):         1,
		},
		StartTime: time.Now().Add(-time.Minute),
		EndTime:   time.Now(),
		Duration:  time.Minute,
	}

	// Create test config
	reportConfig := &models.ReportConfig{
		Paths:     []string{"./test"},
		FileTypes: []string{".go"},
		Thresholds: &models.Thresholds{
			FunctionLines:        25,
			CyclomaticComplexity: 8,
			Parameters:           4,
			NestingDepth:         3,
			ClassLines:           150,
		},
		HTMLSettings: &models.HTMLOptions{
			AutoRefresh:     true,
			RefreshInterval: 10,
			Theme:           "auto",
		},
	}

	// Create and return test report
	return models.NewReport(summary, []*models.ScanResult{scanResult}, reportConfig)
}

func TestHTMLReporter_InteractiveFeatures(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "interactive-report.html")

	config := &HTMLConfig{
		OutputPath:      outputPath,
		AutoRefresh:     true,
		RefreshInterval: 10,
		Theme:           "light",
	}

	reporter, err := NewHTMLReporter(config)
	if err != nil {
		t.Fatalf("Failed to create HTML reporter: %v", err)
	}

	report := createTestReport()

	err = reporter.Generate(report)
	if err != nil {
		t.Fatalf("Failed to generate HTML report: %v", err)
	}

	// Read and verify content contains interactive elements
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated HTML file: %v", err)
	}

	htmlContent := string(content)

	// Verify filter and search controls are present
	interactiveElements := []string{
		"id=\"searchInput\"",
		"id=\"severityFilter\"",
		"id=\"typeFilter\"",
		"id=\"sortBy\"",
		"id=\"expandAll\"",
		"id=\"collapseAll\"",
		"id=\"clearSearch\"",
		"id=\"visibleFilesCount\"",
		"id=\"visibleViolationsCount\"",
	}

	for _, element := range interactiveElements {
		if !strings.Contains(htmlContent, element) {
			t.Errorf("Generated HTML missing interactive element: %s", element)
		}
	}

	// Verify data attributes for filtering are present
	dataAttributes := []string{
		"data-file-path=",
		"data-file-name=",
		"data-violation-count=",
		"data-search-content=",
		"data-severity=",
		"data-type=",
		"data-line=",
		"file-item",
		"violation-card",
	}

	for _, attribute := range dataAttributes {
		if !strings.Contains(htmlContent, attribute) {
			t.Errorf("Generated HTML missing data attribute: %s", attribute)
		}
	}

	// Verify CSS classes for filtering are present
	cssClasses := []string{
		".filtered-hidden",
		".search-highlight",
		".no-visible-violations",
		".filter-active",
		".sort-indicator",
	}

	for _, cssClass := range cssClasses {
		if !strings.Contains(htmlContent, cssClass) {
			t.Errorf("Generated HTML missing CSS class: %s", cssClass)
		}
	}

	// Verify JavaScript functionality is present
	jsFeatures := []string{
		"class ViolationFilter",
		"applyFilters()",
		"applySorting()",
		"expandAll()",
		"collapseAll()",
		"highlightSearchTerms",
		"updateCounts",
		"addEventListener",
	}

	for _, jsFeature := range jsFeatures {
		if !strings.Contains(htmlContent, jsFeature) {
			t.Errorf("Generated HTML missing JavaScript feature: %s", jsFeature)
		}
	}
}

func TestTemplateFunctions(t *testing.T) {
	funcs := getTemplateFunctions()

	// Test new 'lower' function
	lowerFunc, exists := funcs["lower"]
	if !exists {
		t.Error("Template function 'lower' not found")
	} else {
		result := lowerFunc.(func(string) string)("TEST STRING")
		if result != "test string" {
			t.Errorf("Lower function failed: expected 'test string', got '%s'", result)
		}
	}

	// Test existing functions still work
	testCases := []struct {
		name     string
		function string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "add function",
			function: "add",
			input:    []int{5, 3},
			expected: 8,
		},
		{
			name:     "percentage function",
			function: "percentage",
			input:    []int{25, 100},
			expected: 25.0,
		},
		{
			name:     "basename function",
			function: "basename",
			input:    "/path/to/file.go",
			expected: "file.go",
		},
		{
			name:     "replace function",
			function: "replace",
			input:    []string{"hello world", "world", "golang"},
			expected: "hello golang",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn, exists := funcs[tc.function]
			if !exists {
				t.Errorf("Template function '%s' not found", tc.function)
				return
			}

			var result interface{}
			switch tc.function {
			case "add":
				inputs := tc.input.([]int)
				result = fn.(func(int, int) int)(inputs[0], inputs[1])
			case "percentage":
				inputs := tc.input.([]int)
				result = fn.(func(int, int) float64)(inputs[0], inputs[1])
			case "basename":
				result = fn.(func(string) string)(tc.input.(string))
			case "replace":
				inputs := tc.input.([]string)
				result = fn.(func(string, string, string) string)(inputs[0], inputs[1], inputs[2])
			}

			if result != tc.expected {
				t.Errorf("Function '%s' failed: expected %v, got %v", tc.function, tc.expected, result)
			}
		})
	}
}