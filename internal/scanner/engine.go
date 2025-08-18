// Package scanner provides the core scanning engine for GoClean.
// It orchestrates file discovery, parsing, AST analysis, and violation detection
// across multiple files and directories with concurrent processing capabilities.
package scanner

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/violations"
)

// Scanning progress constants
const (
	ProgressUpdateInterval = 10  // Update progress every N files
	PercentageMultiplier   = 100 // For percentage calculations
)

// Engine is the main scanning orchestrator
type Engine struct {
	fileWalker           *FileWalker
	parser               *Parser
	violationDetector    *ViolationDetector
	verbose              bool
	maxWorkers           int
	progressFn           func(string)
	realTimeMode         bool
	workerBufferSize     int
	rustOptimizer        *RustPerformanceOptimizer
	enableRustOptimization bool
}

// NewEngine creates a new scanning engine
func NewEngine(includePaths []string, excludePatterns []string, fileTypes []string, verbose bool) *Engine {
	numCPU := runtime.NumCPU()
	rustOptimizer := NewRustPerformanceOptimizer(verbose)
	
	return &Engine{
		fileWalker:             NewFileWalker(includePaths, excludePatterns, fileTypes, verbose),
		parser:                 NewParserWithOptimizer(verbose, rustOptimizer),
		violationDetector:      NewViolationDetector(violations.DefaultDetectorConfig()),
		verbose:                verbose,
		maxWorkers:             numCPU, // Default to number of CPU cores
		workerBufferSize:       numCPU * 2, // Buffer size for better throughput
		rustOptimizer:          rustOptimizer,
		enableRustOptimization: true,
	}
}

// NewEngineWithConfig creates a new scanning engine with test file configuration
func NewEngineWithConfig(includePaths []string, excludePatterns []string, fileTypes []string, verbose bool,
	skipTestFiles bool, aggressiveMode bool, customTestPatterns []string) *Engine {
	numCPU := runtime.NumCPU()
	rustOptimizer := NewRustPerformanceOptimizer(verbose)
	
	// Create detector config with test file awareness
	detectorConfig := violations.DefaultDetectorConfig()
	detectorConfig.AggressiveMode = aggressiveMode
	
	return &Engine{
		fileWalker:             NewFileWalkerWithConfig(includePaths, excludePatterns, fileTypes, verbose, skipTestFiles, aggressiveMode, customTestPatterns),
		parser:                 NewParserWithOptimizer(verbose, rustOptimizer),
		violationDetector:      NewViolationDetector(detectorConfig),
		verbose:                verbose,
		maxWorkers:             numCPU,
		workerBufferSize:       numCPU * 2,
		rustOptimizer:          rustOptimizer,
		enableRustOptimization: true,
	}
}

// SetMaxWorkers sets the maximum number of concurrent workers
func (e *Engine) SetMaxWorkers(workers int) {
	if workers > 0 {
		e.maxWorkers = workers
		// Update Rust optimizer configuration if available
		if e.rustOptimizer != nil {
			e.rustOptimizer.SetWorkerConfiguration(workers, workers*2)
		}
	}
}

// SetMaxFileSize sets the maximum file size limit for scanning
func (e *Engine) SetMaxFileSize(maxSizeStr string) error {
	if e.fileWalker != nil {
		return e.fileWalker.SetMaxFileSize(maxSizeStr)
	}
	return nil
}

// EnableRustOptimization enables or disables Rust-specific performance optimizations
func (e *Engine) EnableRustOptimization(enabled bool) {
	e.enableRustOptimization = enabled
}

// SetRustCacheConfig configures the Rust AST cache
func (e *Engine) SetRustCacheConfig(maxSize int, ttl time.Duration) {
	if e.rustOptimizer != nil {
		e.rustOptimizer.SetCacheConfiguration(maxSize, ttl)
	}
}

