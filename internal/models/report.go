// Package models defines data structures for representing code analysis reports and violations.
package models

import (
	"time"
)

// Report represents a complete scan report
type Report struct {
	ID          string         `json:"id"`
	GeneratedAt time.Time      `json:"generated_at"`
	Config      *ReportConfig  `json:"config"`
	Summary     *ScanSummary   `json:"summary"`
	Files       []*ScanResult  `json:"files"`
	Statistics  *Statistics    `json:"statistics"`
}

// ReportConfig contains configuration settings used for this report
type ReportConfig struct {
	Paths        []string     `json:"paths"`
	FileTypes    []string     `json:"file_types"`
	Thresholds   *Thresholds  `json:"thresholds"`
	HTMLSettings *HTMLOptions `json:"html_settings"`
}

// Thresholds contains the clean code thresholds used for this scan
type Thresholds struct {
	FunctionLines        int `json:"function_lines"`
	CyclomaticComplexity int `json:"cyclomatic_complexity"`
	Parameters           int `json:"parameters"`
	NestingDepth         int `json:"nesting_depth"`
	ClassLines           int `json:"class_lines"`
}

// HTMLOptions contains HTML-specific report options
type HTMLOptions struct {
	AutoRefresh     bool   `json:"auto_refresh"`
	RefreshInterval int    `json:"refresh_interval"`
	Theme           string `json:"theme"`
}

// Statistics provides detailed statistics about violations and files
type Statistics struct {
	ViolationsByType     map[ViolationType]int    `json:"violations_by_type"`
	ViolationsBySeverity map[Severity]int         `json:"violations_by_severity"`
	FilesByLanguage      map[string]int           `json:"files_by_language"`
	TopViolatedFiles     []*FileViolationSummary  `json:"top_violated_files"`
	ViolationTrends      []*ViolationTrend        `json:"violation_trends"`
}

// FileViolationSummary summarizes violations for a single file
type FileViolationSummary struct {
	File            string                   `json:"file"`
	TotalViolations int                      `json:"total_violations"`
	ViolationsByType map[ViolationType]int   `json:"violations_by_type"`
	ViolationsBySeverity map[Severity]int    `json:"violations_by_severity"`
	Lines           int                      `json:"lines"`
}

// ViolationTrend represents historical trend data (for future use)
type ViolationTrend struct {
	Date        time.Time `json:"date"`
	Type        ViolationType `json:"type"`
	Count       int       `json:"count"`
}

// FileTreeNode represents a file tree structure for navigation
type FileTreeNode struct {
	Name        string          `json:"name"`
	Path        string          `json:"path"`
	Type        string          `json:"type"` // "file" or "directory"
	Size        int64           `json:"size,omitempty"`
	Violations  int             `json:"violations"`
	Children    []*FileTreeNode `json:"children,omitempty"`
	Language    string          `json:"language,omitempty"`
	Scanned     bool            `json:"scanned"`
}

// NewReport creates a new report with the given data
func NewReport(summary *ScanSummary, files []*ScanResult, config *ReportConfig) *Report {
	report := &Report{
		ID:          generateReportID(),
		GeneratedAt: time.Now(),
		Config:      config,
		Summary:     summary,
		Files:       files,
	}

	// Calculate statistics
	report.Statistics = calculateStatistics(files)

	return report
}

// generateReportID generates a unique ID for the report
func generateReportID() string {
	return time.Now().Format("20060102-150405")
}

