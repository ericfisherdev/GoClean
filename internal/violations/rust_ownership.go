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

// RustOwnershipDetector detects ownership and borrowing violations in Rust code
type RustOwnershipDetector struct {
	config        *DetectorConfig
	codeExtractor *CodeExtractor
}

// NewRustOwnershipDetector creates a new Rust ownership violation detector
func NewRustOwnershipDetector(config *DetectorConfig) *RustOwnershipDetector {
	if config == nil {
		config = DefaultDetectorConfig()
	}
	return &RustOwnershipDetector{
		config:        config,
		codeExtractor: NewCodeExtractor(),
	}
}

// Name returns the name of this detector
func (d *RustOwnershipDetector) Name() string {
	return "Rust Ownership Analysis"
}

// Description returns a description of what this detector checks for
func (d *RustOwnershipDetector) Description() string {
	return "Detects ownership and borrowing violations in Rust code including unnecessary clones, inefficient borrowing patterns, and complex lifetime annotations"
}

// Detect analyzes Rust code for ownership violations
func (d *RustOwnershipDetector) Detect(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation {
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

	// Analyze content for ownership violations
	violations = append(violations, d.detectUnnecessaryClones(fileInfo.Path, lines)...)
	violations = append(violations, d.detectInefficientBorrowing(fileInfo.Path, lines)...)
	violations = append(violations, d.detectComplexLifetimes(fileInfo.Path, lines)...)
	violations = append(violations, d.detectMoveSemanticsViolations(fileInfo.Path, lines)...)
	violations = append(violations, d.detectBorrowCheckerBypass(fileInfo.Path, lines)...)

	return violations
}

// detectUnnecessaryClones identifies unnecessary use of .clone()
func (d *RustOwnershipDetector) detectUnnecessaryClones(filePath string, lines []string) []*models.Violation {
	var violations []*models.Violation

	// Pattern to find all clone usages
	clonePattern := regexp.MustCompile(`(\w+)\.clone\(\)`)

	for lineNum, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Skip comments and empty lines
		if strings.HasPrefix(trimmedLine, "//") || trimmedLine == "" {
			continue
		}

		// Find all clone usages in the line
		if clonePattern.MatchString(trimmedLine) {
			// Analyze if this clone is likely unnecessary
			if d.isLikelyUnnecessaryClone(trimmedLine) {
				context := d.getCloneContext(trimmedLine)
				codeSnippet := d.extractCodeSnippet(filePath, lineNum+1, lineNum+1)
				violations = append(violations, &models.Violation{
					Type:        models.ViolationTypeRustUnnecessaryClone,
					Severity:    models.SeverityLow,
					Message:     fmt.Sprintf("Unnecessary clone() usage%s - consider borrowing with & instead", context),
					File:        filePath,
					Line:        lineNum + 1,
					Column:      strings.Index(line, ".clone()") + 1,
					Rule:        "rust-unnecessary-clone",
					Suggestion:  d.getCloneSuggestion(trimmedLine),
					CodeSnippet: codeSnippet,
				})
			}
		}
	}

	return violations
}

