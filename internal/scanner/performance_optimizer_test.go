//go:build !cgo || no_rust
// +build !cgo no_rust

package scanner

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// TestMemoryPoolingBasic tests the memory pooling functionality
func TestMemoryPoolingBasic(t *testing.T) {
	pool := NewMemoryPool()

	// Test string slice pooling
	slice1 := pool.GetStringSlice()
	slice1 = append(slice1, "test1", "test2", "test3")
	pool.PutStringSlice(slice1)

	slice2 := pool.GetStringSlice()
	if len(slice2) != 0 {
		t.Errorf("Expected empty slice from pool, got length %d", len(slice2))
	}
	if cap(slice2) < 3 {
		t.Errorf("Expected capacity >= 3 from reused slice, got %d", cap(slice2))
	}

	// Test buffer pooling
	buffer1 := pool.GetBuffer()
	buffer1 = append(buffer1, []byte("test data")...)
	pool.PutBuffer(buffer1)

	buffer2 := pool.GetBuffer()
	if len(buffer2) != 0 {
		t.Errorf("Expected empty buffer from pool, got length %d", len(buffer2))
	}
	if cap(buffer2) < 9 {
		t.Errorf("Expected capacity >= 9 from reused buffer, got %d", cap(buffer2))
	}
}

// TestPerformanceProfilerBasic tests the performance profiling functionality
func TestPerformanceProfilerBasic(t *testing.T) {
	profiler := NewPerformanceProfiler()

	// Create test parsing function
	testCode := []byte(`
fn fibonacci(n: u32) -> u32 {
    match n {
        0 => 0,
        1 => 1,
        _ => fibonacci(n - 1) + fibonacci(n - 2),
    }
}
`)

	parseFunc := func(content []byte) error {
		// Simulate parsing work
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	// Run multiple benchmarks
	for i := 0; i < 5; i++ {
		result := profiler.BenchmarkParsingOperation("test_parse", testCode, parseFunc)
		if !result.Success {
			t.Errorf("Benchmark %d failed: %s", i, result.Error)
		}
		if result.Duration < 10*time.Millisecond {
			t.Errorf("Benchmark %d duration too short: %v", i, result.Duration)
		}
	}

	// Test benchmark summary
	summary := profiler.GetBenchmarkSummary()
	benchmarkCount := summary["benchmark_count"].(int)
	if benchmarkCount != 5 {
		t.Errorf("Expected 5 benchmarks, got %d", benchmarkCount)
	}

	successRate := summary["success_rate_percent"].(float64)
	if successRate != 100.0 {
		t.Errorf("Expected 100%% success rate, got %.2f%%", successRate)
	}

	// Test recent benchmarks retrieval
	recent := profiler.GetRecentBenchmarks(3)
	if len(recent) != 3 {
		t.Errorf("Expected 3 recent benchmarks, got %d", len(recent))
	}

	// Test clearing benchmarks
	profiler.ClearBenchmarks()
	summary = profiler.GetBenchmarkSummary()
	benchmarkCount = summary["benchmark_count"].(int)
	if benchmarkCount != 0 {
		t.Errorf("Expected 0 benchmarks after clearing, got %d", benchmarkCount)
	}
}

// TestRustPerformanceOptimizer tests the performance optimizer functionality
func TestRustPerformanceOptimizer(t *testing.T) {
	optimizer := NewRustPerformanceOptimizer(true)

	// Test cache configuration
	optimizer.SetCacheConfiguration(500, 15*time.Minute)
	
	// Test worker configuration
	optimizer.SetWorkerConfiguration(4, 8)

	// Test AST info pooling
	astInfo := optimizer.GetASTInfo()
	if astInfo == nil {
		t.Error("Expected AST info from pool")
	}

	// Add some data to the AST info
	astInfo.FilePath = "test.rs"
	astInfo.Functions = append(astInfo.Functions, &types.RustFunctionInfo{
		Name: "test_function",
		LineCount: 10,
	})

	// Return to pool
	optimizer.PutASTInfo(astInfo)

	// Get another AST info (should be reused)
	astInfo2 := optimizer.GetASTInfo()
	if astInfo2.FilePath != "" {
		t.Error("Expected reset AST info from pool")
	}
	if len(astInfo2.Functions) != 0 {
		t.Error("Expected empty functions slice from reset AST info")
	}

	// Test scan result pooling
	result := optimizer.GetScanResult()
	if result == nil {
		t.Error("Expected scan result from pool")
	}

	// Add some violations
	result.Violations = append(result.Violations, &models.Violation{
		Type:     "test_violation",
		Severity: models.SeverityHigh,
		Message:  "Test violation message",
	})

	optimizer.PutScanResult(result)

	// Get another result (should be reused and reset)
	result2 := optimizer.GetScanResult()
	if len(result2.Violations) != 0 {
		t.Error("Expected empty violations slice from reset scan result")
	}
}

// TestCacheOperations tests the AST caching functionality
func TestCacheOperations(t *testing.T) {
	optimizer := NewRustPerformanceOptimizer(false)

	// Create test AST info
	astInfo := &types.RustASTInfo{
		FilePath: "test.rs",
		Functions: []*types.RustFunctionInfo{
			{
				Name:      "test_func",
				LineCount: 20,
				IsPublic:  true,
			},
		},
	}

	// Test content hash calculation
	content := []byte("fn test() { println!(\"Hello\"); }")
	hash := optimizer.CalculateContentHash(content)
	if hash == 0 {
		t.Error("Expected non-zero hash")
	}

	// Test caching
	optimizer.CacheAST("test.rs", astInfo, hash)

	// Test cache retrieval
	cachedAST := optimizer.GetCachedAST("test.rs", hash)
	if cachedAST == nil {
		t.Error("Expected cached AST to be found")
	}
	if cachedAST.FilePath != "test.rs" {
		t.Errorf("Expected cached AST file path to be 'test.rs', got '%s'", cachedAST.FilePath)
	}

	// Test cache miss with different hash
	cachedAST = optimizer.GetCachedAST("test.rs", hash+1)
	if cachedAST != nil {
		t.Error("Expected cache miss with different hash")
	}

	// Test cache miss with different file
	cachedAST = optimizer.GetCachedAST("other.rs", hash)
	if cachedAST != nil {
		t.Error("Expected cache miss with different file")
	}

	// Test performance metrics
	metrics := optimizer.GetPerformanceMetrics()
	if metrics["cache_hits"].(int64) != 1 {
		t.Errorf("Expected 1 cache hit, got %d", metrics["cache_hits"].(int64))
	}
	if metrics["cache_misses"].(int64) != 2 {
		t.Errorf("Expected 2 cache misses, got %d", metrics["cache_misses"].(int64))
	}

	// Test cache cleanup
	optimizer.CleanupCache()
	
	// Test cache clearing
	optimizer.ClearCache()
	cachedAST = optimizer.GetCachedAST("test.rs", hash)
	if cachedAST != nil {
		t.Error("Expected cache to be cleared")
	}
}

// TestMemoryUsageEstimation tests memory usage estimation functionality
func TestMemoryUsageEstimation(t *testing.T) {
	optimizer := NewRustPerformanceOptimizer(false)

	// Add some test data to cache
	for i := 0; i < 10; i++ {
		astInfo := &types.RustASTInfo{
			FilePath: fmt.Sprintf("test%d.rs", i),
			Functions: []*types.RustFunctionInfo{
				{Name: "func1", LineCount: 10},
				{Name: "func2", LineCount: 20},
			},
			Structs: []*types.RustStructInfo{
				{Name: "Struct1", FieldCount: 5},
			},
		}
		optimizer.CacheAST(fmt.Sprintf("test%d.rs", i), astInfo, uint64(i))
	}

	// Test memory usage estimation
	usage := optimizer.EstimateMemoryUsage()
	
	cacheEntries := usage["cache_entries"].(int)
	if cacheEntries != 10 {
		t.Errorf("Expected 10 cache entries, got %d", cacheEntries)
	}

	memoryBytes := usage["estimated_cache_memory_bytes"].(int)
	if memoryBytes <= 0 {
		t.Error("Expected positive memory usage estimation")
	}

	memoryMB := usage["estimated_cache_memory_mb"].(float64)
	if memoryMB <= 0 {
		t.Error("Expected positive memory usage in MB")
	}

	avgMemoryPerEntry := usage["avg_memory_per_entry_bytes"].(int)
	if avgMemoryPerEntry <= 0 {
		t.Error("Expected positive average memory per entry")
	}
}

// TestConcurrentCacheAccess tests thread safety of cache operations
func TestConcurrentCacheAccess(t *testing.T) {
	optimizer := NewRustPerformanceOptimizer(false)

	const numGoroutines = 10
	const operationsPerGoroutine = 100

	// Run concurrent cache operations
	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for j := 0; j < operationsPerGoroutine; j++ {
				filePath := fmt.Sprintf("test_%d_%d.rs", goroutineID, j)
				astInfo := &types.RustASTInfo{
					FilePath: filePath,
					Functions: []*types.RustFunctionInfo{
						{Name: "test_func", LineCount: 10},
					},
				}
				
				hash := uint64(goroutineID*1000 + j)
				
				// Cache the AST
				optimizer.CacheAST(filePath, astInfo, hash)
				
				// Try to retrieve it
				cachedAST := optimizer.GetCachedAST(filePath, hash)
				if cachedAST == nil {
					t.Errorf("Failed to retrieve cached AST for %s", filePath)
					return
				}
				
				// Verify the content
				if cachedAST.FilePath != filePath {
					t.Errorf("Cached AST file path mismatch: expected %s, got %s", filePath, cachedAST.FilePath)
					return
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify final state
	metrics := optimizer.GetPerformanceMetrics()
	totalOperations := int64(numGoroutines * operationsPerGoroutine)
	
	cacheHits := metrics["cache_hits"].(int64)
	if cacheHits != totalOperations {
		t.Errorf("Expected %d cache hits, got %d", totalOperations, cacheHits)
	}

	// Test memory usage after concurrent operations
	usage := optimizer.EstimateMemoryUsage()
	cacheEntries := usage["cache_entries"].(int)
	if cacheEntries != numGoroutines*operationsPerGoroutine {
		t.Errorf("Expected %d cache entries, got %d", numGoroutines*operationsPerGoroutine, cacheEntries)
	}
}

// BenchmarkCacheOperations benchmarks cache performance
func BenchmarkCacheOperations(b *testing.B) {
	optimizer := NewRustPerformanceOptimizer(false)
	
	astInfo := &types.RustASTInfo{
		FilePath: "benchmark.rs",
		Functions: []*types.RustFunctionInfo{
			{Name: "bench_func", LineCount: 50, IsPublic: true},
		},
	}
	
	hash := uint64(12345)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filePath := fmt.Sprintf("bench_%d.rs", i%1000) // Cycle through 1000 files
		optimizer.CacheAST(filePath, astInfo, hash)
		optimizer.GetCachedAST(filePath, hash)
	}
}

// BenchmarkMemoryPooling benchmarks memory pooling performance
func BenchmarkMemoryPooling(b *testing.B) {
	pool := NewMemoryPool()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test string slice pooling
		slice := pool.GetStringSlice()
		slice = append(slice, "test1", "test2", "test3")
		pool.PutStringSlice(slice)
		
		// Test buffer pooling
		buffer := pool.GetBuffer()
		buffer = append(buffer, []byte("test data")...)
		pool.PutBuffer(buffer)
	}
}

