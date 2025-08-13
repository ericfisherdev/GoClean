package models

import (
	"testing"
	"time"
)

func TestNewReport(t *testing.T) {
	summary, files, config := createTestReportData()

	report := NewReport(summary, files, config)

	if report.ID == "" {
		t.Error("Report ID should not be empty")
	}

	if report.GeneratedAt.IsZero() {
		t.Error("GeneratedAt should be set")
	}

	if report.Config != config {
		t.Error("Report config not set correctly")
	}

	if report.Summary != summary {
		t.Error("Report summary not set correctly")
	}

	if len(report.Files) != len(files) {
		t.Errorf("Expected %d files, got %d", len(files), len(report.Files))
	}

	if report.Statistics == nil {
		t.Error("Statistics should be calculated")
	}
}

func TestReport_BuildFileTree(t *testing.T) {
	summary, files, config := createTestReportData()
	report := NewReport(summary, files, config)

	fileTree := report.BuildFileTree()

	if fileTree == nil {
		t.Fatal("File tree should not be nil")
	}

	if fileTree.Name != "root" {
		t.Errorf("Root node name should be 'root', got '%s'", fileTree.Name)
	}

	if fileTree.Type != "directory" {
		t.Errorf("Root node type should be 'directory', got '%s'", fileTree.Type)
	}

	if len(fileTree.Children) == 0 {
		t.Error("Root node should have children")
	}

	// Check first child (should be our test file)
	firstChild := fileTree.Children[0]
	if firstChild.Name != "test.go" {
		t.Errorf("First child name should be 'test.go', got '%s'", firstChild.Name)
	}

	if firstChild.Type != "file" {
		t.Errorf("File node type should be 'file', got '%s'", firstChild.Type)
	}

	if firstChild.Violations != 2 {
		t.Errorf("File node should have 2 violations, got %d", firstChild.Violations)
	}
}

func TestReport_GetViolationsByFile(t *testing.T) {
	summary, files, config := createTestReportData()
	report := NewReport(summary, files, config)

	violationsByFile := report.GetViolationsByFile()

	if len(violationsByFile) != 1 {
		t.Errorf("Expected 1 file with violations, got %d", len(violationsByFile))
	}

	violations, exists := violationsByFile["test.go"]
	if !exists {
		t.Error("Expected violations for test.go")
	}

	if len(violations) != 2 {
		t.Errorf("Expected 2 violations for test.go, got %d", len(violations))
	}
}

func TestCalculateStatistics(t *testing.T) {
	_, files, _ := createTestReportData()

	stats := calculateStatistics(files)

	if stats == nil {
		t.Fatal("Statistics should not be nil")
	}

	// Check violations by type
	if len(stats.ViolationsByType) != 2 {
		t.Errorf("Expected 2 violation types, got %d", len(stats.ViolationsByType))
	}

	if stats.ViolationsByType[ViolationTypeFunctionLength] != 1 {
		t.Errorf("Expected 1 function length violation, got %d", stats.ViolationsByType[ViolationTypeFunctionLength])
	}

	if stats.ViolationsByType[ViolationTypeNaming] != 1 {
		t.Errorf("Expected 1 naming violation, got %d", stats.ViolationsByType[ViolationTypeNaming])
	}

	// Check violations by severity
	if len(stats.ViolationsBySeverity) != 2 {
		t.Errorf("Expected 2 severity levels, got %d", len(stats.ViolationsBySeverity))
	}

	if stats.ViolationsBySeverity[SeverityCritical] != 1 {
		t.Errorf("Expected 1 critical violation, got %d", stats.ViolationsBySeverity[SeverityCritical])
	}

	if stats.ViolationsBySeverity[SeverityMedium] != 1 {
		t.Errorf("Expected 1 medium violation, got %d", stats.ViolationsBySeverity[SeverityMedium])
	}

	// Check files by language
	if len(stats.FilesByLanguage) != 1 {
		t.Errorf("Expected 1 language, got %d", len(stats.FilesByLanguage))
	}

	if stats.FilesByLanguage["Go"] != 1 {
		t.Errorf("Expected 1 Go file, got %d", stats.FilesByLanguage["Go"])
	}

	// Check top violated files
	if len(stats.TopViolatedFiles) != 1 {
		t.Errorf("Expected 1 top violated file, got %d", len(stats.TopViolatedFiles))
	}

	topFile := stats.TopViolatedFiles[0]
	if topFile.File != "test.go" {
		t.Errorf("Expected top violated file to be 'test.go', got '%s'", topFile.File)
	}

	if topFile.TotalViolations != 2 {
		t.Errorf("Expected 2 total violations, got %d", topFile.TotalViolations)
	}

	if topFile.Lines != 50 {
		t.Errorf("Expected 50 lines, got %d", topFile.Lines)
	}
}

