package reporters

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ericfisherdev/goclean/internal/models"
)

// ConsoleReporter outputs reports to the console
type ConsoleReporter struct {
	verbose bool
	colors  bool
}

// NewConsoleReporter creates a new console reporter
func NewConsoleReporter(verbose, colors bool) *ConsoleReporter {
	return &ConsoleReporter{
		verbose: verbose,
		colors:  colors,
	}
}

// Generate prints a report to the console
func (c *ConsoleReporter) Generate(report *models.Report) error {
	c.printHeader(report)
	c.printSummary(report.Summary)
	c.printStatistics(report.Statistics)
	
	if c.verbose {
		c.printDetailedViolations(report)
	} else {
		c.printTopViolatedFiles(report.Statistics.TopViolatedFiles)
	}
	
	c.printFooter(report)
	return nil
}

// printHeader prints the report header
func (c *ConsoleReporter) printHeader(report *models.Report) {
	fmt.Println(c.colorize("=== GoClean Code Analysis Report ===", "header"))
	fmt.Printf("Generated: %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Report ID: %s\n\n", report.ID)
}

// printSummary prints the scan summary
func (c *ConsoleReporter) printSummary(summary *models.ScanSummary) {
	fmt.Println(c.colorize("📊 SCAN SUMMARY", "section"))
	fmt.Println(strings.Repeat("-", 40))
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "Total Files:\t%d\n", summary.TotalFiles)
	fmt.Fprintf(w, "Scanned Files:\t%d\n", summary.ScannedFiles)
	fmt.Fprintf(w, "Skipped Files:\t%d\n", summary.SkippedFiles)
	fmt.Fprintf(w, "Total Violations:\t%s\n", c.colorizeViolationCount(summary.TotalViolations))
	fmt.Fprintf(w, "Scan Duration:\t%v\n", summary.Duration.Round(time.Millisecond))
	w.Flush()
	fmt.Println()
}

// printStatistics prints violation statistics
func (c *ConsoleReporter) printStatistics(stats *models.Statistics) {
	if len(stats.ViolationsByType) == 0 {
		fmt.Println(c.colorize("✅ No violations found! Your code follows clean code principles.", "success"))
		return
	}

	fmt.Println(c.colorize("📈 VIOLATIONS BY TYPE", "section"))
	fmt.Println(strings.Repeat("-", 40))
	
	// Sort violation types by count (descending)
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
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	for _, tc := range sortedTypes {
		fmt.Fprintf(w, "%s:\t%s\n", 
			tc.Type.GetDisplayName(), 
			c.colorizeViolationCount(tc.Count))
	}
	w.Flush()
	fmt.Println()

	fmt.Println(c.colorize("🚨 VIOLATIONS BY SEVERITY", "section"))
	fmt.Println(strings.Repeat("-", 40))
	
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	severities := []models.Severity{models.SeverityCritical, models.SeverityHigh, models.SeverityMedium, models.SeverityLow}
	for _, severity := range severities {
		if count, exists := stats.ViolationsBySeverity[severity]; exists && count > 0 {
			fmt.Fprintf(w, "%s:\t%s\n", 
				severity.String(), 
				c.colorizeBySeverity(fmt.Sprintf("%d", count), severity))
		}
	}
	w.Flush()
	fmt.Println()
}

// printTopViolatedFiles prints the most violated files
func (c *ConsoleReporter) printTopViolatedFiles(topFiles []*models.FileViolationSummary) {
	if len(topFiles) == 0 {
		return
	}

	fmt.Println(c.colorize("🔥 TOP VIOLATED FILES", "section"))
	fmt.Println(strings.Repeat("-", 70))
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "FILE\tVIOLATIONS\tLINES\tRATE\n")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", 
		strings.Repeat("-", 30),
		strings.Repeat("-", 10), 
		strings.Repeat("-", 5), 
		strings.Repeat("-", 8))
	
	displayCount := len(topFiles)
	if displayCount > 10 {
		displayCount = 10
	}
	
	for i := 0; i < displayCount; i++ {
		file := topFiles[i]
		rate := float64(file.TotalViolations) / float64(file.Lines) * 100
		
		fileName := file.File
		if len(fileName) > 30 {
			fileName = "..." + fileName[len(fileName)-27:]
		}
		
		fmt.Fprintf(w, "%s\t%s\t%d\t%.1f%%\n",
			fileName,
			c.colorizeViolationCount(file.TotalViolations),
			file.Lines,
			rate)
	}
	w.Flush()
	fmt.Println()
}

