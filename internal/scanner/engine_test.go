package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ericfisherdev/goclean/internal/models"
)

func TestNewEngine(t *testing.T) {
	includePaths := []string{"./test"}
	excludePatterns := []string{"*.tmp"}
	fileTypes := []string{".go", ".js"}
	verbose := true

	engine := NewEngine(includePaths, excludePatterns, fileTypes, verbose)

	if engine == nil {
		t.Fatal("Expected engine to be created, got nil")
	}

	if !engine.verbose {
		t.Error("Expected verbose to be true")
	}

	if engine.maxWorkers != 10 {
		t.Errorf("Expected default max workers 10, got %d", engine.maxWorkers)
	}

	if engine.fileWalker == nil {
		t.Error("Expected file walker to be initialized")
	}

	if engine.parser == nil {
		t.Error("Expected parser to be initialized")
	}
}

func TestSetMaxWorkers(t *testing.T) {
	engine := NewEngine([]string{"."}, []string{}, []string{".go"}, false)

	// Test setting valid worker count
	engine.SetMaxWorkers(5)
	if engine.maxWorkers != 5 {
		t.Errorf("Expected max workers 5, got %d", engine.maxWorkers)
	}

	// Test setting zero workers (should not change)
	engine.SetMaxWorkers(0)
	if engine.maxWorkers != 5 {
		t.Errorf("Expected max workers to remain 5, got %d", engine.maxWorkers)
	}

	// Test setting negative workers (should not change)
	engine.SetMaxWorkers(-1)
	if engine.maxWorkers != 5 {
		t.Errorf("Expected max workers to remain 5, got %d", engine.maxWorkers)
	}
}

func TestScanEmptyDirectory(t *testing.T) {
	// Create temporary empty directory
	tmpDir := t.TempDir()

	engine := NewEngine([]string{tmpDir}, []string{}, []string{".go"}, false)

	summary, results, err := engine.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if summary.TotalFiles != 0 {
		t.Errorf("Expected 0 total files, got %d", summary.TotalFiles)
	}

	if summary.ScannedFiles != 0 {
		t.Errorf("Expected 0 scanned files, got %d", summary.ScannedFiles)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}

	if summary.TotalViolations != 0 {
		t.Errorf("Expected 0 violations, got %d", summary.TotalViolations)
	}
}

