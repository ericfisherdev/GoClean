package scanner

import (
	"bufio"
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
	return p.parseFileLineByLine(fileInfo)
}

// parseGoFileWithAST performs AST-based parsing for Go files
func (p *Parser) parseGoFileWithAST(fileInfo *models.FileInfo) (*models.ScanResult, error) {
	// Perform AST analysis
	astInfo, err := p.astAnalyzer.AnalyzeGoFile(fileInfo.Path)
	if err != nil {
		// Fall back to line-by-line parsing if AST fails
		if p.verbose {
			fmt.Printf("AST parsing failed for %s, falling back to line parsing: %v\n", fileInfo.Path, err)
		}
		return p.parseFileLineByLine(fileInfo)
	}
	
	// Extract metrics from AST
	metrics := p.extractMetricsFromAST(astInfo, fileInfo)
	
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
func (p *Parser) parseFileLineByLine(fileInfo *models.FileInfo) (*models.ScanResult, error) {
	file, err := os.Open(fileInfo.Path)
	if err != nil {
		return nil, fmt.Errorf("cannot open file %s: %w", fileInfo.Path, err)
	}
	defer file.Close()
	
	// Initialize metrics
	metrics := &models.FileMetrics{}
	
	// Read file line by line
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		
		// Analyze line
		p.analyzeLine(line, lineNumber, metrics, fileInfo)
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", fileInfo.Path, err)
	}
	
	// Update file info
	fileInfo.Lines = lineNumber
	fileInfo.Scanned = true
	
	// Update metrics
	metrics.TotalLines = lineNumber
	
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

// extractMetricsFromAST extracts file metrics from AST information
func (p *Parser) extractMetricsFromAST(astInfo *GoASTInfo, fileInfo *models.FileInfo) *models.FileMetrics {
	// Count lines by reading the file (needed for accurate line counts)
	file, err := os.Open(fileInfo.Path)
	if err != nil {
		// Return basic metrics if file can't be opened
		return &models.FileMetrics{
			TotalLines:    1,
			FunctionCount: len(astInfo.Functions),
		}
	}
	defer file.Close()
	
	metrics := &models.FileMetrics{}
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	
	// Count lines and basic metrics
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		
		if len(trimmed) == 0 {
			metrics.BlankLines++
		} else if p.isCommentLine(trimmed, fileInfo.Language) {
			metrics.CommentLines++
		} else {
			metrics.CodeLines++
		}
	}
	
	// Set total lines
	metrics.TotalLines = lineNumber
	
	// Extract AST-based metrics
	metrics.FunctionCount = len(astInfo.Functions)
	metrics.ClassCount = len(astInfo.Types) // Types in Go can be structs/interfaces
	
	return metrics
}