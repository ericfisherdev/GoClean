package testutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ericfisherdev/goclean/internal/models"
)

func TestCreateTempDir(t *testing.T) {
	dir := CreateTempDir(t)
	
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Expected temp directory to be created, but it doesn't exist")
	}
}

func TestCreateTestFile(t *testing.T) {
	dir := CreateTempDir(t)
	content := "package main\n\nfunc main() {}\n"
	
	filePath := CreateTestFile(t, dir, "test.go", content)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Expected test file to be created, but it doesn't exist")
	}
	
	// Check file content
	readContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	
	if string(readContent) != content {
		t.Errorf("Expected file content %s, got %s", content, string(readContent))
	}
}

func TestCreateTestFileWithSubdirectory(t *testing.T) {
	dir := CreateTempDir(t)
	content := "test content"
	
	filePath := CreateTestFile(t, dir, "subdir/test.txt", content)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Expected test file in subdirectory to be created, but it doesn't exist")
	}
	
	// Check if subdirectory was created
	subDir := filepath.Dir(filePath)
	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		t.Errorf("Expected subdirectory to be created, but it doesn't exist")
	}
}

func TestCreateTestConfig(t *testing.T) {
	cfg := CreateTestConfig()
	
	// Check basic settings
	if len(cfg.Scan.Paths) == 0 {
		t.Error("Expected scan paths to be set")
	}
	
	if cfg.Thresholds.FunctionLines != 10 {
		t.Errorf("Expected function lines threshold 10, got %d", cfg.Thresholds.FunctionLines)
	}
	
	if cfg.Thresholds.CyclomaticComplexity != 5 {
		t.Errorf("Expected cyclomatic complexity threshold 5, got %d", cfg.Thresholds.CyclomaticComplexity)
	}
	
	if cfg.Output.HTML.Path != "./test-report.html" {
		t.Errorf("Expected HTML path './test-report.html', got %s", cfg.Output.HTML.Path)
	}
}

func TestAssertViolationCount(t *testing.T) {
	violations := []models.Violation{
		CreateSampleViolation(models.ViolationTypeFunctionLength, models.SeverityHigh, "file1.go", 1),
		CreateSampleViolation(models.ViolationTypeNaming, models.SeverityMedium, "file2.go", 2),
	}
	
	// This should pass
	AssertViolationCount(t, violations, 2)
}

func TestAssertViolationType(t *testing.T) {
	violation := CreateSampleViolation(models.ViolationTypeFunctionLength, models.SeverityHigh, "test.go", 1)
	
	// This should pass
	AssertViolationType(t, violation, models.ViolationTypeFunctionLength)
}

func TestAssertViolationSeverity(t *testing.T) {
	violation := CreateSampleViolation(models.ViolationTypeFunctionLength, models.SeverityHigh, "test.go", 1)
	
	// This should pass
	AssertViolationSeverity(t, violation, models.SeverityHigh)
}

func TestCreateSampleViolation(t *testing.T) {
	violationType := models.ViolationTypeFunctionLength
	severity := models.SeverityHigh
	file := "test.go"
	line := 42
	
	violation := CreateSampleViolation(violationType, severity, file, line)
	
	if violation.Type != violationType {
		t.Errorf("Expected violation type %s, got %s", violationType, violation.Type)
	}
	
	if violation.Severity != severity {
		t.Errorf("Expected violation severity %s, got %s", severity, violation.Severity)
	}
	
	if violation.File != file {
		t.Errorf("Expected violation file %s, got %s", file, violation.File)
	}
	
	if violation.Line != line {
		t.Errorf("Expected violation line %d, got %d", line, violation.Line)
	}
	
	if violation.Message == "" {
		t.Error("Expected violation message to be set")
	}
}

func TestMockProgressCallback(t *testing.T) {
	callback := MockProgressCallback()
	
	// Should not panic when called
	callback(1, 10)
	callback(5, 10)
	callback(10, 10)
}

func TestFileExists(t *testing.T) {
	dir := CreateTempDir(t)
	filePath := CreateTestFile(t, dir, "exists.txt", "content")
	
	// Existing file
	if !FileExists(t, filePath) {
		t.Error("Expected FileExists to return true for existing file")
	}
	
	// Non-existing file
	nonExistentPath := filepath.Join(dir, "does_not_exist.txt")
	if FileExists(t, nonExistentPath) {
		t.Error("Expected FileExists to return false for non-existing file")
	}
}

func TestAssertFileExists(t *testing.T) {
	dir := CreateTempDir(t)
	filePath := CreateTestFile(t, dir, "exists.txt", "content")
	
	// This should pass (no panic/error)
	AssertFileExists(t, filePath)
}

func TestAssertFileNotExists(t *testing.T) {
	dir := CreateTempDir(t)
	nonExistentPath := filepath.Join(dir, "does_not_exist.txt")
	
	// This should pass (no panic/error)
	AssertFileNotExists(t, nonExistentPath)
}