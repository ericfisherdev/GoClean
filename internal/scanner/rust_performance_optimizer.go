package scanner

import (
	"hash/fnv"
	"math"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// RustASTCacheEntry represents a cached AST result
type RustASTCacheEntry struct {
	ASTInfo   *types.RustASTInfo
	Timestamp time.Time
	Hash      uint64
}

// RustPerformanceOptimizer provides performance optimizations for Rust parsing
type RustPerformanceOptimizer struct {
	// AST Caching
	astCache      map[string]*RustASTCacheEntry
	cacheMutex    sync.RWMutex
	cacheMaxSize  int
	cacheTTL      time.Duration
	
	// Worker Pool Configuration
	maxWorkers       int
	workerBufferSize int
	
	// Memory Pool for reusing objects
	astInfoPool   sync.Pool
	resultPool    sync.Pool
	
	// Performance metrics
	cacheHits     int64
	cacheMisses   int64
	metricsMutex  sync.RWMutex
	
	verbose bool
}

// NewRustPerformanceOptimizer creates a new performance optimizer
func NewRustPerformanceOptimizer(verbose bool) *RustPerformanceOptimizer {
	numCPU := runtime.NumCPU()
	
	optimizer := &RustPerformanceOptimizer{
		astCache:         make(map[string]*RustASTCacheEntry),
		cacheMaxSize:     1000, // Cache up to 1000 AST results
		cacheTTL:         30 * time.Minute, // Cache for 30 minutes
		maxWorkers:       numCPU,
		workerBufferSize: numCPU * 2,
		verbose:          verbose,
	}
	
	// Initialize object pools
	optimizer.astInfoPool = sync.Pool{
		New: func() interface{} {
			return &types.RustASTInfo{
				Functions: make([]*types.RustFunctionInfo, 0, 10),
				Structs:   make([]*types.RustStructInfo, 0, 5),
				Enums:     make([]*types.RustEnumInfo, 0, 5),
				Traits:    make([]*types.RustTraitInfo, 0, 5),
				Impls:     make([]*types.RustImplInfo, 0, 10),
				Modules:   make([]*types.RustModuleInfo, 0, 3),
				Constants: make([]*types.RustConstantInfo, 0, 5),
				Uses:      make([]*types.RustUseInfo, 0, 20),
				Macros:    make([]*types.RustMacroInfo, 0, 2),
			}
		},
	}
	
	optimizer.resultPool = sync.Pool{
		New: func() interface{} {
			return &models.ScanResult{
				Violations: make([]*models.Violation, 0, 50),
			}
		},
	}
	
	return optimizer
}

// SetCacheConfiguration configures the AST cache
func (opt *RustPerformanceOptimizer) SetCacheConfiguration(maxSize int, ttl time.Duration) {
	opt.cacheMutex.Lock()
	defer opt.cacheMutex.Unlock()
	
	opt.cacheMaxSize = maxSize
	opt.cacheTTL = ttl
}

// SetWorkerConfiguration configures the worker pool
func (opt *RustPerformanceOptimizer) SetWorkerConfiguration(maxWorkers, bufferSize int) {
	opt.maxWorkers = maxWorkers
	opt.workerBufferSize = bufferSize
}

// GetCachedAST retrieves AST from cache if available and valid
func (opt *RustPerformanceOptimizer) GetCachedAST(filePath string, contentHash uint64) *types.RustASTInfo {
	opt.cacheMutex.RLock()
	defer opt.cacheMutex.RUnlock()
	
	entry, exists := opt.astCache[filePath]
	if !exists {
		opt.incrementCacheMisses()
		return nil
	}
	
	// Check if cache entry is expired
	if time.Since(entry.Timestamp) > opt.cacheTTL {
		// Don't remove here, cleanup will handle it
		opt.incrementCacheMisses()
		return nil
	}
	
	// Check if content hash matches (file hasn't changed)
	if entry.Hash != contentHash {
		opt.incrementCacheMisses()
		return nil
	}
	
	opt.incrementCacheHits()
	return entry.ASTInfo
}

// CacheAST stores AST result in cache
func (opt *RustPerformanceOptimizer) CacheAST(filePath string, astInfo *types.RustASTInfo, contentHash uint64) {
	opt.cacheMutex.Lock()
	defer opt.cacheMutex.Unlock()
	
	// Check cache size and cleanup if necessary
	if len(opt.astCache) >= opt.cacheMaxSize {
		opt.cleanupCacheUnsafe()
	}
	
	opt.astCache[filePath] = &RustASTCacheEntry{
		ASTInfo:   astInfo,
		Timestamp: time.Now(),
		Hash:      contentHash,
	}
}

// CleanupCache removes expired entries from cache
func (opt *RustPerformanceOptimizer) CleanupCache() {
	opt.cacheMutex.Lock()
	defer opt.cacheMutex.Unlock()
	opt.cleanupCacheUnsafe()
}

// cleanupCacheUnsafe removes expired entries (must be called with write lock)
func (opt *RustPerformanceOptimizer) cleanupCacheUnsafe() {
	now := time.Now()
	for filePath, entry := range opt.astCache {
		if now.Sub(entry.Timestamp) > opt.cacheTTL {
			delete(opt.astCache, filePath)
		}
	}
	
	// If still too many entries, remove oldest ones
	if len(opt.astCache) >= opt.cacheMaxSize {
		type cacheItem struct {
			path      string
			timestamp time.Time
		}
		
		var items []cacheItem
		for path, entry := range opt.astCache {
			items = append(items, cacheItem{path, entry.Timestamp})
		}
		
		// Sort by timestamp (oldest first)
		sort.Slice(items, func(i, j int) bool {
			return items[i].timestamp.Before(items[j].timestamp)
		})
		
		// Remove oldest entries until we're under the limit
		removeCount := len(opt.astCache) - opt.cacheMaxSize/2 // Remove half when cleaning
		for i := 0; i < removeCount && i < len(items); i++ {
			delete(opt.astCache, items[i].path)
		}
	}
}

// GetASTInfo gets a reusable AST info object from the pool
func (opt *RustPerformanceOptimizer) GetASTInfo() *types.RustASTInfo {
	astInfo := opt.astInfoPool.Get().(*types.RustASTInfo)
	
	// Reset the object
	astInfo.FilePath = ""
	astInfo.CrateName = ""
	astInfo.Functions = astInfo.Functions[:0]
	astInfo.Structs = astInfo.Structs[:0]
	astInfo.Enums = astInfo.Enums[:0]
	astInfo.Traits = astInfo.Traits[:0]
	astInfo.Impls = astInfo.Impls[:0]
	astInfo.Modules = astInfo.Modules[:0]
	astInfo.Constants = astInfo.Constants[:0]
	astInfo.Uses = astInfo.Uses[:0]
	astInfo.Macros = astInfo.Macros[:0]
	
	return astInfo
}

// PutASTInfo returns an AST info object to the pool
func (opt *RustPerformanceOptimizer) PutASTInfo(astInfo *types.RustASTInfo) {
	// Don't return to pool if slices are too large (memory pressure)
	const maxSliceSize = 100
	
	if len(astInfo.Functions) > maxSliceSize ||
		len(astInfo.Structs) > maxSliceSize ||
		len(astInfo.Uses) > maxSliceSize {
		return
	}
	
	opt.astInfoPool.Put(astInfo)
}

// GetScanResult gets a reusable scan result object from the pool
func (opt *RustPerformanceOptimizer) GetScanResult() *models.ScanResult {
	result := opt.resultPool.Get().(*models.ScanResult)
	
	// Reset the object
	result.File = nil
	result.Violations = result.Violations[:0]
	result.Metrics = nil
	result.ASTInfo = nil
	result.RustASTInfo = nil
	
	return result
}

// PutScanResult returns a scan result object to the pool
func (opt *RustPerformanceOptimizer) PutScanResult(result *models.ScanResult) {
	// Don't return to pool if violations slice is too large
	if len(result.Violations) > 200 {
		return
	}
	
	opt.resultPool.Put(result)
}

// CalculateContentHash computes a hash of file content for cache validation
func (opt *RustPerformanceOptimizer) CalculateContentHash(content []byte) uint64 {
	h := fnv.New64a()
	h.Write(content)
	return h.Sum64()
}

// GetPerformanceMetrics returns current performance metrics
func (opt *RustPerformanceOptimizer) GetPerformanceMetrics() map[string]interface{} {
	opt.metricsMutex.RLock()
	defer opt.metricsMutex.RUnlock()
	
	opt.cacheMutex.RLock()
	cacheSize := len(opt.astCache)
	opt.cacheMutex.RUnlock()
	
	hitRate := float64(0)
	totalRequests := opt.cacheHits + opt.cacheMisses
	if totalRequests > 0 {
		hitRate = float64(opt.cacheHits) / float64(totalRequests) * 100
	}
	
	return map[string]interface{}{
		"cache_hits":      opt.cacheHits,
		"cache_misses":    opt.cacheMisses,
		"cache_hit_rate":  hitRate,
		"cache_size":      cacheSize,
		"max_workers":     opt.maxWorkers,
		"buffer_size":     opt.workerBufferSize,
		"cache_max_size":  opt.cacheMaxSize,
		"cache_ttl_minutes": opt.cacheTTL.Minutes(),
	}
}

// ResetMetrics resets performance metrics
func (opt *RustPerformanceOptimizer) ResetMetrics() {
	opt.metricsMutex.Lock()
	defer opt.metricsMutex.Unlock()
	
	opt.cacheHits = 0
	opt.cacheMisses = 0
}

// ClearCache clears the entire AST cache
func (opt *RustPerformanceOptimizer) ClearCache() {
	opt.cacheMutex.Lock()
	defer opt.cacheMutex.Unlock()
	
	opt.astCache = make(map[string]*RustASTCacheEntry)
}

// incrementCacheHits increments cache hit counter (thread-safe)
func (opt *RustPerformanceOptimizer) incrementCacheHits() {
	opt.metricsMutex.Lock()
	defer opt.metricsMutex.Unlock()
	opt.cacheHits++
}

// incrementCacheMisses increments cache miss counter (thread-safe)
func (opt *RustPerformanceOptimizer) incrementCacheMisses() {
	opt.metricsMutex.Lock()
	defer opt.metricsMutex.Unlock()
	opt.cacheMisses++
}

// ProcessRustFilesInParallel processes multiple Rust files using worker pools
func (opt *RustPerformanceOptimizer) ProcessRustFilesInParallel(
	files []*models.FileInfo,
	analyzer *RustASTAnalyzer,
	processFunc func(*models.FileInfo, *types.RustASTInfo) (*models.ScanResult, error),
) ([]*models.ScanResult, error) {
	
	filesChan := make(chan *models.FileInfo, opt.workerBufferSize)
	resultsChan := make(chan *models.ScanResult, opt.workerBufferSize)
	errorsChan := make(chan error, opt.workerBufferSize)
	
	var wg sync.WaitGroup
	
	// Start workers
	for i := 0; i < opt.maxWorkers; i++ {
		wg.Add(1)
		go opt.rustParsingWorker(&wg, filesChan, resultsChan, errorsChan, analyzer, processFunc)
	}
	
	// Send files to workers
	go func() {
		defer close(filesChan)
		for _, file := range files {
			filesChan <- file
		}
	}()
	
	// Wait for workers to complete
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errorsChan)
	}()
	
	// Collect results
	var results []*models.ScanResult
	var errors []error
	
	for result := range resultsChan {
		results = append(results, result)
	}
	
	for err := range errorsChan {
		errors = append(errors, err)
	}
	
	if len(errors) > 0 {
		// Return first error for simplicity
		return results, errors[0]
	}
	
	return results, nil
}

