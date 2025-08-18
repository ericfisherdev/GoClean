package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/testutils"
)

// BenchmarkRustVsGoParsingPerformance compares Rust and Go parsing performance
func BenchmarkRustVsGoParsingPerformance(b *testing.B) {
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
			b.Run("Go", func(b *testing.B) {
				benchmarkGoParsingPerformance(b, bm.fileCount, bm.linesPerFile)
			})
			b.Run("Rust", func(b *testing.B) {
				benchmarkRustParsingPerformance(b, bm.fileCount, bm.linesPerFile)
			})
		})
	}
}

func benchmarkGoParsingPerformance(b *testing.B, fileCount, linesPerFile int) {
	helper := testutils.NewBenchmarkHelper(b)
	files := helper.CreateBenchmarkFiles(b, fileCount, linesPerFile)
	
	analyzer := NewASTAnalyzer(false)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	var totalFiles int
	for i := 0; i < b.N; i++ {
		for _, filePath := range files {
			content, err := os.ReadFile(filePath)
			if err != nil {
				b.Fatalf("Failed to read file %s: %v", filePath, err)
			}
			
			_, err = analyzer.AnalyzeGoFile(filePath, content)
			if err != nil {
				b.Fatalf("Failed to parse Go file %s: %v", filePath, err)
			}
			totalFiles++
		}
	}
	
	b.ReportMetric(float64(totalFiles)/b.Elapsed().Seconds(), "files/sec")
}

func benchmarkRustParsingPerformance(b *testing.B, fileCount, linesPerFile int) {
	helper := testutils.NewBenchmarkHelper(b)
	
	// Generate Rust content
	rustContent := generateRustFileContent(linesPerFile)
	files := helper.CreateRustBenchmarkFiles(b, fileCount, rustContent)
	
	analyzer := &RustASTAnalyzer{verbose: false}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	var totalFiles int
	for i := 0; i < b.N; i++ {
		for _, file := range files {
			content := []byte(rustContent)
			
			_, err := analyzer.AnalyzeRustFile(file.Path, content)
			if err != nil {
				b.Fatalf("Failed to analyze Rust file %s: %v", file.Path, err)
			}
			totalFiles++
		}
	}
	
	b.ReportMetric(float64(totalFiles)/b.Elapsed().Seconds(), "files/sec")
}

// BenchmarkRealWorldProject tests performance on the real Helix project
func BenchmarkRealWorldProject(b *testing.B) {
	// Get paths from environment variables
	helixPath := os.Getenv("HELIX_PATH")
	if helixPath == "" {
		b.Skipf("HELIX_PATH environment variable not set - skipping Helix benchmarks")
		return
	}
	
	// Check if Helix path exists
	if _, err := os.Stat(helixPath); os.IsNotExist(err) {
		b.Skipf("Helix path does not exist: %s - skipping Helix benchmarks", helixPath)
		return
	}
	
	b.Run("RustFiles_Helix", func(b *testing.B) {
		benchmarkRealWorldRustFiles(b, helixPath)
	})
	
	// For comparison, we'll use the GoClean codebase for Go files
	goCleanPath := os.Getenv("GOCLEAN_PATH")
	if goCleanPath == "" {
		b.Skipf("GOCLEAN_PATH environment variable not set - skipping GoClean benchmarks")
		return
	}
	
	// Check if GoClean path exists
	if _, err := os.Stat(goCleanPath); os.IsNotExist(err) {
		b.Skipf("GoClean path does not exist: %s - skipping GoClean benchmarks", goCleanPath)
		return
	}
	
	b.Run("GoFiles_GoClean", func(b *testing.B) {
		benchmarkRealWorldGoFiles(b, goCleanPath)
	})
}

