// Package testutils provides test utilities and helpers for GoClean testing.
package testutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ericfisherdev/goclean/internal/config"
	"github.com/ericfisherdev/goclean/internal/models"
)

// CreateTempDir creates a temporary directory for testing
func CreateTempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

// CreateTestFile creates a test file with the given content
func CreateTestFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	filePath := filepath.Join(dir, filename)
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return filePath
}

// CreateTestConfig creates a test configuration with common settings
func CreateTestConfig() *config.Config {
	cfg := config.GetDefaultConfig()
	cfg.Scan.Paths = []string{"./"}
	cfg.Thresholds.FunctionLines = 10
	cfg.Thresholds.CyclomaticComplexity = 5
	cfg.Thresholds.Parameters = 3
	cfg.Thresholds.NestingDepth = 2
	cfg.Output.HTML.Path = "./test-report.html"
	cfg.Output.Markdown.Enabled = true
	cfg.Output.Markdown.Path = "./test-report.md"
	return cfg
}

// AssertViolationCount checks if the number of violations matches expected
func AssertViolationCount(t *testing.T, violations []models.Violation, expected int) {
	t.Helper()
	if len(violations) != expected {
		t.Errorf("Expected %d violations, got %d", expected, len(violations))
	}
}

// AssertViolationType checks if a violation has the expected type
func AssertViolationType(t *testing.T, violation models.Violation, expectedType models.ViolationType) {
	t.Helper()
	if violation.Type != expectedType {
		t.Errorf("Expected violation type '%s', got '%s'", expectedType, violation.Type)
	}
}

// AssertViolationSeverity checks if a violation has the expected severity
func AssertViolationSeverity(t *testing.T, violation models.Violation, expectedSeverity models.Severity) {
	t.Helper()
	if violation.Severity != expectedSeverity {
		t.Errorf("Expected violation severity '%s', got '%s'", expectedSeverity, violation.Severity)
	}
}

// CreateSampleViolation creates a sample violation for testing
func CreateSampleViolation(violationType models.ViolationType, severity models.Severity, file string, line int) models.Violation {
	return models.Violation{
		ID:       "test-id",
		Type:     violationType,
		Message:  "Sample violation message",
		Severity: severity,
		File:     file,
		Line:     line,
		Column:   1,
		Rule:     "test-rule",
	}
}

// MockProgressCallback creates a mock progress callback for testing
func MockProgressCallback() func(current, total int) {
	return func(current, total int) {
		// Mock implementation - does nothing in tests
	}
}

// FileExists checks if a file exists
func FileExists(t *testing.T, path string) bool {
	t.Helper()
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// AssertFileExists checks if a file exists and fails the test if it doesn't
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	if !FileExists(t, path) {
		t.Errorf("Expected file %s to exist", path)
	}
}

// AssertFileNotExists checks if a file doesn't exist and fails the test if it does
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if FileExists(t, path) {
		t.Errorf("Expected file %s to not exist", path)
	}
}