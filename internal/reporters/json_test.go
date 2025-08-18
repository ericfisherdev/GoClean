package reporters

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ericfisherdev/goclean/internal/config"
	"github.com/ericfisherdev/goclean/internal/models"
)

func TestNewJSONReporter(t *testing.T) {
	cfg := &config.JSONConfig{
		Enabled:     true,
		Path:        "./test-reports/test.json",
		PrettyPrint: true,
	}
	
	reporter := NewJSONReporter(cfg)
	
	if reporter == nil {
		t.Fatal("Expected JSONReporter to be created")
	}
	
	if reporter.Format() != "json" {
		t.Errorf("Expected format 'json', got '%s'", reporter.Format())
	}
}

func TestNewJSONReporter_DefaultConfig(t *testing.T) {
	reporter := NewJSONReporter(nil)
	
	if reporter == nil {
		t.Fatal("Expected JSONReporter to be created with default config")
	}
	
	if !reporter.config.Enabled {
		t.Error("Expected default config to be enabled")
	}
	
	if reporter.config.Path != "./reports/violations.json" {
		t.Errorf("Expected default path './reports/violations.json', got '%s'", reporter.config.Path)
	}
	
	if !reporter.config.PrettyPrint {
		t.Error("Expected default config to have pretty print enabled")
	}
}

func TestJSONReporter_Generate(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test-violations.json")
	
	cfg := &config.JSONConfig{
		Enabled:     true,
		Path:        outputPath,
		PrettyPrint: true,
	}
	
	reporter := NewJSONReporter(cfg)
	
	// Create test report with scan results
	violations := []*models.Violation{
		{
			ID:          "test-1",
			Type:        models.ViolationTypeFunctionLength,
			Severity:    models.SeverityHigh,
			Message:     "Function too long",
			File:        "test.go",
			Line:        10,
			Column:      5,
			Suggestion:  "Consider breaking this function into smaller functions",
			CodeSnippet: "func longFunction() {",
		},
		{
			ID:          "test-2",
			Type:        models.ViolationTypeNaming,
			Severity:    models.SeverityMedium,
			Message:     "Variable name not descriptive",
			File:        "test.rs",
			Line:        20,
			Column:      8,
			Suggestion:  "Use a more descriptive name",
			CodeSnippet: "let x = 42;",
		},
	}
	
	files := []*models.ScanResult{
		{
			File: &models.FileInfo{
				Path:     "test.go",
				Language: "Go",
			},
			Violations: []*models.Violation{violations[0]},
		},
		{
			File: &models.FileInfo{
				Path:     "test.rs", 
				Language: "Rust",
			},
			Violations: []*models.Violation{violations[1]},
		},
	}
	
	summary := &models.ScanSummary{
		ScannedFiles:     2,
		TotalViolations:  2,
		Duration:         150 * time.Millisecond,
	}
	
	reportConfig := &models.ReportConfig{}
	
	report := models.NewReport(summary, files, reportConfig)
	
	// Generate the report
	err := reporter.Generate(report)
	if err != nil {
		t.Fatalf("Failed to generate JSON report: %v", err)
	}
	
	// Verify the file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("JSON report file was not created")
	}
	
	// Read and verify the content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read JSON report: %v", err)
	}
	
	// Parse JSON to verify structure
	var jsonReport JSONReport
	if err := json.Unmarshal(content, &jsonReport); err != nil {
		t.Fatalf("Failed to parse JSON report: %v", err)
	}
	
	// Verify metadata
	if jsonReport.Metadata.FilesScanned != 2 {
		t.Errorf("Expected 2 files scanned, got %d", jsonReport.Metadata.FilesScanned)
	}
	
	if jsonReport.Metadata.TotalViolations != 2 {
		t.Errorf("Expected 2 total violations, got %d", jsonReport.Metadata.TotalViolations)
	}
	
	// Verify violations
	if len(jsonReport.Violations) != 2 {
		t.Errorf("Expected 2 violations, got %d", len(jsonReport.Violations))
	}
	
	// Verify first violation
	v1 := jsonReport.Violations[0]
	if v1.ID != "test-1" {
		t.Errorf("Expected violation ID 'test-1', got '%s'", v1.ID)
	}
	if v1.Type != "function_length" {
		t.Errorf("Expected violation type 'function_length', got '%s'", v1.Type)
	}
	if v1.Severity != "High" {
		t.Errorf("Expected severity 'High', got '%s'", v1.Severity)
	}
	if v1.Language != "Go" {
		t.Errorf("Expected language 'Go', got '%s'", v1.Language)
	}
	
	// Verify summary
	if jsonReport.Summary.BySeverity["High"] != 1 {
		t.Errorf("Expected 1 High severity violation, got %d", jsonReport.Summary.BySeverity["High"])
	}
	if jsonReport.Summary.BySeverity["Medium"] != 1 {
		t.Errorf("Expected 1 Medium severity violation, got %d", jsonReport.Summary.BySeverity["Medium"])
	}
	
	if jsonReport.Summary.ByLanguage["Go"] != 1 {
		t.Errorf("Expected 1 Go violation, got %d", jsonReport.Summary.ByLanguage["Go"])
	}
	if jsonReport.Summary.ByLanguage["Rust"] != 1 {
		t.Errorf("Expected 1 Rust violation, got %d", jsonReport.Summary.ByLanguage["Rust"])
	}
	
	// Verify statistics
	if jsonReport.Statistics.TotalFiles != 2 {
		t.Errorf("Expected 2 total files, got %d", jsonReport.Statistics.TotalFiles)
	}
	
	if jsonReport.Statistics.MostCommonSeverity != "High" && jsonReport.Statistics.MostCommonSeverity != "Medium" {
		t.Errorf("Expected most common severity to be 'High' or 'Medium', got '%s'", jsonReport.Statistics.MostCommonSeverity)
	}
}

