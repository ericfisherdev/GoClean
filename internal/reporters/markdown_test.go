package reporters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ericfisherdev/goclean/internal/models"
)

func TestNewMarkdownReporter(t *testing.T) {
	config := &MarkdownConfig{
		OutputPath:      "/tmp/test.md",
		IncludeExamples: true,
	}

	reporter := NewMarkdownReporter(config)
	if reporter == nil {
		t.Fatal("Expected reporter to be non-nil")
	}
	if reporter.config != config {
		t.Errorf("Expected config to be %v, got %v", config, reporter.config)
	}
}

func TestMarkdownReporter_Generate(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "report.md")

	config := &MarkdownConfig{
		OutputPath:      outputPath,
		IncludeExamples: true,
	}

	reporter := NewMarkdownReporter(config)

	report := createTestReport()

	err := reporter.Generate(report)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "# GoClean Code Analysis Report") {
		t.Error("Expected content to contain report header")
	}
	if !strings.Contains(contentStr, "## Executive Summary") {
		t.Error("Expected content to contain executive summary")
	}
	if !strings.Contains(contentStr, "## Violation Statistics") {
		t.Error("Expected content to contain violation statistics")
	}
	if !strings.Contains(contentStr, "## Detailed Violations") {
		t.Error("Expected content to contain detailed violations")
	}
	if !strings.Contains(contentStr, "## Recommendations") {
		t.Error("Expected content to contain recommendations")
	}
}

func TestMarkdownReporter_GenerateWithNoViolations(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "report.md")

	config := &MarkdownConfig{
		OutputPath:      outputPath,
		IncludeExamples: false,
	}

	reporter := NewMarkdownReporter(config)

	report := &models.Report{
		ID:          "test-report",
		GeneratedAt: time.Now(),
		Summary: &models.ScanSummary{
			TotalFiles:      10,
			ScannedFiles:    10,
			SkippedFiles:    0,
			TotalViolations: 0,
			Duration:        time.Second,
		},
		Statistics: &models.Statistics{
			ViolationsByType:     make(map[models.ViolationType]int),
			ViolationsBySeverity: make(map[models.Severity]int),
			TopViolatedFiles:     []*models.FileViolationSummary{},
		},
		Files: []*models.ScanResult{},
	}

	err := reporter.Generate(report)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "🎉 **Excellent!** No clean code violations were found") {
		t.Error("Expected content to contain no violations message")
	}
	if !strings.Contains(contentStr, "🎉 **Congratulations!** Your code demonstrates excellent adherence") {
		t.Error("Expected content to contain congratulations message")
	}
}

func TestMarkdownReporter_GenerateMarkdown(t *testing.T) {
	config := &MarkdownConfig{
		OutputPath:      "/tmp/test.md",
		IncludeExamples: true,
	}

	reporter := NewMarkdownReporter(config)
	report := createTestReport()

	content := reporter.generateMarkdown(report)

	expectedStrings := []string{
		"# GoClean Code Analysis Report",
		"## Executive Summary",
		"## Violation Statistics",
		"### Violations by Type",
		"### Violations by Severity",
		"## Most Violated Files",
		"## Detailed Violations",
		"## Recommendations",
		"## About This Report",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(content, expected) {
			t.Errorf("Expected content to contain %q", expected)
		}
	}
}

func TestMarkdownReporter_WriteHeader(t *testing.T) {
	config := &MarkdownConfig{
		OutputPath:      "/tmp/test.md",
		IncludeExamples: false,
	}

	reporter := NewMarkdownReporter(config)
	report := createMarkdownTestReport()

	var md strings.Builder
	reporter.writeHeader(&md, report)

	content := md.String()
	expectedStrings := []string{
		"# GoClean Code Analysis Report",
		"**Generated:**",
		"**Report ID:** test-report",
		"**Scanned Paths:**",
		"- `./src`",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(content, expected) {
			t.Errorf("Expected content to contain %q", expected)
		}
	}
}