// calculateStatistics calculates detailed statistics from scan results
func calculateStatistics(files []*ScanResult) *Statistics {
	stats := &Statistics{
		ViolationsByType:     make(map[ViolationType]int),
		ViolationsBySeverity: make(map[Severity]int),
		FilesByLanguage:      make(map[string]int),
		TopViolatedFiles:     make([]*FileViolationSummary, 0),
		ViolationTrends:      make([]*ViolationTrend, 0),
	}

	fileSummaries := make(map[string]*FileViolationSummary)

	// Process each file
	for _, file := range files {
		if file.File == nil {
			continue
		}

		// Count files by language
		stats.FilesByLanguage[file.File.Language]++

		// Process violations
		fileSummary := &FileViolationSummary{
			File:                 file.File.Path,
			TotalViolations:     len(file.Violations),
			ViolationsByType:    make(map[ViolationType]int),
			ViolationsBySeverity: make(map[Severity]int),
			Lines:               file.File.Lines,
		}

		for _, violation := range file.Violations {
			// Count by type
			stats.ViolationsByType[violation.Type]++
			fileSummary.ViolationsByType[violation.Type]++

			// Count by severity
			stats.ViolationsBySeverity[violation.Severity]++
			fileSummary.ViolationsBySeverity[violation.Severity]++
		}

		if fileSummary.TotalViolations > 0 {
			fileSummaries[file.File.Path] = fileSummary
		}
	}

	// Convert file summaries to slice and sort by violation count
	for _, summary := range fileSummaries {
		stats.TopViolatedFiles = append(stats.TopViolatedFiles, summary)
	}

	// Sort top violated files (simple sort by total violations)
	for i := 0; i < len(stats.TopViolatedFiles)-1; i++ {
		for j := 0; j < len(stats.TopViolatedFiles)-i-1; j++ {
			if stats.TopViolatedFiles[j].TotalViolations < stats.TopViolatedFiles[j+1].TotalViolations {
				stats.TopViolatedFiles[j], stats.TopViolatedFiles[j+1] = 
					stats.TopViolatedFiles[j+1], stats.TopViolatedFiles[j]
			}
		}
	}

	// Limit to top 10
	if len(stats.TopViolatedFiles) > 10 {
		stats.TopViolatedFiles = stats.TopViolatedFiles[:10]
	}

	return stats
}

// BuildFileTree creates a hierarchical file tree structure
func (r *Report) BuildFileTree() *FileTreeNode {
	root := &FileTreeNode{
		Name:     "root",
		Path:     "",
		Type:     "directory",
		Children: make([]*FileTreeNode, 0),
	}

	// Build tree structure from files
	for _, file := range r.Files {
		if file.File == nil {
			continue
		}

		r.addFileToTree(root, file)
	}

	return root
}

// addFileToTree adds a file to the file tree structure
func (r *Report) addFileToTree(root *FileTreeNode, scanResult *ScanResult) {
	// This would implement the logic to build a hierarchical tree
	// For now, we'll add a simple flat structure
	fileNode := &FileTreeNode{
		Name:       scanResult.File.Name,
		Path:       scanResult.File.Path,
		Type:       "file",
		Size:       scanResult.File.Size,
		Violations: len(scanResult.Violations),
		Language:   scanResult.File.Language,
		Scanned:    scanResult.File.Scanned,
	}

	root.Children = append(root.Children, fileNode)
}

// GetViolationsByFile returns violations grouped by file
func (r *Report) GetViolationsByFile() map[string][]*Violation {
	result := make(map[string][]*Violation)

	for _, file := range r.Files {
		if file.File != nil && len(file.Violations) > 0 {
			result[file.File.Path] = file.Violations
		}
	}

	return result
}

// GetSeverityColor returns the CSS class for a severity level
func (s Severity) GetColor() string {
	switch s {
	case SeverityLow:
		return "text-success"
	case SeverityMedium:
		return "text-warning"
	case SeverityHigh:
		return "text-danger"
	case SeverityCritical:
		return "text-danger fw-bold"
	default:
		return "text-muted"
	}
}

// GetTypeDisplayName returns a human-readable name for violation types
func (vt ViolationType) GetDisplayName() string {
	switch vt {
	case ViolationTypeFunctionLength:
		return "Long Functions"
	case ViolationTypeCyclomaticComplexity:
		return "Complex Functions"
	case ViolationTypeParameterCount:
		return "Too Many Parameters"
	case ViolationTypeNestingDepth:
		return "Deep Nesting"
	case ViolationTypeNaming:
		return "Naming Convention"
	case ViolationTypeClassSize:
		return "Large Classes"
	case ViolationTypeMissingDocumentation:
		return "Missing Documentation"
	case ViolationTypeMagicNumbers:
		return "Magic Numbers"
	case ViolationTypeDuplication:
		return "Code Duplication"
	default:
		return string(vt)
	}
}