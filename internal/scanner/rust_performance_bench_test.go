package scanner

import (
	"fmt"
	"testing"
	"time"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/testutils"
	"github.com/ericfisherdev/goclean/internal/types"
)

// BenchmarkRustASTAnalyzer_WithOptimization compares performance with and without optimization
func BenchmarkRustASTAnalyzer_WithOptimization(b *testing.B) {
	benchmarks := []struct {
		name         string
		optimization bool
	}{
		{"WithoutOptimization", false},
		{"WithOptimization", true},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			benchmarkRustAnalyzerOptimization(b, bm.optimization)
		})
	}
}

func benchmarkRustAnalyzerOptimization(b *testing.B, withOptimization bool) {
	helper := testutils.NewBenchmarkHelper(b)
	
	// Create sample Rust content
	rustContent := `
fn main() {
    println!("Hello, world!");
}

pub struct Config {
    pub name: String,
    pub value: i32,
}

impl Config {
    pub fn new(name: String, value: i32) -> Self {
        Config { name, value }
    }
    
    pub fn update(&mut self, new_value: i32) {
        self.value = new_value;
    }
}

pub enum Status {
    Active,
    Inactive,
    Pending,
}

pub trait Processor {
    fn process(&self, data: &str) -> String;
}
`

	// Create multiple test files
	rustFiles := helper.CreateRustBenchmarkFiles(b, 50, rustContent)
	
	var analyzer *RustASTAnalyzer
	if withOptimization {
		optimizer := NewRustPerformanceOptimizer(false)
		analyzer = NewRustASTAnalyzerWithOptimizer(false, optimizer)
	} else {
		analyzer = &RustASTAnalyzer{verbose: false, optimizer: nil}
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		for _, file := range rustFiles {
			content := []byte(rustContent)
			_, err := analyzer.AnalyzeRustFile(file.Path, content)
			if err != nil {
				b.Fatalf("Failed to analyze Rust file %s: %v", file.Path, err)
			}
		}
	}
}

// BenchmarkRustPerformanceOptimizer_CachingBenefit measures caching benefit
func BenchmarkRustPerformanceOptimizer_CachingBenefit(b *testing.B) {
	optimizer := NewRustPerformanceOptimizer(false)
	
	// Create a realistic Rust AST for caching
	testAST := &types.RustASTInfo{
		FilePath:  "large_file.rs",
		CrateName: "large_crate",
		Functions: make([]*types.RustFunctionInfo, 100),
		Structs:   make([]*types.RustStructInfo, 50),
		Enums:     make([]*types.RustEnumInfo, 25),
		Traits:    make([]*types.RustTraitInfo, 20),
		Impls:     make([]*types.RustImplInfo, 80),
		Modules:   make([]*types.RustModuleInfo, 10),
		Constants: make([]*types.RustConstantInfo, 30),
		Uses:      make([]*types.RustUseInfo, 200),
		Macros:    make([]*types.RustMacroInfo, 15),
	}
	
	// Populate with realistic data
	for i := 0; i < 100; i++ {
		testAST.Functions[i] = &types.RustFunctionInfo{
			Name:        fmt.Sprintf("function_%d", i),
			StartLine:   i * 10,
			EndLine:     i*10 + 8,
			IsPublic:    i%2 == 0,
			Complexity:  i%10 + 1,
			LineCount:   8,
			Parameters:  make([]types.RustParameterInfo, i%5),
			ReturnType:  "Result<(), Error>",
			Visibility:  "pub",
		}
	}
	
	for i := 0; i < 50; i++ {
		testAST.Structs[i] = &types.RustStructInfo{
			Name:       fmt.Sprintf("Struct%d", i),
			StartLine:  i * 20,
			EndLine:    i*20 + 15,
			IsPublic:   i%3 == 0,
			FieldCount: i%10,
			Visibility: "pub",
		}
	}
	
	contentHash := uint64(123456789)
	
	// Cache the AST
	optimizer.CacheAST("large_file.rs", testAST, contentHash)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	// Benchmark cache retrieval
	for i := 0; i < b.N; i++ {
		cachedAST := optimizer.GetCachedAST("large_file.rs", contentHash)
		if cachedAST == nil {
			b.Fatal("Expected cached AST to be available")
		}
	}
}

// BenchmarkRustPerformanceOptimizer_MemoryPools measures pool allocation efficiency
func BenchmarkRustPerformanceOptimizer_MemoryPools(b *testing.B) {
	optimizer := NewRustPerformanceOptimizer(false)
	
	b.Run("ASTInfoPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			astInfo := optimizer.GetASTInfo()
			// Simulate some usage
			astInfo.FilePath = "test.rs"
			astInfo.Functions = append(astInfo.Functions, &types.RustFunctionInfo{
				Name: "test_func",
			})
			optimizer.PutASTInfo(astInfo)
		}
	})
	
	b.Run("ScanResultPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			result := optimizer.GetScanResult()
			// Simulate some usage
			result.Violations = append(result.Violations, &models.Violation{
				Type:    "TEST_VIOLATION",
				Message: "Test message",
			})
			optimizer.PutScanResult(result)
		}
	})
}

