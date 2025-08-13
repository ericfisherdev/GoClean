package reporters

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
)

func TestNewConsoleReporter(t *testing.T) {
	reporter := NewConsoleReporter(true, true)

	if !reporter.verbose {
		t.Error("Verbose flag not set correctly")
	}

	if !reporter.colors {
		t.Error("Colors flag not set correctly")
	}

	// Test with different settings
	reporter2 := NewConsoleReporter(false, false)
	if reporter2.verbose {
		t.Error("Verbose flag should be false")
	}

	if reporter2.colors {
		t.Error("Colors flag should be false")
	}
}

func TestConsoleReporter_Generate(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	reporter := NewConsoleReporter(false, false) // No verbose, no colors for cleaner output
	report := createTestReport()

	err := reporter.Generate(report)
	if err != nil {
		t.Fatalf("Failed to generate console report: %v", err)
	}

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout

	output, _ := io.ReadAll(r)
	outputStr := string(output)

	// Verify essential elements are present
	expectedElements := []string{
		"GoClean Code Analysis Report",
		"SCAN SUMMARY",
		"Total Files:",
		"Total Violations:",
		"VIOLATIONS BY TYPE",
		"Long Functions:",
		"Naming Convention:",
		"TOP VIOLATED FILES",
		"test.go",
	}

	for _, element := range expectedElements {
		if !strings.Contains(outputStr, element) {
			t.Errorf("Console output missing expected element: %s", element)
		}
	}
}

func TestConsoleReporter_GenerateVerbose(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	reporter := NewConsoleReporter(true, false) // Verbose mode
	report := createTestReport()

	err := reporter.Generate(report)
	if err != nil {
		t.Fatalf("Failed to generate verbose console report: %v", err)
	}

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout

	output, _ := io.ReadAll(r)
	outputStr := string(output)

	// Verify verbose elements are present
	verboseElements := []string{
		"DETAILED VIOLATIONS",
		"Test Violation",
		"Poor naming convention",
		"Break down this function",
		"Use descriptive variable names",
		"Line 10",
		"Line 5",
	}

	for _, element := range verboseElements {
		if !strings.Contains(outputStr, element) {
			t.Errorf("Verbose console output missing expected element: %s", element)
		}
	}
}

func TestConsoleReporter_GenerateWithColors(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	reporter := NewConsoleReporter(false, true) // Colors enabled
	report := createTestReport()

	err := reporter.Generate(report)
	if err != nil {
		t.Fatalf("Failed to generate console report with colors: %v", err)
	}

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout

	output, _ := io.ReadAll(r)
	outputStr := string(output)

	// Check for ANSI color codes
	colorCodes := []string{
		"\033[1;36m", // Cyan bold (header)
		"\033[1;34m", // Blue bold (section)
		"\033[0m",    // Reset
	}

	hasColors := false
	for _, colorCode := range colorCodes {
		if strings.Contains(outputStr, colorCode) {
			hasColors = true
			break
		}
	}

	if !hasColors {
		t.Error("Console output should contain ANSI color codes when colors are enabled")
	}
}

func TestConsoleReporter_GenerateNoViolations(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	reporter := NewConsoleReporter(false, false)
	
	// Create report with no violations
	report := createTestReportNoViolations()

	err := reporter.Generate(report)
	if err != nil {
		t.Fatalf("Failed to generate console report: %v", err)
	}

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout

	output, _ := io.ReadAll(r)
	outputStr := string(output)

	// Verify success message is present
	successElements := []string{
		"No violations found!",
		"follows clean code principles",
	}

	for _, element := range successElements {
		if !strings.Contains(outputStr, element) {
			t.Errorf("Console output missing success element: %s", element)
		}
	}

	// Should not contain violation sections
	violationSections := []string{
		"VIOLATIONS BY TYPE",
		"VIOLATIONS BY SEVERITY",
		"TOP VIOLATED FILES",
	}

	for _, section := range violationSections {
		if strings.Contains(outputStr, section) {
			t.Errorf("Console output should not contain violation section when no violations: %s", section)
		}
	}
}

func TestConsoleReporter_Colorize(t *testing.T) {
	reporter := NewConsoleReporter(false, true) // Colors enabled

	// Test colorization
	colorized := reporter.colorize("test", "header")
	if !strings.Contains(colorized, "\033[1;36m") || !strings.Contains(colorized, "\033[0m") {
		t.Error("Colorize should add color codes when colors enabled")
	}

	// Test without colors
	reporterNoColor := NewConsoleReporter(false, false)
	plain := reporterNoColor.colorize("test", "header")
	if plain != "test" {
		t.Error("Colorize should return plain text when colors disabled")
	}
}

func TestConsoleReporter_SeverityIcons(t *testing.T) {
	reporter := NewConsoleReporter(false, false)

	icons := map[models.Severity]string{
		models.SeverityLow:      "ℹ️",
		models.SeverityMedium:   "⚠️",
		models.SeverityHigh:     "🚨",
		models.SeverityCritical: "💥",
	}

	for severity, expectedIcon := range icons {
		icon := reporter.getSeverityIcon(severity)
		if icon != expectedIcon {
			t.Errorf("Wrong icon for severity %s: expected %s, got %s", severity, expectedIcon, icon)
		}
	}
}

func createTestReportNoViolations() *models.Report {
	// Create test file info with no violations
	fileInfo := &models.FileInfo{
		Path:      "clean.go",
		Name:      "clean.go",
		Extension: ".go",
		Size:      512,
		Lines:     25,
		Language:  "Go",
		Scanned:   true,
	}

	metrics := &models.FileMetrics{
		TotalLines:      25,
		CodeLines:       20,
		CommentLines:    3,
		BlankLines:      2,
		FunctionCount:   2,
		ClassCount:      0,
		ComplexityScore: 3,
	}

	scanResult := &models.ScanResult{
		File:       fileInfo,
		Violations: []*models.Violation{}, // No violations
		Metrics:    metrics,
	}

	summary := &models.ScanSummary{
		TotalFiles:       1,
		ScannedFiles:     1,
		SkippedFiles:     0,
		TotalViolations:  0,
		ViolationsByType: map[string]int{},
	}

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
	}

	return models.NewReport(summary, []*models.ScanResult{scanResult}, reportConfig)
}