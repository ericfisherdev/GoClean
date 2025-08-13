package violations

import (
	"fmt"
	"go/ast"
	"go/token"
	"regexp"
	"strings"

	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/types"
)

// TodoTrackerDetector tracks TODO, FIXME, and other technical debt markers
type TodoTrackerDetector struct {
	config      *DetectorConfig
	markerPattern *regexp.Regexp
}

// NewTodoTrackerDetector creates a new TODO/FIXME tracker detector
func NewTodoTrackerDetector(config *DetectorConfig) *TodoTrackerDetector {
	// Pattern to detect technical debt markers
	markerPattern := regexp.MustCompile(`(?i)\b(TODO|FIXME|HACK|XXX|BUG|OPTIMIZE|REFACTOR)\b\s*:?\s*(.*)`)
	
	return &TodoTrackerDetector{
		config:        config,
		markerPattern: markerPattern,
	}
}

// Name returns the name of this detector
func (d *TodoTrackerDetector) Name() string {
	return "Technical Debt Tracker"
}

// Description returns a description of what this detector checks for
func (d *TodoTrackerDetector) Description() string {
	return "Tracks TODO, FIXME, and other technical debt markers in code"
}

// Detect analyzes the provided file information and returns violations
func (d *TodoTrackerDetector) Detect(fileInfo *models.FileInfo, astInfo interface{}) []*models.Violation {
	var violations []*models.Violation
	
	if astInfo == nil {
		return violations
	}
	
	// Type assertion to get types.GoASTInfo
	goAstInfo, ok := astInfo.(*types.GoASTInfo)
	if !ok || goAstInfo == nil || goAstInfo.AST == nil {
		return violations
	}
	
	// Check all comments for technical debt markers
	for _, commentGroup := range goAstInfo.AST.Comments {
		violations = append(violations, d.checkCommentGroupForMarkers(commentGroup, goAstInfo.FileSet, fileInfo.Path)...)
	}
	
	return violations
}

// checkCommentGroupForMarkers checks a comment group for technical debt markers
func (d *TodoTrackerDetector) checkCommentGroupForMarkers(group *ast.CommentGroup, fset *token.FileSet, filePath string) []*models.Violation {
	var violations []*models.Violation
	
	if group == nil {
		return violations
	}
	
	for _, comment := range group.List {
		text := comment.Text
		
		// Remove comment markers
		if strings.HasPrefix(text, "//") {
			text = strings.TrimPrefix(text, "//")
		} else if strings.HasPrefix(text, "/*") && strings.HasSuffix(text, "*/") {
			text = strings.TrimPrefix(text, "/*")
			text = strings.TrimSuffix(text, "*/")
		}
		
		text = strings.TrimSpace(text)
		
		// Check for technical debt markers
		if matches := d.markerPattern.FindStringSubmatch(text); len(matches) > 0 {
			marker := strings.ToUpper(matches[1])
			description := strings.TrimSpace(matches[2])
			
			pos := fset.Position(comment.Pos())
			
			violation := &models.Violation{
				Type:        models.ViolationTypeTodo,
				Severity:    d.classifyMarkerSeverity(marker),
				File:        filePath,
				Line:        pos.Line,
				Column:      pos.Column,
				Message:     fmt.Sprintf("%s marker found: %s", marker, d.getMarkerDescription(marker)),
				Suggestion:  d.getMarkerSuggestion(marker),
				CodeSnippet: fmt.Sprintf("%s: %s", marker, description),
			}
			
			violations = append(violations, violation)
		}
	}
	
	return violations
}

// classifyMarkerSeverity classifies the severity based on the marker type
func (d *TodoTrackerDetector) classifyMarkerSeverity(marker string) models.Severity {
	switch marker {
	case "BUG", "FIXME":
		return models.SeverityHigh
	case "HACK", "XXX":
		return models.SeverityMedium
	case "TODO", "OPTIMIZE", "REFACTOR":
		return models.SeverityLow
	default:
		return models.SeverityInfo
	}
}

// getMarkerDescription provides a description for each marker type
func (d *TodoTrackerDetector) getMarkerDescription(marker string) string {
	switch marker {
	case "TODO":
		return "Pending task or feature"
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
	default:
		return "Technical debt marker"
	}
}

// getMarkerSuggestion provides suggestions for each marker type
func (d *TodoTrackerDetector) getMarkerSuggestion(marker string) string {
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
	default:
		return "Address this technical debt item"
	}
}