// rustParsingWorker is a worker function for parallel Rust file processing
func (opt *RustPerformanceOptimizer) rustParsingWorker(
	wg *sync.WaitGroup,
	filesChan <-chan *models.FileInfo,
	resultsChan chan<- *models.ScanResult,
	errorsChan chan<- error,
	analyzer *RustASTAnalyzer,
	processFunc func(*models.FileInfo, *types.RustASTInfo) (*models.ScanResult, error),
) {
	defer wg.Done()
	
	for file := range filesChan {
		// Process the file
		result, err := opt.processRustFileOptimized(file, analyzer, processFunc)
		if err != nil {
			errorsChan <- err
			continue
		}
		
		resultsChan <- result
	}
}

// processRustFileOptimized processes a single Rust file with caching and optimization
func (opt *RustPerformanceOptimizer) processRustFileOptimized(
	file *models.FileInfo,
	analyzer *RustASTAnalyzer,
	processFunc func(*models.FileInfo, *types.RustASTInfo) (*models.ScanResult, error),
) (*models.ScanResult, error) {
	
	// Read file content
	content, err := readFileForHashing(file.Path)
	if err != nil {
		return nil, err
	}
	
	// Calculate content hash for cache validation
	contentHash := opt.CalculateContentHash(content)
	
	// Try to get from cache first
	var astInfo *types.RustASTInfo
	if cachedAST := opt.GetCachedAST(file.Path, contentHash); cachedAST != nil {
		astInfo = cachedAST
	} else {
		// Parse the file
		astInfo, err = analyzer.AnalyzeRustFile(file.Path, content)
		if err != nil {
			return nil, err
		}
		
		// Cache the result
		opt.CacheAST(file.Path, astInfo, contentHash)
	}
	
	// Process the file using the provided function
	return processFunc(file, astInfo)
}

