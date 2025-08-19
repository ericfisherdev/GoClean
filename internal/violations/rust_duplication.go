// Package violations provides detectors for various clean code violations in Rust source code.
package violations

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// Rust-specific duplication detection constants
const (
	RustMinFunctionLinesForDuplication = 5
	RustSameFileLineThreshold          = 15
	RustMinCodeBlockSize               = 3
)

// RustDuplicationDetector detects code duplication in Rust code
type RustDuplicationDetector struct {
	config        *DetectorConfig
	codeExtractor *CodeExtractor
	hashCache     map[string][]RustCodeBlock
	mutex         sync.RWMutex
	// Compiled regex patterns for Rust code normalization
	commentPattern    *regexp.Regexp
	whitespacePattern *regexp.Regexp
	stringPattern     *regexp.Regexp
}

// RustCodeBlock represents a block of Rust code for duplication analysis
type RustCodeBlock struct {
	File      string
	StartLine int
	EndLine   int
	Content   string
	Hash      string
	Type      string // "function", "impl", "trait", "module"
	Name      string
}

// NewRustDuplicationDetector creates a new Rust duplication detector
func NewRustDuplicationDetector(config *DetectorConfig) *RustDuplicationDetector {
	if config == nil {
		config = DefaultDetectorConfig()
	}

	return &RustDuplicationDetector{
		config:        config,
		codeExtractor: NewCodeExtractor(),
		hashCache:     make(map[string][]RustCodeBlock),
		// Rust-specific regex patterns
		commentPattern:    regexp.MustCompile(`//.*|/\*[\s\S]*?\*/`),
		whitespacePattern: regexp.MustCompile(`\s+`),
		stringPattern:     regexp.MustCompile(`"[^"]*"|'[^']*'`),
	}
}

// Name returns the name of this detector
func (d *RustDuplicationDetector) Name() string {
	return "Rust Code Duplication Detector"
}

// Description returns a description of what this detector checks for
func (d *RustDuplicationDetector) Description() string {
	return "Detects duplicate code blocks in Rust code including functions, impl blocks, and traits"
}

