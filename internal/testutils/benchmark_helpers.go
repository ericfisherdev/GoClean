package testutils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/ericfisherdev/goclean/internal/models"
)

// BenchmarkHelper provides utilities for benchmarking operations
type BenchmarkHelper struct {
	TempDir string
}

// NewBenchmarkHelper creates a new benchmark helper
func NewBenchmarkHelper(b *testing.B) *BenchmarkHelper {
	b.Helper()
	return &BenchmarkHelper{
		TempDir: b.TempDir(),
	}
}

// CreateBenchmarkFiles creates multiple test files for benchmarking
func (h *BenchmarkHelper) CreateBenchmarkFiles(b *testing.B, count int, linesPerFile int) []string {
	b.Helper()
	var files []string
	
	for i := 0; i < count; i++ {
		content := h.generateFileContent(linesPerFile)
		fileName := fmt.Sprintf("bench_file_%d.go", i)
		filePath := filepath.Join(h.TempDir, fileName)
		
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			b.Fatalf("Failed to create benchmark file: %v", err)
		}
		
		files = append(files, filePath)
	}
	
	return files
}

// generateFileContent generates Go code content with specified number of lines
func (h *BenchmarkHelper) generateFileContent(lines int) string {
	content := "package benchmark\n\nimport \"fmt\"\n\n"
	
	// Create a function with many lines
	content += "func BenchmarkFunction() {\n"
	for i := 0; i < lines-6; i++ {
		content += fmt.Sprintf("\tfmt.Println(\"Line %d\")\n", i+1)
	}
	content += "}\n"
	
	return content
}

// CreateLargeTestFile creates a single large file for testing
func (h *BenchmarkHelper) CreateLargeTestFile(b *testing.B, lines int) string {
	b.Helper()
	content := h.generateFileContent(lines)
	filePath := filepath.Join(h.TempDir, "large_file.go")
	
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		b.Fatalf("Failed to create large test file: %v", err)
	}
	
	return filePath
}

// TimeOperation measures the time taken for an operation
func TimeOperation(operation func()) time.Duration {
	start := time.Now()
	operation()
	return time.Since(start)
}

// CreateViolationBatch creates a batch of violations for benchmarking
func CreateViolationBatch(count int) []models.Violation {
	violations := make([]models.Violation, count)
	
	violationTypes := []models.ViolationType{
		models.ViolationTypeFunctionLength,
		models.ViolationTypeParameterCount,
		models.ViolationTypeCyclomaticComplexity,
		models.ViolationTypeNestingDepth,
		models.ViolationTypeNaming,
	}
	
	severities := []models.Severity{
		models.SeverityLow,
		models.SeverityMedium,
		models.SeverityHigh,
		models.SeverityCritical,
	}
	
	for i := 0; i < count; i++ {
		violations[i] = models.Violation{
			ID:       fmt.Sprintf("violation-%d", i),
			Type:     violationTypes[i%len(violationTypes)],
			Message:  fmt.Sprintf("Violation message %d", i),
			Severity: severities[i%len(severities)],
			File:     fmt.Sprintf("file_%d.go", i%10),
			Line:     i%100 + 1,
			Column:   1,
			Rule:     fmt.Sprintf("rule-%d", i%5),
		}
	}
	
	return violations
}

// MeasureMemoryUsage provides a simple way to measure memory before and after an operation
type MemoryUsage struct {
	Before uint64
	After  uint64
}

// StartMemoryMeasurement begins memory usage measurement
func StartMemoryMeasurement() *MemoryUsage {
	return &MemoryUsage{
		Before: getCurrentMemoryUsage(),
	}
}

// StopMemoryMeasurement completes memory usage measurement
func (m *MemoryUsage) StopMemoryMeasurement() {
	m.After = getCurrentMemoryUsage()
}

// Delta returns the memory usage difference
func (m *MemoryUsage) Delta() int64 {
	return int64(m.After) - int64(m.Before)
}

// getCurrentMemoryUsage gets current memory usage using runtime.ReadMemStats
func getCurrentMemoryUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// CreateRustBenchmarkFiles creates multiple Rust test files for benchmarking
func (h *BenchmarkHelper) CreateRustBenchmarkFiles(b *testing.B, count int, content string) []*models.FileInfo {
	b.Helper()
	var files []*models.FileInfo
	
	for i := 0; i < count; i++ {
		fileName := fmt.Sprintf("bench_file_%d.rs", i)
		filePath := filepath.Join(h.TempDir, fileName)
		
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			b.Fatalf("Failed to create Rust benchmark file: %v", err)
		}
		
		fileInfo := &models.FileInfo{
			Path:     filePath,
			Name:     fileName,
			Language: "Rust",
			Size:     int64(len(content)),
		}
		
		files = append(files, fileInfo)
	}
	
	return files
}