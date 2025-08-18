package reporters

import (
	"fmt"
	"io"
	"testing"

	"github.com/ericfisherdev/goclean/internal/config"
	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/testutils"
)

// BenchmarkHTMLReporting tests HTML report generation performance
func BenchmarkHTMLReporting(b *testing.B) {
	violationCounts := []int{100, 500, 1000, 5000}
	
	for _, count := range violationCounts {
		b.Run(fmt.Sprintf("Violations_%d", count), func(b *testing.B) {
			benchmarkHTMLReporter(b, count)
		})
	}
}

func benchmarkHTMLReporter(b *testing.B, violationCount int) {
	violations := testutils.CreateViolationBatch(violationCount)
	
	// Convert to proper violation pointers
	violationPtrs := make([]*models.Violation, len(violations))
	for i := range violations {
		violationPtrs[i] = &violations[i]
	}
	
	// Create scan results
	scanResults := []*models.ScanResult{
		{
			File: &models.FileInfo{
				Path:     "test/file.go",
				Name:     "file.go",
				Language: "go",
				Lines:    1000,
			},
			Violations: violationPtrs,
			Metrics: &models.FileMetrics{
				TotalLines: 1000,
				CodeLines:  800,
			},
		},
	}
	
	summary := &models.ScanSummary{
		TotalFiles:      1,
		TotalViolations: len(violations),
		ScannedFiles:    1,
	}
	
	config := &HTMLConfig{
		OutputPath:      b.TempDir() + "/report.html",
		AutoRefresh:     false,
		RefreshInterval: 10,
		Theme:           "light",
	}
	
	reporter, err := NewHTMLReporter(config)
	if err != nil {
		b.Fatalf("Failed to create HTML reporter: %v", err)
	}
	
	// Create full report
	report := models.NewReport(summary, scanResults, &models.ReportConfig{
		Paths:     []string{"test/"},
		FileTypes: []string{".go"},
	})
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		err := reporter.Generate(report)
		if err != nil {
			b.Fatalf("HTML report generation failed: %v", err)
		}
	}
}

// BenchmarkMarkdownReporting tests Markdown report generation performance  
func BenchmarkMarkdownReporting(b *testing.B) {
	violationCounts := []int{100, 500, 1000, 5000}
	
	for _, count := range violationCounts {
		b.Run(fmt.Sprintf("Violations_%d", count), func(b *testing.B) {
			benchmarkMarkdownReporter(b, count)
		})
	}
}

func benchmarkMarkdownReporter(b *testing.B, violationCount int) {
	violations := testutils.CreateViolationBatch(violationCount)
	
	// Convert to proper violation pointers
	violationPtrs := make([]*models.Violation, len(violations))
	for i := range violations {
		violationPtrs[i] = &violations[i]
	}
	
	scanResults := []*models.ScanResult{
		{
			File: &models.FileInfo{
				Path:     "test/file.go",
				Name:     "file.go",
				Language: "go",
				Lines:    1000,
			},
			Violations: violationPtrs,
			Metrics: &models.FileMetrics{
				TotalLines: 1000,
				CodeLines:  800,
			},
		},
	}
	
	summary := &models.ScanSummary{
		TotalFiles:      1,
		TotalViolations: len(violations),
		ScannedFiles:    1,
	}
	
	config := &MarkdownConfig{
		OutputPath:      b.TempDir() + "/report.md",
		IncludeExamples: true,
	}
	
	reporter := NewMarkdownReporter(config)
	
	// Create full report
	report := models.NewReport(summary, scanResults, &models.ReportConfig{
		Paths:     []string{"test/"},
		FileTypes: []string{".go"},
	})
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		err := reporter.Generate(report)
		if err != nil {
			b.Fatalf("Markdown report generation failed: %v", err)
		}
	}
}

// BenchmarkConsoleReporting tests console output performance
func BenchmarkConsoleReporting(b *testing.B) {
	violations := testutils.CreateViolationBatch(1000)
	
	// Convert to proper violation pointers
	violationPtrs := make([]*models.Violation, len(violations))
	for i := range violations {
		violationPtrs[i] = &violations[i]
	}
	
	scanResults := []*models.ScanResult{
		{
			File: &models.FileInfo{
				Path:     "test/file.go",
				Name:     "file.go",
				Language: "go",
				Lines:    1000,
			},
			Violations: violationPtrs,
			Metrics: &models.FileMetrics{
				TotalLines: 1000,
				CodeLines:  800,
			},
		},
	}
	
	summary := &models.ScanSummary{
		TotalFiles:      1,
		TotalViolations: len(violations),
		ScannedFiles:    1,
	}
	
	// Create console config with equivalent settings and output to io.Discard
	consoleConfig := &config.ConsoleConfig{
		Verbose: false,
		Colored: true,
		Output:  io.Discard,
	}
	reporter := NewConsoleReporter(consoleConfig)
	
	// Create full report
	report := models.NewReport(summary, scanResults, &models.ReportConfig{
		Paths:     []string{"test/"},
		FileTypes: []string{".go"},
	})
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		err := reporter.Generate(report)
		if err != nil {
			b.Fatalf("Console report generation failed: %v", err)
		}
	}
}

