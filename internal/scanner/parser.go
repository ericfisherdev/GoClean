package scanner

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/ericfisherdev/goclean/internal/models"
)

// Parser handles file parsing and basic analysis
type Parser struct {
	verbose     bool
	astAnalyzer *ASTAnalyzer
}

// NewParser creates a new Parser instance
func NewParser(verbose bool) *Parser {
	return &Parser{
		verbose:     verbose,
		astAnalyzer: NewASTAnalyzer(verbose),
	}
}

// ParseFile reads and analyzes a single file
func (p *Parser) ParseFile(fileInfo *models.FileInfo) (*models.ScanResult, error) {
	if p.verbose {
		fmt.Printf("Parsing file: %s\n", fileInfo.Path)
	}
	
	// Use AST parsing for Go files
	if fileInfo.Language == "Go" {
		return p.parseGoFileWithAST(fileInfo)
	}
	
	// Fall back to line-by-line parsing for other languages
	return p.parseFileLineByLine(fileInfo, nil)
}

// parseGoFileWithAST performs AST-based parsing for Go files
func (p *Parser) parseGoFileWithAST(fileInfo *models.FileInfo) (*models.ScanResult, error) {
	content, err := p.readFileOptimized(fileInfo.Path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s: %w", fileInfo.Path, err)
	}

	// Perform AST analysis
	astInfo, err := p.astAnalyzer.AnalyzeGoFile(fileInfo.Path, content)
	if err != nil {
		// Fall back to line-by-line parsing if AST fails
		if p.verbose {
			fmt.Printf("AST parsing failed for %s, falling back to line parsing: %v\n", fileInfo.Path, err)
		}
		return p.parseFileLineByLine(fileInfo, content)
	}

	// Extract metrics from content
	metrics := p.extractMetricsFromContent(content, fileInfo.Language)
	metrics.FunctionCount = len(astInfo.Functions)
	metrics.ClassCount = len(astInfo.Types) // In Go, types are structs/interfaces

	// Update file info with AST data
	fileInfo.Lines = metrics.TotalLines
	fileInfo.Scanned = true

	// Create scan result with AST information
	result := &models.ScanResult{
		File:       fileInfo,
		Violations: []*models.Violation{}, // Will be populated by violation detectors
		Metrics:    metrics,
		ASTInfo:    astInfo, // Store AST info for violation detection
	}

	if p.verbose {
		fmt.Printf("AST parsed %s: %d lines, %d functions, %d types\n",
			fileInfo.Path, metrics.TotalLines, len(astInfo.Functions), len(astInfo.Types))
	}

	return result, nil
}

// parseFileLineByLine performs traditional line-by-line parsing
func (p *Parser) parseFileLineByLine(fileInfo *models.FileInfo, content []byte) (*models.ScanResult, error) {
	var err error
	if content == nil {
		content, err = p.readFileOptimized(fileInfo.Path)
		if err != nil {
			return nil, fmt.Errorf("cannot read file %s: %w", fileInfo.Path, err)
		}
	}

	// Initialize metrics
	metrics := p.extractMetricsFromContent(content, fileInfo.Language)

	// Update file info
	fileInfo.Lines = metrics.TotalLines
	fileInfo.Scanned = true

	// Create scan result
	result := &models.ScanResult{
		File:       fileInfo,
		Violations: []*models.Violation{}, // Will be populated by violation detectors
		Metrics:    metrics,
	}

	if p.verbose {
		fmt.Printf("Line parsed %s: %d lines, %d code lines, %d comment lines\n",
			fileInfo.Path, metrics.TotalLines, metrics.CodeLines, metrics.CommentLines)
	}

	return result, nil
}

// analyzeLine analyzes a single line and updates metrics
func (p *Parser) analyzeLine(line string, lineNumber int, metrics *models.FileMetrics, fileInfo *models.FileInfo) {
	trimmed := strings.TrimSpace(line)
	
	// Count line types
	if len(trimmed) == 0 {
		metrics.BlankLines++
		return
	}
	
	if p.isCommentLine(trimmed, fileInfo.Language) {
		metrics.CommentLines++
		return
	}
	
	metrics.CodeLines++
	
	// Basic function detection (simplified)
	if p.isFunctionDeclaration(trimmed, fileInfo.Language) {
		metrics.FunctionCount++
	}
	
	// Basic class detection (simplified)
	if p.isClassDeclaration(trimmed, fileInfo.Language) {
		metrics.ClassCount++
	}
}

