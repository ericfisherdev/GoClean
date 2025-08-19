package reporters

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
	"unicode"

	"github.com/ericfisherdev/goclean/internal/config"
	"github.com/ericfisherdev/goclean/internal/models"
)

// Console formatting constants
const (
	SeparatorLength       = 40
	FileSeparatorLength   = 70
	FinalSeparatorLength  = 50
	DetailSeparatorLength = 80
	TabwriterMinWidth     = 0
	TabwriterTabWidth     = 0
	TabwriterPadding      = 3
	TabwriterPadChar      = ' '
	TabwriterFlags        = 0
	MaxDisplayedFiles     = 10
	MaxFileNameLength     = 30
	FileNameTruncateChars = 27
	LowViolationThreshold = 5
)

// titleCase converts a string to title case, replacing the deprecated strings.Title
func titleCase(s string) string {
	if s == "" {
		return s
	}
	
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			for j := 1; j < len(runes); j++ {
				runes[j] = unicode.ToLower(runes[j])
			}
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

// typeCount represents a violation type with its count
type typeCount struct {
	Type  models.ViolationType
	Count int
}

// ConsoleReporter outputs reports to the console
type ConsoleReporter struct {
	verbose bool
	colors  bool
	output  io.Writer
}

// NewConsoleReporter creates a new console reporter with config
func NewConsoleReporter(cfg *config.ConsoleConfig) *ConsoleReporter {
	var output io.Writer = os.Stdout
	if cfg.Output != nil {
		if w, ok := cfg.Output.(io.Writer); ok {
			output = w
		}
	}
	
	return &ConsoleReporter{
		verbose: cfg.Verbose,
		colors:  cfg.Colored,
		output:  output,
	}
}

// NewConsoleReporterLegacy creates a new console reporter (compatibility wrapper)
func NewConsoleReporterLegacy(verbose, colors bool) *ConsoleReporter {
	cfg := &config.ConsoleConfig{
		Verbose: verbose,
		Colored: colors,
		Output:  os.Stdout,
	}
	return NewConsoleReporter(cfg)
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
	fmt.Fprintln(c.output, c.colorize("=== GoClean Code Analysis Report ===", "header"))
	fmt.Fprintf(c.output, "Generated: %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(c.output, "Report ID: %s\n\n", report.ID)
}

// printSummary prints the scan summary
func (c *ConsoleReporter) printSummary(summary *models.ScanSummary) {
	fmt.Println(c.colorize("📊 SCAN SUMMARY", "section"))
	fmt.Println(strings.Repeat("-", SeparatorLength))
	
	w := tabwriter.NewWriter(os.Stdout, TabwriterMinWidth, TabwriterTabWidth, TabwriterPadding, TabwriterPadChar, TabwriterFlags)
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
	fmt.Println(strings.Repeat("-", SeparatorLength))
	
	// Sort violation types by count (descending)
	
	var sortedTypes []typeCount
	var rustViolations []typeCount
	for vtype, count := range stats.ViolationsByType {
		if models.IsRustSpecificViolation(vtype) {
			rustViolations = append(rustViolations, typeCount{Type: vtype, Count: count})
		} else {
			sortedTypes = append(sortedTypes, typeCount{Type: vtype, Count: count})
		}
	}
	
	sort.Slice(sortedTypes, func(i, j int) bool {
		return sortedTypes[i].Count > sortedTypes[j].Count
	})
	sort.Slice(rustViolations, func(i, j int) bool {
		return rustViolations[i].Count > rustViolations[j].Count
	})
	
	w := tabwriter.NewWriter(os.Stdout, TabwriterMinWidth, TabwriterTabWidth, TabwriterPadding, TabwriterPadChar, TabwriterFlags)
	
	// Print general violations
	for _, tc := range sortedTypes {
		fmt.Fprintf(w, "%s:\t%s\n", 
			tc.Type.GetDisplayName(), 
			c.colorizeViolationCount(tc.Count))
	}
	
	// Print Rust-specific violations with category indicators
	for _, tc := range rustViolations {
		category := models.GetRustViolationCategory(tc.Type)
		fmt.Fprintf(w, "%s 🦀 [%s]:\t%s\n", 
			tc.Type.GetDisplayName(), 
			titleCase(string(category)), 
			c.colorizeViolationCount(tc.Count))
	}
	
	w.Flush()
	fmt.Println()
	
	// Print Rust violation categories summary if there are Rust violations
	if len(rustViolations) > 0 {
		c.printRustCategorySummary(rustViolations)
	}

	fmt.Println(c.colorize("🚨 VIOLATIONS BY SEVERITY", "section"))
	fmt.Println(strings.Repeat("-", SeparatorLength))
	
	w = tabwriter.NewWriter(os.Stdout, TabwriterMinWidth, TabwriterTabWidth, TabwriterPadding, TabwriterPadChar, TabwriterFlags)
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
	fmt.Println(strings.Repeat("-", FileSeparatorLength))
	
	w := tabwriter.NewWriter(os.Stdout, TabwriterMinWidth, TabwriterTabWidth, TabwriterPadding, TabwriterPadChar, TabwriterFlags)
	fmt.Fprintf(w, "FILE\tVIOLATIONS\tLINES\tRATE\n")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", 
		strings.Repeat("-", MaxFileNameLength),
		strings.Repeat("-", MaxDisplayedFiles), 
		strings.Repeat("-", LowViolationThreshold), 
		strings.Repeat("-", config.DefaultCyclomaticComplexity))
	
	displayCount := len(topFiles)
	if displayCount > MaxDisplayedFiles {
		displayCount = MaxDisplayedFiles
	}
	
	for i := 0; i < displayCount; i++ {
		file := topFiles[i]
		rate := float64(file.TotalViolations) / float64(file.Lines) * 100
		
		fileName := file.File
		if len(fileName) > MaxFileNameLength {
			fileName = "..." + fileName[len(fileName)-FileNameTruncateChars:]
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
	fmt.Println(strings.Repeat("=", DetailSeparatorLength))
	
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
		fmt.Println(strings.Repeat("-", DetailSeparatorLength))
		
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
	
	// Add Rust category indicator if it's a Rust-specific violation
	categoryInfo := ""
	if models.IsRustSpecificViolation(v.Type) {
		category := models.GetRustViolationCategory(v.Type)
		categoryInfo = fmt.Sprintf(" [🦀 %s]", titleCase(string(category)))
	}
	
	fmt.Fprintf(c.output, "  %s %s [Line %d] %s%s\n",
		severityIcon,
		c.colorize(v.Severity.String(), severityColor),
		v.Line,
		v.Type.GetDisplayName(),
		c.colorize(categoryInfo, "rust-category"))
	
	fmt.Fprintf(c.output, "    %s\n", v.Message)
	
	if v.Description != "" {
		fmt.Fprintf(c.output, "    %s\n", c.colorize(v.Description, "description"))
	}
	
	if v.Suggestion != "" {
		fmt.Fprintf(c.output, "    💡 %s\n", c.colorize(v.Suggestion, "suggestion"))
	}
	
	if c.verbose && v.CodeSnippet != "" {
		fmt.Fprintf(c.output, "    Code:\n")
		lines := strings.Split(v.CodeSnippet, "\n")
		for i, line := range lines {
			if strings.TrimSpace(line) != "" {
				fmt.Fprintf(c.output, "      %d: %s\n", v.Line+i, line)
			}
		}
	}
	
	fmt.Fprintln(c.output)
}

// printFooter prints the report footer
func (c *ConsoleReporter) printFooter(report *models.Report) {
	fmt.Println(strings.Repeat("=", FinalSeparatorLength))
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
		"header":       "\033[1;36m",  // Cyan bold
		"section":      "\033[1;34m",  // Blue bold
		"success":      "\033[1;32m",  // Green bold
		"warning":      "\033[1;33m",  // Yellow bold
		"error":        "\033[1;31m",  // Red bold
		"filename":     "\033[1;37m",  // White bold
		"description":  "\033[0;90m",  // Dark gray
		"suggestion":   "\033[0;36m",  // Cyan
		"file":         "\033[0;35m",  // Magenta
		"rust-category": "\033[0;33m",  // Yellow for Rust categories
		"reset":        "\033[0m",     // Reset
	}
	
	if color, exists := colors[colorType]; exists {
		return color + text + colors["reset"]
	}
	
	return text
}

// printRustCategorySummary prints a summary of Rust violations by category
func (c *ConsoleReporter) printRustCategorySummary(rustViolations []typeCount) {
	fmt.Fprintln(c.output, c.colorize("🦀 RUST VIOLATIONS BY CATEGORY", "section"))
	fmt.Fprintln(c.output, strings.Repeat("-", SeparatorLength))
	
	// Group by category
	categoryCount := make(map[models.RustViolationCategory]int)
	for _, tv := range rustViolations {
		category := models.GetRustViolationCategory(tv.Type)
		categoryCount[category] += tv.Count
	}
	
	// Sort categories by count
	type categoryInfo struct {
		Category models.RustViolationCategory
		Count    int
	}
	
	var categories []categoryInfo
	for cat, count := range categoryCount {
		categories = append(categories, categoryInfo{Category: cat, Count: count})
	}
	
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Count > categories[j].Count
	})
	
	w := tabwriter.NewWriter(c.output, TabwriterMinWidth, TabwriterTabWidth, TabwriterPadding, TabwriterPadChar, TabwriterFlags)
	for _, cat := range categories {
		fmt.Fprintf(w, "%s:\t%s\n", 
			titleCase(string(cat.Category)), 
			c.colorizeViolationCount(cat.Count))
	}
	w.Flush()
	fmt.Fprintln(c.output)
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
	case count < LowViolationThreshold:
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