// readFileForHashing reads file content for hash calculation
func readFileForHashing(filePath string) ([]byte, error) {
	// Use the same method as the parser for consistency
	parser := &Parser{}
	return parser.readFileOptimized(filePath)
}

// EstimateMemoryUsage estimates memory usage for current cache state
func (opt *RustPerformanceOptimizer) EstimateMemoryUsage() map[string]interface{} {
	opt.cacheMutex.RLock()
	defer opt.cacheMutex.RUnlock()
	
	const (
		// Rough estimates in bytes
		astInfoBaseSize     = 200
		functionInfoSize    = 150
		structInfoSize      = 100
		enumInfoSize        = 80
		traitInfoSize       = 100
		implInfoSize        = 80
		moduleInfoSize      = 80
		constantInfoSize    = 100
		useInfoSize         = 80
		macroInfoSize       = 100
		cacheEntryOverhead  = 64
	)
	
	totalMemory := 0
	for _, entry := range opt.astCache {
		entrySize := astInfoBaseSize + cacheEntryOverhead
		entrySize += len(entry.ASTInfo.Functions) * functionInfoSize
		entrySize += len(entry.ASTInfo.Structs) * structInfoSize
		entrySize += len(entry.ASTInfo.Enums) * enumInfoSize
		entrySize += len(entry.ASTInfo.Traits) * traitInfoSize
		entrySize += len(entry.ASTInfo.Impls) * implInfoSize
		entrySize += len(entry.ASTInfo.Modules) * moduleInfoSize
		entrySize += len(entry.ASTInfo.Constants) * constantInfoSize
		entrySize += len(entry.ASTInfo.Uses) * useInfoSize
		entrySize += len(entry.ASTInfo.Macros) * macroInfoSize
		
		totalMemory += entrySize
	}
	
	return map[string]interface{}{
		"estimated_cache_memory_bytes": totalMemory,
		"estimated_cache_memory_mb":    float64(totalMemory) / (1024 * 1024),
		"cache_entries":                len(opt.astCache),
		"avg_memory_per_entry_bytes":   func() int {
			if len(opt.astCache) == 0 {
				return 0
			}
			return totalMemory / len(opt.astCache)
		}(),
	}
}