// detectInefficientBorrowing identifies inefficient borrowing patterns
func (d *RustOwnershipDetector) detectInefficientBorrowing(filePath string, lines []string) []*models.Violation {
	var violations []*models.Violation

	// Patterns for inefficient borrowing
	inefficientPatterns := []*regexp.Regexp{
		// Multiple dereferences that could be simplified
		regexp.MustCompile(`\*\*+\w+`),
		// Unnecessary reference-dereference chains
		regexp.MustCompile(`&\*(\w+)`),
		// Complex borrowing in simple contexts
		regexp.MustCompile(`&&&+\w+`),
		// Borrowing numeric literals
		regexp.MustCompile(`&\d+\b`),
	}

	for lineNum, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Skip comments and empty lines
		if strings.HasPrefix(trimmedLine, "//") || trimmedLine == "" {
			continue
		}

		for _, pattern := range inefficientPatterns {
			if pattern.MatchString(trimmedLine) {
				context := d.getBorrowingContext(trimmedLine)
				codeSnippet := d.extractCodeSnippet(filePath, lineNum+1, lineNum+1)
				
				violations = append(violations, &models.Violation{
					Type:        models.ViolationTypeRustInefficientBorrowing,
					Severity:    models.SeverityLow,
					Message:     fmt.Sprintf("Inefficient borrowing pattern detected: %s", context),
					File:        filePath,
					Line:        lineNum + 1,
					Column:      d.findPatternColumn(line, pattern),
					Rule:        "rust-inefficient-borrowing",
					Suggestion:  d.getBorrowingSuggestion(trimmedLine),
					CodeSnippet: codeSnippet,
				})
			}
		}
	}

	return violations
}

// detectComplexLifetimes identifies overly complex lifetime annotations
func (d *RustOwnershipDetector) detectComplexLifetimes(filePath string, lines []string) []*models.Violation {
	var violations []*models.Violation

	// Pattern for lifetime annotations
	lifetimePattern := regexp.MustCompile(`'[a-zA-Z_][a-zA-Z0-9_]*`)
	complexLifetimePatterns := []*regexp.Regexp{
		// Functions with many lifetime parameters
		regexp.MustCompile(`fn\s+\w+<[^>]*'[^>]*'[^>]*'[^>]*>`),
		// Nested lifetime annotations
		regexp.MustCompile(`<[^>]*<[^>]*'[a-zA-Z_][a-zA-Z0-9_]*[^>]*>[^>]*>`),
		// Very long lifetime parameter lists
		regexp.MustCompile(`<[^>]{50,}>`),
	}

	for lineNum, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Skip comments and empty lines
		if strings.HasPrefix(trimmedLine, "//") || trimmedLine == "" {
			continue
		}

		// Count lifetime annotations in a single line
		lifetimeMatches := lifetimePattern.FindAllString(trimmedLine, -1)
		if len(lifetimeMatches) > 3 {
			codeSnippet := d.extractCodeSnippet(filePath, lineNum+1, lineNum+1)
			violations = append(violations, &models.Violation{
				Type:        models.ViolationTypeRustComplexLifetime,
				Severity:    models.SeverityMedium,
				Message:     fmt.Sprintf("Complex lifetime annotations (%d lifetimes in one line) - consider simplifying", len(lifetimeMatches)),
				File:        filePath,
				Line:        lineNum + 1,
				Column:      strings.Index(line, "'") + 1,
				Rule:        "rust-complex-lifetime",
				Suggestion:  "Consider breaking down complex lifetime relationships or using lifetime elision where possible",
				CodeSnippet: codeSnippet,
			})
		}

		// Check for other complex lifetime patterns
		for _, pattern := range complexLifetimePatterns {
			if pattern.MatchString(trimmedLine) {
				codeSnippet := d.extractCodeSnippet(filePath, lineNum+1, lineNum+1)
				violations = append(violations, &models.Violation{
					Type:        models.ViolationTypeRustComplexLifetime,
					Severity:    models.SeverityMedium,
					Message:     "Overly complex lifetime annotations detected",
					File:        filePath,
					Line:        lineNum + 1,
					Column:      d.findPatternColumn(line, pattern),
					Rule:        "rust-complex-lifetime",
					Suggestion:  "Consider restructuring code to reduce lifetime complexity or using lifetime elision",
					CodeSnippet: codeSnippet,
				})
			}
		}
	}

	return violations
}

