// Package violations provides detectors for various clean code violations in Go source code.
package violations

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// CommentedCodeDetector detects blocks of commented-out code
type CommentedCodeDetector struct {
	config      *DetectorConfig
	codePattern *regexp.Regexp
}

// NewCommentedCodeDetector creates a new commented code detector
func NewCommentedCodeDetector(config *DetectorConfig) *CommentedCodeDetector {
	// Pattern to detect likely code in comments
	codePattern := regexp.MustCompile(`(?i)(func |if |for |switch |case |return |var |const |type |import |package |fmt\.|log\.|:=|\{|\}|\(|\))`)
	
	return &CommentedCodeDetector{
		config:      config,
		codePattern: codePattern,
	}
}

// Name returns the name of this detector
func (d *CommentedCodeDetector) Name() string {
	return "Commented Code Detector"
}

// Description returns a description of what this detector checks for
func (d *CommentedCodeDetector) Description() string {
	return "Detects blocks of commented-out code that should be removed"
}

// Detect analyzes the provided file information and returns violations
func (d *CommentedCodeDetector) Detect(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation {
	var violations []*models.Violation
	
	if astInfo == nil {
		return violations
	}
	
	// Type assertion to get types.GoASTInfo
	goAstInfo, ok := astInfo.(*types.GoASTInfo)
	if !ok || goAstInfo == nil || goAstInfo.AST == nil {
		return violations
	}
	
	// Extract all comments from the AST
	for _, commentGroup := range goAstInfo.AST.Comments {
		if violation := d.checkCommentGroup(commentGroup, goAstInfo.FileSet, fileInfo.Path); violation != nil {
			violations = append(violations, violation)
		}
	}
	
	return violations
}

// checkCommentGroup checks if a comment group contains commented-out code
func (d *CommentedCodeDetector) checkCommentGroup(group *ast.CommentGroup, fset *token.FileSet, filePath string) *models.Violation {
	if group == nil {
		return nil
	}
	
	// Combine all comments in the group
	var commentText strings.Builder
	startLine := 0
	
	for i, comment := range group.List {
		text := comment.Text
		
		// Remove comment markers
		if strings.HasPrefix(text, "//") {
			text = strings.TrimPrefix(text, "//")
		} else if strings.HasPrefix(text, "/*") && strings.HasSuffix(text, "*/") {
			text = strings.TrimPrefix(text, "/*")
			text = strings.TrimSuffix(text, "*/")
		}
		
		text = strings.TrimSpace(text)
		
		// Skip obvious documentation patterns
		if d.isDocumentation(text) {
			return nil
		}
		
		commentText.WriteString(text)
		if i < len(group.List)-1 {
			commentText.WriteString("\n")
		}
		
		if startLine == 0 {
			pos := fset.Position(comment.Pos())
			startLine = pos.Line
		}
	}
	
	fullText := commentText.String()
	
	// Check if this looks like code
	if d.looksLikeCode(fullText) {
		// Count the number of code-like patterns
		matches := d.codePattern.FindAllString(fullText, -1)
		if len(matches) >= 3 { // At least 3 code patterns to be considered commented code
			snippet := fullText
			if len(snippet) > 100 {
				snippet = snippet[:97] + "..."
			}
			
			return &models.Violation{
				Type:        models.ViolationTypeCommentedCode,
				Severity:    models.SeverityLow,
				File:        filePath,
				Line:        startLine,
				Column:      0,
				Message:     "Block of commented-out code detected",
				Suggestion:  "Remove commented-out code. Use version control to preserve old code if needed",
				CodeSnippet: snippet,
			}
		}
	}
	
	return nil
}

// looksLikeCode checks if text looks like commented-out code
func (d *CommentedCodeDetector) looksLikeCode(text string) bool {
	// Check for common code patterns
	if d.codePattern.MatchString(text) {
		// Additional checks to reduce false positives
		
		// If it's very short, it's probably not code
		if len(text) < 20 {
			return false
		}
		
		// Check for balanced braces/parens (common in code)
		openBraces := strings.Count(text, "{")
		closeBraces := strings.Count(text, "}")
		openParens := strings.Count(text, "(")
		closeParens := strings.Count(text, ")")
		
		// If we have balanced braces or parens, it's likely code
		if (openBraces > 0 && openBraces == closeBraces) ||
		   (openParens > 0 && openParens == closeParens) {
			return true
		}
		
		// Check for assignment operators
		if strings.Contains(text, ":=") || strings.Contains(text, "=") {
			return true
		}
		
		// Check for function/method calls
		if regexp.MustCompile(`\w+\(`).MatchString(text) {
			return true
		}
		
		return true
	}
	
	return false
}

// isDocumentation checks if text is likely documentation rather than commented code
func (d *CommentedCodeDetector) isDocumentation(text string) bool {
	// Common documentation patterns
	docPatterns := []string{
		"TODO:", "FIXME:", "NOTE:", "HACK:", "XXX:", "BUG:",
		"Copyright", "License", "Author",
		"@param", "@return", "@throws", // Javadoc style
		"Example:", "Usage:", "Description:",
	}
	
	textLower := strings.ToLower(text)
	for _, pattern := range docPatterns {
		if strings.Contains(textLower, strings.ToLower(pattern)) {
			return true
		}
	}
	
	// Check if it's a sentence (starts with capital, ends with period)
	if len(text) > 0 {
		firstChar := text[0]
		lastChar := text[len(text)-1]
		if firstChar >= 'A' && firstChar <= 'Z' && lastChar == '.' {
			// Check if it contains mostly words rather than code
			wordPattern := regexp.MustCompile(`\b[a-zA-Z]+\b`)
			words := wordPattern.FindAllString(text, -1)
			if len(words) > 3 && float64(len(strings.Join(words, ""))) / float64(len(text)) > 0.5 {
				return true
			}
		}
	}
	
	return false
}