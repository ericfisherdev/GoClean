package violations

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// RustTodoTrackerDetector tracks TODO, FIXME, and other technical debt markers in Rust code
type RustTodoTrackerDetector struct {
	config        *DetectorConfig
	markerPattern *regexp.Regexp
}

// NewRustTodoTrackerDetector creates a new Rust TODO/FIXME tracker detector
func NewRustTodoTrackerDetector(config *DetectorConfig) *RustTodoTrackerDetector {
	// Pattern to detect technical debt markers in Rust comments
	// Includes Rust-specific markers and common ones
	markerPattern := regexp.MustCompile(`(?i)\b(TODO|FIXME|HACK|XXX|BUG|OPTIMIZE|REFACTOR|PANIC|UNWRAP|UNIMPLEMENTED|UNREACHABLE)\b\s*:?\s*(.*)`)
	
	return &RustTodoTrackerDetector{
		config:        config,
		markerPattern: markerPattern,
	}
}

// Name returns the name of this detector
func (d *RustTodoTrackerDetector) Name() string {
	return "Rust Technical Debt Tracker"
}

// Description returns a description of what this detector checks for
func (d *RustTodoTrackerDetector) Description() string {
	return "Tracks TODO, FIXME, and other technical debt markers in Rust code comments"
}

// Detect analyzes the provided file information and returns violations
func (d *RustTodoTrackerDetector) Detect(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation {
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
	violations = append(violations, d.analyzeCommentsForMarkers(lines, fileInfo.Path)...)
	
	return violations
}

// readFileContent reads the content of a file (simplified implementation)
func (d *RustTodoTrackerDetector) readFileContent(filePath string) (string, error) {
	// In a real implementation, we would read the file here
	// For now, we'll return empty to avoid file system access in this context
	// TODO: Implement proper file reading or pass content through the interface
	return "", nil
}

// analyzeCommentsForMarkers analyzes lines for technical debt markers
func (d *RustTodoTrackerDetector) analyzeCommentsForMarkers(lines []string, filePath string) []*models.Violation {
	var violations []*models.Violation
	var inBlockComment bool
	
	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)
		
		// Handle single-line comments
		if strings.HasPrefix(trimmed, "//") {
			commentText := strings.TrimPrefix(trimmed, "//")
			commentText = strings.TrimSpace(commentText)
			
			// Skip doc comments (they usually contain legitimate TODOs)
			if !strings.HasPrefix(trimmed, "///") && !strings.HasPrefix(trimmed, "//!") {
				if violation := d.checkForMarkers(commentText, lineNum, filePath); violation != nil {
					violations = append(violations, violation)
				}
			}
		} else if strings.Contains(trimmed, "/*") && strings.Contains(trimmed, "*/") {
			// Single-line block comment
			start := strings.Index(trimmed, "/*")
			end := strings.Index(trimmed, "*/") + 2
			if start < end {
				commentText := trimmed[start+2 : end-2]
				commentText = strings.TrimSpace(commentText)
				
				if violation := d.checkForMarkers(commentText, lineNum, filePath); violation != nil {
					violations = append(violations, violation)
				}
			}
		} else if strings.Contains(trimmed, "/*") {
			// Start of multi-line block comment
			inBlockComment = true
			
			start := strings.Index(trimmed, "/*")
			commentText := trimmed[start+2:]
			commentText = strings.TrimSpace(commentText)
			
			if violation := d.checkForMarkers(commentText, lineNum, filePath); violation != nil {
				violations = append(violations, violation)
			}
		} else if inBlockComment {
			if strings.Contains(trimmed, "*/") {
				// End of multi-line block comment
				inBlockComment = false
				end := strings.Index(trimmed, "*/")
				commentText := trimmed[:end]
				commentText = strings.TrimSpace(commentText)
				
				if violation := d.checkForMarkers(commentText, lineNum, filePath); violation != nil {
					violations = append(violations, violation)
				}
			} else {
				// Middle of multi-line block comment
				if violation := d.checkForMarkers(trimmed, lineNum, filePath); violation != nil {
					violations = append(violations, violation)
				}
			}
		}
		
		// Also check for Rust-specific markers in code (panic!, unimplemented!, unreachable!)
		if !strings.HasPrefix(trimmed, "//") && !inBlockComment {
			if violation := d.checkForRustSpecificMarkers(trimmed, lineNum, filePath); violation != nil {
				violations = append(violations, violation)
			}
		}
	}
	
	return violations
}