// Detect analyzes the provided Rust file information and returns violations
func (d *RustDuplicationDetector) Detect(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation {
	var violations []*models.Violation

	if astInfo == nil {
		return violations
	}

	// Type assertion to get types.RustASTInfo
	rustAstInfo, ok := astInfo.(*types.RustASTInfo)
	if !ok || rustAstInfo == nil {
		return violations
	}

	// Read the file content for analysis
	content, err := d.readFileContent(fileInfo.Path)
	if err != nil {
		return violations
	}

	// Analyze functions for duplication
	violations = append(violations, d.analyzeFunctionDuplication(rustAstInfo, fileInfo.Path, content)...)

	// Analyze implementation blocks for duplication
	violations = append(violations, d.analyzeImplDuplication(rustAstInfo, fileInfo.Path, content)...)

	// Analyze code patterns for duplication
	violations = append(violations, d.analyzeCodePatternDuplication(fileInfo.Path, content)...)

	return violations
}

// readFileContent reads the content of a file
func (d *RustDuplicationDetector) readFileContent(filePath string) (string, error) {
	if d.codeExtractor == nil {
		return "", fmt.Errorf("code extractor not available")
	}

	content, err := d.codeExtractor.ExtractSnippet(filePath, 1, -1) // Read entire file
	if err != nil {
		return "", err
	}

	return content, nil
}

// analyzeFunctionDuplication analyzes functions for code duplication
func (d *RustDuplicationDetector) analyzeFunctionDuplication(rustAstInfo *types.RustASTInfo, filePath, content string) []*models.Violation {
	var violations []*models.Violation

	if rustAstInfo.Functions == nil {
		return violations
	}

	lines := strings.Split(content, "\n")

	for _, function := range rustAstInfo.Functions {
		if function == nil {
			continue
		}

		// Only check functions with sufficient lines
		if function.EndLine-function.StartLine < RustMinFunctionLinesForDuplication {
			continue
		}

		// Extract function content
		functionContent := d.extractLines(lines, function.StartLine, function.EndLine)
		normalized := d.normalizeRustCode(functionContent)
		hash := d.hashRustCode(normalized)

		block := RustCodeBlock{
			File:      filePath,
			StartLine: function.StartLine,
			EndLine:   function.EndLine,
			Content:   functionContent,
			Hash:      hash,
			Type:      "function",
			Name:      function.Name,
		}

		// Check for duplicates
		violations = append(violations, d.checkForDuplicates(block)...)
	}

	return violations
}

// analyzeImplDuplication analyzes implementation blocks for code duplication
func (d *RustDuplicationDetector) analyzeImplDuplication(rustAstInfo *types.RustASTInfo, filePath, content string) []*models.Violation {
	var violations []*models.Violation

	if rustAstInfo.Impls == nil {
		return violations
	}

	lines := strings.Split(content, "\n")

	for _, impl := range rustAstInfo.Impls {
		if impl == nil {
			continue
		}

		// Only check implementation blocks with sufficient lines
		if impl.EndLine-impl.StartLine < RustMinFunctionLinesForDuplication {
			continue
		}

		// Extract impl content
		implContent := d.extractLines(lines, impl.StartLine, impl.EndLine)
		normalized := d.normalizeRustCode(implContent)
		hash := d.hashRustCode(normalized)

		block := RustCodeBlock{
			File:      filePath,
			StartLine: impl.StartLine,
			EndLine:   impl.EndLine,
			Content:   implContent,
			Hash:      hash,
			Type:      "impl",
			Name:      impl.TargetType,
		}

		// Check for duplicates
		violations = append(violations, d.checkForDuplicates(block)...)
	}

	return violations
}

// analyzeCodePatternDuplication analyzes general code patterns for duplication
func (d *RustDuplicationDetector) analyzeCodePatternDuplication(filePath, content string) []*models.Violation {
	var violations []*models.Violation

	lines := strings.Split(content, "\n")

	// Look for duplicated code patterns using a sliding window approach
	for i := 0; i < len(lines)-RustMinCodeBlockSize; i++ {
		// Extract a block of lines
		blockLines := lines[i : i+RustMinCodeBlockSize]
		blockContent := strings.Join(blockLines, "\n")

		// Skip if block is mostly empty or comments
		if d.shouldSkipBlock(blockContent) {
			continue
		}

		normalized := d.normalizeRustCode(blockContent)
		hash := d.hashRustCode(normalized)

		block := RustCodeBlock{
			File:      filePath,
			StartLine: i + 1,
			EndLine:   i + RustMinCodeBlockSize,
			Content:   blockContent,
			Hash:      hash,
			Type:      "pattern",
			Name:      fmt.Sprintf("code_block_%d_%d", i+1, i+RustMinCodeBlockSize),
		}

		// Check for duplicates (only report significant patterns)
		if len(strings.TrimSpace(normalized)) > 50 { // Only substantial code blocks
			violations = append(violations, d.checkForDuplicates(block)...)
		}
	}

	return violations
}

// shouldSkipBlock determines if a code block should be skipped from duplication analysis
func (d *RustDuplicationDetector) shouldSkipBlock(content string) bool {
	lines := strings.Split(content, "\n")
	nonEmptyLines := 0
	commentLines := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		nonEmptyLines++

		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			commentLines++
		}
	}

	// Skip if mostly empty or comments
	if nonEmptyLines < 2 || float64(commentLines)/float64(nonEmptyLines) > 0.8 {
		return true
	}

	// Skip common Rust patterns that are naturally repetitive
	contentLower := strings.ToLower(content)
	skipPatterns := []string{
		"use ", "extern crate", "#[derive", "#[cfg", "mod ", "pub mod",
		"println!", "eprintln!", "dbg!", "panic!", "unimplemented!",
		"todo!", "unreachable!", "assert!", "assert_eq!", "assert_ne!",
	}

	for _, pattern := range skipPatterns {
		if strings.Contains(contentLower, pattern) {
			return true
		}
	}

	return false
}

// extractLines extracts a range of lines from the file content
func (d *RustDuplicationDetector) extractLines(lines []string, startLine, endLine int) string {
	if startLine < 1 || startLine > endLine {
		return ""
	}

	// Convert to 0-based indexing
	start := startLine - 1
	end := endLine
	if end > len(lines) {
		end = len(lines)
	}
	if start >= len(lines) {
		return ""
	}

	return strings.Join(lines[start:end], "\n")
}

