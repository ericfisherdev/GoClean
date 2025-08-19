// Package violations provides detectors for various clean code violations in Rust source code.
package violations

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// RustErrorHandlingDetector detects error handling violations in Rust code
type RustErrorHandlingDetector struct {
	config        *DetectorConfig
	codeExtractor *CodeExtractor
}

// NewRustErrorHandlingDetector creates a new Rust error handling violation detector
func NewRustErrorHandlingDetector(config *DetectorConfig) *RustErrorHandlingDetector {
	if config == nil {
		config = DefaultDetectorConfig()
	}
	return &RustErrorHandlingDetector{
		config:        config,
		codeExtractor: NewCodeExtractor(),
	}
}

// Name returns the name of this detector
func (d *RustErrorHandlingDetector) Name() string {
	return "Rust Error Handling Analysis"
}

// Description returns a description of what this detector checks for
func (d *RustErrorHandlingDetector) Description() string {
	return "Detects error handling violations in Rust code including overuse of unwrap() and expect(), missing error propagation, and panic-prone patterns"
}

// Detect analyzes Rust code for error handling violations
func (d *RustErrorHandlingDetector) Detect(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation {
	var violations []*models.Violation

	if astInfo == nil {
		return violations
	}

	// Type assertion to get types.RustASTInfo
	rustAstInfo, ok := astInfo.(*types.RustASTInfo)
	if !ok || rustAstInfo == nil {
		// Not a Rust file or invalid AST info
		return violations
	}

	// Read the source file for pattern analysis
	content, err := d.readFileContent(fileInfo.Path)
	if err != nil {
		return violations
	}

	lines := strings.Split(content, "\n")

	// Analyze content for error handling violations
	violations = append(violations, d.detectUnwrapOveruse(fileInfo.Path, lines)...)
	violations = append(violations, d.detectMissingErrorPropagation(fileInfo.Path, lines)...)
	violations = append(violations, d.detectInconsistentErrorTypes(fileInfo.Path, lines)...)
	violations = append(violations, d.detectPanicProneCode(fileInfo.Path, lines)...)
	violations = append(violations, d.detectUnhandledResults(fileInfo.Path, lines)...)
	violations = append(violations, d.detectImproperExpect(fileInfo.Path, lines)...)

	return violations
}

// detectUnwrapOveruse identifies overuse of unwrap() calls
func (d *RustErrorHandlingDetector) detectUnwrapOveruse(filePath string, lines []string) []*models.Violation {
	var violations []*models.Violation

	// Patterns for unwrap usage
	unwrapPattern := regexp.MustCompile(`(\w+)\.unwrap\(\)`)
	
	// Count unwraps in the file
	unwrapCount := 0
	unwrapLines := []int{}

	for lineNum, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Skip comments and empty lines
		if strings.HasPrefix(trimmedLine, "//") || trimmedLine == "" {
			continue
		}

		if unwrapPattern.MatchString(trimmedLine) {
			unwrapCount++
			unwrapLines = append(unwrapLines, lineNum+1)

			// Check if this specific unwrap is problematic
			if d.isProblematicUnwrap(trimmedLine) {
				codeSnippet := d.extractCodeSnippet(filePath, lineNum+1, lineNum+1)
				violations = append(violations, &models.Violation{
					Type:        models.ViolationTypeRustOveruseUnwrap,
					Severity:    d.getUnwrapSeverity(trimmedLine),
					Message:     fmt.Sprintf("Potentially dangerous unwrap() usage - consider proper error handling"),
					File:        filePath,
					Line:        lineNum + 1,
					Column:      strings.Index(line, ".unwrap()") + 1,
					Rule:        "rust-overuse-unwrap",
					Suggestion:  d.getUnwrapSuggestion(trimmedLine),
					CodeSnippet: codeSnippet,
				})
			}
		}
	}

	// If there are too many unwraps in the file overall
	if unwrapCount > d.config.RustConfig.MaxLifetimeParams*2 { // Reusing config value as threshold
		if len(unwrapLines) > 0 {
			firstLine := unwrapLines[0]
			codeSnippet := d.extractCodeSnippet(filePath, firstLine, firstLine)
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeRustOveruseUnwrap,
				Severity:    models.SeverityMedium,
				Message:     fmt.Sprintf("File contains too many unwrap() calls (%d) - consider using proper error handling", unwrapCount),
				File:        filePath,
				Line:        firstLine,
				Column:      1,
				Rule:        "rust-unwrap-count",
				Suggestion:  "Refactor to use pattern matching, if let, or ? operator for error handling",
				CodeSnippet: codeSnippet,
			})
		}
	}

	return violations
}