// detectMoveSemanticsViolations identifies improper use of move semantics
func (d *RustOwnershipDetector) detectMoveSemanticsViolations(filePath string, lines []string) []*models.Violation {
	var violations []*models.Violation

	moveSemanticsPatterns := []*regexp.Regexp{
		// Using move when copy would be more appropriate
		regexp.MustCompile(`move\s*\|\|\s*\{\s*\w+\s*\}`),
		// Moving values unnecessarily in simple operations
		regexp.MustCompile(`let\s+\w+\s*=\s*move\s+\w+;`),
		// Move in closure that doesn't need it
		regexp.MustCompile(`\|[^|]*\|\s*move\s*\{[^}]*\w+\s*\}`),
	}

	for lineNum, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Skip comments and empty lines
		if strings.HasPrefix(trimmedLine, "//") || trimmedLine == "" {
			continue
		}

		for _, pattern := range moveSemanticsPatterns {
			if pattern.MatchString(trimmedLine) {
				// Additional context checks
				if d.isLikelyUnnecessaryMove(trimmedLine) {
					codeSnippet := d.extractCodeSnippet(filePath, lineNum+1, lineNum+1)
					violations = append(violations, &models.Violation{
						Type:        models.ViolationTypeRustMoveSemanticsViolation,
						Severity:    models.SeverityLow,
						Message:     "Potentially unnecessary move semantics usage",
						File:        filePath,
						Line:        lineNum + 1,
						Column:      strings.Index(line, "move") + 1,
						Rule:        "rust-move-semantics",
						Suggestion:  "Consider if borrowing or copying would be more appropriate than moving",
						CodeSnippet: codeSnippet,
					})
				}
			}
		}
	}

	return violations
}

// detectBorrowCheckerBypass identifies attempts to bypass the borrow checker
func (d *RustOwnershipDetector) detectBorrowCheckerBypass(filePath string, lines []string) []*models.Violation {
	var violations []*models.Violation

	bypassPatterns := []*regexp.Regexp{
		// Using unsafe blocks (potential bypass)
		regexp.MustCompile(`unsafe\s*\{`),
		// Transmute for ownership manipulation
		regexp.MustCompile(`transmute\s*\(`),
		// Raw pointer manipulation to bypass borrow checker
		regexp.MustCompile(`\*const\s+\w+\s*as\s*\*mut`),
		// Using from_raw unnecessarily
		regexp.MustCompile(`from_raw\s*\(`),
	}

	for lineNum, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Skip comments and empty lines
		if strings.HasPrefix(trimmedLine, "//") || trimmedLine == "" {
			continue
		}

		for _, pattern := range bypassPatterns {
			if pattern.MatchString(trimmedLine) {
				severity := models.SeverityHigh
				message := "Potential borrow checker bypass detected"
				
				if strings.Contains(trimmedLine, "transmute") {
					severity = models.SeverityHigh
					message = "Using transmute to bypass ownership rules - highly dangerous"
				} else if strings.Contains(trimmedLine, "unsafe") {
					severity = models.SeverityMedium
					message = "Using unsafe to bypass borrowing rules - ensure this is necessary"
				}

				codeSnippet := d.extractCodeSnippet(filePath, lineNum+1, lineNum+1)
				violations = append(violations, &models.Violation{
					Type:        models.ViolationTypeRustBorrowCheckerBypass,
					Severity:    severity,
					Message:     message,
					File:        filePath,
					Line:        lineNum + 1,
					Column:      d.findPatternColumn(line, pattern),
					Rule:        "rust-borrow-checker-bypass",
					Suggestion:  "Work with the borrow checker instead of bypassing it. Consider redesigning the data flow.",
					CodeSnippet: codeSnippet,
				})
			}
		}
	}

	return violations
}

// Helper methods