// normalizeRustCode normalizes Rust code for comparison
func (d *RustDuplicationDetector) normalizeRustCode(code string) string {
	// Split into lines first to preserve line structure
	lines := strings.Split(code, "\n")
	var normalizedLines []string

	for _, line := range lines {
		// Remove comments from each line
		if idx := strings.Index(line, "//"); idx >= 0 {
			line = line[:idx]
		}
		
		// Remove block comments (simplified approach)
		line = strings.ReplaceAll(line, "/*", "")
		line = strings.ReplaceAll(line, "*/", "")

		// Replace string literals with placeholders
		line = d.stringPattern.ReplaceAllString(line, `"STRING"`)

		// Remove common Rust-specific tokens that don't affect logic
		replacements := map[string]string{
			"mut ":    "",
			"&mut ":   "&",
			"& ":      "&", 
			"pub ":    "",
			"unsafe ": "",
			"async ":  "",
			"const ":  "",
			"static ": "",
		}

		for old, new := range replacements {
			line = strings.ReplaceAll(line, old, new)
		}

		// Normalize whitespace within the line
		line = d.whitespacePattern.ReplaceAllString(line, " ")
		line = strings.TrimSpace(line)

		// Skip empty lines and standalone braces
		if line == "" || line == "{" || line == "}" {
			continue
		}

		normalizedLines = append(normalizedLines, line)
	}

	return strings.Join(normalizedLines, "\n")
}

// hashRustCode creates a hash of the normalized Rust code
func (d *RustDuplicationDetector) hashRustCode(code string) string {
	hasher := md5.New()
	hasher.Write([]byte(code))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// checkForDuplicates checks if a code block is a duplicate and returns violations
func (d *RustDuplicationDetector) checkForDuplicates(block RustCodeBlock) []*models.Violation {
	var violations []*models.Violation

	d.mutex.Lock()
	defer d.mutex.Unlock()

	existing, found := d.hashCache[block.Hash]
	if found {
		for _, existingBlock := range existing {
			// Don't report duplicates in the same file within threshold lines
			if existingBlock.File == block.File &&
				d.abs(existingBlock.StartLine-block.StartLine) < RustSameFileLineThreshold {
				continue
			}

			// Create violation
			violation := &models.Violation{
				Type:     models.ViolationTypeDuplication,
				Severity: d.classifyRustDuplicationSeverity(block),
				File:     block.File,
				Line:     block.StartLine,
				Column:   1,
				Message: fmt.Sprintf("Duplicate Rust %s found (lines %d-%d). Similar code in %s:%d-%d",
					block.Type, block.StartLine, block.EndLine,
					existingBlock.File, existingBlock.StartLine, existingBlock.EndLine),
				Rule:       "rust-code-duplication",
				Suggestion: d.getRustDuplicationSuggestion(block),
				CodeSnippet: d.truncateRustCode(block.Content, 5),
			}
			violations = append(violations, violation)
		}
	}

	// Add to cache
	d.hashCache[block.Hash] = append(d.hashCache[block.Hash], block)

	return violations
}

// classifyRustDuplicationSeverity determines the severity of a duplication violation
func (d *RustDuplicationDetector) classifyRustDuplicationSeverity(block RustCodeBlock) models.Severity {
	lineCount := block.EndLine - block.StartLine

	// Larger duplications are more severe
	if lineCount > 50 {
		return models.SeverityHigh
	}
	if lineCount > 20 {
		return models.SeverityMedium
	}

	// Function duplications are more concerning than pattern duplications
	if block.Type == "function" || block.Type == "impl" {
		return models.SeverityMedium
	}

	return models.SeverityLow
}

// getRustDuplicationSuggestion provides Rust-specific suggestions for duplication violations
func (d *RustDuplicationDetector) getRustDuplicationSuggestion(block RustCodeBlock) string {
	switch block.Type {
	case "function":
		return "Consider extracting duplicate function logic into a shared function, trait method, or generic function"
	case "impl":
		return "Consider creating a shared trait or extracting common implementation patterns into a macro or generic implementation"
	case "trait":
		return "Consider creating a super-trait or extracting common trait functionality"
	case "pattern":
		return "Consider extracting duplicate code patterns into a function, macro, or using Rust's powerful macro system"
	default:
		return "Consider extracting duplicate code into a shared component using Rust's module system, traits, or macros"
	}
}

// truncateRustCode truncates Rust code for display in violation reports
func (d *RustDuplicationDetector) truncateRustCode(code string, maxLines int) string {
	lines := strings.Split(code, "\n")
	if len(lines) <= maxLines {
		return code
	}

	truncated := strings.Join(lines[:maxLines], "\n")
	return truncated + "\n... (truncated)"
}

// abs returns the absolute value of an integer
func (d *RustDuplicationDetector) abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Reset clears the duplication cache (used when starting a new scan)
func (d *RustDuplicationDetector) Reset() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.hashCache = make(map[string][]RustCodeBlock)
}