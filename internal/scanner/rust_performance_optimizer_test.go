package scanner

import (
	"fmt"
	"testing"
	"time"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

func TestNewRustPerformanceOptimizer(t *testing.T) {
	optimizer := NewRustPerformanceOptimizer(false)
	
	if optimizer == nil {
		t.Fatal("Expected optimizer to be created")
	}
	
	if optimizer.cacheMaxSize != 1000 {
		t.Errorf("Expected cache max size to be 1000, got %d", optimizer.cacheMaxSize)
	}
	
	if optimizer.cacheTTL != 30*time.Minute {
		t.Errorf("Expected cache TTL to be 30 minutes, got %v", optimizer.cacheTTL)
	}
	
	if optimizer.maxWorkers <= 0 {
		t.Error("Expected max workers to be greater than 0")
	}
}

func TestRustPerformanceOptimizer_CacheOperations(t *testing.T) {
	optimizer := NewRustPerformanceOptimizer(false)
	
	// Test cache miss
	astInfo := optimizer.GetCachedAST("test.rs", 12345)
	if astInfo != nil {
		t.Error("Expected cache miss for non-existent file")
	}
	
	// Create test AST info
	testAST := &types.RustASTInfo{
		FilePath:  "test.rs",
		CrateName: "test_crate",
		Functions: []*types.RustFunctionInfo{
			{
				Name:      "test_func",
				IsPublic:  true,
				StartLine: 1,
				EndLine:   10,
			},
		},
	}
	
	// Cache the AST
	contentHash := uint64(12345)
	optimizer.CacheAST("test.rs", testAST, contentHash)
	
	// Test cache hit
	cachedAST := optimizer.GetCachedAST("test.rs", contentHash)
	if cachedAST == nil {
		t.Fatal("Expected cache hit for cached file")
	}
	
	if cachedAST.FilePath != testAST.FilePath {
		t.Errorf("Expected file path %s, got %s", testAST.FilePath, cachedAST.FilePath)
	}
	
	if len(cachedAST.Functions) != len(testAST.Functions) {
		t.Errorf("Expected %d functions, got %d", len(testAST.Functions), len(cachedAST.Functions))
	}
	
	// Test cache miss with different hash
	cachedAST = optimizer.GetCachedAST("test.rs", 54321)
	if cachedAST != nil {
		t.Error("Expected cache miss for different content hash")
	}
}

func TestRustPerformanceOptimizer_CacheExpiration(t *testing.T) {
	optimizer := NewRustPerformanceOptimizer(false)
	
	// Set short TTL for testing
	optimizer.SetCacheConfiguration(100, 1*time.Millisecond)
	
	testAST := &types.RustASTInfo{
		FilePath: "test.rs",
		Functions: []*types.RustFunctionInfo{
			{Name: "test_func", StartLine: 1, EndLine: 5},
		},
	}
	
	contentHash := uint64(12345)
	optimizer.CacheAST("test.rs", testAST, contentHash)
	
	// Should be available immediately
	cachedAST := optimizer.GetCachedAST("test.rs", contentHash)
	if cachedAST == nil {
		t.Error("Expected cache hit immediately after caching")
	}
	
	// Wait for expiration
	time.Sleep(2 * time.Millisecond)
	
	// Should be expired now
	cachedAST = optimizer.GetCachedAST("test.rs", contentHash)
	if cachedAST != nil {
		t.Error("Expected cache miss after expiration")
	}
}

func TestRustPerformanceOptimizer_MemoryPools(t *testing.T) {
	optimizer := NewRustPerformanceOptimizer(false)
	
	// Test AST info pool
	astInfo1 := optimizer.GetASTInfo()
	if astInfo1 == nil {
		t.Fatal("Expected AST info from pool")
	}
	
	// Add some data
	astInfo1.FilePath = "test.rs"
	astInfo1.Functions = append(astInfo1.Functions, &types.RustFunctionInfo{
		Name: "test_func",
	})
	
	// Return to pool
	optimizer.PutASTInfo(astInfo1)
	
	// Get again - should be reset
	astInfo2 := optimizer.GetASTInfo()
	if astInfo2.FilePath != "" {
		t.Error("Expected AST info to be reset when retrieved from pool")
	}
	
	if len(astInfo2.Functions) != 0 {
		t.Error("Expected functions slice to be reset")
	}
	
	// Test scan result pool
	result1 := optimizer.GetScanResult()
	if result1 == nil {
		t.Fatal("Expected scan result from pool")
	}
	
	// Add some data
	result1.Violations = append(result1.Violations, &models.Violation{
		Type:    "TEST_VIOLATION",
		Message: "Test message",
	})
	
	// Return to pool
	optimizer.PutScanResult(result1)
	
	// Get again - should be reset
	result2 := optimizer.GetScanResult()
	if len(result2.Violations) != 0 {
		t.Error("Expected violations slice to be reset")
	}
}

func TestRustPerformanceOptimizer_ContentHashing(t *testing.T) {
	optimizer := NewRustPerformanceOptimizer(false)
	
	content1 := []byte("fn main() { println!(\"Hello, world!\"); }")
	content2 := []byte("fn main() { println!(\"Hello, Rust!\"); }")
	content3 := []byte("fn main() { println!(\"Hello, world!\"); }")
	
	hash1 := optimizer.CalculateContentHash(content1)
	hash2 := optimizer.CalculateContentHash(content2)
	hash3 := optimizer.CalculateContentHash(content3)
	
	if hash1 == 0 {
		t.Error("Expected non-zero hash for content1")
	}
	
	if hash2 == 0 {
		t.Error("Expected non-zero hash for content2")
	}
	
	if hash1 == hash2 {
		t.Error("Expected different hashes for different content")
	}
	
	if hash1 != hash3 {
		t.Error("Expected same hashes for identical content")
	}
}

func TestRustPerformanceOptimizer_PerformanceMetrics(t *testing.T) {
	optimizer := NewRustPerformanceOptimizer(false)
	
	// Initial metrics
	metrics := optimizer.GetPerformanceMetrics()
	if metrics["cache_hits"].(int64) != 0 {
		t.Error("Expected 0 initial cache hits")
	}
	
	if metrics["cache_misses"].(int64) != 0 {
		t.Error("Expected 0 initial cache misses")
	}
	
	// Create cache miss
	optimizer.GetCachedAST("test.rs", 12345)
	
	// Create cache hit
	testAST := &types.RustASTInfo{FilePath: "test.rs"}
	optimizer.CacheAST("test.rs", testAST, 12345)
	optimizer.GetCachedAST("test.rs", 12345)
	
	metrics = optimizer.GetPerformanceMetrics()
	if metrics["cache_hits"].(int64) != 1 {
		t.Errorf("Expected 1 cache hit, got %d", metrics["cache_hits"].(int64))
	}
	
	if metrics["cache_misses"].(int64) != 1 {
		t.Errorf("Expected 1 cache miss, got %d", metrics["cache_misses"].(int64))
	}
	
	hitRate := metrics["cache_hit_rate"].(float64)
	if hitRate != 50.0 {
		t.Errorf("Expected 50%% hit rate, got %.1f%%", hitRate)
	}
	
	// Reset metrics
	optimizer.ResetMetrics()
	metrics = optimizer.GetPerformanceMetrics()
	if metrics["cache_hits"].(int64) != 0 {
		t.Error("Expected 0 cache hits after reset")
	}
	
	if metrics["cache_misses"].(int64) != 0 {
		t.Error("Expected 0 cache misses after reset")
	}
}

func TestRustPerformanceOptimizer_CacheCleanup(t *testing.T) {
	optimizer := NewRustPerformanceOptimizer(false)
	
	// Set small cache size for testing
	optimizer.SetCacheConfiguration(2, 1*time.Hour)
	
	// Fill cache beyond capacity
	for i := 0; i < 5; i++ {
		testAST := &types.RustASTInfo{
			FilePath: fmt.Sprintf("test%d.rs", i),
		}
		optimizer.CacheAST(fmt.Sprintf("test%d.rs", i), testAST, uint64(i))
	}
	
	// Cache should be limited in size
	metrics := optimizer.GetPerformanceMetrics()
	cacheSize := metrics["cache_size"].(int)
	if cacheSize > 2 {
		t.Errorf("Expected cache size <= 2, got %d", cacheSize)
	}
	
	// Clear entire cache
	optimizer.ClearCache()
	metrics = optimizer.GetPerformanceMetrics()
	cacheSize = metrics["cache_size"].(int)
	if cacheSize != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cacheSize)
	}
}