func TestMarkdownReporter_WriteSummary(t *testing.T) {
	config := &MarkdownConfig{
		OutputPath:      "/tmp/test.md",
		IncludeExamples: false,
	}

	reporter := NewMarkdownReporter(config)

	t.Run("with violations", func(t *testing.T) {
		summary := &models.ScanSummary{
			TotalFiles:      100,
			ScannedFiles:    90,
			SkippedFiles:    10,
			TotalViolations: 25,
			Duration:        2 * time.Second,
		}

		var md strings.Builder
		reporter.writeSummary(&md, summary)

		content := md.String()
		expectedStrings := []string{
			"## Executive Summary",
			"📊 **Analysis Results:** 25 violations found across 90 files",
			"### Key Metrics",
			"| Total Files | 100 |",
			"| Total Violations | **25** |",
		}

		for _, expected := range expectedStrings {
			if !strings.Contains(content, expected) {
				t.Errorf("Expected content to contain %q", expected)
			}
		}
	})

	t.Run("without violations", func(t *testing.T) {
		summary := &models.ScanSummary{
			TotalFiles:      100,
			ScannedFiles:    100,
			SkippedFiles:    0,
			TotalViolations: 0,
			Duration:        1 * time.Second,
		}

		var md strings.Builder
		reporter.writeSummary(&md, summary)

		content := md.String()
		if !strings.Contains(content, "🎉 **Excellent!** No clean code violations were found") {
			t.Error("Expected content to contain no violations message")
		}
		if !strings.Contains(content, "Your code follows clean code principles") {
			t.Error("Expected content to contain clean code message")
		}
	})
}

func TestMarkdownReporter_WriteStatistics(t *testing.T) {
	config := &MarkdownConfig{
		OutputPath:      "/tmp/test.md",
		IncludeExamples: false,
	}

	reporter := NewMarkdownReporter(config)

	stats := &models.Statistics{
		ViolationsByType: map[models.ViolationType]int{
			models.ViolationTypeFunctionLength:      10,
			models.ViolationTypeCyclomaticComplexity: 5,
			models.ViolationTypeParameterCount:       3,
		},
		ViolationsBySeverity: map[models.Severity]int{
			models.SeverityCritical: 2,
			models.SeverityHigh:     5,
			models.SeverityMedium:   8,
			models.SeverityLow:      3,
		},
	}

	var md strings.Builder
	reporter.writeStatistics(&md, stats)

	content := md.String()
	expectedStrings := []string{
		"## Violation Statistics",
		"### Violations by Type",
		"| Violation Type | Count | Percentage |",
		"### Violations by Severity",
		"| Severity | Count | Status |",
		"Critical",
		"High",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(content, expected) {
			t.Errorf("Expected content to contain %q", expected)
		}
	}
}

func TestMarkdownReporter_WriteTopViolatedFiles(t *testing.T) {
	config := &MarkdownConfig{
		OutputPath:      "/tmp/test.md",
		IncludeExamples: false,
	}

	reporter := NewMarkdownReporter(config)

	topFiles := []*models.FileViolationSummary{
		{
			File:            "/src/main.go",
			TotalViolations: 10,
			Lines:           200,
			ViolationsBySeverity: map[models.Severity]int{
				models.SeverityHigh:   5,
				models.SeverityMedium: 5,
			},
		},
		{
			File:            "/src/utils.go",
			TotalViolations: 5,
			Lines:           100,
			ViolationsBySeverity: map[models.Severity]int{
				models.SeverityLow: 5,
			},
		},
	}

	var md strings.Builder
	reporter.writeTopViolatedFiles(&md, topFiles)

	content := md.String()
	expectedStrings := []string{
		"## Most Violated Files",
		"| File | Violations | Lines | Violations/100 Lines | Severity Distribution |",
		"`main.go`",
		"`utils.go`",
		"High:5",
		"Medium:5",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(content, expected) {
			t.Errorf("Expected content to contain %q", expected)
		}
	}
}

func TestMarkdownReporter_WriteDetailedViolations(t *testing.T) {
	config := &MarkdownConfig{
		OutputPath:      "/tmp/test.md",
		IncludeExamples: true,
	}

	reporter := NewMarkdownReporter(config)
	report := createMarkdownTestReport()

	var md strings.Builder
	reporter.writeDetailedViolations(&md, report)

	content := md.String()
	expectedStrings := []string{
		"## Detailed Violations",
		"### Per-File Breakdown",
		"📁 /test/file.go",
		"**2 violations found**",
		"**Line 10**",
		"💡 **Suggestion:**",
		"```go",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(content, expected) {
			t.Errorf("Expected content to contain %q", expected)
		}
	}
}

