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