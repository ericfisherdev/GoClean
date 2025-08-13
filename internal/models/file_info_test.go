package models

import (
	"testing"
	"time"
)

func TestFileInfoConstruction(t *testing.T) {
	now := time.Now()
	
	fileInfo := FileInfo{
		Path:         "/path/to/test.go",
		Name:         "test.go",
		Extension:    ".go",
		Size:         1024,
		Lines:        50,
		ModifiedTime: now,
		Language:     "Go",
		Scanned:      true,
		Error:        "",
	}

	// Test all fields are set correctly
	if fileInfo.Path != "/path/to/test.go" {
		t.Errorf("Expected path '/path/to/test.go', got %q", fileInfo.Path)
	}

	if fileInfo.Name != "test.go" {
		t.Errorf("Expected name 'test.go', got %q", fileInfo.Name)
	}

	if fileInfo.Extension != ".go" {
		t.Errorf("Expected extension '.go', got %q", fileInfo.Extension)
	}

	if fileInfo.Size != 1024 {
		t.Errorf("Expected size 1024, got %d", fileInfo.Size)
	}

	if fileInfo.Lines != 50 {
		t.Errorf("Expected lines 50, got %d", fileInfo.Lines)
	}

	if !fileInfo.ModifiedTime.Equal(now) {
		t.Errorf("Expected modified time %v, got %v", now, fileInfo.ModifiedTime)
	}

	if fileInfo.Language != "Go" {
		t.Errorf("Expected language 'Go', got %q", fileInfo.Language)
	}

	if !fileInfo.Scanned {
		t.Error("Expected scanned to be true")
	}

	if fileInfo.Error != "" {
		t.Errorf("Expected empty error, got %q", fileInfo.Error)
	}
}

func TestFileInfoWithError(t *testing.T) {
	fileInfo := FileInfo{
		Path:    "/path/to/broken.go",
		Name:    "broken.go",
		Scanned: false,
		Error:   "file not found",
	}

	if fileInfo.Scanned {
		t.Error("File with error should not be marked as scanned")
	}

	if fileInfo.Error != "file not found" {
		t.Errorf("Expected error 'file not found', got %q", fileInfo.Error)
	}
}

func TestScanResultConstruction(t *testing.T) {
	fileInfo := &FileInfo{
		Path:      "test.go",
		Name:      "test.go",
		Extension: ".go",
		Language:  "Go",
		Scanned:   true,
	}

	violations := []*Violation{
		{
			ID:       "violation-1",
			Type:     ViolationTypeFunctionLength,
			Severity: SeverityHigh,
			Message:  "Function too long",
		},
	}

	metrics := &FileMetrics{
		TotalLines:      100,
		CodeLines:       80,
		CommentLines:    15,
		BlankLines:      5,
		FunctionCount:   3,
		ClassCount:      1,
		ComplexityScore: 15,
	}

	scanResult := ScanResult{
		File:       fileInfo,
		Violations: violations,
		Metrics:    metrics,
	}

	if scanResult.File != fileInfo {
		t.Error("ScanResult file should match provided FileInfo")
	}

	if len(scanResult.Violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(scanResult.Violations))
	}

	if scanResult.Violations[0].ID != "violation-1" {
		t.Errorf("Expected violation ID 'violation-1', got %q", scanResult.Violations[0].ID)
	}

	if scanResult.Metrics != metrics {
		t.Error("ScanResult metrics should match provided FileMetrics")
	}
}

func TestFileMetricsConstruction(t *testing.T) {
	metrics := FileMetrics{
		TotalLines:      100,
		CodeLines:       70,
		CommentLines:    20,
		BlankLines:      10,
		FunctionCount:   5,
		ClassCount:      2,
		ComplexityScore: 25,
	}

	if metrics.TotalLines != 100 {
		t.Errorf("Expected total lines 100, got %d", metrics.TotalLines)
	}

	if metrics.CodeLines != 70 {
		t.Errorf("Expected code lines 70, got %d", metrics.CodeLines)
	}

	if metrics.CommentLines != 20 {
		t.Errorf("Expected comment lines 20, got %d", metrics.CommentLines)
	}

	if metrics.BlankLines != 10 {
		t.Errorf("Expected blank lines 10, got %d", metrics.BlankLines)
	}

	if metrics.FunctionCount != 5 {
		t.Errorf("Expected function count 5, got %d", metrics.FunctionCount)
	}

	if metrics.ClassCount != 2 {
		t.Errorf("Expected class count 2, got %d", metrics.ClassCount)
	}

	if metrics.ComplexityScore != 25 {
		t.Errorf("Expected complexity score 25, got %d", metrics.ComplexityScore)
	}

	// Test that line counts add up
	totalCalculated := metrics.CodeLines + metrics.CommentLines + metrics.BlankLines
	if totalCalculated != metrics.TotalLines {
		t.Errorf("Line counts don't add up: %d + %d + %d = %d, expected %d",
			metrics.CodeLines, metrics.CommentLines, metrics.BlankLines,
			totalCalculated, metrics.TotalLines)
	}
}

