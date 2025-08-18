package violations

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ericfisherdev/goclean/internal/models"
)

// Package-level cache for clippy violations to improve performance
var (
	clippyCache     map[string][]*models.Violation
	clippyCacheLock sync.RWMutex
)

// Initialize the cache map
func init() {
	clippyCache = make(map[string][]*models.Violation)
}

// ClippyIntegrator integrates rust-clippy lints into GoClean's violation detection
type ClippyIntegrator struct {
	config *DetectorConfig
}

// ClippyMessage represents a cargo clippy JSON message
type ClippyMessage struct {
	Reason      string         `json:"reason"`
	Message     ClippyDiagnostic `json:"message"`
	PackageID   string         `json:"package_id"`
	Target      ClippyTarget   `json:"target"`
	Manifest_path string       `json:"manifest_path"`
}

// ClippyDiagnostic represents the diagnostic information from clippy
type ClippyDiagnostic struct {
	Message  string              `json:"message"`
	Code     *ClippyCode         `json:"code"`
	Level    string              `json:"level"`
	Spans    []ClippySpan        `json:"spans"`
	Children []ClippyDiagnostic  `json:"children"`
	Rendered string              `json:"rendered"`
}

// ClippyCode represents the lint code information
type ClippyCode struct {
	Code        string `json:"code"`
	Explanation string `json:"explanation"`
}

// ClippySpan represents source location information
type ClippySpan struct {
	FileName    string `json:"file_name"`
	ByteStart   int    `json:"byte_start"`
	ByteEnd     int    `json:"byte_end"`
	LineStart   int    `json:"line_start"`
	LineEnd     int    `json:"line_end"`
	ColumnStart int    `json:"column_start"`
	ColumnEnd   int    `json:"column_end"`
	IsPrimary   bool   `json:"is_primary"`
	Text        []ClippyText `json:"text"`
	Label       string `json:"label"`
	SuggestionApplicability string `json:"suggestion_applicability"`
}

// ClippyText represents the source text at a span
type ClippyText struct {
	Text      string `json:"text"`
	Highlight_start int `json:"highlight_start"`
	Highlight_end   int `json:"highlight_end"`
}

// ClippyTarget represents the build target information
type ClippyTarget struct {
	Name  string   `json:"name"`
	Kind  []string `json:"kind"`
	CrateTypes []string `json:"crate_types"`
	RequiredFeatures []string `json:"required-features"`
	SrcPath string `json:"src_path"`
	Edition string `json:"edition"`
}

// ClippyLintCategory represents the categories of clippy lints
type ClippyLintCategory string

const (
	ClippyCorrectness ClippyLintCategory = "clippy::correctness"
	ClippySuspicious  ClippyLintCategory = "clippy::suspicious"
	ClippyStyle       ClippyLintCategory = "clippy::style"
	ClippyComplexity  ClippyLintCategory = "clippy::complexity"
	ClippyPerf        ClippyLintCategory = "clippy::perf"
)

// NewClippyIntegrator creates a new clippy integrator
func NewClippyIntegrator(config *DetectorConfig) *ClippyIntegrator {
	if config == nil {
		config = DefaultDetectorConfig()
	}
	return &ClippyIntegrator{
		config: config,
	}
}

// Name returns the name of this detector
func (c *ClippyIntegrator) Name() string {
	return "ClippyIntegrator"
}

// Description returns the description of this detector
func (c *ClippyIntegrator) Description() string {
	return "Integrates rust-clippy lints to detect Rust code issues and suggestions"
}