func benchmarkRealWorldRustFiles(b *testing.B, projectPath string) {
	// Discover all Rust files in the Helix project
	rustFiles, err := discoverRustFiles(projectPath)
	if err != nil {
		b.Fatalf("Failed to discover Rust files: %v", err)
	}
	
	if len(rustFiles) == 0 {
		b.Skip("No Rust files found in project")
	}
	
	b.Logf("Found %d Rust files in Helix project", len(rustFiles))
	
	analyzer := &RustASTAnalyzer{verbose: false}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	var totalFiles int
	var totalBytes int64
	
	for i := 0; i < b.N; i++ {
		for _, file := range rustFiles {
			content, err := os.ReadFile(file.Path)
			if err != nil {
				b.Fatalf("Failed to read file %s: %v", file.Path, err)
			}
			
			_, err = analyzer.AnalyzeRustFile(file.Path, content)
			if err != nil {
				b.Fatalf("Failed to analyze Rust file %s: %v", file.Path, err)
			}
			
			totalFiles++
			totalBytes += int64(len(content))
		}
	}
	
	elapsed := b.Elapsed()
	b.ReportMetric(float64(totalFiles)/elapsed.Seconds(), "files/sec")
	b.ReportMetric(float64(totalBytes)/elapsed.Seconds(), "bytes/sec")
	b.ReportMetric(float64(totalBytes)/(1024*1024), "total_MB")
}

func benchmarkRealWorldGoFiles(b *testing.B, projectPath string) {
	// Discover all Go files in the GoClean project
	goFiles, err := discoverGoFiles(projectPath)
	if err != nil {
		b.Fatalf("Failed to discover Go files: %v", err)
	}
	
	if len(goFiles) == 0 {
		b.Skip("No Go files found in project")
	}
	
	b.Logf("Found %d Go files in GoClean project", len(goFiles))
	
	analyzer := NewASTAnalyzer(false)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	var totalFiles int
	var totalBytes int64
	
	for i := 0; i < b.N; i++ {
		for _, filePath := range goFiles {
			content, err := os.ReadFile(filePath)
			if err != nil {
				b.Fatalf("Failed to read file %s: %v", filePath, err)
			}
			
			_, err = analyzer.AnalyzeGoFile(filePath, content)
			if err != nil {
				b.Fatalf("Failed to parse Go file %s: %v", filePath, err)
			}
			
			totalFiles++
			totalBytes += int64(len(content))
		}
	}
	
	elapsed := b.Elapsed()
	b.ReportMetric(float64(totalFiles)/elapsed.Seconds(), "files/sec")
	b.ReportMetric(float64(totalBytes)/elapsed.Seconds(), "bytes/sec")
	b.ReportMetric(float64(totalBytes)/(1024*1024), "total_MB")
}

// BenchmarkParsingMemoryComparison compares memory usage between Rust and Go parsing
func BenchmarkParsingMemoryComparison(b *testing.B) {
	fileCount := 100
	linesPerFile := 200
	
	b.Run("GoMemoryUsage", func(b *testing.B) {
		benchmarkGoMemoryUsage(b, fileCount, linesPerFile)
	})
	
	b.Run("RustMemoryUsage", func(b *testing.B) {
		benchmarkRustMemoryUsage(b, fileCount, linesPerFile)
	})
}

func benchmarkGoMemoryUsage(b *testing.B, fileCount, linesPerFile int) {
	helper := testutils.NewBenchmarkHelper(b)
	files := helper.CreateBenchmarkFiles(b, fileCount, linesPerFile)
	
	analyzer := NewASTAnalyzer(false)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		memUsage := testutils.StartMemoryMeasurement()
		
		for _, filePath := range files {
			content, err := os.ReadFile(filePath)
			if err != nil {
				b.Fatalf("Failed to read file %s: %v", filePath, err)
			}
			
			_, err = analyzer.AnalyzeGoFile(filePath, content)
			if err != nil {
				b.Fatalf("Failed to parse Go file %s: %v", filePath, err)
			}
		}
		
		memUsage.StopMemoryMeasurement()
		b.ReportMetric(float64(memUsage.Delta()), "bytes/op")
	}
}

