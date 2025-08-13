package scanner

import (
	"fmt"
	"sync"
	"time"

	"github.com/ericfisherdev/goclean/internal/models"
)

// Engine is the main scanning orchestrator
type Engine struct {
	fileWalker *FileWalker
	parser     *Parser
	verbose    bool
	maxWorkers int
}

// NewEngine creates a new scanning engine
func NewEngine(includePaths []string, excludePatterns []string, fileTypes []string, verbose bool) *Engine {
	return &Engine{
		fileWalker: NewFileWalker(includePaths, excludePatterns, fileTypes, verbose),
		parser:     NewParser(verbose),
		verbose:    verbose,
		maxWorkers: 10, // Default number of concurrent workers
	}
}

// SetMaxWorkers sets the maximum number of concurrent workers
func (e *Engine) SetMaxWorkers(workers int) {
	if workers > 0 {
		e.maxWorkers = workers
	}
}

// Scan performs the complete scanning operation
func (e *Engine) Scan() (*models.ScanSummary, []*models.ScanResult, error) {
	startTime := time.Now()
	
	if e.verbose {
		fmt.Println("Starting file discovery...")
	}
	
	// Discover files
	files, err := e.fileWalker.Walk()
	if err != nil {
		return nil, nil, fmt.Errorf("file discovery failed: %w", err)
	}
	
	if len(files) == 0 {
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
	
	if e.verbose {
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
	
	if e.verbose {
		fmt.Printf("Scan completed in %v\n", summary.Duration)
		fmt.Printf("Files scanned: %d/%d\n", summary.ScannedFiles, summary.TotalFiles)
		fmt.Printf("Total violations: %d\n", summary.TotalViolations)
	}
	
	return summary, results, nil
}

// scanFiles scans multiple files concurrently
func (e *Engine) scanFiles(files []*models.FileInfo) ([]*models.ScanResult, error) {
	// Create channels for work distribution
	filesChan := make(chan *models.FileInfo, len(files))
	resultsChan := make(chan *models.ScanResult, len(files))
	errorsChan := make(chan error, len(files))
	
	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < e.maxWorkers; i++ {
		wg.Add(1)
		go e.worker(&wg, filesChan, resultsChan, errorsChan)
	}
	
	// Send files to workers
	for _, file := range files {
		filesChan <- file
	}
	close(filesChan)
	
	// Wait for workers to complete
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errorsChan)
	}()
	
	// Collect results
	var results []*models.ScanResult
	var errors []error
	
	// Collect results
	for result := range resultsChan {
		results = append(results, result)
	}
	
	// Collect errors
	for err := range errorsChan {
		errors = append(errors, err)
	}
	
	// Report errors if any
	if len(errors) > 0 && e.verbose {
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
	
	return results, nil
}

// worker processes files from the input channel
func (e *Engine) worker(wg *sync.WaitGroup, filesChan <-chan *models.FileInfo, resultsChan chan<- *models.ScanResult, errorsChan chan<- error) {
	defer wg.Done()
	
	for file := range filesChan {
		result, err := e.parser.ParseFile(file)
		if err != nil {
			// Mark file with error
			file.Error = err.Error()
			file.Scanned = false
			
			// Create result with error
			result = &models.ScanResult{
				File:       file,
				Violations: []*models.Violation{},
				Metrics:    &models.FileMetrics{},
			}
			
			errorsChan <- fmt.Errorf("failed to parse %s: %w", file.Path, err)
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