func TestMarkdownReporter_WriteRecommendations(t *testing.T) {
	config := &MarkdownConfig{
		OutputPath:      "/tmp/test.md",
		IncludeExamples: false,
	}

	reporter := NewMarkdownReporter(config)

	testCases := []struct {
		name             string
		violationType    models.ViolationType
		expectedContent  []string
	}{
		{
			name:          "function length violations",
			violationType: models.ViolationTypeFunctionLength,
			expectedContent: []string{
				"## Recommendations",
				"🔧 **Break down long functions**",
				"📏 **Aim for functions under 20-25 lines**",
			},
		},
		{
			name:          "cyclomatic complexity violations",
			violationType: models.ViolationTypeCyclomaticComplexity,
			expectedContent: []string{
				"🌀 **Reduce cyclomatic complexity**",
				"🏗️ **Use early returns**",
			},
		},
		{
			name:          "parameter count violations",
			violationType: models.ViolationTypeParameterCount,
			expectedContent: []string{
				"📦 **Group related parameters**",
				"🔧 **Apply builder pattern**",
			},
		},
		{
			name:          "naming violations",
			violationType: models.ViolationTypeNaming,
			expectedContent: []string{
				"📝 **Improve naming conventions**",
				"🚫 **Avoid abbreviations**",
			},
		},
		{
			name:          "missing documentation",
			violationType: models.ViolationTypeMissingDocumentation,
			expectedContent: []string{
				"📚 **Add documentation**",
				"💬 **Write meaningful comments**",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			report := &models.Report{
				Summary: &models.ScanSummary{TotalViolations: 10},
				Statistics: &models.Statistics{
					ViolationsByType: map[models.ViolationType]int{
						tc.violationType: 10,
					},
				},
			}

			var md strings.Builder
			reporter.writeRecommendations(&md, report)

			content := md.String()
			for _, expected := range tc.expectedContent {
				if !strings.Contains(content, expected) {
					t.Errorf("Expected content to contain %q", expected)
				}
			}
			if !strings.Contains(content, "🧪 **Write unit tests**") {
				t.Error("Expected content to contain unit test recommendation")
			}
			if !strings.Contains(content, "🔍 **Regular code reviews**") {
				t.Error("Expected content to contain code review recommendation")
			}
		})
	}
}

func TestMarkdownReporter_WriteFooter(t *testing.T) {
	config := &MarkdownConfig{
		OutputPath:      "/tmp/test.md",
		IncludeExamples: false,
	}

	reporter := NewMarkdownReporter(config)

	t.Run("no violations", func(t *testing.T) {
		report := &models.Report{
			GeneratedAt: time.Now(),
			Summary: &models.ScanSummary{
				TotalViolations: 0,
				Duration:        time.Second,
			},
			Config: &models.ReportConfig{
				Thresholds: &models.Thresholds{
					FunctionLines:        25,
					CyclomaticComplexity: 10,
					Parameters:           5,
					NestingDepth:         3,
					ClassLines:           200,
				},
			},
		}

		var md strings.Builder
		reporter.writeFooter(&md, report)

		content := md.String()
		if !strings.Contains(content, "## About This Report") {
			t.Error("Expected content to contain about section")
		}
		if !strings.Contains(content, "### Thresholds Used") {
			t.Error("Expected content to contain thresholds section")
		}
		if !strings.Contains(content, "| Function Lines | 25 |") {
			t.Error("Expected content to contain function lines threshold")
		}
		if !strings.Contains(content, "🎉 **Congratulations!** Your code demonstrates excellent adherence") {
			t.Error("Expected content to contain congratulations message")
		}
	})

	t.Run("few violations", func(t *testing.T) {
		report := &models.Report{
			GeneratedAt: time.Now(),
			Summary: &models.ScanSummary{
				TotalViolations: 5,
				Duration:        time.Second,
			},
		}

		var md strings.Builder
		reporter.writeFooter(&md, report)

		content := md.String()
		if !strings.Contains(content, "✨ **Good job!** You have relatively few violations") {
			t.Error("Expected content to contain good job message")
		}
	})

	t.Run("many violations", func(t *testing.T) {
		report := &models.Report{
			GeneratedAt: time.Now(),
			Summary: &models.ScanSummary{
				TotalViolations: 50,
				Duration:        time.Second,
			},
		}

		var md strings.Builder
		reporter.writeFooter(&md, report)

		content := md.String()
		if !strings.Contains(content, "💪 **Improvement opportunity!**") {
			t.Error("Expected content to contain improvement message")
		}
	})
}