func TestRustPerformanceOptimizer_WorkerConfiguration(t *testing.T) {
	optimizer := NewRustPerformanceOptimizer(false)
	
	// Test initial configuration
	metrics := optimizer.GetPerformanceMetrics()
	initialWorkers := metrics["max_workers"].(int)
	initialBuffer := metrics["buffer_size"].(int)
	
	if initialWorkers <= 0 {
		t.Error("Expected positive number of workers")
	}
	
	if initialBuffer <= 0 {
		t.Error("Expected positive buffer size")
	}
	
	// Update configuration
	optimizer.SetWorkerConfiguration(8, 16)
	
	metrics = optimizer.GetPerformanceMetrics()
	if metrics["max_workers"].(int) != 8 {
		t.Errorf("Expected 8 workers, got %d", metrics["max_workers"].(int))
	}
	
	if metrics["buffer_size"].(int) != 16 {
		t.Errorf("Expected 16 buffer size, got %d", metrics["buffer_size"].(int))
	}
}

func TestRustPerformanceOptimizer_MemoryEstimation(t *testing.T) {
	optimizer := NewRustPerformanceOptimizer(false)
	
	// Test with empty cache
	memStats := optimizer.EstimateMemoryUsage()
	if memStats["estimated_cache_memory_bytes"].(int) != 0 {
		t.Error("Expected 0 memory usage for empty cache")
	}
	
	if memStats["cache_entries"].(int) != 0 {
		t.Error("Expected 0 cache entries for empty cache")
	}
	
	// Add some entries
	for i := 0; i < 3; i++ {
		testAST := &types.RustASTInfo{
			FilePath: fmt.Sprintf("test%d.rs", i),
			Functions: []*types.RustFunctionInfo{
				{Name: "func1", StartLine: 1, EndLine: 10},
				{Name: "func2", StartLine: 11, EndLine: 20},
			},
			Structs: []*types.RustStructInfo{
				{Name: "Struct1", StartLine: 25, EndLine: 30},
			},
		}
		optimizer.CacheAST(fmt.Sprintf("test%d.rs", i), testAST, uint64(i))
	}
	
	memStats = optimizer.EstimateMemoryUsage()
	memoryBytes := memStats["estimated_cache_memory_bytes"].(int)
	cacheEntries := memStats["cache_entries"].(int)
	
	if memoryBytes <= 0 {
		t.Error("Expected positive memory usage for populated cache")
	}
	
	if cacheEntries != 3 {
		t.Errorf("Expected 3 cache entries, got %d", cacheEntries)
	}
	
	if memStats["avg_memory_per_entry_bytes"].(int) <= 0 {
		t.Error("Expected positive average memory per entry")
	}
	
	memoryMB := memStats["estimated_cache_memory_mb"].(float64)
	if memoryMB <= 0 {
		t.Error("Expected positive memory usage in MB")
	}
}

