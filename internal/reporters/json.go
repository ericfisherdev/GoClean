package reporters

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ericfisherdev/goclean/internal/config"
	"github.com/ericfisherdev/goclean/internal/models"
)

// JSONReporter generates JSON reports for analysis results
type JSONReporter struct {
	config *config.JSONConfig
}

// NewJSONReporter creates a new JSON reporter with the given configuration
func NewJSONReporter(cfg *config.JSONConfig) *JSONReporter {
	if cfg == nil {
		cfg = &config.JSONConfig{
			Enabled:     true,
			Path:        "./reports/violations.json",
			PrettyPrint: true,
		}
	}
	return &JSONReporter{
		config: cfg,
	}
}

// Generate generates a JSON report from the report data
func (r *JSONReporter) Generate(report *models.Report) error {
	if !r.config.Enabled {
		return nil // Skip generation if disabled
	}

	// Create the report directory if it doesn't exist
	dir := filepath.Dir(r.config.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	// Extract all violations from scan results
	allViolations := r.extractViolations(report.Files)
	
	// Create file to language mapping
	fileLanguages := r.createFileLanguageMap(report.Files)

	// Create the JSON report data structure
	jsonData := &JSONReport{
		Metadata: JSONMetadata{
			GeneratedAt:     report.GeneratedAt,
			GoCleanVersion:  "dev", // TODO: Get actual version
			FilesScanned:    report.Summary.ScannedFiles,
			TotalViolations: report.Summary.TotalViolations,
			ScanDuration:    report.Summary.Duration,
		},
		Summary: r.generateSummary(allViolations, fileLanguages),
		Violations: r.convertViolations(allViolations, fileLanguages),
		Statistics: r.generateStatistics(allViolations, fileLanguages),
	}

	// Marshal to JSON
	var jsonBytes []byte
	var err error
	if r.config.PrettyPrint {
		jsonBytes, err = json.MarshalIndent(jsonData, "", "  ")
	} else {
		jsonBytes, err = json.Marshal(jsonData)
	}
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(r.config.Path, jsonBytes, 0644); err != nil {
		return fmt.Errorf("failed to write JSON report: %w", err)
	}

	return nil
}

// Format returns the format name of this reporter
func (r *JSONReporter) Format() string {
	return "json"
}

// JSONReport represents the structure of the JSON report
type JSONReport struct {
	Metadata   JSONMetadata        `json:"metadata"`
	Summary    JSONSummary         `json:"summary"`
	Violations []JSONViolation     `json:"violations"`
	Statistics JSONStatistics      `json:"statistics"`
}

// JSONMetadata contains metadata about the report
type JSONMetadata struct {
	GeneratedAt     time.Time     `json:"generated_at"`
	GoCleanVersion  string        `json:"goclean_version"`
	FilesScanned    int           `json:"files_scanned"`
	TotalViolations int           `json:"total_violations"`
	ScanDuration    time.Duration `json:"scan_duration"`
}

// JSONSummary contains a summary of violations by severity and type
type JSONSummary struct {
	BySeverity map[string]int `json:"by_severity"`
	ByType     map[string]int `json:"by_type"`
	ByLanguage map[string]int `json:"by_language"`
}

// JSONViolation represents a violation in the JSON report
type JSONViolation struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Message     string `json:"message"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Column      int    `json:"column"`
	Language    string `json:"language"`
	Suggestion  string `json:"suggestion,omitempty"`
	CodeSnippet string `json:"code_snippet,omitempty"`
}

// JSONStatistics contains statistical information about the violations
type JSONStatistics struct {
	TotalFiles       int     `json:"total_files"`
	FilesWithViolations int  `json:"files_with_violations"`
	AverageViolationsPerFile float64 `json:"average_violations_per_file"`
	MostCommonViolationType string  `json:"most_common_violation_type"`
	MostCommonSeverity      string  `json:"most_common_severity"`
	LanguageBreakdown       map[string]JSONLanguageStats `json:"language_breakdown"`
}

// JSONLanguageStats contains statistics for a specific language
type JSONLanguageStats struct {
	FilesScanned int     `json:"files_scanned"`
	Violations   int     `json:"violations"`
	AvgPerFile   float64 `json:"avg_per_file"`
}

// extractViolations extracts all violations from scan results
func (r *JSONReporter) extractViolations(files []*models.ScanResult) []*models.Violation {
	var allViolations []*models.Violation
	for _, file := range files {
		allViolations = append(allViolations, file.Violations...)
	}
	return allViolations
}

// createFileLanguageMap creates a mapping from file paths to languages
func (r *JSONReporter) createFileLanguageMap(files []*models.ScanResult) map[string]string {
	fileLanguages := make(map[string]string)
	for _, file := range files {
		if file.File != nil {
			fileLanguages[file.File.Path] = file.File.Language
		}
	}
	return fileLanguages
}

// generateSummary creates a summary of violations by severity, type, and language
func (r *JSONReporter) generateSummary(violations []*models.Violation, fileLanguages map[string]string) JSONSummary {
	summary := JSONSummary{
		BySeverity: make(map[string]int),
		ByType:     make(map[string]int),
		ByLanguage: make(map[string]int),
	}

	for _, violation := range violations {
		// Count by severity
		summary.BySeverity[violation.Severity.String()]++
		
		// Count by type
		summary.ByType[string(violation.Type)]++
		
		// Count by language (get from file info)
		if language, exists := fileLanguages[violation.File]; exists {
			summary.ByLanguage[language]++
		}
	}

	return summary
}

// convertViolations converts models.Violation to JSONViolation
func (r *JSONReporter) convertViolations(violations []*models.Violation, fileLanguages map[string]string) []JSONViolation {
	jsonViolations := make([]JSONViolation, len(violations))
	
	for i, v := range violations {
		language := "Unknown"
		if lang, exists := fileLanguages[v.File]; exists {
			language = lang
		}
		
		jsonViolations[i] = JSONViolation{
			ID:          v.ID,
			Type:        string(v.Type),
			Severity:    v.Severity.String(),
			Message:     v.Message,
			File:        v.File,
			Line:        v.Line,
			Column:      v.Column,
			Language:    language,
			Suggestion:  v.Suggestion,
			CodeSnippet: v.CodeSnippet,
			// Note: Metadata removed as it doesn't exist in models.Violation
		}
	}
	
	return jsonViolations
}

// generateStatistics generates statistical information about the violations
func (r *JSONReporter) generateStatistics(violations []*models.Violation, fileLanguages map[string]string) JSONStatistics {
	stats := JSONStatistics{
		LanguageBreakdown: make(map[string]JSONLanguageStats),
	}

	// Track files and violations by language
	filesByLanguage := make(map[string]map[string]bool)
	violationsByLanguage := make(map[string]int)
	violationsByType := make(map[string]int)
	violationsBySeverity := make(map[string]int)

	for _, v := range violations {
		language := "Unknown"
		if lang, exists := fileLanguages[v.File]; exists {
			language = lang
		}
		
		// Track files by language
		if filesByLanguage[language] == nil {
			filesByLanguage[language] = make(map[string]bool)
		}
		filesByLanguage[language][v.File] = true
		
		// Count violations by language, type, and severity
		violationsByLanguage[language]++
		violationsByType[string(v.Type)]++
		violationsBySeverity[v.Severity.String()]++
	}

	// Calculate language breakdown
	totalFiles := 0
	for language, files := range filesByLanguage {
		fileCount := len(files)
		totalFiles += fileCount
		
		violationCount := violationsByLanguage[language]
		avgPerFile := 0.0
		if fileCount > 0 {
			avgPerFile = float64(violationCount) / float64(fileCount)
		}
		
		stats.LanguageBreakdown[language] = JSONLanguageStats{
			FilesScanned: fileCount,
			Violations:   violationCount,
			AvgPerFile:   avgPerFile,
		}
	}

	// Set overall statistics
	stats.TotalFiles = totalFiles
	stats.FilesWithViolations = totalFiles
	if totalFiles > 0 {
		stats.AverageViolationsPerFile = float64(len(violations)) / float64(totalFiles)
	}

	// Find most common violation type and severity
	stats.MostCommonViolationType = findMostCommon(violationsByType)
	stats.MostCommonSeverity = findMostCommon(violationsBySeverity)

	return stats
}

// findMostCommon finds the key with the highest value in a map
func findMostCommon(counts map[string]int) string {
	maxCount := 0
	mostCommon := ""
	
	for key, count := range counts {
		if count > maxCount {
			maxCount = count
			mostCommon = key
		}
	}
	
	return mostCommon
}