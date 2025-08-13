package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ericfisherdev/goclean/internal/config"
	"github.com/ericfisherdev/goclean/internal/reporters"
	"github.com/ericfisherdev/goclean/internal/scanner"
	"github.com/ericfisherdev/goclean/internal/testutils"
)

// TestSimpleEndToEndWorkflow tests the basic scanning and reporting workflow
func TestSimpleEndToEndWorkflow(t *testing.T) {
	// Create test directory
	tempDir := testutils.CreateTempDir(t)
	reportsDir := filepath.Join(tempDir, "reports")
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		t.Fatalf("Failed to create reports directory: %v", err)
	}
	
	// Create test files with violations
	testFiles := map[string]string{
		"good.go": `package main

// SimpleFunction is a clean, simple function
func SimpleFunction() {
	println("Hello, World!")
}
`,
		"bad.go": `package main

// VeryLongFunctionWithViolations has multiple issues
func VeryLongFunctionWithViolations(a, b, c, d, e string) {
	if a != "" {
		if b != "" {
			if c != "" {
				if d != "" {
					println("Deep nesting")
				}
			}
		}
	}
	println("Line 1")
	println("Line 2")
	println("Line 3")
	println("Line 4")
	println("Line 5")
	println("Line 6")
	println("Line 7")
	println("Line 8")
	println("Line 9")
	println("Line 10")
	println("Line 11")
	println("Line 12")
	println("Line 13")
	println("Line 14")
	println("Line 15")
	println("Line 16")
	println("Line 17")
	println("Line 18")
	println("Line 19")
	println("Line 20")
}
`,
	}
	
	// Write test files
	for filename, content := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file %s: %v", filename, err)
		}
	}
	
	// Test 1: Scanner functionality
	t.Run("Scanner", func(t *testing.T) {
		// Create scanner
		eng := scanner.NewEngine([]string{tempDir}, []string{}, []string{".go"}, false)
		if eng == nil {
			t.Fatal("Failed to create scanner engine")
		}
		
		// Perform scan
		summary, results, err := eng.Scan()
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}
		
		// Verify results
		if summary.ScannedFiles != 2 {
			t.Errorf("Expected 2 files scanned, got %d", summary.ScannedFiles)
		}
		
		if summary.TotalViolations == 0 {
			t.Error("Expected violations to be found")
		}
		
		if len(results) == 0 {
			t.Error("Expected scan results")
		}
		
		t.Logf("Scanned %d files, found %d violations", summary.ScannedFiles, summary.TotalViolations)
	})
	
	// Test 2: Configuration loading
	t.Run("Configuration", func(t *testing.T) {
		// Create configuration file
		configContent := `
scan:
  paths: ["` + tempDir + `"]
  file_types: [".go"]
thresholds:
  function_lines: 15
  parameters: 3
output:
  html:
    path: "` + filepath.Join(reportsDir, "test.html") + `"
  markdown:
    path: "` + filepath.Join(reportsDir, "test.md") + `"
`
		configPath := filepath.Join(tempDir, "goclean.yaml")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}
		
		// Load configuration  
		cfg, err := config.Load(configPath)
		if err != nil {
			t.Fatalf("Failed to load configuration: %v", err)
		}
		
		// Verify configuration
		if len(cfg.Scan.Paths) == 0 {
			t.Error("Expected scan paths to be set")
		}
		
		if cfg.Thresholds.FunctionLines != 15 {
			t.Errorf("Expected function lines threshold 15, got %d", cfg.Thresholds.FunctionLines)
		}
		
		t.Logf("Configuration loaded successfully with %d scan paths", len(cfg.Scan.Paths))
	})
	
	// Test 3: Report generation
	t.Run("Reports", func(t *testing.T) {
		// Create configuration for reports
		cfg := &config.Config{
			Output: config.OutputConfig{
				HTML: config.HTMLConfig{
					Path: filepath.Join(reportsDir, "integration.html"),
				},
				Markdown: config.MarkdownConfig{
					Enabled: true,
					Path:    filepath.Join(reportsDir, "integration.md"),
				},
			},
		}
		
		// Create scanner and get results
		eng := scanner.NewEngine([]string{tempDir}, []string{}, []string{".go"}, false)
		summary, results, err := eng.Scan()
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}
		
		// Create report manager
		reportManager, err := reporters.NewManager(cfg)
		if err != nil {
			t.Fatalf("Failed to create report manager: %v", err)
		}
		
		// Generate reports using the actual API
		err = reportManager.GenerateReports(summary, results)
		if err != nil {
			t.Logf("Report generation failed: %v", err)
		}
		
		// Verify reports were created
		htmlPath := cfg.Output.HTML.Path
		if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
			t.Errorf("HTML report was not created at %s", htmlPath)
		}
		
		markdownPath := cfg.Output.Markdown.Path
		if _, err := os.Stat(markdownPath); os.IsNotExist(err) {
			t.Errorf("Markdown report was not created at %s", markdownPath)
		}
		
		// Verify report content
		if htmlContent, err := os.ReadFile(htmlPath); err == nil {
			if len(htmlContent) < 100 {
				t.Error("HTML report appears to be too small")
			}
			t.Logf("HTML report generated with %d bytes", len(htmlContent))
		}
		
		if markdownContent, err := os.ReadFile(markdownPath); err == nil {
			if len(markdownContent) < 50 {
				t.Error("Markdown report appears to be too small")
			}
			t.Logf("Markdown report generated with %d bytes", len(markdownContent))
		}
	})
}