// GetRustPerformanceMetrics returns performance metrics for Rust parsing
func (e *Engine) GetRustPerformanceMetrics() map[string]interface{} {
	if e.rustOptimizer != nil {
		return e.rustOptimizer.GetPerformanceMetrics()
	}
	return nil
}

// GetRustMemoryUsage returns memory usage estimates for Rust parsing cache
func (e *Engine) GetRustMemoryUsage() map[string]interface{} {
	if e.rustOptimizer != nil {
		return e.rustOptimizer.EstimateMemoryUsage()
	}
	return map[string]interface{}{
		"estimated_cache_memory_bytes": 0,
		"estimated_cache_memory_mb":    0.0,
		"cache_entries":                0,
		"avg_memory_per_entry_bytes":   0,
	}
}

// ClearRustCache clears the Rust AST cache
func (e *Engine) ClearRustCache() {
	if e.rustOptimizer != nil {
		e.rustOptimizer.ClearCache()
	}
}

// CleanupRustCache removes expired entries from the Rust cache
func (e *Engine) CleanupRustCache() {
	if e.rustOptimizer != nil {
		e.rustOptimizer.CleanupCache()
	}
}

// Scan performs the complete scanning operation, with optional progress reporting.
func (e *Engine) Scan() (*models.ScanSummary, []*models.ScanResult, error) {
	startTime := time.Now()

	// Reset violation detector caches for new scan
	e.violationDetector.ResetDuplicationCache()
	
	// Cleanup expired Rust cache entries if optimization is enabled
	if e.enableRustOptimization && e.rustOptimizer != nil {
		e.rustOptimizer.CleanupCache()
	}

	if e.progressFn != nil {
		e.progressFn("Starting file discovery...")
	} else if e.verbose {
		fmt.Println("Starting file discovery...")
	}

	// Discover files
	files, err := e.fileWalker.Walk()
	if err != nil {
		return nil, nil, fmt.Errorf("file discovery failed: %w", err)
	}

	if len(files) == 0 {
		if e.progressFn != nil {
			e.progressFn("No files found to scan")
		}
		summary := &models.ScanSummary{
			TotalFiles:       0,
			ScannedFiles:     0,
			SkippedFiles:     0,
			TotalViolations:  0,
			ViolationsByType: make(map[string]int),
			StartTime:        startTime,
			EndTime:          time.Now(),
			Duration:         time.Since(startTime),
		}
		return summary, []*models.ScanResult{}, nil
	}

	if e.progressFn != nil {
		e.progressFn(fmt.Sprintf("Scanning %d files with %d workers...", len(files), e.maxWorkers))
	} else if e.verbose {
		fmt.Printf("Scanning %d files with %d workers...\n", len(files), e.maxWorkers)
	}

	// Scan files concurrently
	results, err := e.scanFiles(files)
	if err != nil {
		return nil, nil, fmt.Errorf("file scanning failed: %w", err)
	}

	endTime := time.Now()

	// Generate summary
	summary := e.generateSummary(files, results, startTime, endTime)

	if e.progressFn != nil {
		e.progressFn(fmt.Sprintf("Scan completed: %d violations found in %v",
			summary.TotalViolations, summary.Duration.Round(time.Millisecond)))
	} else if e.verbose {
		fmt.Printf("Scan completed in %v\n", summary.Duration)
		fmt.Printf("Files scanned: %d/%d\n", summary.ScannedFiles, summary.TotalFiles)
		fmt.Printf("Total violations: %d\n", summary.TotalViolations)
	}

	return summary, results, nil
}