// MemoryPool provides enhanced memory pooling for CGO operations
type MemoryPool struct {
	stringPool    sync.Pool
	bufferPool    sync.Pool
	maxBufferSize int
}

// NewMemoryPool creates a new memory pool for efficient memory management
func NewMemoryPool() *MemoryPool {
	return &MemoryPool{
		maxBufferSize: 1024 * 1024, // 1MB max buffer size
		stringPool: sync.Pool{
			New: func() interface{} {
				// Pre-allocate small strings to reduce allocations
				return make([]string, 0, 10)
			},
		},
		bufferPool: sync.Pool{
			New: func() interface{} {
				// Pre-allocate 64KB buffers
				return make([]byte, 0, 64*1024)
			},
		},
	}
}

// GetStringSlice gets a reusable string slice from the pool
func (mp *MemoryPool) GetStringSlice() []string {
	slice := mp.stringPool.Get().([]string)
	return slice[:0] // Reset length but keep capacity
}

// PutStringSlice returns a string slice to the pool
func (mp *MemoryPool) PutStringSlice(slice []string) {
	if cap(slice) > 100 { // Don't pool excessively large slices
		return
	}
	mp.stringPool.Put(slice)
}

// GetBuffer gets a reusable byte buffer from the pool
func (mp *MemoryPool) GetBuffer() []byte {
	buffer := mp.bufferPool.Get().([]byte)
	return buffer[:0] // Reset length but keep capacity
}

// PutBuffer returns a byte buffer to the pool
func (mp *MemoryPool) PutBuffer(buffer []byte) {
	if cap(buffer) > mp.maxBufferSize {
		return // Don't pool excessively large buffers
	}
	mp.bufferPool.Put(buffer)
}

// BenchmarkResult represents the result of a performance benchmark
type BenchmarkResult struct {
	Operation      string        `json:"operation"`
	FileSize       int64         `json:"file_size_bytes"`
	Duration       time.Duration `json:"duration"`
	ThroughputMBps float64       `json:"throughput_mbps"`
	MemoryUsed     int64         `json:"memory_used_bytes"`
	Success        bool          `json:"success"`
	Error          string        `json:"error,omitempty"`
	Timestamp      time.Time     `json:"timestamp"`
}

// PerformanceProfiler provides detailed performance profiling for Rust parsing
type PerformanceProfiler struct {
	benchmarks []BenchmarkResult
	mutex      sync.RWMutex
	memPool    *MemoryPool
}

