package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

// BenchmarkSuite runs comprehensive benchmarks for the entire application
func BenchmarkSuite(b *testing.B) {
	b.Run("Engine", func(b *testing.B) {
		runEngineBenchmarks(b)
	})
	
	b.Run("Violations", func(b *testing.B) {
		runViolationBenchmarks(b)
	})
	
	b.Run("Reporters", func(b *testing.B) {
		runReporterBenchmarks(b)
	})
}

func runEngineBenchmarks(b *testing.B) {
	cmd := exec.Command("go", "test", "-bench=BenchmarkEngine", "-benchmem", "./internal/scanner")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		b.Errorf("Engine benchmarks failed: %v\nOutput: %s", err, output)
	}
	b.Logf("Engine benchmark results:\n%s", output)
}

func runViolationBenchmarks(b *testing.B) {
	cmd := exec.Command("go", "test", "-bench=BenchmarkViolation", "-benchmem", "./internal/violations")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		b.Errorf("Violation benchmarks failed: %v\nOutput: %s", err, output)
	}
	b.Logf("Violation benchmark results:\n%s", output)
}

func runReporterBenchmarks(b *testing.B) {
	cmd := exec.Command("go", "test", "-bench=BenchmarkHTML|BenchmarkMarkdown|BenchmarkConsole|BenchmarkReport", "-benchmem", "./internal/reporters")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		b.Errorf("Reporter benchmarks failed: %v\nOutput: %s", err, output)
	}
	b.Logf("Reporter benchmark results:\n%s", output)
}

// BenchmarkOverallPerformance tests end-to-end performance targets
func BenchmarkOverallPerformance(b *testing.B) {
	b.Run("StartupTime", func(b *testing.B) {
		benchmarkStartupTime(b)
	})
	
	b.Run("ScanningSpeed", func(b *testing.B) {
		benchmarkScanningSpeed(b)
	})
	
	b.Run("MemoryUsage", func(b *testing.B) {
		benchmarkMemoryUsage(b)
	})
	
	b.Run("ReportGeneration", func(b *testing.B) {
		benchmarkReportGeneration(b)
	})
}

func benchmarkStartupTime(b *testing.B) {
	targetStartupTime := 500 * time.Millisecond
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		start := time.Now()
		
		cmd := exec.Command("./goclean", "--help")
		cmd.Dir = "."
		err := cmd.Run()
		if err != nil {
			b.Fatalf("Failed to run goclean: %v", err)
		}
		
		elapsed := time.Since(start)
		if elapsed > targetStartupTime {
			b.Logf("Warning: Startup time %v exceeds target %v", elapsed, targetStartupTime)
		}
		
		b.ReportMetric(float64(elapsed.Nanoseconds()), "ns/startup")
	}
}

func benchmarkScanningSpeed(b *testing.B) {
	// Create test files for scanning
	testDir := b.TempDir()
	fileCount := 100
	
	// Create test Go files
	for i := 0; i < fileCount; i++ {
		content := fmt.Sprintf(`package test%d

import "fmt"

func TestFunction%d() {
	fmt.Println("Test function %d")
}
`, i, i, i)
		
		filePath := fmt.Sprintf("%s/test%d.go", testDir, i)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
	}
	
	targetSpeed := 1000.0 // files per second
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		start := time.Now()
		
		cmd := exec.Command("./goclean", "scan", "--path", testDir, "--output-format", "console")
		cmd.Dir = "."
		err := cmd.Run()
		if err != nil {
			b.Fatalf("Failed to run scan: %v", err)
		}
		
		elapsed := time.Since(start)
		filesPerSecond := float64(fileCount) / elapsed.Seconds()
		
		if filesPerSecond < targetSpeed {
			b.Logf("Warning: Scanning speed %.2f files/sec is below target %.2f files/sec", 
				filesPerSecond, targetSpeed)
		}
		
		b.ReportMetric(filesPerSecond, "files/sec")
	}
}

func benchmarkMemoryUsage(b *testing.B) {
	// Create a larger test directory
	testDir := b.TempDir()
	fileCount := 1000
	
	for i := 0; i < fileCount; i++ {
		content := generateLargeGoFile(i, 50)
		filePath := fmt.Sprintf("%s/large%d.go", testDir, i)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		cmd := exec.Command("./goclean", "scan", "--path", testDir, "--output-format", "console")
		cmd.Dir = "."
		err := cmd.Run()
		if err != nil {
			b.Fatalf("Failed to run scan: %v", err)
		}
	}
}

func benchmarkReportGeneration(b *testing.B) {
	targetReportTime := 2 * time.Second
	
	testDir := b.TempDir()
	fileCount := 500
	
	for i := 0; i < fileCount; i++ {
		content := generateLargeGoFile(i, 100)
		filePath := fmt.Sprintf("%s/report%d.go", testDir, i)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		start := time.Now()
		
		outputPath := fmt.Sprintf("%s/report-%d.html", b.TempDir(), i)
		cmd := exec.Command("./goclean", "scan", "--path", testDir, 
			"--output-format", "html", "--output-path", outputPath)
		cmd.Dir = "."
		err := cmd.Run()
		if err != nil {
			b.Fatalf("Failed to generate report: %v", err)
		}
		
		elapsed := time.Since(start)
		
		if elapsed > targetReportTime {
			b.Logf("Warning: Report generation time %v exceeds target %v", 
				elapsed, targetReportTime)
		}
		
		b.ReportMetric(float64(elapsed.Nanoseconds()), "ns/report")
	}
}

func generateLargeGoFile(index, lines int) string {
	content := fmt.Sprintf("package large%d\n\nimport \"fmt\"\n\n", index)
	
	content += fmt.Sprintf("func LargeFunction%d() {\n", index)
	for i := 0; i < lines; i++ {
		content += fmt.Sprintf("\tfmt.Println(\"Line %d in function %d\")\n", i, index)
	}
	content += "}\n\n"
	
	// Add some violations
	content += "func x() {}\n" // Short name violation
	content += "func func_with_underscores() {}\n" // Naming convention violation
	content += "func parameterHeavy(a, b, c, d, e, f, g int) {}\n" // Too many parameters
	
	return content
}