func TestScanSummaryConstruction(t *testing.T) {
	startTime := time.Now().Add(-time.Minute)
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	violationsByType := map[string]int{
		"function_length": 5,
		"naming":         3,
		"complexity":     2,
	}

	summary := ScanSummary{
		TotalFiles:       20,
		ScannedFiles:     18,
		SkippedFiles:     2,
		TotalViolations:  10,
		ViolationsByType: violationsByType,
		StartTime:        startTime,
		EndTime:          endTime,
		Duration:         duration,
	}

	if summary.TotalFiles != 20 {
		t.Errorf("Expected total files 20, got %d", summary.TotalFiles)
	}

	if summary.ScannedFiles != 18 {
		t.Errorf("Expected scanned files 18, got %d", summary.ScannedFiles)
	}

	if summary.SkippedFiles != 2 {
		t.Errorf("Expected skipped files 2, got %d", summary.SkippedFiles)
	}

	if summary.TotalViolations != 10 {
		t.Errorf("Expected total violations 10, got %d", summary.TotalViolations)
	}

	// Test that scanned + skipped equals total
	if summary.ScannedFiles+summary.SkippedFiles != summary.TotalFiles {
		t.Errorf("Scanned (%d) + Skipped (%d) should equal Total (%d)",
			summary.ScannedFiles, summary.SkippedFiles, summary.TotalFiles)
	}

	// Test violations by type
	if len(summary.ViolationsByType) != 3 {
		t.Errorf("Expected 3 violation types, got %d", len(summary.ViolationsByType))
	}

	expectedTotal := 0
	for _, count := range summary.ViolationsByType {
		expectedTotal += count
	}

	if expectedTotal != summary.TotalViolations {
		t.Errorf("Sum of violations by type (%d) should equal total violations (%d)",
			expectedTotal, summary.TotalViolations)
	}

	// Test timing
	if !summary.StartTime.Equal(startTime) {
		t.Errorf("Expected start time %v, got %v", startTime, summary.StartTime)
	}

	if !summary.EndTime.Equal(endTime) {
		t.Errorf("Expected end time %v, got %v", endTime, summary.EndTime)
	}

	if summary.Duration != duration {
		t.Errorf("Expected duration %v, got %v", duration, summary.Duration)
	}

	if summary.EndTime.Before(summary.StartTime) {
		t.Error("End time should be after start time")
	}
}

func TestEmptyScanSummary(t *testing.T) {
	summary := ScanSummary{
		TotalFiles:       0,
		ScannedFiles:     0,
		SkippedFiles:     0,
		TotalViolations:  0,
		ViolationsByType: make(map[string]int),
		StartTime:        time.Now(),
		EndTime:          time.Now(),
		Duration:         0,
	}

	if summary.TotalFiles != 0 {
		t.Errorf("Expected 0 total files, got %d", summary.TotalFiles)
	}

	if summary.TotalViolations != 0 {
		t.Errorf("Expected 0 violations, got %d", summary.TotalViolations)
	}

	if len(summary.ViolationsByType) != 0 {
		t.Errorf("Expected empty violations map, got %d entries", len(summary.ViolationsByType))
	}
}

func TestFileInfoJSONTags(t *testing.T) {
	// Test that we can create FileInfo with all fields
	fileInfo := FileInfo{
		Path:         "test.go",
		Name:         "test.go",
		Extension:    ".go",
		Size:         1024,
		Lines:        50,
		ModifiedTime: time.Now(),
		Language:     "Go",
		Scanned:      true,
		Error:        "", // Should be omitted in JSON if empty due to omitempty tag
	}

	// Basic validation
	if fileInfo.Path == "" {
		t.Error("FileInfo Path should not be empty")
	}

	if fileInfo.Extension == "" {
		t.Error("FileInfo Extension should not be empty")
	}

	if fileInfo.Language == "" {
		t.Error("FileInfo Language should not be empty")
	}
}

func TestScanResultWithEmptyViolations(t *testing.T) {
	fileInfo := &FileInfo{
		Path:     "clean.go",
		Language: "Go",
		Scanned:  true,
	}

	metrics := &FileMetrics{
		TotalLines: 10,
		CodeLines:  8,
	}

	result := ScanResult{
		File:       fileInfo,
		Violations: []*Violation{}, // Empty violations slice
		Metrics:    metrics,
	}

	if len(result.Violations) != 0 {
		t.Errorf("Expected 0 violations, got %d", len(result.Violations))
	}

	// Should still have valid file and metrics
	if result.File == nil {
		t.Error("ScanResult should have valid file info")
	}

	if result.Metrics == nil {
		t.Error("ScanResult should have valid metrics")
	}
}

func TestFileMetricsEdgeCases(t *testing.T) {
	// Test zero values
	metrics := FileMetrics{}

	if metrics.TotalLines != 0 {
		t.Errorf("Expected 0 total lines for zero value, got %d", metrics.TotalLines)
	}

	if metrics.FunctionCount != 0 {
		t.Errorf("Expected 0 function count for zero value, got %d", metrics.FunctionCount)
	}

	// Test with only comments
	commentOnlyMetrics := FileMetrics{
		TotalLines:   5,
		CommentLines: 5,
		CodeLines:    0,
		BlankLines:   0,
	}

	if commentOnlyMetrics.CodeLines != 0 {
		t.Errorf("Expected 0 code lines for comment-only file, got %d", commentOnlyMetrics.CodeLines)
	}

	if commentOnlyMetrics.CommentLines != 5 {
		t.Errorf("Expected 5 comment lines, got %d", commentOnlyMetrics.CommentLines)
	}
}