func benchmarkRustMemoryUsage(b *testing.B, fileCount, linesPerFile int) {
	helper := testutils.NewBenchmarkHelper(b)
	rustContent := generateRustFileContent(linesPerFile)
	files := helper.CreateRustBenchmarkFiles(b, fileCount, rustContent)
	
	analyzer := &RustASTAnalyzer{verbose: false}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		memUsage := testutils.StartMemoryMeasurement()
		
		for _, file := range files {
			content := []byte(rustContent)
			
			_, err := analyzer.AnalyzeRustFile(file.Path, content)
			if err != nil {
				b.Fatalf("Failed to analyze Rust file %s: %v", file.Path, err)
			}
		}
		
		memUsage.StopMemoryMeasurement()
		b.ReportMetric(float64(memUsage.Delta()), "bytes/op")
	}
}

// BenchmarkLargeRustProject tests performance regression detection for large projects
func BenchmarkLargeRustProject(b *testing.B) {
	helixPath := "/run/media/esfisher/NovusLocus/Dev Projects/helix"
	
	// Performance targets based on Go parsing performance
	targetFilesPerSecond := 500.0 // Minimum acceptable files per second
	targetMemoryPerFile := int64(5 * 1024 * 1024) // 5MB per file maximum
	
	rustFiles, err := discoverRustFiles(helixPath)
	if err != nil {
		b.Fatalf("Failed to discover Rust files: %v", err)
	}
	
	if len(rustFiles) == 0 {
		b.Skip("No Rust files found for performance regression test")
	}
	
	// Limit to first 50 files for performance regression testing
	if len(rustFiles) > 50 {
		rustFiles = rustFiles[:50]
	}
	
	analyzer := &RustASTAnalyzer{verbose: false}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		start := time.Now()
		var memBefore runtime.MemStats
		runtime.ReadMemStats(&memBefore)
		
		for _, file := range rustFiles {
			content, err := os.ReadFile(file.Path)
			if err != nil {
				b.Fatalf("Failed to read file %s: %v", file.Path, err)
			}
			
			_, err = analyzer.AnalyzeRustFile(file.Path, content)
			if err != nil {
				b.Fatalf("Failed to analyze Rust file %s: %v", file.Path, err)
			}
		}
		
		elapsed := time.Since(start)
		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)
		
		filesPerSecond := float64(len(rustFiles)) / elapsed.Seconds()
		memoryPerFile := int64(memAfter.Alloc-memBefore.Alloc) / int64(len(rustFiles))
		
		// Performance regression checks
		if filesPerSecond < targetFilesPerSecond {
			b.Logf("WARNING: Performance regression detected. Files/sec: %.2f (target: %.2f)", 
				filesPerSecond, targetFilesPerSecond)
		}
		
		if memoryPerFile > targetMemoryPerFile {
			b.Logf("WARNING: Memory usage regression detected. Memory/file: %d bytes (target: %d bytes)", 
				memoryPerFile, targetMemoryPerFile)
		}
		
		b.ReportMetric(filesPerSecond, "files/sec")
		b.ReportMetric(float64(memoryPerFile), "memory_bytes/file")
	}
}

// BenchmarkParsingAccuracy compares parsing accuracy between Rust and Go
func BenchmarkParsingAccuracy(b *testing.B) {
	b.Run("RustAccuracy", func(b *testing.B) {
		benchmarkRustParsingAccuracy(b)
	})
	
	b.Run("GoAccuracy", func(b *testing.B) {
		benchmarkGoParsingAccuracy(b)
	})
}