// BenchmarkReportManager tests report manager performance with multiple formats
func BenchmarkReportManager(b *testing.B) {
	violations := testutils.CreateViolationBatch(500)
	
	// Convert to proper violation pointers
	violationPtrs := make([]*models.Violation, len(violations))
	for i := range violations {
		violationPtrs[i] = &violations[i]
	}
	
	scanResults := []*models.ScanResult{
		{
			File: &models.FileInfo{
				Path:     "test/file.go",
				Name:     "file.go",
				Language: "go",
				Lines:    1000,
			},
			Violations: violationPtrs,
			Metrics: &models.FileMetrics{
				TotalLines: 1000,
				CodeLines:  800,
			},
		},
	}
	
	summary := &models.ScanSummary{
		TotalFiles:      1,
		TotalViolations: len(violations),
		ScannedFiles:    1,
	}
	
	tempDir := b.TempDir()
	
	// Create individual reporters
	htmlReporter, err := NewHTMLReporter(&HTMLConfig{
		OutputPath:      tempDir + "/report.html",
		AutoRefresh:     false,
		RefreshInterval: 10,
		Theme:           "light",
	})
	if err != nil {
		b.Fatalf("Failed to create HTML reporter: %v", err)
	}
	
	markdownReporter := NewMarkdownReporter(&MarkdownConfig{
		OutputPath:      tempDir + "/report.md",
		IncludeExamples: false,
	})
	
	consoleReporter := NewConsoleReporter(&config.ConsoleConfig{
		Verbose: false,
		Colored: false,
		Output:  io.Discard, // Suppress output during benchmarks
	})
	
	// Create full report
	report := models.NewReport(summary, scanResults, &models.ReportConfig{
		Paths:     []string{"test/"},
		FileTypes: []string{".go"},
	})
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// Run all reporters
		err := htmlReporter.Generate(report)
		if err != nil {
			b.Fatalf("HTML reporter failed: %v", err)
		}
		
		err = markdownReporter.Generate(report)
		if err != nil {
			b.Fatalf("Markdown reporter failed: %v", err)
		}
		
		err = consoleReporter.Generate(report)
		if err != nil {
			b.Fatalf("Console reporter failed: %v", err)
		}
	}
}

// BenchmarkTemplateRendering tests HTML template rendering performance
func BenchmarkTemplateRendering(b *testing.B) {
	violations := testutils.CreateViolationBatch(1000)
	
	// Convert to proper violation pointers
	violationPtrs := make([]*models.Violation, len(violations))
	for i := range violations {
		violationPtrs[i] = &violations[i]
	}
	
	scanResults := []*models.ScanResult{
		{
			File: &models.FileInfo{
				Path:     "test/file.go",
				Name:     "file.go",
				Language: "go",
				Lines:    1000,
			},
			Violations: violationPtrs,
			Metrics: &models.FileMetrics{
				TotalLines: 1000,
				CodeLines:  800,
			},
		},
	}
	
	summary := &models.ScanSummary{
		TotalFiles:      1,
		TotalViolations: len(violations),
		ScannedFiles:    1,
	}
	
	config := &HTMLConfig{
		OutputPath:      b.TempDir() + "/template_test.html",
		AutoRefresh:     false,
		RefreshInterval: 10,
		Theme:           "light",
	}
	
	reporter, err := NewHTMLReporter(config)
	if err != nil {
		b.Fatalf("Failed to create HTML reporter: %v", err)
	}
	
	// Create full report
	report := models.NewReport(summary, scanResults, &models.ReportConfig{
		Paths:     []string{"test/"},
		FileTypes: []string{".go"},
	})
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// Focus on template rendering by generating the report
		err := reporter.Generate(report)
		if err != nil {
			b.Fatalf("Template rendering failed: %v", err)
		}
	}
}