// NewPerformanceProfiler creates a new performance profiler
func NewPerformanceProfiler() *PerformanceProfiler {
	return &PerformanceProfiler{
		benchmarks: make([]BenchmarkResult, 0, 100),
		memPool:    NewMemoryPool(),
	}
}

// BenchmarkParsingOperation benchmarks a single parsing operation
func (pp *PerformanceProfiler) BenchmarkParsingOperation(
	operation string,
	content []byte,
	parseFunc func([]byte) error,
) BenchmarkResult {
	
	var memStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memStats)
	startMem := memStats.Alloc
	
	start := time.Now()
	err := parseFunc(content)
	duration := time.Since(start)
	
	runtime.ReadMemStats(&memStats)
	endMem := memStats.Alloc
	
	var memUsed int64
	if endMem > startMem {
		memUsed = int64(endMem - startMem)
	}
	
	fileSize := int64(len(content))
	throughput := float64(0)
	if duration.Nanoseconds() > 0 {
		throughput = float64(fileSize) / duration.Seconds() / (1024 * 1024) // MB/s
	}
	
	result := BenchmarkResult{
		Operation:      operation,
		FileSize:       fileSize,
		Duration:       duration,
		ThroughputMBps: throughput,
		MemoryUsed:     memUsed,
		Success:        err == nil,
		Timestamp:      time.Now(),
	}
	
	if err != nil {
		result.Error = err.Error()
	}
	
	pp.mutex.Lock()
	pp.benchmarks = append(pp.benchmarks, result)
	// Keep only last 100 benchmarks to prevent unbounded growth
	if len(pp.benchmarks) > 100 {
		pp.benchmarks = pp.benchmarks[len(pp.benchmarks)-100:]
	}
	pp.mutex.Unlock()
	
	return result
}

// GetBenchmarkSummary returns a summary of recent benchmark results
func (pp *PerformanceProfiler) GetBenchmarkSummary() map[string]interface{} {
	pp.mutex.RLock()
	defer pp.mutex.RUnlock()
	
	if len(pp.benchmarks) == 0 {
		return map[string]interface{}{
			"benchmark_count": 0,
		}
	}
	
	var (
		totalDuration     time.Duration
		totalThroughput   float64
		totalMemory       int64
		successCount      int
		avgFileSize       int64
		maxThroughput     float64
		minThroughput     float64 = math.MaxFloat64
	)
	
	for _, bench := range pp.benchmarks {
		totalDuration += bench.Duration
		totalThroughput += bench.ThroughputMBps
		totalMemory += bench.MemoryUsed
		avgFileSize += bench.FileSize
		
		if bench.Success {
			successCount++
		}
		
		if bench.ThroughputMBps > maxThroughput {
			maxThroughput = bench.ThroughputMBps
		}
		
		if bench.ThroughputMBps < minThroughput && bench.ThroughputMBps > 0 {
			minThroughput = bench.ThroughputMBps
		}
	}
	
	benchmarkCount := len(pp.benchmarks)
	successRate := float64(successCount) / float64(benchmarkCount) * 100
	
	return map[string]interface{}{
		"benchmark_count":       benchmarkCount,
		"success_rate_percent":  successRate,
		"avg_duration_ms":       float64(totalDuration.Nanoseconds()) / float64(benchmarkCount) / 1e6,
		"avg_throughput_mbps":   totalThroughput / float64(benchmarkCount),
		"max_throughput_mbps":   maxThroughput,
		"min_throughput_mbps":   func() float64 {
			if minThroughput == float64(^uint64(0)>>1) {
				return 0
			}
			return minThroughput
		}(),
		"avg_memory_kb":         float64(totalMemory) / float64(benchmarkCount) / 1024,
		"avg_file_size_kb":      float64(avgFileSize) / float64(benchmarkCount) / 1024,
		"total_duration_ms":     float64(totalDuration.Nanoseconds()) / 1e6,
	}
}

// GetRecentBenchmarks returns the most recent benchmark results
func (pp *PerformanceProfiler) GetRecentBenchmarks(count int) []BenchmarkResult {
	pp.mutex.RLock()
	defer pp.mutex.RUnlock()
	
	if count <= 0 || len(pp.benchmarks) == 0 {
		return nil
	}
	
	if count > len(pp.benchmarks) {
		count = len(pp.benchmarks)
	}
	
	start := len(pp.benchmarks) - count
	result := make([]BenchmarkResult, count)
	copy(result, pp.benchmarks[start:])
	
	return result
}

// ClearBenchmarks clears all stored benchmark results
func (pp *PerformanceProfiler) ClearBenchmarks() {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	
	pp.benchmarks = pp.benchmarks[:0]
}