func benchmarkRustParsingAccuracy(b *testing.B) {
	// Use a complex Rust file to test parsing accuracy
	complexRustContent := `
use std::collections::HashMap;
use std::sync::{Arc, Mutex};

#[derive(Debug, Clone)]
pub struct ComplexStruct<T> 
where 
    T: Clone + Send + Sync + 'static,
{
    pub field1: T,
    pub field2: Option<String>,
    pub field3: Arc<Mutex<HashMap<String, i32>>>,
}

impl<T> ComplexStruct<T> 
where 
    T: Clone + Send + Sync + 'static,
{
    pub fn new(field1: T) -> Self {
        Self {
            field1,
            field2: None,
            field3: Arc::new(Mutex::new(HashMap::new())),
        }
    }
    
    pub async fn complex_method(&self, param1: &str, param2: i32) -> Result<String, Box<dyn std::error::Error>> {
        let mut map = self.field3.lock().unwrap();
        map.insert(param1.to_string(), param2);
        
        match self.field2.as_ref() {
            Some(value) => Ok(format!("{}-{}", value, param2)),
            None => Err("Field2 is None".into()),
        }
    }
}

pub trait ComplexTrait {
    type AssociatedType;
    
    fn trait_method(&self) -> Self::AssociatedType;
    fn default_impl(&self) -> String {
        "default".to_string()
    }
}

macro_rules! complex_macro {
    ($name:ident, $type:ty) => {
        impl ComplexTrait for $name {
            type AssociatedType = $type;
            
            fn trait_method(&self) -> Self::AssociatedType {
                Default::default()
            }
        }
    };
}

#[cfg(test)]
mod tests {
    use super::*;
    
    #[tokio::test]
    async fn test_complex_functionality() {
        let instance = ComplexStruct::new(42);
        let result = instance.complex_method("test", 100).await;
        assert!(result.is_err());
    }
}
`
	
	helper := testutils.NewBenchmarkHelper(b)
	files := helper.CreateRustBenchmarkFiles(b, 1, complexRustContent)
	
	analyzer := &RustASTAnalyzer{verbose: false}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		for _, file := range files {
			astInfo, err := analyzer.AnalyzeRustFile(file.Path, []byte(complexRustContent))
			if err != nil {
				b.Fatalf("Failed to analyze complex Rust file: %v", err)
			}
			
			// Verify parsing accuracy
			expectedStructs := 1
			expectedFunctions := 3 // new, complex_method, trait_method (from macro)
			expectedTraits := 1
			expectedImpls := 2 // manual impl + macro impl
			expectedMacros := 1
			
			if len(astInfo.Structs) != expectedStructs {
				b.Errorf("Expected %d structs, got %d", expectedStructs, len(astInfo.Structs))
			}
			
			if len(astInfo.Functions) < expectedFunctions {
				b.Errorf("Expected at least %d functions, got %d", expectedFunctions, len(astInfo.Functions))
			}
			
			if len(astInfo.Traits) != expectedTraits {
				b.Errorf("Expected %d traits, got %d", expectedTraits, len(astInfo.Traits))
			}
			
			if len(astInfo.Impls) < expectedImpls {
				b.Errorf("Expected at least %d impls, got %d", expectedImpls, len(astInfo.Impls))
			}
			
			if len(astInfo.Macros) < expectedMacros {
				b.Errorf("Expected at least %d macros, got %d", expectedMacros, len(astInfo.Macros))
			}
			
			// Report accuracy metrics
			b.ReportMetric(float64(len(astInfo.Structs)), "structs_detected")
			b.ReportMetric(float64(len(astInfo.Functions)), "functions_detected")
			b.ReportMetric(float64(len(astInfo.Traits)), "traits_detected")
			b.ReportMetric(float64(len(astInfo.Impls)), "impls_detected")
			b.ReportMetric(float64(len(astInfo.Macros)), "macros_detected")
		}
	}
}