// detectMissingErrorPropagation identifies places where ? operator should be used
func (d *RustErrorHandlingDetector) detectMissingErrorPropagation(filePath string, lines []string) []*models.Violation {
	var violations []*models.Violation

	// Simpler patterns for error propagation opportunities
	errorPropagationPatterns := []*regexp.Regexp{
		// match with error return
		regexp.MustCompile(`return\s+Err\(`),
		// if let with error return
		regexp.MustCompile(`if\s+let\s+Err`),
	}

	// Look for functions that return Result
	hasResultFunction := false
	for _, line := range lines {
		if strings.Contains(line, "fn ") && strings.Contains(line, "-> Result") {
			hasResultFunction = true
			break
		}
	}
	
	for lineNum, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Skip comments and empty lines
		if strings.HasPrefix(trimmedLine, "//") || trimmedLine == "" {
			continue
		}

		// Look for error propagation opportunities
		for _, pattern := range errorPropagationPatterns {
			if pattern.MatchString(trimmedLine) && hasResultFunction {
				codeSnippet := d.extractCodeSnippet(filePath, lineNum+1, lineNum+1)
				violations = append(violations, &models.Violation{
					Type:        models.ViolationTypeRustMissingErrorPropagation,
					Severity:    models.SeverityMedium,
					Message:     "Consider using ? operator for error propagation instead of explicit match/if let",
					File:        filePath,
					Line:        lineNum + 1,
					Column:      d.findPatternColumn(line, pattern),
					Rule:        "rust-missing-error-propagation",
					Suggestion:  "Replace explicit error handling with ? operator for cleaner error propagation",
					CodeSnippet: codeSnippet,
				})
			}
		}
	}

	return violations
}

// detectInconsistentErrorTypes identifies inconsistent error type usage
func (d *RustErrorHandlingDetector) detectInconsistentErrorTypes(filePath string, lines []string) []*models.Violation {
	var violations []*models.Violation

	// Track different error types used in the file
	errorTypes := make(map[string][]int) // error type -> line numbers
	resultPattern := regexp.MustCompile(`Result<[^,]+,\s*([^>]+)>`)
	errorReturnPattern := regexp.MustCompile(`Err\(([^)]+)\(`)

	for lineNum, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Skip comments and empty lines
		if strings.HasPrefix(trimmedLine, "//") || trimmedLine == "" {
			continue
		}

		// Find Result type declarations
		if matches := resultPattern.FindStringSubmatch(trimmedLine); matches != nil {
			errorType := strings.TrimSpace(matches[1])
			errorTypes[errorType] = append(errorTypes[errorType], lineNum+1)
		}

		// Find Err returns
		if matches := errorReturnPattern.FindStringSubmatch(trimmedLine); matches != nil {
			errorType := strings.TrimSpace(matches[1])
			if !strings.Contains(errorType, "::") { // Simple error type
				errorTypes[errorType] = append(errorTypes[errorType], lineNum+1)
			}
		}
	}

	// Check for inconsistent error types (more than 3 different types in one file)
	if len(errorTypes) > 3 {
		for errorType, lineNums := range errorTypes {
			if len(lineNums) > 0 {
				codeSnippet := d.extractCodeSnippet(filePath, lineNums[0], lineNums[0])
				violations = append(violations, &models.Violation{
					Type:        models.ViolationTypeRustInconsistentErrorType,
					Severity:    models.SeverityLow,
					Message:     fmt.Sprintf("Inconsistent error types detected (using %s among %d different error types) - consider using a unified error type", errorType, len(errorTypes)),
					File:        filePath,
					Line:        lineNums[0],
					Column:      1,
					Rule:        "rust-inconsistent-error-type",
					Suggestion:  "Use a consistent error type throughout the module, consider using anyhow or thiserror crates",
					CodeSnippet: codeSnippet,
				})
				break // Only report once per file
			}
		}
	}

	return violations
}