func TestCalculateStatisticsNoViolations(t *testing.T) {
	// Create files with no violations
	fileInfo := &FileInfo{
		Path:     "clean.go",
		Name:     "clean.go",
		Language: "Go",
		Lines:    25,
		Scanned:  true,
	}

	metrics := &FileMetrics{
		TotalLines:    25,
		FunctionCount: 2,
	}

	scanResult := &ScanResult{
		File:       fileInfo,
		Violations: []*Violation{}, // No violations
		Metrics:    metrics,
	}

	files := []*ScanResult{scanResult}
	stats := calculateStatistics(files)

	if len(stats.ViolationsByType) != 0 {
		t.Errorf("Expected 0 violation types, got %d", len(stats.ViolationsByType))
	}

	if len(stats.ViolationsBySeverity) != 0 {
		t.Errorf("Expected 0 severity levels, got %d", len(stats.ViolationsBySeverity))
	}

	if len(stats.TopViolatedFiles) != 0 {
		t.Errorf("Expected 0 top violated files, got %d", len(stats.TopViolatedFiles))
	}

	if stats.FilesByLanguage["Go"] != 1 {
		t.Errorf("Expected 1 Go file, got %d", stats.FilesByLanguage["Go"])
	}
}

func TestSeverityGetColor(t *testing.T) {
	colorTests := map[Severity]string{
		SeverityLow:      "text-success",
		SeverityMedium:   "text-warning", 
		SeverityHigh:     "text-danger",
		SeverityCritical: "text-danger fw-bold",
	}

	for severity, expectedColor := range colorTests {
		color := severity.GetColor()
		if color != expectedColor {
			t.Errorf("Expected color '%s' for severity %s, got '%s'", expectedColor, severity, color)
		}
	}
}

func TestViolationTypeGetDisplayName(t *testing.T) {
	displayNameTests := map[ViolationType]string{
		ViolationTypeFunctionLength:         "Long Functions",
		ViolationTypeCyclomaticComplexity:   "Complex Functions", 
		ViolationTypeParameterCount:         "Too Many Parameters",
		ViolationTypeNestingDepth:           "Deep Nesting",
		ViolationTypeNaming:                 "Naming Convention",
		ViolationTypeClassSize:              "Large Classes",
		ViolationTypeMissingDocumentation:   "Missing Documentation",
		ViolationTypeMagicNumbers:           "Magic Numbers",
		ViolationTypeDuplication:            "Code Duplication",
	}

	for violationType, expectedName := range displayNameTests {
		name := violationType.GetDisplayName()
		if name != expectedName {
			t.Errorf("Expected display name '%s' for type %s, got '%s'", expectedName, violationType, name)
		}
	}
}

func TestGenerateReportID(t *testing.T) {
	id1 := generateReportID()
	time.Sleep(time.Second) // Ensure different timestamps (seconds precision)
	id2 := generateReportID()

	if id1 == "" {
		t.Error("Report ID should not be empty")
	}

	if id2 == "" {
		t.Error("Report ID should not be empty")
	}

	if id1 == id2 {
		t.Error("Report IDs should be unique")
	}

	// Check format (should be timestamp-based)
	if len(id1) != 15 { // Format: 20060102-150405 (15 characters)
		t.Errorf("Report ID should be 15 characters, got %d", len(id1))
	}
}

func createTestReportData() (*ScanSummary, []*ScanResult, *ReportConfig) {
	// Create test violations
	violations := []*Violation{
		{
			ID:       "test-1",
			Type:     ViolationTypeFunctionLength,
			Severity: SeverityCritical,
			Message:  "Function too long",
			File:     "test.go",
			Line:     10,
		},
		{
			ID:       "test-2", 
			Type:     ViolationTypeNaming,
			Severity: SeverityMedium,
			Message:  "Poor variable name",
			File:     "test.go",
			Line:     5,
		},
	}

	// Create test file info
	fileInfo := &FileInfo{
		Path:     "test.go",
		Name:     "test.go",
		Language: "Go",
		Lines:    50,
		Scanned:  true,
	}

	// Create test metrics
	metrics := &FileMetrics{
		TotalLines:    50,
		FunctionCount: 3,
	}

	// Create test scan result
	scanResult := &ScanResult{
		File:       fileInfo,
		Violations: violations,
		Metrics:    metrics,
	}

	// Create test summary
	summary := &ScanSummary{
		TotalFiles:      1,
		ScannedFiles:    1,
		TotalViolations: 2,
		ViolationsByType: map[string]int{
			string(ViolationTypeFunctionLength): 1,
			string(ViolationTypeNaming):         1,
		},
	}

	// Create test config
	config := &ReportConfig{
		Paths:     []string{"./test"},
		FileTypes: []string{".go"},
		Thresholds: &Thresholds{
			FunctionLines: 25,
		},
	}

	return summary, []*ScanResult{scanResult}, config
}