func (d *RustOwnershipDetector) isLikelyUnnecessaryClone(line string) bool {
	// Skip if this appears to be a necessary clone for threading/async
	if strings.Contains(line, "thread") || strings.Contains(line, "spawn") || 
	   strings.Contains(line, "async") {
		return false
	}
	
	// Clone in simple assignment without complex ownership transfer
	if strings.Contains(line, "let") && strings.Contains(line, ".clone()") {
		// Only flag if not in move context
		if !strings.Contains(line, "move") {
			return true
		}
	}
	
	// Clone in function call context (common pattern)
	if strings.Contains(line, ".clone()") && strings.Contains(line, "(") && 
	   !strings.Contains(line, "collect") && !strings.Contains(line, "to_owned") {
		return true
	}
	
	// Clone in iterator chains
	if strings.Contains(line, ".iter()") && strings.Contains(line, ".clone()") {
		return true
	}
	
	// Clone in return statements
	if strings.Contains(line, "return") && strings.Contains(line, ".clone()") {
		return true
	}
	
	return false
}

func (d *RustOwnershipDetector) getCloneContext(line string) string {
	if strings.Contains(line, "let") {
		return " for simple assignment"
	}
	if strings.Contains(line, "return") {
		return " in return statement"
	}
	if strings.Contains(line, "(") {
		return " in function call"
	}
	if strings.Contains(line, ".iter()") {
		return " in iterator context"
	}
	return ""
}

func (d *RustOwnershipDetector) getBorrowingContext(line string) string {
	if strings.Contains(line, "**") {
		return "multiple dereferences"
	}
	if strings.Contains(line, "&*") {
		return "reference-dereference chain"
	}
	if strings.Contains(line, "&&&") {
		return "multiple references"
	}
	if strings.Contains(line, "&\"") || strings.Contains(line, "&'") || 
	   strings.Contains(line, "&[0-9]") {
		return "borrowing literal value"
	}
	return "complex borrowing"
}

func (d *RustOwnershipDetector) isLikelyUnnecessaryMove(line string) bool {
	// Move in simple closures
	if strings.Contains(line, "move") && strings.Contains(line, "||") {
		// Check if the closure body is simple
		if strings.Count(line, "\n") == 0 && len(line) < 50 {
			return true
		}
	}
	
	// Move in simple assignment
	if strings.Contains(line, "let") && strings.Contains(line, "move") && 
	   !strings.Contains(line, "thread") && !strings.Contains(line, "async") {
		return true
	}
	
	return false
}

func (d *RustOwnershipDetector) findPatternColumn(line string, pattern *regexp.Regexp) int {
	match := pattern.FindStringIndex(line)
	if match != nil {
		return match[0] + 1
	}
	return 1
}

func (d *RustOwnershipDetector) getCloneSuggestion(line string) string {
	if strings.Contains(line, "let") {
		return "Consider using borrowing (&) instead of cloning for simple assignments"
	}
	if strings.Contains(line, "return") {
		return "Consider returning a reference or using move semantics instead of cloning"
	}
	if strings.Contains(line, ".iter()") {
		return "Use references in iterator chains instead of cloning elements"
	}
	return "Consider whether borrowing (&) would be sufficient instead of cloning"
}

func (d *RustOwnershipDetector) getBorrowingSuggestion(line string) string {
	if strings.Contains(line, "**") {
		return "Simplify multiple dereferences by restructuring data access"
	}
	if strings.Contains(line, "&*") {
		return "Remove unnecessary reference-dereference chain"
	}
	if strings.Contains(line, "&&&") {
		return "Reduce multiple reference levels by simplifying data structure access"
	}
	return "Simplify borrowing pattern for better readability and performance"
}

// extractCodeSnippet extracts a code snippet for the violation with context
func (d *RustOwnershipDetector) extractCodeSnippet(filePath string, startLine, endLine int) string {
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
func (d *RustOwnershipDetector) generateFallbackSnippet(startLine, endLine int) string {
	if endLine <= startLine {
		return fmt.Sprintf("Line %d: <code snippet unavailable>", startLine)
	}
	return fmt.Sprintf("Lines %d-%d: <code snippet unavailable>", startLine, endLine)
}

// readFileContent reads and returns the content of a file
func (d *RustOwnershipDetector) readFileContent(filePath string) (string, error) {
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