// detectPanicProneCode identifies code patterns that can panic
func (d *RustErrorHandlingDetector) detectPanicProneCode(filePath string, lines []string) []*models.Violation {
	var violations []*models.Violation

	panicPatterns := []*regexp.Regexp{
		// Direct panic calls
		regexp.MustCompile(`panic!\s*\(`),
		// Array/slice indexing without bounds check
		regexp.MustCompile(`\w+\[\w+\]`),
		// Division that could panic on zero
		regexp.MustCompile(`\w+\s*/\s*\w+`),
		// Unwrap on known panic-prone operations
		regexp.MustCompile(`(to_string|parse|from_str).*\.unwrap\(\)`),
	}

	for lineNum, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Skip comments and empty lines
		if strings.HasPrefix(trimmedLine, "//") || trimmedLine == "" {
			continue
		}

		for _, pattern := range panicPatterns {
			if pattern.MatchString(trimmedLine) {
				// Skip array indexing if it's an assignment
				if strings.Contains(pattern.String(), `\[\w+\]`) && strings.Contains(trimmedLine, "=") && 
				   strings.Index(trimmedLine, "[") < strings.Index(trimmedLine, "=") {
					continue
				}
				
				severity := d.getPanicSeverity(trimmedLine, pattern)
				message := d.getPanicMessage(trimmedLine, pattern)
				
				codeSnippet := d.extractCodeSnippet(filePath, lineNum+1, lineNum+1)
				violations = append(violations, &models.Violation{
					Type:        models.ViolationTypeRustPanicProneCode,
					Severity:    severity,
					Message:     message,
					File:        filePath,
					Line:        lineNum + 1,
					Column:      d.findPatternColumn(line, pattern),
					Rule:        "rust-panic-prone-code",
					Suggestion:  d.getPanicSuggestion(trimmedLine, pattern),
					CodeSnippet: codeSnippet,
				})
			}
		}
	}

	return violations
}

// detectUnhandledResults identifies Result types that aren't properly handled
func (d *RustErrorHandlingDetector) detectUnhandledResults(filePath string, lines []string) []*models.Violation {
	var violations []*models.Violation

	// Pattern for unhandled Result calls
	unhandledPattern := regexp.MustCompile(`(\w+)\(.*\);`)

	for lineNum, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Skip comments and empty lines
		if strings.HasPrefix(trimmedLine, "//") || trimmedLine == "" {
			continue
		}

		// Look for function calls ending with semicolon (potential unhandled Results)
		if unhandledPattern.MatchString(trimmedLine) {
			// Check if this looks like a Result-returning function call
			if d.looksLikeResultCall(trimmedLine) {
				codeSnippet := d.extractCodeSnippet(filePath, lineNum+1, lineNum+1)
				violations = append(violations, &models.Violation{
					Type:        models.ViolationTypeRustUnhandledResult,
					Severity:    models.SeverityMedium,
					Message:     "Potentially unhandled Result - consider using ? operator or explicit error handling",
					File:        filePath,
					Line:        lineNum + 1,
					Column:      1,
					Rule:        "rust-unhandled-result",
					Suggestion:  "Handle the Result with match, if let, or ? operator",
					CodeSnippet: codeSnippet,
				})
			}
		}
	}

	return violations
}

// detectImproperExpect identifies expect() calls without descriptive messages
func (d *RustErrorHandlingDetector) detectImproperExpect(filePath string, lines []string) []*models.Violation {
	var violations []*models.Violation

	// Pattern for expect usage
	expectPattern := regexp.MustCompile(`\.expect\s*\(\s*"([^"]*)"\s*\)`)

	for lineNum, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Skip comments and empty lines
		if strings.HasPrefix(trimmedLine, "//") || trimmedLine == "" {
			continue
		}

		if matches := expectPattern.FindStringSubmatch(trimmedLine); matches != nil {
			message := matches[1]
			
			// Check if the expect message is descriptive enough
			if len(message) < 5 || message == "error" || message == "failed" {
				codeSnippet := d.extractCodeSnippet(filePath, lineNum+1, lineNum+1)
				violations = append(violations, &models.Violation{
					Type:        models.ViolationTypeRustImproperExpect,
					Severity:    models.SeverityLow,
					Message:     fmt.Sprintf("expect() call has non-descriptive message: \"%s\"", message),
					File:        filePath,
					Line:        lineNum + 1,
					Column:      strings.Index(line, ".expect") + 1,
					Rule:        "rust-improper-expect",
					Suggestion:  "Provide a descriptive message explaining why the operation should not fail",
					CodeSnippet: codeSnippet,
				})
			}
		}
	}

	return violations
}

// Helper methods

func (d *RustErrorHandlingDetector) isProblematicUnwrap(line string) bool {
	// Don't flag unwrap_or as problematic
	if strings.Contains(line, "unwrap_or") {
		return false
	}
	
	// Unwrap in main/test functions is usually okay
	if strings.Contains(line, "fn main") || strings.Contains(line, "#[test]") {
		return false
	}
	
	// Unwrap in loop (likely to panic)
	if strings.Contains(line, "for ") || strings.Contains(line, "while ") {
		return true
	}
	
	// Most other unwraps should be flagged
	return true
}