// TestMultiLanguageScanning tests scanning different file types
func TestMultiLanguageScanning(t *testing.T) {
	tempDir := testutils.CreateTempDir(t)
	
	// Create files for different languages
	testFiles := map[string]string{
		"test.go": `package main

func LongGoFunction() {
	println("Line 1")
	println("Line 2")
	println("Line 3")
	println("Line 4")
	println("Line 5")
	println("Line 6")
	println("Line 7")
	println("Line 8")
	println("Line 9")
	println("Line 10")
	println("Line 11")
	println("Line 12")
	println("Line 13")
	println("Line 14")
	println("Line 15")
	println("Line 16")
	println("Line 17")
	println("Line 18")
	println("Line 19")
	println("Line 20")
	println("Line 21")
	println("Line 22")
	println("Line 23")
	println("Line 24")
	println("Line 25")
}
`,
		"test.js": `function longJavaScriptFunction() {
    console.log("Line 1");
    console.log("Line 2");
    console.log("Line 3");
    console.log("Line 4");
    console.log("Line 5");
    console.log("Line 6");
    console.log("Line 7");
    console.log("Line 8");
    console.log("Line 9");
    console.log("Line 10");
    console.log("Line 11");
    console.log("Line 12");
    console.log("Line 13");
    console.log("Line 14");
    console.log("Line 15");
    console.log("Line 16");
    console.log("Line 17");
    console.log("Line 18");
    console.log("Line 19");
    console.log("Line 20");
    console.log("Line 21");
    console.log("Line 22");
    console.log("Line 23");
    console.log("Line 24");
    console.log("Line 25");
}
`,
		"test.py": `def long_python_function():
    print("Line 1")
    print("Line 2")
    print("Line 3")
    print("Line 4")
    print("Line 5")
    print("Line 6")
    print("Line 7")
    print("Line 8")
    print("Line 9")
    print("Line 10")
    print("Line 11")
    print("Line 12")
    print("Line 13")
    print("Line 14")
    print("Line 15")
    print("Line 16")
    print("Line 17")
    print("Line 18")
    print("Line 19")
    print("Line 20")
    print("Line 21")
    print("Line 22")
    print("Line 23")
    print("Line 24")
    print("Line 25")
`,
	}
	
	// Write test files
	for filename, content := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file %s: %v", filename, err)
		}
	}
	
	// Test scanning all languages
	eng := scanner.NewEngine([]string{tempDir}, []string{}, []string{".go", ".js", ".py"}, false)
	summary, results, err := eng.Scan()
	if err != nil {
		t.Fatalf("Multi-language scan failed: %v", err)
	}
	
	// Verify all files were scanned
	if summary.ScannedFiles != 3 {
		t.Errorf("Expected 3 files scanned, got %d", summary.ScannedFiles)
	}
	
	// Check that violations were found (long functions should be detected)
	if summary.TotalViolations == 0 {
		t.Error("Expected violations to be found in long functions")
	}
	
	// Verify we have results for each file
	if len(results) != 3 {
		t.Errorf("Expected 3 scan results, got %d", len(results))
	}
	
	t.Logf("Multi-language scan completed: %d files, %d violations", 
		summary.ScannedFiles, summary.TotalViolations)
}

// TestErrorHandling tests how the system handles various error conditions
func TestErrorHandling(t *testing.T) {
	// Test scanning non-existent directory
	t.Run("NonExistentDirectory", func(t *testing.T) {
		eng := scanner.NewEngine([]string{"/nonexistent/path"}, []string{}, []string{".go"}, false)
		summary, results, err := eng.Scan()
		
		// Should not panic, but should handle the error gracefully
		if err != nil {
			t.Logf("Expected error for non-existent directory: %v", err)
		}
		
		// Should not crash
		if summary != nil && summary.ScannedFiles > 0 {
			t.Error("Should not have scanned any files from non-existent directory")
		}
		
		if results != nil && len(results) > 0 {
			t.Error("Should not have results from non-existent directory")
		}
	})
	
	// Test scanning empty directory
	t.Run("EmptyDirectory", func(t *testing.T) {
		tempDir := testutils.CreateTempDir(t)
		
		eng := scanner.NewEngine([]string{tempDir}, []string{}, []string{".go"}, false)
		summary, results, err := eng.Scan()
		
		// Should succeed with no files
		if err != nil {
			t.Errorf("Unexpected error scanning empty directory: %v", err)
		}
		
		if summary.ScannedFiles != 0 {
			t.Errorf("Expected 0 files scanned in empty directory, got %d", summary.ScannedFiles)
		}
		
		if len(results) != 0 {
			t.Errorf("Expected 0 results in empty directory, got %d", len(results))
		}
		
		t.Log("Empty directory handled correctly")
	})
}