func TestMarkdownReporter_GetSeverityStatus(t *testing.T) {
	config := &MarkdownConfig{
		OutputPath:      "/tmp/test.md",
		IncludeExamples: false,
	}

	reporter := NewMarkdownReporter(config)

	tests := []struct {
		severity models.Severity
		expected string
	}{
		{models.SeverityLow, "ℹ️ Info"},
		{models.SeverityMedium, "⚠️ Warning"},
		{models.SeverityHigh, "🚨 High Priority"},
		{models.SeverityCritical, "💥 Critical"},
		{models.Severity(99), "❓ Unknown"},
	}

	for _, test := range tests {
		result := reporter.getSeverityStatus(test.severity)
		if result != test.expected {
			t.Errorf("Expected %q for severity %v, got %q", test.expected, test.severity, result)
		}
	}
}

func TestMarkdownReporter_CreateOutputDirectory(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "nested", "dir", "report.md")

	config := &MarkdownConfig{
		OutputPath:      outputPath,
		IncludeExamples: false,
	}

	reporter := NewMarkdownReporter(config)
	report := createTestReport()

	err := reporter.Generate(report)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Expected file to exist at %s", outputPath)
	}
}

func TestMarkdownReporter_GenerateError(t *testing.T) {
	config := &MarkdownConfig{
		OutputPath:      "/invalid\x00path/report.md",
		IncludeExamples: false,
	}

	reporter := NewMarkdownReporter(config)
	report := createTestReport()

	err := reporter.Generate(report)
	if err == nil {
		t.Error("Expected an error, got nil")
	}
}

func createMarkdownTestReport() *models.Report {
	return &models.Report{
		ID:          "test-report",
		GeneratedAt: time.Now(),
		Summary: &models.ScanSummary{
			TotalFiles:      10,
			ScannedFiles:    9,
			SkippedFiles:    1,
			TotalViolations: 2,
			Duration:        time.Second,
		},
		Statistics: &models.Statistics{
			ViolationsByType: map[models.ViolationType]int{
				models.ViolationTypeFunctionLength: 1,
				models.ViolationTypeParameterCount: 1,
			},
			ViolationsBySeverity: map[models.Severity]int{
				models.SeverityMedium: 1,
				models.SeverityHigh:   1,
			},
			TopViolatedFiles: []*models.FileViolationSummary{
				{
					File:            "/test/file.go",
					TotalViolations: 2,
					Lines:           100,
					ViolationsBySeverity: map[models.Severity]int{
						models.SeverityMedium: 1,
						models.SeverityHigh:   1,
					},
				},
			},
		},
		Files: []*models.ScanResult{
			{
				File: &models.FileInfo{
					Path:  "/test/file.go",
					Lines: 100,
					Size:  1024,
				},
				Violations: []*models.Violation{
					{
						Type:        models.ViolationTypeFunctionLength,
						Severity:    models.SeverityMedium,
						Message:     "Function is too long",
						Description: "Function exceeds the maximum allowed lines",
						Suggestion:  "Break down the function into smaller ones",
						File:        "/test/file.go",
						Line:        10,
						Column:      1,
						CodeSnippet: "func longFunction() {\n  // lots of code\n}",
					},
					{
						Type:        models.ViolationTypeParameterCount,
						Severity:    models.SeverityHigh,
						Message:     "Too many parameters",
						Description: "Function has too many parameters",
						Suggestion:  "Use a struct to group parameters",
						File:        "/test/file.go",
						Line:        50,
						Column:      1,
						CodeSnippet: "func manyParams(a, b, c, d, e, f int) {}",
					},
				},
			},
		},
		Config: &models.ReportConfig{
			Paths:     []string{"./src"},
			FileTypes: []string{".go"},
			Thresholds: &models.Thresholds{
				FunctionLines:        25,
				CyclomaticComplexity: 10,
				Parameters:           5,
				NestingDepth:         3,
				ClassLines:           200,
			},
		},
	}
}