// isCommentLine checks if a line is a comment
func (p *Parser) isCommentLine(line, language string) bool {
	switch language {
	case "Go", "JavaScript", "TypeScript", "Java", "C", "C++", "C#", "Rust", "Swift", "Kotlin", "Scala":
		return strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*/") || strings.HasPrefix(line, "*")
	case "Python", "Ruby":
		return strings.HasPrefix(line, "#")
	case "PHP":
		return strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "/*")
	default:
		return false
	}
}

// isFunctionDeclaration checks if a line contains a function declaration (basic detection)
func (p *Parser) isFunctionDeclaration(line, language string) bool {
	switch language {
	case "Go":
		return strings.Contains(line, "func ") && (strings.Contains(line, "(") || strings.Contains(line, "{"))
	case "JavaScript", "TypeScript":
		return (strings.Contains(line, "function ") || 
				strings.Contains(line, "() =>") || 
				strings.Contains(line, ") => {") ||
				(strings.Contains(line, ":") && strings.Contains(line, "{"))) && 
			   !strings.HasPrefix(strings.TrimSpace(line), "//")
	case "Python":
		return strings.HasPrefix(line, "def ") || strings.Contains(line, " def ")
	case "Java", "C#":
		return (strings.Contains(line, "(") && strings.Contains(line, ")") && 
				(strings.Contains(line, "public") || strings.Contains(line, "private") || 
				 strings.Contains(line, "protected") || strings.Contains(line, "static"))) &&
			   !strings.Contains(line, "class")
	case "C", "C++":
		return strings.Contains(line, "(") && strings.Contains(line, ")") && 
			   !strings.Contains(line, "class") && !strings.Contains(line, "struct") &&
			   !strings.HasPrefix(strings.TrimSpace(line), "//") &&
			   p.looksLikeFunctionSignature(line)
	default:
		return false
	}
}

// isClassDeclaration checks if a line contains a class declaration
func (p *Parser) isClassDeclaration(line, language string) bool {
	switch language {
	case "Go":
		return strings.Contains(line, "type ") && strings.Contains(line, "struct")
	case "JavaScript", "TypeScript":
		return strings.Contains(line, "class ")
	case "Python":
		return strings.HasPrefix(line, "class ") || strings.Contains(line, " class ")
	case "Java", "C#", "C++", "Swift", "Kotlin", "Scala":
		return strings.Contains(line, "class ")
	default:
		return false
	}
}

// looksLikeFunctionSignature performs additional checks for C/C++ function signatures
func (p *Parser) looksLikeFunctionSignature(line string) bool {
	// Basic heuristics for C/C++ function signatures
	// This is simplified and can be enhanced with proper AST parsing
	
	// Must have parentheses
	if !strings.Contains(line, "(") || !strings.Contains(line, ")") {
		return false
	}
	
	// Should not be a macro or preprocessor directive
	if strings.HasPrefix(strings.TrimSpace(line), "#") {
		return false
	}
	
	// Should have some identifier before the parentheses
	parenIndex := strings.Index(line, "(")
	beforeParen := strings.TrimSpace(line[:parenIndex])
	
	// Must have at least one identifier
	if len(beforeParen) == 0 {
		return false
	}
	
	// Should end with an identifier (function name)
	fields := strings.Fields(beforeParen)
	if len(fields) == 0 {
		return false
	}
	
	lastField := fields[len(fields)-1]
	
	// Function name should start with letter or underscore
	if len(lastField) == 0 || (!unicode.IsLetter(rune(lastField[0])) && lastField[0] != '_') {
		return false
	}
	
	return true
}

// extractMetricsFromContent analyzes file content and extracts metrics.
func (p *Parser) extractMetricsFromContent(content []byte, language string) *models.FileMetrics {
	metrics := &models.FileMetrics{}
	scanner := bufio.NewScanner(bytes.NewReader(content))

	for scanner.Scan() {
		line := scanner.Text()
		metrics.TotalLines++
		trimmed := strings.TrimSpace(line)

		if len(trimmed) == 0 {
			metrics.BlankLines++
			continue
		}

		if p.isCommentLine(trimmed, language) {
			metrics.CommentLines++
			continue
		}

		metrics.CodeLines++

		// For non-Go files, perform basic detection.
		// For Go, we use more accurate AST-based counts.
		if language != "Go" {
			if p.isFunctionDeclaration(trimmed, language) {
				metrics.FunctionCount++
			}
			if p.isClassDeclaration(trimmed, language) {
				metrics.ClassCount++
			}
		}
	}
	return metrics
}

// readFileOptimized reads a file using the standard library for simplicity and efficiency.
func (p *Parser) readFileOptimized(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}