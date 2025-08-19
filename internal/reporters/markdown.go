// Package reporters provides output formatters for clean code analysis reports.
package reporters

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ericfisherdev/goclean/internal/models"
)

// MarkdownReporter generates markdown reports
type MarkdownReporter struct {
	config *MarkdownConfig
}

// MarkdownConfig contains markdown-specific configuration
type MarkdownConfig struct {
	OutputPath      string
	IncludeExamples bool
}

// NewMarkdownReporter creates a new markdown reporter
func NewMarkdownReporter(config *MarkdownConfig) *MarkdownReporter {
	return &MarkdownReporter{
		config: config,
	}
}

// Generate creates a markdown report from the provided data
func (m *MarkdownReporter) Generate(report *models.Report) error {
	// Ensure output directory exists
	outputDir := filepath.Dir(m.config.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	file, err := os.Create(m.config.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Generate markdown content
	content := m.generateMarkdown(report)
	
	// Write to file
	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write markdown content: %w", err)
	}

	return nil
}

// generateMarkdown creates the markdown content for the report
func (m *MarkdownReporter) generateMarkdown(report *models.Report) string {
	var md strings.Builder

	// Header
	m.writeHeader(&md, report)
	
	// Executive Summary
	m.writeSummary(&md, report.Summary)
	
	// Statistics
	m.writeStatistics(&md, report.Statistics)
	
	// Top Violated Files
	if len(report.Statistics.TopViolatedFiles) > 0 {
		m.writeTopViolatedFiles(&md, report.Statistics.TopViolatedFiles)
	}
	
	// Detailed Violations
	m.writeDetailedViolations(&md, report)
	
	// Recommendations
	m.writeRecommendations(&md, report)
	
	// Footer
	m.writeFooter(&md, report)

	return md.String()
}

// writeHeader writes the markdown header
func (m *MarkdownReporter) writeHeader(md *strings.Builder, report *models.Report) {
	md.WriteString("# GoClean Code Analysis Report\n\n")
	md.WriteString(fmt.Sprintf("**Generated:** %s  \n", report.GeneratedAt.Format("2006-01-02 15:04:05")))
	md.WriteString(fmt.Sprintf("**Report ID:** %s  \n", report.ID))
	
	if report.Config != nil {
		md.WriteString("**Scanned Paths:**\n")
		for _, path := range report.Config.Paths {
			md.WriteString(fmt.Sprintf("- `%s`\n", path))
		}
	}
	
	md.WriteString("\n---\n\n")
}

// writeSummary writes the executive summary section
func (m *MarkdownReporter) writeSummary(md *strings.Builder, summary *models.ScanSummary) {
	md.WriteString("## Executive Summary\n\n")
	
	if summary.TotalViolations == 0 {
		md.WriteString("🎉 **Excellent!** No clean code violations were found in your codebase.\n\n")
		md.WriteString("Your code follows clean code principles and is well-structured.\n\n")
	} else {
		md.WriteString(fmt.Sprintf("📊 **Analysis Results:** %d violations found across %d files\n\n", 
			summary.TotalViolations, summary.ScannedFiles))
		
		md.WriteString("### Key Metrics\n\n")
		md.WriteString("| Metric | Value |\n")
		md.WriteString("|--------|-------|\n")
		md.WriteString(fmt.Sprintf("| Total Files | %d |\n", summary.TotalFiles))
		md.WriteString(fmt.Sprintf("| Scanned Files | %d |\n", summary.ScannedFiles))
		md.WriteString(fmt.Sprintf("| Skipped Files | %d |\n", summary.SkippedFiles))
		md.WriteString(fmt.Sprintf("| Total Violations | **%d** |\n", summary.TotalViolations))
		md.WriteString(fmt.Sprintf("| Scan Duration | %v |\n", summary.Duration.Round(time.Millisecond)))
	}
	
	md.WriteString("\n")
}

// writeStatistics writes the statistics section
func (m *MarkdownReporter) writeStatistics(md *strings.Builder, stats *models.Statistics) {
	if len(stats.ViolationsByType) == 0 {
		return
	}

	md.WriteString("## Violation Statistics\n\n")
	
	// Violations by Type
	md.WriteString("### Violations by Type\n\n")
	md.WriteString("| Violation Type | Count | Percentage |\n")
	md.WriteString("|----------------|-------|------------|\n")
	
	// Calculate total for percentages
	total := 0
	for _, count := range stats.ViolationsByType {
		total += count
	}
	
	// Sort by count (descending)
	type typeCount struct {
		Type  models.ViolationType
		Count int
	}
	
	var sortedTypes []typeCount
	for vtype, count := range stats.ViolationsByType {
		sortedTypes = append(sortedTypes, typeCount{Type: vtype, Count: count})
	}
	
	sort.Slice(sortedTypes, func(i, j int) bool {
		return sortedTypes[i].Count > sortedTypes[j].Count
	})
	
	for _, tc := range sortedTypes {
		percentage := float64(tc.Count) / float64(total) * 100
		md.WriteString(fmt.Sprintf("| %s | %d | %.1f%% |\n", 
			tc.Type.GetDisplayName(), tc.Count, percentage))
	}
	
	md.WriteString("\n")
	
	// Violations by Severity
	md.WriteString("### Violations by Severity\n\n")
	md.WriteString("| Severity | Count | Status |\n")
	md.WriteString("|----------|-------|--------|\n")
	
	severities := []models.Severity{models.SeverityCritical, models.SeverityHigh, models.SeverityMedium, models.SeverityLow}
	for _, severity := range severities {
		if count, exists := stats.ViolationsBySeverity[severity]; exists && count > 0 {
			status := m.getSeverityStatus(severity)
			md.WriteString(fmt.Sprintf("| %s | %d | %s |\n", 
				severity.String(), count, status))
		}
	}
	
	md.WriteString("\n")
}

// writeTopViolatedFiles writes the most violated files section
func (m *MarkdownReporter) writeTopViolatedFiles(md *strings.Builder, topFiles []*models.FileViolationSummary) {
	md.WriteString("## Most Violated Files\n\n")
	md.WriteString("The following files have the highest number of violations and should be prioritized for refactoring:\n\n")
	
	md.WriteString("| File | Violations | Lines | Violations/100 Lines | Severity Distribution |\n")
	md.WriteString("|------|------------|-------|----------------------|----------------------|\n")
	
	displayCount := len(topFiles)
	if displayCount > 10 {
		displayCount = 10
	}
	
	for i := 0; i < displayCount; i++ {
		file := topFiles[i]
		rate := float64(file.TotalViolations) / float64(file.Lines) * 100
		
		// Create severity distribution
		var severityParts []string
		severities := []models.Severity{models.SeverityCritical, models.SeverityHigh, models.SeverityMedium, models.SeverityLow}
		for _, severity := range severities {
			if count, exists := file.ViolationsBySeverity[severity]; exists && count > 0 {
				severityParts = append(severityParts, fmt.Sprintf("%s:%d", severity.String(), count))
			}
		}
		severityDist := strings.Join(severityParts, ", ")
		
		md.WriteString(fmt.Sprintf("| `%s` | %d | %d | %.1f | %s |\n",
			filepath.Base(file.File), file.TotalViolations, file.Lines, rate, severityDist))
	}
	
	md.WriteString("\n")
}

// writeDetailedViolations writes detailed violations section
func (m *MarkdownReporter) writeDetailedViolations(md *strings.Builder, report *models.Report) {
	violationsByFile := report.GetViolationsByFile()
	if len(violationsByFile) == 0 {
		return
	}

	md.WriteString("## Detailed Violations\n\n")
	md.WriteString("### Per-File Breakdown\n\n")
	
	// Sort files by violation count
	type fileViolations struct {
		File       string
		Violations []*models.Violation
	}
	
	var sortedFiles []fileViolations
	for file, violations := range violationsByFile {
		sortedFiles = append(sortedFiles, fileViolations{
			File:       file,
			Violations: violations,
		})
	}
	
	sort.Slice(sortedFiles, func(i, j int) bool {
		return len(sortedFiles[i].Violations) > len(sortedFiles[j].Violations)
	})
	
	for _, fv := range sortedFiles {
		md.WriteString(fmt.Sprintf("#### 📁 %s\n\n", fv.File))
		md.WriteString(fmt.Sprintf("**%d violations found**\n\n", len(fv.Violations)))
		
		// Sort violations by line number
		sort.Slice(fv.Violations, func(i, j int) bool {
			return fv.Violations[i].Line < fv.Violations[j].Line
		})
		
		// Group violations by type for cleaner presentation
		violationsByType := make(map[models.ViolationType][]*models.Violation)
		for _, v := range fv.Violations {
			violationsByType[v.Type] = append(violationsByType[v.Type], v)
		}
		
		for vtype, violations := range violationsByType {
			// Add Rust-specific violation categorization if applicable
			if models.IsRustSpecificViolation(vtype) {
				category := models.GetRustViolationCategory(vtype)
				md.WriteString(fmt.Sprintf("**%s (%d) - %s Category**\n\n", vtype.GetDisplayName(), len(violations), strings.Title(string(category))))
			} else {
				md.WriteString(fmt.Sprintf("**%s (%d)**\n\n", vtype.GetDisplayName(), len(violations)))
			}
			
			for _, violation := range violations {
				md.WriteString(fmt.Sprintf("- **Line %d** (%s): %s\n", 
					violation.Line, violation.Severity.String(), violation.Message))
				
				if violation.Description != "" {
					md.WriteString(fmt.Sprintf("  - *%s*\n", violation.Description))
				}
				
				if violation.Suggestion != "" {
					md.WriteString(fmt.Sprintf("  - 💡 **Suggestion:** %s\n", violation.Suggestion))
				}
				
				if m.config.IncludeExamples && violation.CodeSnippet != "" {
					md.WriteString("  - **Code:**\n")
					language := m.detectLanguageFromFile(fv.File)
					md.WriteString(fmt.Sprintf("    ```%s\n", language))
					lines := strings.Split(violation.CodeSnippet, "\n")
					for _, line := range lines {
						if strings.TrimSpace(line) != "" {
							md.WriteString(fmt.Sprintf("    %s\n", line))
						}
					}
					md.WriteString("    ```\n")
				}
				
				md.WriteString("\n")
			}
		}
		
		md.WriteString("---\n\n")
	}
}

// writeRecommendations writes actionable recommendations
func (m *MarkdownReporter) writeRecommendations(md *strings.Builder, report *models.Report) {
	if report.Summary.TotalViolations == 0 {
		return
	}

	md.WriteString("## Recommendations\n\n")
	md.WriteString("Based on the analysis results, here are prioritized recommendations for improving your code quality:\n\n")
	
	// Analyze most common violation types and provide specific recommendations
	var recommendations []string
	
	// Check for most common violation types
	maxCount := 0
	var topViolationType models.ViolationType
	for vtype, count := range report.Statistics.ViolationsByType {
		if count > maxCount {
			maxCount = count
			topViolationType = vtype
		}
	}
	
	// Provide specific recommendations based on top violation types
	switch topViolationType {
	case models.ViolationTypeFunctionLength:
		recommendations = append(recommendations, 
			"🔧 **Break down long functions** - Extract smaller, focused functions with single responsibilities",
			"📏 **Aim for functions under 20-25 lines** - This improves readability and testability")
			
	case models.ViolationTypeCyclomaticComplexity:
		recommendations = append(recommendations,
			"🌀 **Reduce cyclomatic complexity** - Simplify conditional logic and nested statements",
			"🏗️ **Use early returns** - Eliminate deep nesting with guard clauses")
			
	case models.ViolationTypeParameterCount:
		recommendations = append(recommendations,
			"📦 **Group related parameters** - Use structs or configuration objects",
			"🔧 **Apply builder pattern** - For functions with many optional parameters")
			
	case models.ViolationTypeNaming:
		recommendations = append(recommendations,
			"📝 **Improve naming conventions** - Use descriptive, intention-revealing names",
			"🚫 **Avoid abbreviations** - Write full, meaningful variable and function names")
			
	case models.ViolationTypeMissingDocumentation:
		recommendations = append(recommendations,
			"📚 **Add documentation** - Document all public functions and types",
			"💬 **Write meaningful comments** - Explain the why, not the what")
	}
	
	// Add general recommendations
	recommendations = append(recommendations,
		"🧪 **Write unit tests** - Test each function/method independently",
		"🔍 **Regular code reviews** - Catch violations early in the development process",
		"⚙️ **Integrate with CI/CD** - Run GoClean automatically on every commit",
		"📊 **Monitor trends** - Track improvement over time")
	
	for _, rec := range recommendations {
		md.WriteString(fmt.Sprintf("### %s\n\n", rec))
	}
}

// writeFooter writes the report footer
func (m *MarkdownReporter) writeFooter(md *strings.Builder, report *models.Report) {
	md.WriteString("---\n\n")
	md.WriteString("## About This Report\n\n")
	md.WriteString("This report was generated by **GoClean**, a static code analysis tool focused on clean code principles.\n\n")
	
	if report.Config != nil && report.Config.Thresholds != nil {
		md.WriteString("### Thresholds Used\n\n")
		md.WriteString("| Metric | Threshold |\n")
		md.WriteString("|--------|----------|\n")
		md.WriteString(fmt.Sprintf("| Function Lines | %d |\n", report.Config.Thresholds.FunctionLines))
		md.WriteString(fmt.Sprintf("| Cyclomatic Complexity | %d |\n", report.Config.Thresholds.CyclomaticComplexity))
		md.WriteString(fmt.Sprintf("| Parameters | %d |\n", report.Config.Thresholds.Parameters))
		md.WriteString(fmt.Sprintf("| Nesting Depth | %d |\n", report.Config.Thresholds.NestingDepth))
		md.WriteString(fmt.Sprintf("| Class Lines | %d |\n", report.Config.Thresholds.ClassLines))
		md.WriteString("\n")
	}
	
	md.WriteString(fmt.Sprintf("**Report generated in:** %v  \n", report.Summary.Duration.Round(time.Millisecond)))
	md.WriteString(fmt.Sprintf("**Timestamp:** %s  \n", report.GeneratedAt.Format(time.RFC3339)))
	
	// Add some final motivational message based on results
	if report.Summary.TotalViolations == 0 {
		md.WriteString("\n🎉 **Congratulations!** Your code demonstrates excellent adherence to clean code principles. Keep up the great work!\n")
	} else if report.Summary.TotalViolations < 10 {
		md.WriteString("\n✨ **Good job!** You have relatively few violations. A small amount of refactoring will make your code even cleaner.\n")
	} else {
		md.WriteString("\n💪 **Improvement opportunity!** Focus on the recommendations above to significantly enhance your code quality.\n")
	}
}

// getSeverityStatus returns a status emoji/text for severity levels
func (m *MarkdownReporter) getSeverityStatus(severity models.Severity) string {
	switch severity {
	case models.SeverityLow:
		return "ℹ️ Info"
	case models.SeverityMedium:
		return "⚠️ Warning"
	case models.SeverityHigh:
		return "🚨 High Priority"
	case models.SeverityCritical:
		return "💥 Critical"
	default:
		return "❓ Unknown"
	}
}

// detectLanguageFromFile detects programming language from file extension for syntax highlighting
func (m *MarkdownReporter) detectLanguageFromFile(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".rs":
		return "rust"
	case ".go":
		return "go"
	case ".js", ".jsx":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".cs":
		return "csharp"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".cxx", ".hpp":
		return "cpp"
	case ".php":
		return "php"
	case ".rb":
		return "ruby"
	case ".swift":
		return "swift"
	case ".kt", ".kts":
		return "kotlin"
	case ".scala":
		return "scala"
	default:
		return "text"
	}
}