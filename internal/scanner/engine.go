package scanner

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/violations"
)

// Engine is the main scanning orchestrator
type Engine struct {
	fileWalker         *FileWalker
	parser             *Parser
	violationDetector  *ViolationDetector
	verbose            bool
	maxWorkers         int
	progressFn         func(string)
	realTimeMode       bool
	workerBufferSize   int
}

// NewEngine creates a new scanning engine
func NewEngine(includePaths []string, excludePatterns []string, fileTypes []string, verbose bool) *Engine {
	numCPU := runtime.NumCPU()
	return &Engine{
		fileWalker:        NewFileWalker(includePaths, excludePatterns, fileTypes, verbose),
		parser:            NewParser(verbose),
		violationDetector: NewViolationDetector(violations.DefaultDetectorConfig()),
		verbose:           verbose,
		maxWorkers:        numCPU, // Default to number of CPU cores
		workerBufferSize:  numCPU * 2, // Buffer size for better throughput
	}
}

// SetMaxWorkers sets the maximum number of concurrent workers
func (e *Engine) SetMaxWorkers(workers int) {
	if workers > 0 {
		e.maxWorkers = workers
	}
}

// Scan performs the complete scanning operation, with optional progress reporting.
func (e *Engine) Scan() (*models.ScanSummary, []*models.ScanResult, error) {
	startTime := time.Now()

	// Reset violation detector caches for new scan
	e.violationDetector.ResetDuplicationCache()

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
		if e.progressFn != nil && processed%10 == 0 { // Update every 10 files
			percentage := float64(processed) / float64(total) * 100
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
			errorsChan <- fmt.Errorf("failed to parse %s: %w", file.Path, err)
		} else {
			e.violationDetector.DetectViolations(result)
		}
		resultsChan <- result
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