// Detect runs clippy analysis on the Rust project and converts results to GoClean violations
func (c *ClippyIntegrator) Detect(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation {
	// Only process Rust files
	if fileInfo.Language != "rust" {
		return []*models.Violation{}
	}

	// Check if clippy integration is enabled
	if c.config.ClippyConfig == nil || !c.config.ClippyConfig.Enabled {
		return []*models.Violation{}
	}

	// Find the project root (directory containing Cargo.toml)
	projectRoot := c.findCargoProjectRoot(fileInfo.Path)
	if projectRoot == "" {
		// No Cargo.toml found, skip clippy integration
		return []*models.Violation{}
	}

	// Create cache key based on project root and file path
	cacheKey := fmt.Sprintf("%s:%s", projectRoot, fileInfo.Path)
	
	// Check cache first (read lock)
	clippyCacheLock.RLock()
	if cachedViolations, exists := clippyCache[cacheKey]; exists {
		clippyCacheLock.RUnlock()
		return cachedViolations
	}
	clippyCacheLock.RUnlock()

	// Run clippy and get JSON output
	clippyOutput, err := c.runClippy(projectRoot)
	if err != nil {
		// Log error but don't fail the detection
		if c.config.Verbose {
			fmt.Printf("Clippy integration failed: %v\n", err)
		}
		return []*models.Violation{}
	}

	// Parse clippy output and convert to violations
	violations := c.parseClippyOutput(clippyOutput, fileInfo.Path)
	
	// Store in cache (write lock)
	clippyCacheLock.Lock()
	clippyCache[cacheKey] = violations
	clippyCacheLock.Unlock()
	
	return violations
}

// findCargoProjectRoot finds the directory containing Cargo.toml starting from the given path
func (c *ClippyIntegrator) findCargoProjectRoot(filePath string) string {
	dir := filepath.Dir(filePath)
	
	// Look for Cargo.toml up the directory tree
	for {
		cargoPath := filepath.Join(dir, "Cargo.toml")
		if _, err := os.Stat(cargoPath); err == nil {
			return dir
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}
	
	return ""
}

// runClippy executes cargo clippy with JSON output and core lint categories
func (c *ClippyIntegrator) runClippy(projectRoot string) ([]byte, error) {
	// Build the clippy command with specific lint categories
	args := []string{
		"clippy",
		"--message-format=json",
		"--",
		"-W", "clippy::correctness",
		"-W", "clippy::suspicious", 
		"-W", "clippy::style",
		"-W", "clippy::complexity",
		"-W", "clippy::perf",
	}
	
	cmd := exec.Command("cargo", args...)
	cmd.Dir = projectRoot
	
	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	
	return output, err
}

// parseClippyOutput parses the JSON output from clippy and converts to GoClean violations
func (c *ClippyIntegrator) parseClippyOutput(output []byte, targetFile string) []*models.Violation {
	var violations []*models.Violation
	
	// Split output into lines since cargo outputs one JSON object per line
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		var message ClippyMessage
		if err := json.Unmarshal([]byte(line), &message); err != nil {
			// Skip invalid JSON lines
			continue
		}
		
		// Only process compiler-message entries
		if message.Reason != "compiler-message" {
			continue
		}
		
		// Convert clippy diagnostic to GoClean violation
		violation := c.convertDiagnosticToViolation(message.Message, targetFile)
		if violation != nil {
			violations = append(violations, violation)
		}
	}
	
	return violations
}

// convertDiagnosticToViolation converts a clippy diagnostic to a GoClean violation
func (c *ClippyIntegrator) convertDiagnosticToViolation(diagnostic ClippyDiagnostic, targetFile string) *models.Violation {
	// Skip non-warning/error diagnostics
	if diagnostic.Level != "warning" && diagnostic.Level != "error" {
		return nil
	}
	
	// Find the primary span for this diagnostic
	var primarySpan *ClippySpan
	for _, span := range diagnostic.Spans {
		if span.IsPrimary {
			primarySpan = &span
			break
		}
	}
	
	if primarySpan == nil {
		return nil
	}
	
	// Only include violations for the target file we're analyzing
	if !strings.HasSuffix(primarySpan.FileName, filepath.Base(targetFile)) {
		return nil
	}
	
	// Only create violations for diagnostics with clippy:: prefix
	if diagnostic.Code == nil || !strings.HasPrefix(diagnostic.Code.Code, "clippy::") {
		return nil
	}
	
	// Handle potential nil diagnostic.Code safely
	var codeStr string
	if diagnostic.Code != nil {
		codeStr = diagnostic.Code.Code
	} else {
		codeStr = "unknown"
	}
	
	// Map clippy severity to GoClean severity
	severity := c.mapClippySeverity(diagnostic.Level, codeStr)
	
	// Determine violation type based on clippy lint category
	violationType := c.mapClippyLintToViolationType(codeStr)
	
	// Create the violation with clear clippy attribution
	violation := &models.Violation{
		Type:        violationType,
		Severity:    severity,
		File:        primarySpan.FileName,
		Line:        primarySpan.LineStart,
		Column:      primarySpan.ColumnStart,
		Message:     fmt.Sprintf("Detected by rust-clippy: %s", diagnostic.Message),
		Description: "Violation detected by rust-clippy static analysis tool",
		Rule:        codeStr,
		Suggestion:  c.generateClippySuggestion(diagnostic),
		CodeSnippet: c.extractCodeSnippet(primarySpan),
	}
	
	return violation
}

// mapClippySeverity maps clippy diagnostic levels to GoClean severity
func (c *ClippyIntegrator) mapClippySeverity(level string, lintCode string) models.Severity {
	// Map based on diagnostic level first
	switch level {
	case "error":
		return models.SeverityHigh
	case "warning":
		// Further classify warnings based on lint category
		if strings.Contains(lintCode, "correctness") {
			return models.SeverityHigh
		} else if strings.Contains(lintCode, "suspicious") {
			return models.SeverityMedium
		} else if strings.Contains(lintCode, "complexity") || strings.Contains(lintCode, "perf") {
			return models.SeverityMedium
		} else if strings.Contains(lintCode, "style") {
			return models.SeverityLow
		}
		return models.SeverityMedium
	default:
		return models.SeverityLow
	}
}

// mapClippyLintToViolationType maps clippy lint codes to GoClean violation types
func (c *ClippyIntegrator) mapClippyLintToViolationType(lintCode string) models.ViolationType {
	// Map common clippy lints to appropriate violation types
	switch {
	case strings.Contains(lintCode, "cognitive_complexity") || strings.Contains(lintCode, "too_many_arguments"):
		return models.ViolationTypeCyclomaticComplexity
	case strings.Contains(lintCode, "missing_docs") || strings.Contains(lintCode, "undocumented"):
		return models.ViolationTypeMissingDocumentation
	case strings.Contains(lintCode, "naming") || strings.Contains(lintCode, "wrong_self_convention"):
		return models.ViolationTypeNaming
	case strings.Contains(lintCode, "todo") || strings.Contains(lintCode, "unimplemented") || strings.Contains(lintCode, "panic"):
		return models.ViolationTypeTodo
	case strings.Contains(lintCode, "magic_number") || strings.Contains(lintCode, "approx_constant"):
		return models.ViolationTypeMagicNumber
	case strings.Contains(lintCode, "duplicate") || strings.Contains(lintCode, "similar"):
		return models.ViolationTypeDuplication
	default:
		// Use a generic structure violation type for other clippy lints
		return models.ViolationTypeStructure
	}
}

// generateClippySuggestion creates a suggestion based on clippy's diagnostic and children
func (c *ClippyIntegrator) generateClippySuggestion(diagnostic ClippyDiagnostic) string {
	suggestion := "Consider addressing this clippy lint"
	
	// Check if clippy provided specific suggestions in children diagnostics
	for _, child := range diagnostic.Children {
		if child.Message != "" && strings.Contains(child.Message, "help:") {
			suggestion = strings.TrimPrefix(child.Message, "help: ")
			break
		}
	}
	
	// If we have an explanation, use that
	if diagnostic.Code != nil && diagnostic.Code.Explanation != "" {
		suggestion += ". " + diagnostic.Code.Explanation
	}
	
	return suggestion
}

// extractCodeSnippet extracts the relevant code snippet from the span
func (c *ClippyIntegrator) extractCodeSnippet(span *ClippySpan) string {
	if len(span.Text) == 0 {
		return ""
	}
	
	// Use the first text entry as the code snippet
	return strings.TrimSpace(span.Text[0].Text)
}

// isClippyAvailable checks if cargo clippy is available in the system
func (c *ClippyIntegrator) isClippyAvailable() bool {
	cmd := exec.Command("cargo", "clippy", "--version")
	err := cmd.Run()
	return err == nil
}

// ClearClippyCache clears the global clippy violation cache
func ClearClippyCache() {
	clippyCacheLock.Lock()
	defer clippyCacheLock.Unlock()
	clippyCache = make(map[string][]*models.Violation)
}