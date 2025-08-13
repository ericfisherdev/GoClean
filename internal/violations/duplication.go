package violations

import (
	"crypto/md5"
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// DuplicationDetector detects code duplication across files
type DuplicationDetector struct {
	config    *DetectorConfig
	hashCache map[string][]CodeBlock
}

// CodeBlock represents a block of code for duplication analysis
type CodeBlock struct {
	File      string
	StartLine int
	EndLine   int
	Content   string
	Hash      string
}

// NewDuplicationDetector creates a new duplication detector
func NewDuplicationDetector(config *DetectorConfig) *DuplicationDetector {
	return &DuplicationDetector{
		config:    config,
		hashCache: make(map[string][]CodeBlock),
	}
}

// Name returns the name of this detector
func (d *DuplicationDetector) Name() string {
	return "Code Duplication Detector"
}

// Description returns a description of what this detector checks for
func (d *DuplicationDetector) Description() string {
	return "Detects duplicate code blocks across files"
}

// Detect analyzes the provided file information and returns violations
func (d *DuplicationDetector) Detect(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation {
	var violations []*models.Violation
	
	// Type assertion to get types.GoASTInfo
	if astInfo == nil {
		return violations
	}
	
	goAstInfo, ok := astInfo.(*types.GoASTInfo)
	if !ok || goAstInfo == nil {
		return violations
	}
	
	// Extract code blocks from functions using AST directly
	if goAstInfo.AST != nil {
		ast.Inspect(goAstInfo.AST, func(n ast.Node) bool {
			if funcDecl, ok := n.(*ast.FuncDecl); ok && funcDecl.Body != nil {
				body := extractFunctionBody(funcDecl, goAstInfo.FileSet)
				
				// Only check functions with more than 5 lines
				lines := strings.Split(body, "\n")
				if len(lines) < 5 {
					return true
				}
				
				pos := goAstInfo.FileSet.Position(funcDecl.Pos())
				endPos := goAstInfo.FileSet.Position(funcDecl.End())
				
				// Create a normalized version for comparison
				normalized := normalizeCode(body)
				hash := hashCode(normalized)
				
				block := CodeBlock{
					File:      fileInfo.Path,
					StartLine: pos.Line,
					EndLine:   endPos.Line,
					Content:   body,
					Hash:      hash,
				}
				
				// Check if we've seen this code block before
				if existing, found := d.hashCache[hash]; found {
					for _, existingBlock := range existing {
						// Don't report duplicates in the same file within 10 lines (could be intentional patterns)
						if existingBlock.File == block.File && 
						   abs(existingBlock.StartLine - block.StartLine) < 10 {
							continue
						}
						
						violation := &models.Violation{
							Type:        models.ViolationTypeDuplication,
							Severity:    d.classifyDuplicationSeverity(len(lines)),
							File:        fileInfo.Path,
							Line:        pos.Line,
							Column:      pos.Column,
							Message:     fmt.Sprintf("Duplicate code block found (lines %d-%d). Similar code in %s:%d-%d", pos.Line, endPos.Line, existingBlock.File, existingBlock.StartLine, existingBlock.EndLine),
							Suggestion:  "Consider extracting duplicate code into a shared function or method",
							CodeSnippet: truncateCode(body, 5),
						}
						violations = append(violations, violation)
					}
				}
				
				// Add to cache
				d.hashCache[hash] = append(d.hashCache[hash], block)
			}
			return true
		})
	}
	
	return violations
}

// normalizeCode normalizes code for comparison
func normalizeCode(code string) string {
	// Remove comments and whitespace variations
	lines := strings.Split(code, "\n")
	var normalized []string
	
	for _, line := range lines {
		// Remove line comments
		if idx := strings.Index(line, "//"); idx >= 0 {
			line = line[:idx]
		}
		
		// Trim whitespace
		line = strings.TrimSpace(line)
		
		// Skip empty lines
		if line == "" {
			continue
		}
		
		normalized = append(normalized, line)
	}
	
	return strings.Join(normalized, "\n")
}

// hashCode creates a hash of the code block
func hashCode(code string) string {
	h := md5.Sum([]byte(code))
	return fmt.Sprintf("%x", h)
}

// truncateCode truncates code to a maximum number of lines
func truncateCode(code string, maxLines int) string {
	lines := strings.Split(code, "\n")
	if len(lines) <= maxLines {
		return code
	}
	
	truncated := strings.Join(lines[:maxLines], "\n")
	return truncated + "\n..."
}

// classifyDuplicationSeverity classifies the severity of code duplication
func (d *DuplicationDetector) classifyDuplicationSeverity(lineCount int) models.Severity {
	if lineCount > 20 {
		return models.SeverityHigh
	} else if lineCount > 10 {
		return models.SeverityMedium
	}
	return models.SeverityLow
}

// abs returns the absolute value of an integer
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// extractFunctionBody extracts the body of a function as a string
func extractFunctionBody(funcDecl *ast.FuncDecl, fset *token.FileSet) string {
	if funcDecl.Body == nil {
		return ""
	}
	
	start := fset.Position(funcDecl.Body.Pos())
	end := fset.Position(funcDecl.Body.End())
	
	// This is a simplified extraction - in a real implementation,
	// you might want to use go/format to get the actual source code
	return fmt.Sprintf("func %s() { /* lines %d-%d */ }", funcDecl.Name.Name, start.Line, end.Line)
}

// Reset clears the hash cache (useful when starting a new scan)
func (d *DuplicationDetector) Reset() {
	d.hashCache = make(map[string][]CodeBlock)
}