package reporters

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ericfisherdev/goclean/internal/config"
	"github.com/ericfisherdev/goclean/internal/models"
)

func TestNewManager(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Output: config.OutputConfig{
			HTML: config.HTMLConfig{
				Path:            "./test-report.html",
				AutoRefresh:     true,
				RefreshInterval: 10,
				Theme:           "auto",
			},
			Markdown: config.MarkdownConfig{
				Enabled:         true,
				Path:            "./test-report.md",
				IncludeExamples: true,
			},
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if manager.config != cfg {
		t.Error("Manager config not set correctly")
	}

	// Should have 2 reporters (HTML and Markdown)
	if len(manager.reporters) != 2 {
		t.Errorf("Expected 2 reporters, got %d", len(manager.reporters))
	}

	// Check configured reporters
	configuredTypes := manager.GetConfiguredReporters()
	if len(configuredTypes) != 2 {
		t.Errorf("Expected 2 configured reporter types, got %d", len(configuredTypes))
	}

	expectedTypes := map[string]bool{"HTML": false, "Markdown": false}
	for _, reporterType := range configuredTypes {
		if _, exists := expectedTypes[reporterType]; exists {
			expectedTypes[reporterType] = true
		}
	}

	for reporterType, found := range expectedTypes {
		if !found {
			t.Errorf("Expected reporter type %s not found", reporterType)
		}
	}
}

func TestNewManager_HTMLOnly(t *testing.T) {
	cfg := &config.Config{
		Output: config.OutputConfig{
			HTML: config.HTMLConfig{
				Path:            "./test-report.html",
				AutoRefresh:     false,
				RefreshInterval: 0,
				Theme:           "light",
			},
			Markdown: config.MarkdownConfig{
				Enabled: false, // Markdown disabled
			},
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Should have 1 reporter (HTML only)
	if len(manager.reporters) != 1 {
		t.Errorf("Expected 1 reporter, got %d", len(manager.reporters))
	}

	configuredTypes := manager.GetConfiguredReporters()
	if len(configuredTypes) != 1 || configuredTypes[0] != "HTML" {
		t.Errorf("Expected only HTML reporter, got %v", configuredTypes)
	}
}

func TestNewManager_NoReporters(t *testing.T) {
	cfg := &config.Config{
		Output: config.OutputConfig{
			HTML: config.HTMLConfig{
				Path: "", // No HTML output
			},
			Markdown: config.MarkdownConfig{
				Enabled: false, // No Markdown output
			},
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Should have 0 reporters
	if len(manager.reporters) != 0 {
		t.Errorf("Expected 0 reporters, got %d", len(manager.reporters))
	}

	configuredTypes := manager.GetConfiguredReporters()
	if len(configuredTypes) != 0 {
		t.Errorf("Expected no configured reporters, got %v", configuredTypes)
	}
}

func TestManager_GenerateReports(t *testing.T) {
	// Create temporary directory for test outputs
	tempDir := t.TempDir()
	htmlPath := filepath.Join(tempDir, "test-report.html")
	markdownPath := filepath.Join(tempDir, "test-report.md")

	cfg := &config.Config{
		Scan: config.ScanConfig{
			Paths:     []string{"./test"},
			FileTypes: []string{".go"},
		},
		Thresholds: config.Thresholds{
			FunctionLines:        25,
			CyclomaticComplexity: 8,
			Parameters:           4,
			NestingDepth:         3,
			ClassLines:           150,
		},
		Output: config.OutputConfig{
			HTML: config.HTMLConfig{
				Path:            htmlPath,
				AutoRefresh:     true,
				RefreshInterval: 30,
				Theme:           "auto",
			},
			Markdown: config.MarkdownConfig{
				Enabled:         true,
				Path:            markdownPath,
				IncludeExamples: true,
			},
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create test data
	summary, results := createTestScanData()

	// Generate reports
	err = manager.GenerateReports(summary, results)
	if err != nil {
		t.Fatalf("Failed to generate reports: %v", err)
	}

	// Verify HTML file was created
	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		t.Error("HTML report file was not created")
	}

	// Verify Markdown file was created
	if _, err := os.Stat(markdownPath); os.IsNotExist(err) {
		t.Error("Markdown report file was not created")
	}

	// Verify HTML content
	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read HTML report: %v", err)
	}

	if !strings.Contains(string(htmlContent), "GoClean Code Analysis Report") {
		t.Error("HTML report missing expected content")
	}

	// Verify Markdown content
	markdownContent, err := os.ReadFile(markdownPath)
	if err != nil {
		t.Fatalf("Failed to read Markdown report: %v", err)
	}

	if !strings.Contains(string(markdownContent), "# GoClean Code Analysis Report") {
		t.Error("Markdown report missing expected content")
	}
}

func TestManager_GenerateConsoleReport(t *testing.T) {
	cfg := &config.Config{
		Scan: config.ScanConfig{
			Paths:     []string{"./test"},
			FileTypes: []string{".go"},
		},
		Thresholds: config.Thresholds{
			FunctionLines:        25,
			CyclomaticComplexity: 8,
			Parameters:           4,
			NestingDepth:         3,
			ClassLines:           150,
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create test data
	summary, results := createTestScanData()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = manager.GenerateConsoleReport(summary, results, false, false)
	if err != nil {
		t.Fatalf("Failed to generate console report: %v", err)
	}

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout

	// Read the output (even though we don't use it, to avoid pipe issues)
	_, _ = io.ReadAll(r)

	// The test passes if no error occurred - detailed console output testing is in console_test.go
}

func TestManager_GetOutputPaths(t *testing.T) {
	cfg := &config.Config{
		Output: config.OutputConfig{
			HTML: config.HTMLConfig{
				Path: "./test-report.html",
			},
			Markdown: config.MarkdownConfig{
				Enabled: true,
				Path:    "./test-report.md",
			},
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	htmlPath := manager.GetHTMLOutputPath()
	if htmlPath != "./test-report.html" {
		t.Errorf("Expected HTML path './test-report.html', got '%s'", htmlPath)
	}

	markdownPath := manager.GetMarkdownOutputPath()
	if markdownPath != "./test-report.md" {
		t.Errorf("Expected Markdown path './test-report.md', got '%s'", markdownPath)
	}
}

func TestManager_GetOutputPaths_Disabled(t *testing.T) {
	cfg := &config.Config{
		Output: config.OutputConfig{
			HTML: config.HTMLConfig{
				Path: "", // No HTML output
			},
			Markdown: config.MarkdownConfig{
				Enabled: false, // Markdown disabled
				Path:    "./test-report.md",
			},
		},
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	htmlPath := manager.GetHTMLOutputPath()
	if htmlPath != "" {
		t.Errorf("Expected empty HTML path, got '%s'", htmlPath)
	}

	markdownPath := manager.GetMarkdownOutputPath()
	if markdownPath != "" {
		t.Errorf("Expected empty Markdown path when disabled, got '%s'", markdownPath)
	}
}

func createTestScanData() (*models.ScanSummary, []*models.ScanResult) {
	// Create test violations
	violations := []*models.Violation{
		{
			ID:          "test-1",
			Type:        models.ViolationTypeFunctionLength,
			Severity:    models.SeverityHigh,
			Message:     "Function too long",
			Description: "This function exceeds the recommended length",
			File:        "test.go",
			Line:        10,
			Column:      1,
			Rule:        "function-length",
			Suggestion:  "Break this function into smaller functions",
		},
	}

	// Create test file info
	fileInfo := &models.FileInfo{
		Path:      "test.go",
		Name:      "test.go",
		Extension: ".go",
		Size:      1024,
		Lines:     100,
		Language:  "Go",
		Scanned:   true,
	}

	// Create test metrics
	metrics := &models.FileMetrics{
		TotalLines:      100,
		CodeLines:       80,
		CommentLines:    10,
		BlankLines:      10,
		FunctionCount:   5,
		ClassCount:      0,
		ComplexityScore: 15,
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
		TotalViolations: 1,
		ViolationsByType: map[string]int{
			string(models.ViolationTypeFunctionLength): 1,
		},
	}

	return summary, []*models.ScanResult{scanResult}
}