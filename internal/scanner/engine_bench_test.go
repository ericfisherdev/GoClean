package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ericfisherdev/goclean/internal/testutils"
)

// BenchmarkEngineScanning tests scanning performance with various file counts
func BenchmarkEngineScanning(b *testing.B) {
	benchmarks := []struct {
		name      string
		fileCount int
		linesPerFile int
	}{
		{"Small_10files_50lines", 10, 50},
		{"Medium_100files_100lines", 100, 100},
		{"Large_500files_200lines", 500, 200},
		{"XLarge_1000files_300lines", 1000, 300},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			benchmarkEngineWithFiles(b, bm.fileCount, bm.linesPerFile)
		})
	}
}

func benchmarkEngineWithFiles(b *testing.B, fileCount, linesPerFile int) {
	helper := testutils.NewBenchmarkHelper(b)
	_ = helper.CreateBenchmarkFiles(b, fileCount, linesPerFile)
	
	// Create directory with test files
	testDir := helper.TempDir
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		engine := NewEngine([]string{testDir}, []string{}, []string{".go"}, false)
		_, _, err := engine.Scan()
		if err != nil {
			b.Fatalf("Scan failed: %v", err)
		}
	}
}

// BenchmarkEngineConcurrency tests performance with different worker counts
func BenchmarkEngineConcurrency(b *testing.B) {
	helper := testutils.NewBenchmarkHelper(b)
	_ = helper.CreateBenchmarkFiles(b, 200, 150)
	testDir := helper.TempDir
	
	workerCounts := []int{1, 2, 4, 8, 16}
	
	for _, workers := range workerCounts {
		b.Run(fmt.Sprintf("Workers_%d", workers), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			
			for i := 0; i < b.N; i++ {
				engine := NewEngine([]string{testDir}, []string{}, []string{".go"}, false)
				engine.SetMaxWorkers(workers)
				_, _, err := engine.Scan()
				if err != nil {
					b.Fatalf("Scan failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkEngineMemoryUsage measures memory usage during scanning
func BenchmarkEngineMemoryUsage(b *testing.B) {
	helper := testutils.NewBenchmarkHelper(b)
	_ = helper.CreateBenchmarkFiles(b, 1000, 200)
	testDir := helper.TempDir
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		memUsage := testutils.StartMemoryMeasurement()
		
		engine := NewEngine([]string{testDir}, []string{}, []string{".go"}, false)
		_, _, err := engine.Scan()
		if err != nil {
			b.Fatalf("Scan failed: %v", err)
		}
		
		memUsage.StopMemoryMeasurement()
		b.ReportMetric(float64(memUsage.Delta()), "bytes/op")
	}
}

// BenchmarkEngineStartupTime measures engine initialization time
func BenchmarkEngineStartupTime(b *testing.B) {
	includePaths := []string{"./testdata"}
	excludePatterns := []string{"*.test.go"}
	fileTypes := []string{".go"}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_ = NewEngine(includePaths, excludePatterns, fileTypes, false)
	}
}

// BenchmarkEngineFileWalking tests file discovery performance
func BenchmarkEngineFileWalking(b *testing.B) {
	helper := testutils.NewBenchmarkHelper(b)
	
	// Create nested directory structure
	baseDir := helper.TempDir
	for i := 0; i < 10; i++ {
		subDir := filepath.Join(baseDir, fmt.Sprintf("subdir%d", i))
		os.MkdirAll(subDir, 0755)
		
		// Create a new helper for each subdirectory
		subHelper := &testutils.BenchmarkHelper{TempDir: subDir}
		_ = subHelper.CreateBenchmarkFiles(b, 50, 100)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		engine := NewEngine([]string{baseDir}, []string{}, []string{".go"}, false)
		files, err := engine.fileWalker.Walk()
		if err != nil {
			b.Fatalf("File walking failed: %v", err)
		}
		b.ReportMetric(float64(len(files)), "files/op")
	}
}