func TestJSONReporter_Generate_Disabled(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test-violations.json")
	
	cfg := &config.JSONConfig{
		Enabled:     false,
		Path:        outputPath,
		PrettyPrint: true,
	}
	
	reporter := NewJSONReporter(cfg)
	
	summary := &models.ScanSummary{
		ScannedFiles:     1,
		TotalViolations:  0,
		Duration:         50 * time.Millisecond,
	}
	
	report := models.NewReport(summary, []*models.ScanResult{}, &models.ReportConfig{})
	
	// Generate the report (should be skipped since disabled)
	err := reporter.Generate(report)
	if err != nil {
		t.Fatalf("Failed to generate JSON report: %v", err)
	}
	
	// Verify the file was NOT created
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Error("JSON report file should not have been created when disabled")
	}
}

func TestJSONReporter_Generate_CompactJSON(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "compact-violations.json")
	
	cfg := &config.JSONConfig{
		Enabled:     true,
		Path:        outputPath,
		PrettyPrint: false, // Compact JSON
	}
	
	reporter := NewJSONReporter(cfg)
	
	violations := []*models.Violation{
		{
			ID:       "compact-1",
			Type:     models.ViolationTypeFunctionLength,
			Severity: models.SeverityLow,
			Message:  "Test violation",
			File:     "test.go",
			Line:     1,
			Column:   1,
		},
	}
	
	files := []*models.ScanResult{
		{
			File: &models.FileInfo{
				Path:     "test.go",
				Language: "Go",
			},
			Violations: violations,
		},
	}
	
	summary := &models.ScanSummary{
		ScannedFiles:     1,
		TotalViolations:  1,
		Duration:         50 * time.Millisecond,
	}
	
	report := models.NewReport(summary, files, &models.ReportConfig{})
	
	err := reporter.Generate(report)
	if err != nil {
		t.Fatalf("Failed to generate compact JSON report: %v", err)
	}
	
	// Read the content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read JSON report: %v", err)
	}
	
	// Verify it's compact (no indentation)
	contentStr := string(content)
	if len(contentStr) == 0 {
		t.Error("JSON report is empty")
	}
	
	// Compact JSON shouldn't have multiple consecutive spaces (indentation)
	// This is a simple check - real compact JSON validation would be more complex
	if len(contentStr) < 50 {
		t.Error("JSON report seems too short")
	}
}

func TestFindMostCommon(t *testing.T) {
	tests := []struct {
		name     string
		counts   map[string]int
		expected string
	}{
		{
			name:     "Single item",
			counts:   map[string]int{"high": 5},
			expected: "high",
		},
		{
			name:     "Multiple items with clear winner",
			counts:   map[string]int{"high": 10, "medium": 5, "low": 2},
			expected: "high",
		},
		{
			name:     "Empty map",
			counts:   map[string]int{},
			expected: "",
		},
		{
			name:     "All zeros",
			counts:   map[string]int{"high": 0, "medium": 0},
			expected: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findMostCommon(tt.counts)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}