func (d *RustErrorHandlingDetector) getUnwrapSeverity(line string) models.Severity {
	if strings.Contains(line, "unwrap_or") {
		return models.SeverityLow // unwrap_or is safer
	}
	if strings.Contains(line, "for ") || strings.Contains(line, "while ") {
		return models.SeverityHigh // unwrap in loops is dangerous
	}
	return models.SeverityMedium
}

func (d *RustErrorHandlingDetector) getUnwrapSuggestion(line string) string {
	if strings.Contains(line, "Option") {
		return "Consider using if let Some(...) or match instead of unwrap()"
	}
	if strings.Contains(line, "Result") {
		return "Consider using ? operator, match, or if let Ok(...) instead of unwrap()"
	}
	return "Consider using proper error handling instead of unwrap()"
}

func (d *RustErrorHandlingDetector) getPanicSeverity(line string, pattern *regexp.Regexp) models.Severity {
	patternStr := pattern.String()
	
	if strings.Contains(patternStr, "panic!") {
		return models.SeverityHigh
	}
	if strings.Contains(patternStr, `\[\w+\]`) { // Array indexing
		return models.SeverityMedium
	}
	if strings.Contains(patternStr, "/") { // Division
		return models.SeverityLow
	}
	return models.SeverityMedium
}

func (d *RustErrorHandlingDetector) getPanicMessage(line string, pattern *regexp.Regexp) string {
	patternStr := pattern.String()
	
	if strings.Contains(patternStr, "panic!") {
		return "Direct panic! call - consider using Result or Option for error handling"
	}
	if strings.Contains(patternStr, `\[\w+\]`) {
		return "Array/slice indexing without bounds check - could panic on out-of-bounds access"
	}
	if strings.Contains(patternStr, "/") {
		return "Division operation - could panic on division by zero"
	}
	if strings.Contains(patternStr, "unwrap") {
		return "Unwrap on parsing operation - likely to panic on invalid input"
	}
	return "Potentially panic-prone code pattern detected"
}

func (d *RustErrorHandlingDetector) getPanicSuggestion(line string, pattern *regexp.Regexp) string {
	patternStr := pattern.String()
	
	if strings.Contains(patternStr, "panic!") {
		return "Replace panic! with proper error handling using Result or Option"
	}
	if strings.Contains(patternStr, `\[\w+\]`) {
		return "Use .get() method for safe indexing or check bounds before indexing"
	}
	if strings.Contains(patternStr, "/") {
		return "Check for zero before division or use checked division operations"
	}
	if strings.Contains(patternStr, "unwrap") {
		return "Use proper error handling for parsing operations instead of unwrap()"
	}
	return "Add appropriate bounds checking or error handling"
}

func (d *RustErrorHandlingDetector) looksLikeResultCall(line string) bool {
	// Simple heuristic: function calls that might return Result
	resultHints := []string{"read", "write", "parse", "connect", "open", "create", "send", "recv"}
	
	for _, hint := range resultHints {
		if strings.Contains(line, hint+"(") {
			return true
		}
	}
	
	// Look for common Result-returning patterns
	if strings.Contains(line, "std::") || strings.Contains(line, "::") {
		return true
	}
	
	return false
}

func (d *RustErrorHandlingDetector) findPatternColumn(line string, pattern *regexp.Regexp) int {
	match := pattern.FindStringIndex(line)
	if match != nil {
		return match[0] + 1
	}
	return 1
}

// extractCodeSnippet extracts a code snippet for the violation with context
func (d *RustErrorHandlingDetector) extractCodeSnippet(filePath string, startLine, endLine int) string {
	if d.codeExtractor == nil {
		return d.generateFallbackSnippet(startLine, endLine)
	}

	snippet, err := d.codeExtractor.ExtractSnippet(filePath, startLine, endLine)
	if err != nil {
		return d.generateFallbackSnippet(startLine, endLine)
	}

	return snippet
}

// generateFallbackSnippet creates a simple snippet when file reading fails
func (d *RustErrorHandlingDetector) generateFallbackSnippet(startLine, endLine int) string {
	if endLine <= startLine {
		return fmt.Sprintf("Line %d: <code snippet unavailable>", startLine)
	}
	return fmt.Sprintf("Lines %d-%d: <code snippet unavailable>", startLine, endLine)
}

// readFileContent reads and returns the content of a file
func (d *RustErrorHandlingDetector) readFileContent(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}