// printDetailedViolations prints all violations in detail
func (c *ConsoleReporter) printDetailedViolations(report *models.Report) {
	violationsByFile := report.GetViolationsByFile()
	if len(violationsByFile) == 0 {
		return
	}

	fmt.Println(c.colorize("📋 DETAILED VIOLATIONS", "section"))
	fmt.Println(strings.Repeat("=", 80))
	
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
	
	for i, fv := range sortedFiles {
		if i > 0 {
			fmt.Println()
		}
		
		fmt.Printf("%s %s (%d violations)\n", 
			c.colorize("📁", "file"),
			c.colorize(fv.File, "filename"),
			len(fv.Violations))
		fmt.Println(strings.Repeat("-", 80))
		
		// Sort violations by line number
		sort.Slice(fv.Violations, func(i, j int) bool {
			return fv.Violations[i].Line < fv.Violations[j].Line
		})
		
		for _, violation := range fv.Violations {
			c.printViolation(violation)
		}
	}
}

// printViolation prints a single violation
func (c *ConsoleReporter) printViolation(v *models.Violation) {
	severityIcon := c.getSeverityIcon(v.Severity)
	severityColor := c.getSeverityColor(v.Severity)
	
	fmt.Printf("  %s %s [Line %d] %s\n",
		severityIcon,
		c.colorize(v.Severity.String(), severityColor),
		v.Line,
		v.Type.GetDisplayName())
	
	fmt.Printf("    %s\n", v.Message)
	
	if v.Description != "" {
		fmt.Printf("    %s\n", c.colorize(v.Description, "description"))
	}
	
	if v.Suggestion != "" {
		fmt.Printf("    💡 %s\n", c.colorize(v.Suggestion, "suggestion"))
	}
	
	if c.verbose && v.CodeSnippet != "" {
		fmt.Printf("    Code:\n")
		lines := strings.Split(v.CodeSnippet, "\n")
		for i, line := range lines {
			if strings.TrimSpace(line) != "" {
				fmt.Printf("      %d: %s\n", v.Line+i, line)
			}
		}
	}
	
	fmt.Println()
}

// printFooter prints the report footer
func (c *ConsoleReporter) printFooter(report *models.Report) {
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Report completed in %v\n", report.Summary.Duration.Round(time.Millisecond))
	
	if report.Summary.TotalViolations == 0 {
		fmt.Println(c.colorize("🎉 Great job! No clean code violations found.", "success"))
	} else {
		fmt.Printf("%s %d violations need attention.\n", 
			c.colorize("⚠️", "warning"),
			report.Summary.TotalViolations)
	}
}

// colorize applies color codes to text if colors are enabled
func (c *ConsoleReporter) colorize(text, colorType string) string {
	if !c.colors {
		return text
	}
	
	colors := map[string]string{
		"header":      "\033[1;36m",  // Cyan bold
		"section":     "\033[1;34m",  // Blue bold
		"success":     "\033[1;32m",  // Green bold
		"warning":     "\033[1;33m",  // Yellow bold
		"error":       "\033[1;31m",  // Red bold
		"filename":    "\033[1;37m",  // White bold
		"description": "\033[0;90m",  // Dark gray
		"suggestion":  "\033[0;36m",  // Cyan
		"file":        "\033[0;35m",  // Magenta
		"reset":       "\033[0m",     // Reset
	}
	
	if color, exists := colors[colorType]; exists {
		return color + text + colors["reset"]
	}
	
	return text
}

// colorizeBySeverity applies color based on severity level
func (c *ConsoleReporter) colorizeBySeverity(text string, severity models.Severity) string {
	if !c.colors {
		return text
	}
	
	colorType := c.getSeverityColor(severity)
	return c.colorize(text, colorType)
}

// colorizeViolationCount colors violation counts based on their value
func (c *ConsoleReporter) colorizeViolationCount(count int) string {
	if !c.colors {
		return fmt.Sprintf("%d", count)
	}
	
	text := fmt.Sprintf("%d", count)
	switch {
	case count == 0:
		return c.colorize(text, "success")
	case count < 5:
		return c.colorize(text, "warning")
	default:
		return c.colorize(text, "error")
	}
}

// getSeverityIcon returns an appropriate icon for each severity level
func (c *ConsoleReporter) getSeverityIcon(severity models.Severity) string {
	switch severity {
	case models.SeverityLow:
		return "ℹ️"
	case models.SeverityMedium:
		return "⚠️"
	case models.SeverityHigh:
		return "🚨"
	case models.SeverityCritical:
		return "💥"
	default:
		return "❓"
	}
}

// getSeverityColor returns the appropriate color type for each severity level
func (c *ConsoleReporter) getSeverityColor(severity models.Severity) string {
	switch severity {
	case models.SeverityLow:
		return "success"
	case models.SeverityMedium:
		return "warning"
	case models.SeverityHigh, models.SeverityCritical:
		return "error"
	default:
		return "reset"
	}
}