func benchmarkGoParsingAccuracy(b *testing.B) {
	// Use a complex Go file to test parsing accuracy
	complexGoContent := `
package main

import (
	"context"
	"fmt"
	"sync"
)

type ComplexInterface interface {
	Method1(ctx context.Context, param string) error
	Method2() (int, error)
}

type ComplexStruct struct {
	Field1 string
	Field2 *int
	Field3 chan bool
	mutex  sync.RWMutex
}

func NewComplexStruct() *ComplexStruct {
	return &ComplexStruct{
		Field1: "default",
		Field2: nil,
		Field3: make(chan bool, 1),
	}
}

func (c *ComplexStruct) Method1(ctx context.Context, param string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.Field3 <- true:
		c.Field1 = param
		return nil
	}
}

func (c *ComplexStruct) Method2() (int, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	if c.Field2 == nil {
		return 0, fmt.Errorf("Field2 is nil")
	}
	return *c.Field2, nil
}

func complexFunction(param1 string, param2 int, param3 ...interface{}) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v", r)
		}
	}()
	
	switch param2 {
	case 1, 2, 3:
		result = fmt.Sprintf("low: %s", param1)
	case 4, 5, 6:
		result = fmt.Sprintf("medium: %s", param1)
	default:
		result = fmt.Sprintf("high: %s", param1)
	}
	
	return result, nil
}
`
	
	helper := testutils.NewBenchmarkHelper(b)
	files := helper.CreateBenchmarkFiles(b, 1, 200) // Create one file
	
	// Overwrite with our complex content
	err := os.WriteFile(files[0], []byte(complexGoContent), 0644)
	if err != nil {
		b.Fatalf("Failed to write complex Go content: %v", err)
	}
	
	analyzer := NewASTAnalyzer(false)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		astInfo, err := analyzer.AnalyzeGoFile(files[0], []byte(complexGoContent))
		if err != nil {
			b.Fatalf("Failed to analyze complex Go file: %v", err)
		}
		
		// Verify parsing accuracy
		expectedTypes := 2 // ComplexStruct + ComplexInterface
		expectedFunctions := 4 // NewComplexStruct, Method1, Method2, complexFunction
		
		if len(astInfo.Types) != expectedTypes {
			b.Errorf("Expected %d types, got %d", expectedTypes, len(astInfo.Types))
		}
		
		if len(astInfo.Functions) != expectedFunctions {
			b.Errorf("Expected %d functions, got %d", expectedFunctions, len(astInfo.Functions))
		}
		
		// Count structs and interfaces separately
		structCount := 0
		interfaceCount := 0
		for _, t := range astInfo.Types {
			if t.Kind == "struct" {
				structCount++
			} else if t.Kind == "interface" {
				interfaceCount++
			}
		}
		
		// Report accuracy metrics
		b.ReportMetric(float64(structCount), "structs_detected")
		b.ReportMetric(float64(len(astInfo.Functions)), "functions_detected")
		b.ReportMetric(float64(interfaceCount), "interfaces_detected")
	}
}

// Helper functions

func generateRustFileContent(lines int) string {
	content := `use std::collections::HashMap;

pub struct TestStruct {
    pub field1: String,
    pub field2: i32,
}

impl TestStruct {
    pub fn new(field1: String, field2: i32) -> Self {
        TestStruct { field1, field2 }
    }
    
    pub fn get_field1(&self) -> &str {
        &self.field1
    }
}

pub enum TestEnum {
    Variant1,
    Variant2(i32),
    Variant3 { name: String },
}

pub trait TestTrait {
    fn test_method(&self) -> String;
}

impl TestTrait for TestStruct {
    fn test_method(&self) -> String {
        format!("{}: {}", self.field1, self.field2)
    }
}

`
	
	// Add more lines as needed
	for i := 0; i < lines-25; i++ {
		content += fmt.Sprintf("    // Line %d\n", i+1)
	}
	
	return content
}

func discoverRustFiles(projectPath string) ([]*models.FileInfo, error) {
	var files []*models.FileInfo
	
	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip certain directories
		if info.IsDir() {
			name := info.Name()
			if name == "target" || name == ".git" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		
		// Only include .rs files
		if filepath.Ext(path) == ".rs" {
			fileInfo := &models.FileInfo{
				Path:     path,
				Name:     info.Name(),
				Language: "Rust",
				Size:     info.Size(),
			}
			files = append(files, fileInfo)
		}
		
		return nil
	})
	
	return files, err
}

func discoverGoFiles(projectPath string) ([]string, error) {
	var files []string
	
	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip certain directories
		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || name == ".git" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		
		// Only include .go files, exclude test files for consistency
		if filepath.Ext(path) == ".go" && !strings.HasSuffix(filepath.Base(path), "_test.go") {
			files = append(files, path)
		}
		
		return nil
	})
	
	return files, err
}