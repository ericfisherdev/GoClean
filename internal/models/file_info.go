package models

import (
	"time"
)

// FileInfo represents metadata about a scanned file
type FileInfo struct {
	Path         string    `json:"path"`
	Name         string    `json:"name"`
	Extension    string    `json:"extension"`
	Size         int64     `json:"size"`
	Lines        int       `json:"lines"`
	ModifiedTime time.Time `json:"modified_time"`
	Language     string    `json:"language"`
	Scanned      bool      `json:"scanned"`
	Error        string    `json:"error,omitempty"`
}

// ScanResult represents the result of scanning a single file
type ScanResult struct {
	File       *FileInfo    `json:"file"`
	Violations []*Violation `json:"violations"`
	Metrics    *FileMetrics `json:"metrics"`
}

// FileMetrics contains basic metrics about a file
type FileMetrics struct {
	TotalLines      int `json:"total_lines"`
	CodeLines       int `json:"code_lines"`
	CommentLines    int `json:"comment_lines"`
	BlankLines      int `json:"blank_lines"`
	FunctionCount   int `json:"function_count"`
	ClassCount      int `json:"class_count"`
	ComplexityScore int `json:"complexity_score"`
}

// ScanSummary provides an overview of the entire scan operation
type ScanSummary struct {
	TotalFiles       int           `json:"total_files"`
	ScannedFiles     int           `json:"scanned_files"`
	SkippedFiles     int           `json:"skipped_files"`
	TotalViolations  int           `json:"total_violations"`
	ViolationsByType map[string]int `json:"violations_by_type"`
	StartTime        time.Time     `json:"start_time"`
	EndTime          time.Time     `json:"end_time"`
	Duration         time.Duration `json:"duration"`
}