func TestRustPerformanceOptimizer_ConcurrentAccess(t *testing.T) {
	optimizer := NewRustPerformanceOptimizer(false)
	
	// Test concurrent cache operations
	const numGoroutines = 10
	const operationsPerGoroutine = 100
	
	done := make(chan bool, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			for j := 0; j < operationsPerGoroutine; j++ {
				filePath := fmt.Sprintf("test_%d_%d.rs", id, j)
				testAST := &types.RustASTInfo{
					FilePath: filePath,
					Functions: []*types.RustFunctionInfo{
						{Name: "test_func", StartLine: 1, EndLine: 5},
					},
				}
				
				// Cache operation
				optimizer.CacheAST(filePath, testAST, uint64(id*1000+j))
				
				// Retrieve operation
				cachedAST := optimizer.GetCachedAST(filePath, uint64(id*1000+j))
				if cachedAST == nil {
					t.Errorf("Expected to retrieve cached AST for %s", filePath)
				}
				
				// Pool operations
				astInfo := optimizer.GetASTInfo()
				optimizer.PutASTInfo(astInfo)
				
				result := optimizer.GetScanResult()
				optimizer.PutScanResult(result)
			}
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	// Check final metrics
	metrics := optimizer.GetPerformanceMetrics()
	totalHits := metrics["cache_hits"].(int64)
	totalMisses := metrics["cache_misses"].(int64)
	
	if totalHits+totalMisses != numGoroutines*operationsPerGoroutine {
		t.Errorf("Expected %d total operations, got %d", 
			numGoroutines*operationsPerGoroutine, totalHits+totalMisses)
	}
}

func BenchmarkRustPerformanceOptimizer_CacheOperations(b *testing.B) {
	optimizer := NewRustPerformanceOptimizer(false)
	
	testAST := &types.RustASTInfo{
		FilePath: "bench_test.rs",
		Functions: []*types.RustFunctionInfo{
			{Name: "bench_func", StartLine: 1, EndLine: 10},
		},
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			filePath := fmt.Sprintf("bench_test_%d.rs", i)
			contentHash := uint64(i)
			
			// Cache
			optimizer.CacheAST(filePath, testAST, contentHash)
			
			// Retrieve
			optimizer.GetCachedAST(filePath, contentHash)
			
			i++
		}
	})
}

func BenchmarkRustPerformanceOptimizer_BasicMemoryPools(b *testing.B) {
	optimizer := NewRustPerformanceOptimizer(false)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// AST info pool
			astInfo := optimizer.GetASTInfo()
			optimizer.PutASTInfo(astInfo)
			
			// Scan result pool
			result := optimizer.GetScanResult()
			optimizer.PutScanResult(result)
		}
	})
}