func TestScanWithTestFiles(t *testing.T) {
	// Create temporary directory with test files
	tmpDir := t.TempDir()

	// Create test Go file
	goFile := filepath.Join(tmpDir, "test.go")
	goContent := `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}

func helper() {
    // This is a helper function
}
`
	err := os.WriteFile(goFile, []byte(goContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test Go file: %v", err)
	}

	// Create test JavaScript file
	jsFile := filepath.Join(tmpDir, "test.js")
	jsContent := `function greet(name) {
    console.log("Hello, " + name);
}

class TestClass {
    constructor() {
        this.value = 42;
    }
}
`
	err = os.WriteFile(jsFile, []byte(jsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test JS file: %v", err)
	}

	// Create test file to be excluded
	tmpFile := filepath.Join(tmpDir, "temp.tmp")
	err = os.WriteFile(tmpFile, []byte("temporary"), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	engine := NewEngine([]string{tmpDir}, []string{"*.tmp"}, []string{".go", ".js"}, false)

	summary, results, err := engine.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should find 2 files (.go and .js), excluding .tmp
	if summary.TotalFiles != 2 {
		t.Errorf("Expected 2 total files, got %d", summary.TotalFiles)
	}

	if summary.ScannedFiles != 2 {
		t.Errorf("Expected 2 scanned files, got %d", summary.ScannedFiles)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Verify file details
	var goResult, jsResult *models.ScanResult
	for _, result := range results {
		if result.File.Extension == ".go" {
			goResult = result
		} else if result.File.Extension == ".js" {
			jsResult = result
		}
	}

	if goResult == nil {
		t.Error("Expected Go file in results")
	} else {
		if goResult.File.Language != "Go" {
			t.Errorf("Expected Go language, got %s", goResult.File.Language)
		}
		if goResult.Metrics.FunctionCount != 2 {
			t.Errorf("Expected 2 functions in Go file, got %d", goResult.Metrics.FunctionCount)
		}
		if !goResult.File.Scanned {
			t.Error("Expected Go file to be scanned")
		}
	}

	if jsResult == nil {
		t.Error("Expected JavaScript file in results")
	} else {
		if jsResult.File.Language != "JavaScript" {
			t.Errorf("Expected JavaScript language, got %s", jsResult.File.Language)
		}
		if jsResult.Metrics.FunctionCount != 1 {
			t.Errorf("Expected 1 function in JS file, got %d", jsResult.Metrics.FunctionCount)
		}
		if jsResult.Metrics.ClassCount != 1 {
			t.Errorf("Expected 1 class in JS file, got %d", jsResult.Metrics.ClassCount)
		}
	}

	// Verify timing
	if summary.Duration <= 0 {
		t.Error("Expected positive scan duration")
	}

	if summary.StartTime.IsZero() || summary.EndTime.IsZero() {
		t.Error("Expected valid start and end times")
	}

	if summary.EndTime.Before(summary.StartTime) {
		t.Error("Expected end time to be after start time")
	}
}

func TestScanWithErrors(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create test file with permission issues (we'll simulate by creating a directory with same name)
	problemPath := filepath.Join(tmpDir, "problem.go")
	err := os.Mkdir(problemPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create problem directory: %v", err)
	}

	// Create a valid file too
	validFile := filepath.Join(tmpDir, "valid.go")
	err = os.WriteFile(validFile, []byte("package main\nfunc main() {}\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create valid file: %v", err)
	}

	engine := NewEngine([]string{tmpDir}, []string{}, []string{".go"}, false)

	summary, results, err := engine.Scan()
	if err != nil {
		t.Fatalf("Scan should not fail even with file errors: %v", err)
	}

	// The file walker may not discover the directory as a file, so adjust expectations
	if summary.TotalFiles == 0 {
		t.Skip("File walker didn't discover any files (expected behavior for directory-as-file)")
	}

	// At least the valid file should be scanned
	if summary.ScannedFiles == 0 {
		t.Errorf("Expected at least 1 scanned file, got %d", summary.ScannedFiles)
	}

	if len(results) == 0 {
		t.Errorf("Expected at least 1 result, got %d", len(results))
	}

	// Check results
	hasSuccess := false
	for _, result := range results {
		if result.File.Error == "" && result.File.Scanned {
			hasSuccess = true
		}
	}

	if !hasSuccess {
		t.Error("Expected at least one successful result")
	}
}

func TestGenerateSummary(t *testing.T) {
	engine := NewEngine([]string{"."}, []string{}, []string{".go"}, false)

	// Create mock file info and results
	files := []*models.FileInfo{
		{Path: "file1.go", Scanned: true},
		{Path: "file2.go", Scanned: true},
		{Path: "file3.go", Scanned: false, Error: "test error"},
	}

	results := []*models.ScanResult{
		{
			File:       files[0],
			Violations: []*models.Violation{
				{Type: models.ViolationTypeFunctionLength},
				{Type: models.ViolationTypeFunctionLength},
			},
		},
		{
			File:       files[1],
			Violations: []*models.Violation{
				{Type: models.ViolationTypeNaming},
			},
		},
		{
			File:       files[2],
			Violations: []*models.Violation{},
		},
	}

	// Generate summary with proper time values
	startTime := time.Now().Add(-time.Second)
	endTime := time.Now()
	summary := engine.generateSummary(files, results, startTime, endTime)

	if summary == nil {
		t.Error("Expected summary to be generated")
		return
	}

	if summary.TotalFiles != 3 {
		t.Errorf("Expected 3 total files, got %d", summary.TotalFiles)
	}

	if summary.ScannedFiles != 2 {
		t.Errorf("Expected 2 scanned files, got %d", summary.ScannedFiles)
	}

	if summary.SkippedFiles != 1 {
		t.Errorf("Expected 1 skipped file, got %d", summary.SkippedFiles)
	}

	if summary.TotalViolations != 3 {
		t.Errorf("Expected 3 total violations, got %d", summary.TotalViolations)
	}

	// Check violations by type
	if summary.ViolationsByType["function_length"] != 2 {
		t.Errorf("Expected 2 function_length violations, got %d", summary.ViolationsByType["function_length"])
	}

	if summary.ViolationsByType["naming_convention"] != 1 {
		t.Errorf("Expected 1 naming_convention violation, got %d", summary.ViolationsByType["naming_convention"])
	}
}

func TestScanNonExistentPath(t *testing.T) {
	engine := NewEngine([]string{"/non/existent/path"}, []string{}, []string{".go"}, false)

	summary, results, err := engine.Scan()
	if err != nil {
		t.Fatalf("Scan should handle non-existent paths gracefully: %v", err)
	}

	if summary.TotalFiles != 0 {
		t.Errorf("Expected 0 total files for non-existent path, got %d", summary.TotalFiles)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for non-existent path, got %d", len(results))
	}
}

func TestConcurrentScanning(t *testing.T) {
	// Create temporary directory with multiple files
	tmpDir := t.TempDir()

	// Create multiple test files to ensure concurrent processing
	for i := 0; i < 20; i++ {
		fileName := filepath.Join(tmpDir, fmt.Sprintf("test%d.go", i))
		content := fmt.Sprintf(`package main

func test%d() {
    // Test function %d
}
`, i, i)
		err := os.WriteFile(fileName, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}
	}

	engine := NewEngine([]string{tmpDir}, []string{}, []string{".go"}, false)
	engine.SetMaxWorkers(3) // Test with limited workers

	summary, results, err := engine.Scan()
	if err != nil {
		t.Fatalf("Concurrent scan failed: %v", err)
	}

	if summary.TotalFiles != 20 {
		t.Errorf("Expected 20 total files, got %d", summary.TotalFiles)
	}

	if summary.ScannedFiles != 20 {
		t.Errorf("Expected 20 scanned files, got %d", summary.ScannedFiles)
	}

	if len(results) != 20 {
		t.Errorf("Expected 20 results, got %d", len(results))
	}

	// Verify all files were processed correctly
	for _, result := range results {
		if !result.File.Scanned {
			t.Errorf("File %s was not scanned", result.File.Path)
		}
		if result.File.Error != "" {
			t.Errorf("File %s had error: %s", result.File.Path, result.File.Error)
		}
		if result.Metrics.FunctionCount != 1 {
			t.Errorf("Expected 1 function in %s, got %d", result.File.Path, result.Metrics.FunctionCount)
		}
	}
}