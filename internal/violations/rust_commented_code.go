package violations

import (
	"regexp"
	"strings"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// RustCommentedCodeDetector detects blocks of commented-out Rust code
type RustCommentedCodeDetector struct {
	config      *DetectorConfig
	codePattern *regexp.Regexp
}

// NewRustCommentedCodeDetector creates a new Rust commented code detector
func NewRustCommentedCodeDetector(config *DetectorConfig) *RustCommentedCodeDetector {
	// Pattern to detect likely Rust code in comments
	// Rust-specific keywords and patterns
	codePattern := regexp.MustCompile(`(?i)(fn |if |for |while |loop |match |let |mut |const |static |struct |enum |trait |impl |pub |mod |use |extern |async |await |unsafe |return |\{|\}|\(|\)|::|->|&mut |&|println!|vec!|Some\(|None|Ok\(|Err\()`)
	
	return &RustCommentedCodeDetector{
		config:      config,
		codePattern: codePattern,
	}
}

// Name returns the name of this detector
func (d *RustCommentedCodeDetector) Name() string {
	return "Rust Commented Code Detector"
}

// Description returns a description of what this detector checks for
func (d *RustCommentedCodeDetector) Description() string {
	return "Detects blocks of commented-out Rust code that should be removed"
}

// Detect analyzes the provided file information and returns violations
func (d *RustCommentedCodeDetector) Detect(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation {
	var violations []*models.Violation
	
	if astInfo == nil {
		return violations
	}
	
	// Type assertion to get types.RustASTInfo
	rustAstInfo, ok := astInfo.(*types.RustASTInfo)
	if !ok || rustAstInfo == nil {
		return violations
	}
	
	// Since we don't have access to raw comments in the current RustASTInfo,
	// we'll parse the file content directly for comments
	content, err := d.readFileContent(fileInfo.Path)
	if err != nil {
		return violations
	}
	
	lines := strings.Split(content, "\n")
	violations = append(violations, d.analyzeComments(lines, fileInfo.Path)...)
	
	return violations
}

// readFileContent reads the content of a file (simplified implementation)
func (d *RustCommentedCodeDetector) readFileContent(filePath string) (string, error) {
	// In a real implementation, we would read the file here
	// For now, we'll return empty to avoid file system access in this context
	// TODO: Implement proper file reading or pass content through the interface
	return "", nil
}

// analyzeComments analyzes lines for commented-out Rust code
func (d *RustCommentedCodeDetector) analyzeComments(lines []string, filePath string) []*models.Violation {
	var violations []*models.Violation
	var currentComment strings.Builder
	var commentStartLine int
	var inBlockComment bool
	
	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)
		
		// Handle single-line comments
		if strings.HasPrefix(trimmed, "//") {
			if currentComment.Len() == 0 {
				commentStartLine = lineNum
			}
			
			// Extract comment content
			commentText := strings.TrimPrefix(trimmed, "//")
			commentText = strings.TrimSpace(commentText)
			
			// Skip obvious documentation patterns
			if !d.isDocumentation(commentText) {
				currentComment.WriteString(commentText)
				currentComment.WriteString("\n")
			}
		} else if strings.Contains(trimmed, "/*") && strings.Contains(trimmed, "*/") {
			// Single-line block comment
			start := strings.Index(trimmed, "/*")
			end := strings.Index(trimmed, "*/") + 2
			if start < end {
				commentText := trimmed[start+2 : end-2]
				commentText = strings.TrimSpace(commentText)
				
				if !d.isDocumentation(commentText) && d.looksLikeRustCode(commentText) {
					violation := d.createViolation(commentText, lineNum, filePath)
					if violation != nil {
						violations = append(violations, violation)
					}
				}
			}
		} else if strings.Contains(trimmed, "/*") {
			// Start of multi-line block comment
			inBlockComment = true
			commentStartLine = lineNum
			
			start := strings.Index(trimmed, "/*")
			commentText := trimmed[start+2:]
			commentText = strings.TrimSpace(commentText)
			
			if !d.isDocumentation(commentText) {
				currentComment.WriteString(commentText)
				currentComment.WriteString("\n")
			}
		} else if inBlockComment {
			if strings.Contains(trimmed, "*/") {
				// End of multi-line block comment
				inBlockComment = false
				end := strings.Index(trimmed, "*/")
				commentText := trimmed[:end]
				commentText = strings.TrimSpace(commentText)
				
				if !d.isDocumentation(commentText) {
					currentComment.WriteString(commentText)
				}
				
				// Check accumulated comment
				fullComment := currentComment.String()
				if violation := d.createViolation(fullComment, commentStartLine, filePath); violation != nil {
					violations = append(violations, violation)
				}
				
				currentComment.Reset()
			} else {
				// Middle of multi-line block comment
				if !d.isDocumentation(trimmed) {
					currentComment.WriteString(trimmed)
					currentComment.WriteString("\n")
				}
			}
		} else {
			// Not a comment line, check accumulated single-line comments
			if currentComment.Len() > 0 {
				fullComment := currentComment.String()
				if violation := d.createViolation(fullComment, commentStartLine, filePath); violation != nil {
					violations = append(violations, violation)
				}
				currentComment.Reset()
			}
		}
	}
	
	// Handle any remaining comment at end of file
	if currentComment.Len() > 0 {
		fullComment := currentComment.String()
		if violation := d.createViolation(fullComment, commentStartLine, filePath); violation != nil {
			violations = append(violations, violation)
		}
	}
	
	return violations
}