// checkForMarkers checks a comment text for technical debt markers
func (d *RustTodoTrackerDetector) checkForMarkers(text string, lineNum int, filePath string) *models.Violation {
	// Check for technical debt markers
	if matches := d.markerPattern.FindStringSubmatch(text); len(matches) > 0 {
		marker := strings.ToUpper(matches[1])
		description := strings.TrimSpace(matches[2])
		
		return &models.Violation{
			Type:        models.ViolationTypeTodo,
			Severity:    d.classifyRustMarkerSeverity(marker),
			File:        filePath,
			Line:        lineNum,
			Column:      0,
			Message:     fmt.Sprintf("%s marker found: %s", marker, d.getRustMarkerDescription(marker)),
			Suggestion:  d.getRustMarkerSuggestion(marker),
			CodeSnippet: fmt.Sprintf("%s: %s", marker, description),
		}
	}
	
	return nil
}

// checkForRustSpecificMarkers checks for Rust-specific markers in code
func (d *RustTodoTrackerDetector) checkForRustSpecificMarkers(line string, lineNum int, filePath string) *models.Violation {
	// Check for Rust-specific panic/error macros that often indicate incomplete code
	patterns := map[string]string{
		`panic!\s*\(\s*(?:"([^"]*)")?\s*\)`: "PANIC",
		`unimplemented!\s*\(\s*(?:"([^"]*)")?\s*\)`: "UNIMPLEMENTED",
		`unreachable!\s*\(\s*(?:"([^"]*)")?\s*\)`: "UNREACHABLE",
		`todo!\s*\(\s*(?:"([^"]*)")?\s*\)`: "TODO",
	}
	
	for pattern, marker := range patterns {
		regex := regexp.MustCompile(pattern)
		if matches := regex.FindStringSubmatch(line); len(matches) > 0 {
			description := ""
			if len(matches) > 1 && matches[1] != "" {
				description = matches[1]
			} else {
				description = "No description provided"
			}
			
			return &models.Violation{
				Type:        models.ViolationTypeTodo,
				Severity:    d.classifyRustMarkerSeverity(marker),
				File:        filePath,
				Line:        lineNum,
				Column:      0,
				Message:     fmt.Sprintf("%s macro found: %s", marker, d.getRustMarkerDescription(marker)),
				Suggestion:  d.getRustMarkerSuggestion(marker),
				CodeSnippet: fmt.Sprintf("%s: %s", marker, description),
			}
		}
	}
	
	return nil
}

// classifyRustMarkerSeverity classifies the severity based on the Rust marker type
func (d *RustTodoTrackerDetector) classifyRustMarkerSeverity(marker string) models.Severity {
	switch marker {
	case "BUG", "FIXME", "PANIC":
		return models.SeverityHigh
	case "HACK", "XXX", "UNWRAP", "UNIMPLEMENTED", "UNREACHABLE":
		return models.SeverityMedium
	case "TODO", "OPTIMIZE", "REFACTOR":
		return models.SeverityLow
	default:
		return models.SeverityInfo
	}
}

// getRustMarkerDescription provides a description for each Rust marker type
func (d *RustTodoTrackerDetector) getRustMarkerDescription(marker string) string {
	switch marker {
	case "TODO":
		return "Pending task or feature implementation"
	case "FIXME":
		return "Known issue that needs fixing"
	case "HACK":
		return "Temporary workaround that should be improved"
	case "XXX":
		return "Warning about problematic or dangerous code"
	case "BUG":
		return "Known bug in the code"
	case "OPTIMIZE":
		return "Performance optimization opportunity"
	case "REFACTOR":
		return "Code that needs restructuring"
	case "PANIC":
		return "Code that intentionally panics - may need proper error handling"
	case "UNWRAP":
		return "Use of unwrap() which can panic - consider proper error handling"
	case "UNIMPLEMENTED":
		return "Placeholder for unimplemented functionality"
	case "UNREACHABLE":
		return "Code marked as unreachable - verify logic correctness"
	default:
		return "Technical debt marker"
	}
}

// getRustMarkerSuggestion provides suggestions for each Rust marker type
func (d *RustTodoTrackerDetector) getRustMarkerSuggestion(marker string) string {
	switch marker {
	case "TODO":
		return "Schedule and complete this task, or create a ticket in your issue tracker"
	case "FIXME":
		return "Prioritize fixing this issue as it may cause problems"
	case "HACK":
		return "Replace this workaround with a proper solution"
	case "XXX":
		return "Review and address this problematic code section"
	case "BUG":
		return "Fix this bug and add tests to prevent regression"
	case "OPTIMIZE":
		return "Consider optimizing if this is a performance bottleneck"
	case "REFACTOR":
		return "Plan and execute refactoring to improve code structure"
	case "PANIC":
		return "Replace panic! with proper error handling using Result<T, E> or Option<T>"
	case "UNWRAP":
		return "Replace unwrap() with proper error handling using match, if let, or ?"
	case "UNIMPLEMENTED":
		return "Implement the missing functionality or remove if no longer needed"
	case "UNREACHABLE":
		return "Verify that this code is truly unreachable or handle the case properly"
	default:
		return "Address this technical debt item"
	}
}