// BenchmarkRustPerformanceOptimizer_ParallelProcessing measures parallel processing benefit
func BenchmarkRustPerformanceOptimizer_ParallelProcessing(b *testing.B) {
	optimizer := NewRustPerformanceOptimizer(false)
	analyzer := NewRustASTAnalyzerWithOptimizer(false, optimizer)
	
	// Create test files
	helper := testutils.NewBenchmarkHelper(b)
	files := helper.CreateRustBenchmarkFiles(b, 100, `
fn main() {
    println!("Hello, world!");
}

pub struct TestStruct {
    pub field: i32,
}

impl TestStruct {
    pub fn new(field: i32) -> Self {
        TestStruct { field }
    }
}
`)
	
	processFunc := func(file *models.FileInfo, astInfo *types.RustASTInfo) (*models.ScanResult, error) {
		result := &models.ScanResult{
			File:        file,
			RustASTInfo: astInfo,
			Violations:  []*models.Violation{},
			Metrics: &models.FileMetrics{
				TotalLines:    100,
				CodeLines:     80,
				CommentLines:  15,
				BlankLines:    5,
				FunctionCount: len(astInfo.Functions),
				ClassCount:    len(astInfo.Structs),
			},
		}
		
		// Simulate some processing time
		time.Sleep(10 * time.Microsecond)
		
		return result, nil
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		results, err := optimizer.ProcessRustFilesInParallel(files, analyzer, processFunc)
		if err != nil {
			b.Fatalf("Parallel processing failed: %v", err)
		}
		
		if len(results) != len(files) {
			b.Fatalf("Expected %d results, got %d", len(files), len(results))
		}
	}
}

// BenchmarkRustPerformanceOptimizer_ContentHashing measures hash calculation performance
func BenchmarkRustPerformanceOptimizer_ContentHashing(b *testing.B) {
	optimizer := NewRustPerformanceOptimizer(false)
	
	contents := [][]byte{
		[]byte("fn main() { println!(\"Hello, world!\"); }"),
		[]byte(`
pub struct Config {
    name: String,
    value: i32,
}

impl Config {
    pub fn new(name: String, value: i32) -> Self {
        Config { name, value }
    }
}
`),
		[]byte(`
use std::collections::HashMap;
use std::sync::{Arc, Mutex};

pub trait Processor {
    fn process(&self, data: &str) -> Result<String, ProcessError>;
}

#[derive(Debug)]
pub enum ProcessError {
    InvalidInput,
    ProcessingFailed,
}
`),
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		for _, content := range contents {
			hash := optimizer.CalculateContentHash(content)
			if hash == 0 {
				b.Fatal("Hash should not be zero")
			}
		}
	}
}

// BenchmarkRustPerformanceOptimizer_CacheCleanup measures cleanup performance
func BenchmarkRustPerformanceOptimizer_CacheCleanup(b *testing.B) {
	optimizer := NewRustPerformanceOptimizer(false)
	
	// Fill cache with many entries
	for i := 0; i < 1000; i++ {
		testAST := &types.RustASTInfo{
			FilePath: fmt.Sprintf("test_%d.rs", i),
			Functions: []*types.RustFunctionInfo{
				{Name: fmt.Sprintf("func_%d", i), StartLine: 1, EndLine: 10},
			},
		}
		optimizer.CacheAST(fmt.Sprintf("test_%d.rs", i), testAST, uint64(i))
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		optimizer.CleanupCache()
	}
}

// BenchmarkRustPerformanceOptimizer_MemoryEstimation measures memory estimation performance
func BenchmarkRustPerformanceOptimizer_MemoryEstimation(b *testing.B) {
	optimizer := NewRustPerformanceOptimizer(false)
	
	// Add various cache entries
	for i := 0; i < 100; i++ {
		testAST := &types.RustASTInfo{
			FilePath:  fmt.Sprintf("test_%d.rs", i),
			Functions: make([]*types.RustFunctionInfo, i%20),
			Structs:   make([]*types.RustStructInfo, i%10),
			Enums:     make([]*types.RustEnumInfo, i%5),
		}
		
		// Populate with some data
		for j := 0; j < len(testAST.Functions); j++ {
			testAST.Functions[j] = &types.RustFunctionInfo{
				Name:      fmt.Sprintf("func_%d_%d", i, j),
				StartLine: j * 10,
				EndLine:   j*10 + 8,
			}
		}
		
		optimizer.CacheAST(fmt.Sprintf("test_%d.rs", i), testAST, uint64(i))
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		memStats := optimizer.EstimateMemoryUsage()
		if memStats["cache_entries"].(int) != 100 {
			b.Fatalf("Expected 100 cache entries, got %d", memStats["cache_entries"].(int))
		}
	}
}

// BenchmarkRustEngine_WithOptimizations compares engine performance with optimizations
func BenchmarkRustEngine_WithOptimizations(b *testing.B) {
	helper := testutils.NewBenchmarkHelper(b)
	_ = helper.CreateRustBenchmarkFiles(b, 50, `
fn main() {
    println!("Hello, world!");
}

pub struct Config {
    pub name: String,
    pub value: i32,
}

impl Config {
    pub fn new(name: String, value: i32) -> Self {
        Config { name, value }
    }
}
`)
	
	testDir := helper.TempDir
	
	b.Run("WithRustOptimizations", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			engine := NewEngine([]string{testDir}, []string{}, []string{".rs"}, false)
			engine.EnableRustOptimization(true)
			engine.SetRustCacheConfig(500, 10*time.Minute)
			
			_, _, err := engine.Scan()
			if err != nil {
				b.Fatalf("Scan failed: %v", err)
			}
		}
	})
	
	b.Run("WithoutRustOptimizations", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			engine := NewEngine([]string{testDir}, []string{}, []string{".rs"}, false)
			engine.EnableRustOptimization(false)
			
			_, _, err := engine.Scan()
			if err != nil {
				b.Fatalf("Scan failed: %v", err)
			}
		}
	})
}