// createViolation creates a violation if the comment looks like Rust code
func (d *RustCommentedCodeDetector) createViolation(commentText string, line int, filePath string) *models.Violation {
	if !d.looksLikeRustCode(commentText) {
		return nil
	}
	
	// Count the number of Rust code-like patterns
	matches := d.codePattern.FindAllString(commentText, -1)
	if len(matches) >= 2 { // At least 2 Rust patterns to be considered commented code
		snippet := commentText
		if len(snippet) > 100 {
			snippet = snippet[:97] + "..."
		}
		
		return &models.Violation{
			Type:        models.ViolationTypeCommentedCode,
			Severity:    models.SeverityLow,
			File:        filePath,
			Line:        line,
			Column:      0,
			Message:     "Block of commented-out Rust code detected",
			Suggestion:  "Remove commented-out code. Use version control to preserve old code if needed",
			CodeSnippet: snippet,
		}
	}
	
	return nil
}

// looksLikeRustCode checks if text looks like commented-out Rust code
func (d *RustCommentedCodeDetector) looksLikeRustCode(text string) bool {
	// Check for common Rust patterns
	if d.codePattern.MatchString(text) {
		// Additional checks to reduce false positives
		
		// If it's very short, it's probably not code
		if len(text) < 15 {
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
		
		// Check for Rust-specific assignment patterns
		if strings.Contains(text, "let ") || strings.Contains(text, "= ") {
			return true
		}
		
		// Check for function/method calls with Rust syntax
		if regexp.MustCompile(`\w+\(`).MatchString(text) ||
		   regexp.MustCompile(`\w+::\w+`).MatchString(text) {
			return true
		}
		
		// Check for Rust-specific operators and syntax
		if strings.Contains(text, "->") || strings.Contains(text, "::") ||
		   strings.Contains(text, "&mut") || strings.Contains(text, "&") {
			return true
		}
		
		// Check for Rust macros
		if strings.Contains(text, "!") && regexp.MustCompile(`\w+!`).MatchString(text) {
			return true
		}
		
		return true
	}
	
	return false
}

// isDocumentation checks if text is likely documentation rather than commented code
func (d *RustCommentedCodeDetector) isDocumentation(text string) bool {
	// Common documentation patterns
	docPatterns := []string{
		"TODO:", "FIXME:", "NOTE:", "HACK:", "XXX:", "BUG:",
		"Copyright", "License", "Author",
		"@param", "@return", "@throws", // Javadoc style
		"Example:", "Usage:", "Description:",
		"//!", "///", // Rust doc comments
	}
	
	textLower := strings.ToLower(text)
	for _, pattern := range docPatterns {
		if strings.Contains(textLower, strings.ToLower(pattern)) {
			return true
		}
	}
	
	// Check for Rust doc comment markers
	if strings.HasPrefix(text, "!") || strings.HasPrefix(text, "/") {
		return true
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