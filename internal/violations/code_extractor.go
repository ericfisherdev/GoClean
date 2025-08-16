// Package violations provides violation detection and code analysis capabilities.
// It includes detectors for various clean code violations such as magic numbers,
// function length, cyclomatic complexity, and code duplication.
package violations

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	DefaultContextLines = 3
	MaxContextLines     = 10
)

// CodeExtractor provides functionality to extract code snippets with context
type CodeExtractor struct {
	contextLines int
}

// NewCodeExtractor creates a new code extractor with default context lines
func NewCodeExtractor() *CodeExtractor {
	return &CodeExtractor{
		contextLines: DefaultContextLines,
	}
}

// NewCodeExtractorWithContext creates a new code extractor with specified context lines
func NewCodeExtractorWithContext(contextLines int) *CodeExtractor {
	if contextLines > MaxContextLines {
		contextLines = MaxContextLines
	}
	if contextLines < 0 {
		contextLines = 0
	}
	return &CodeExtractor{
		contextLines: contextLines,
	}
}

// ExtractSnippet extracts a code snippet with context around the specified line
func (e *CodeExtractor) ExtractSnippet(filePath string, targetLine int, endLine int) (string, error) {
	if !filepath.IsAbs(filePath) {
		// Handle relative paths
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path for %s: %w", filePath, err)
		}
		filePath = absPath
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	// Read all lines into memory
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return e.buildSnippet(lines, targetLine, endLine), nil
}

// buildSnippet builds a formatted code snippet with line numbers and highlighting
func (e *CodeExtractor) buildSnippet(lines []string, targetLine int, endLine int) string {
	if len(lines) == 0 {
		return ""
	}

	// Convert to 0-based indexing
	targetIdx := targetLine - 1
	endIdx := endLine - 1

	if endIdx <= 0 || endIdx < targetIdx {
		endIdx = targetIdx
	}

	// Calculate the range including context
	startIdx := targetIdx - e.contextLines
	if startIdx < 0 {
		startIdx = 0
	}

	endContextIdx := endIdx + e.contextLines
	if endContextIdx >= len(lines) {
		endContextIdx = len(lines) - 1
	}

	var snippet strings.Builder

	// Calculate the maximum line number width for alignment
	maxLineNum := endContextIdx + 1
	lineNumWidth := len(strconv.Itoa(maxLineNum))

	for i := startIdx; i <= endContextIdx; i++ {
		lineNum := i + 1
		line := lines[i]

		// Format line number with proper padding
		lineNumStr := fmt.Sprintf("%*d", lineNumWidth, lineNum)

		// Determine if this line is part of the violation
		isViolationLine := lineNum >= targetLine && lineNum <= endLine

		if isViolationLine {
			// Mark violation lines with special formatting
			snippet.WriteString(fmt.Sprintf("→ %s│ %s\n", lineNumStr, line))
		} else {
			// Regular context lines
			snippet.WriteString(fmt.Sprintf("  %s│ %s\n", lineNumStr, line))
		}
	}

	return strings.TrimSuffix(snippet.String(), "\n")
}

// ExtractSnippetWithHighlight extracts a code snippet and returns it with HTML highlighting
func (e *CodeExtractor) ExtractSnippetWithHighlight(filePath string, targetLine int, endLine int) (string, error) {
	snippet, err := e.ExtractSnippet(filePath, targetLine, endLine)
	if err != nil {
		return "", err
	}

	return e.addHTMLHighlighting(snippet, targetLine, endLine), nil
}

// addHTMLHighlighting adds HTML highlighting to violation lines
func (e *CodeExtractor) addHTMLHighlighting(snippet string, targetLine int, endLine int) string {
	lines := strings.Split(snippet, "\n")
	var result strings.Builder

	for _, line := range lines {
		if strings.HasPrefix(line, "→") {
			// This is a violation line, wrap it with highlighting
			result.WriteString(`<span class="violation-line">`)
			result.WriteString(line)
			result.WriteString("</span>\n")
		} else {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return strings.TrimSuffix(result.String(), "\n")
}

// SetContextLines updates the number of context lines to include
func (e *CodeExtractor) SetContextLines(contextLines int) {
	if contextLines > MaxContextLines {
		contextLines = MaxContextLines
	}
	if contextLines < 0 {
		contextLines = 0
	}
	e.contextLines = contextLines
}

// GetContextLines returns the current number of context lines
func (e *CodeExtractor) GetContextLines() int {
	return e.contextLines
}