// scanFiles scans multiple files concurrently and reports progress.
func (e *Engine) scanFiles(files []*models.FileInfo) ([]*models.ScanResult, error) {
	filesChan := make(chan *models.FileInfo, e.workerBufferSize)
	resultsChan := make(chan *models.ScanResult, e.workerBufferSize)
	errorsChan := make(chan error, e.workerBufferSize)

	var wg sync.WaitGroup
	for i := 0; i < e.maxWorkers; i++ {
		wg.Add(1)
		go e.worker(&wg, filesChan, resultsChan, errorsChan)
	}

	go func() {
		for _, file := range files {
			filesChan <- file
		}
		close(filesChan)
	}()

	go func() {
		wg.Wait()
		close(resultsChan)
		close(errorsChan)
	}()

	var results []*models.ScanResult
	var errors []error
	processed := 0
	total := len(files)

	// Collect results
	for result := range resultsChan {
		results = append(results, result)
		processed++
		if e.progressFn != nil && processed%ProgressUpdateInterval == 0 { // Update every N files
			percentage := float64(processed) / float64(total) * PercentageMultiplier
			e.progressFn(fmt.Sprintf("Progress: %d/%d files (%.1f%%)", processed, total, percentage))
		}
	}

	// Collect errors
	for err := range errorsChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		if e.progressFn != nil {
			e.progressFn(fmt.Sprintf("Encountered %d errors during scanning", len(errors)))
		} else if e.verbose {
			fmt.Printf("Encountered %d errors during scanning:\n", len(errors))
			for i, err := range errors {
				if i < 5 { // Limit error output
					fmt.Printf("  - %v\n", err)
				}
			}
			if len(errors) > 5 {
				fmt.Printf("  ... and %d more errors\n", len(errors)-5)
			}
		}
	}

	return results, nil
}

// worker processes files from the input channel.
func (e *Engine) worker(wg *sync.WaitGroup, filesChan <-chan *models.FileInfo, resultsChan chan<- *models.ScanResult, errorsChan chan<- error) {
	defer wg.Done()

	for file := range filesChan {
		if e.progressFn != nil && e.verbose {
			e.progressFn(fmt.Sprintf("Scanning %s...", file.Name))
		}

		result, err := e.parser.ParseFile(file)
		if err != nil {
			file.Error = err.Error()
			file.Scanned = false
			result = &models.ScanResult{
				File:       file,
				Violations: []*models.Violation{},
				Metrics:    &models.FileMetrics{},
			}
		} else {
			e.violationDetector.DetectViolations(result)
		}
		
		// Always send result first to avoid blocking on the errors channel
		resultsChan <- result
		
		// Send error non-blocking (best-effort error reporting)
		if err != nil {
			select {
			case errorsChan <- fmt.Errorf("failed to parse %s: %w", file.Path, err):
				// Error successfully sent to error channel
			default:
				// Error channel is full, but error is already recorded in result.File.Error
				// so we can safely drop this duplicate error report
			}
		}
	}
}

// generateSummary creates a summary of the scan operation
func (e *Engine) generateSummary(files []*models.FileInfo, results []*models.ScanResult, startTime, endTime time.Time) *models.ScanSummary {
	summary := &models.ScanSummary{
		TotalFiles:       len(files),
		ScannedFiles:     0,
		SkippedFiles:     0,
		TotalViolations:  0,
		ViolationsByType: make(map[string]int),
		StartTime:        startTime,
		EndTime:          endTime,
		Duration:         endTime.Sub(startTime),
	}
	
	// Count scanned vs skipped files and violations
	for _, result := range results {
		if result.File.Scanned {
			summary.ScannedFiles++
		} else {
			summary.SkippedFiles++
		}
		
		// Count violations
		for _, violation := range result.Violations {
			summary.TotalViolations++
			summary.ViolationsByType[string(violation.Type)]++
		}
	}
	
	return summary
}

// SetProgressCallback sets a function to be called for progress updates
func (e *Engine) SetProgressCallback(fn func(string)) {
	e.progressFn = fn
}

// EnableRealTimeMode enables real-time report updates during scanning
func (e *Engine) EnableRealTimeMode(enabled bool) {
	e.realTimeMode = enabled
}

// SetViolationDetectorConfig sets a custom violation detector configuration
func (e *Engine) SetViolationDetectorConfig(config *violations.DetectorConfig) {
	e.violationDetector = NewViolationDetector(config)
}

