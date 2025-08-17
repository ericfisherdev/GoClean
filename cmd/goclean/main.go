// Package main provides the GoClean command-line interface for clean code analysis.
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ericfisherdev/goclean/internal/config"
	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/scanner"
	"github.com/ericfisherdev/goclean/internal/reporters"
)

var (
	// Version can be overridden at build time with -ldflags "-X main.Version=x.y.z"
	Version = "2025.08.16.20"
	
	// Global flags
	cfgFile     string
	verbose     bool
	outputPath  string
	format      string
	
	// Scan flags
	paths       []string
	exclude     []string
	fileTypes   []string
	thresholds  map[string]int
	
	// Test file handling flags
	aggressive       bool
	includeTests     bool
	customTestPatterns []string
	
	// Console output flags
	consoleViolations bool

)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goclean",
	Short: "GoClean - Clean Code Analysis Tool",
	Long: `GoClean is a powerful CLI tool for analyzing codebases to identify clean code violations.
It provides real-time HTML reporting and optional markdown output for AI analysis.

Features:
- Multi-language support (Go, JavaScript, TypeScript, Python, Java, C#)
- Real-time HTML dashboard with auto-refresh
- Configurable violation thresholds
- Markdown output for AI analysis
- Comprehensive clean code violation detection`,
	Version: Version,
}

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan [paths...]",
	Short: "Scan codebases for clean code violations",
	Long: `Scan one or more directories or files for clean code violations.
The scan command analyzes your codebase and generates reports based on clean code principles.

Examples:
  goclean scan ./src
  goclean scan ./src ./internal --exclude vendor/,node_modules/
  goclean scan . --format html --output ./reports/report.html
  goclean scan . --verbose --config ./custom-config.yaml
  goclean scan . --console-violations  # AI-friendly output`,
	Args: cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("GoClean v%s - Clean Code Analysis Tool\n", rootCmd.Version)
		fmt.Println("Starting code analysis...")
		
		// Load configuration
		cfg, err := config.Load(cfgFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
			os.Exit(1)
		}
		
		// Override configuration with command-line flags if provided
		if outputPath != "" {
			if format == "html" || format == "" {
				cfg.Output.HTML.Path = outputPath
			} else if format == "markdown" {
				cfg.Output.Markdown.Path = outputPath
				cfg.Output.Markdown.Enabled = true
			}
		}
		
		// Handle test file configuration
		if aggressive || includeTests {
			cfg.Scan.AggressiveMode = &[]bool{true}[0]
			cfg.Scan.SkipTestFiles = &[]bool{false}[0]
		}
		
		// Add custom test patterns if provided
		if len(customTestPatterns) > 0 {
			cfg.Scan.CustomTestPatterns = customTestPatterns
		}
    
		// Merge command-line flags with configuration
		scanPaths := args
		if len(scanPaths) == 0 {
			scanPaths = cfg.Scan.Paths
		}
		
		excludePatterns := exclude
		if len(excludePatterns) == 0 {
			excludePatterns = cfg.Scan.Exclude
		}
		
		fileTypesList := fileTypes
		if len(fileTypesList) == 0 {
			fileTypesList = cfg.Scan.FileTypes
		}
		
		// Display configuration
		if verbose {
			fmt.Printf("Scan paths: %v\n", scanPaths)
			if len(excludePatterns) > 0 {
				fmt.Printf("Exclude patterns: %v\n", excludePatterns)
			}
			if len(fileTypesList) > 0 {
				fmt.Printf("File types: %v\n", fileTypesList)
			}
			if cfgFile != "" {
				fmt.Printf("Configuration file: %s\n", cfgFile)
			}
			fmt.Printf("Output format: %s\n", format)
			if outputPath != "" {
				fmt.Printf("Output path: %s\n", outputPath)
			}
			fmt.Printf("Function lines threshold: %d\n", cfg.Thresholds.FunctionLines)
			fmt.Printf("Cyclomatic complexity threshold: %d\n", cfg.Thresholds.CyclomaticComplexity)
		}
		
		// Create reporter manager
		reporterManager, err := reporters.NewManager(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create reporter manager: %v\n", err)
			os.Exit(1)
		}
		
		// Create and configure scanner engine with test file configuration
		engine := scanner.NewEngineWithConfig(scanPaths, excludePatterns, fileTypesList, verbose,
			cfg.Scan.GetSkipTestFiles(), cfg.Scan.GetAggressiveMode(), cfg.Scan.CustomTestPatterns)
		
		// Set progress callback for real-time updates
		progressCallback := func(message string) {
			if verbose {
				fmt.Printf("Progress: %s\n", message)
			}
		}
		engine.SetProgressCallback(progressCallback)
		
		// Perform scan with progress updates
		summary, results, err := engine.Scan()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Scan failed: %v\n", err)
			os.Exit(1)
		}
		
		// Generate console report
		if consoleViolations {
			// Generate structured violations output for AI agents
			generateConsoleViolationsOutput(summary, results)
		} else {
			fmt.Println()
			err = reporterManager.GenerateConsoleReport(summary, results, verbose, true)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to generate console report: %v\n", err)
			}
		}
		
		// Generate file reports (HTML and/or Markdown)
		configuredReporters := reporterManager.GetConfiguredReporters()
		if len(configuredReporters) > 0 {
			fmt.Printf("\nGenerating reports (%s)...\n", configuredReporters)
			
			err = reporterManager.GenerateReports(summary, results)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to generate reports: %v\n", err)
				os.Exit(1)
			}
			
			// Display output paths
			if htmlPath := reporterManager.GetHTMLOutputPath(); htmlPath != "" {
				fmt.Printf("📊 HTML report generated: %s\n", htmlPath)
				if cfg.Output.HTML.AutoRefresh {
					fmt.Printf("   Auto-refresh enabled (every %d seconds)\n", cfg.Output.HTML.RefreshInterval)
				}
			}
			
			if markdownPath := reporterManager.GetMarkdownOutputPath(); markdownPath != "" {
				fmt.Printf("📝 Markdown report generated: %s\n", markdownPath)
			}
		}
		
		// Exit with appropriate code
		if summary.TotalViolations > 0 {
			fmt.Printf("\n⚠️  Found %d violations. Review the reports for details.\n", summary.TotalViolations)
			os.Exit(1)
		} else {
			fmt.Println("\n🎉 No violations found! Your code follows clean code principles.")
		}
	},
}

// configCmd represents the config command for managing configuration
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage GoClean configuration",
	Long: `Manage GoClean configuration settings.
Use subcommands to view, set, or reset configuration values.`,
}

// configInitCmd initializes a default configuration file
var configInitCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a default configuration file",
	Long: `Initialize a default GoClean configuration file.
If no path is specified, creates goclean.yaml in the current directory.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configPath := "goclean.yaml"
		if len(args) > 0 {
			configPath = args[0]
		}
		
		// Check if config file already exists
		if _, err := os.Stat(configPath); err == nil {
			fmt.Printf("Configuration file already exists: %s\n", configPath)
			fmt.Println("Use --force to overwrite or specify a different path.")
			return
		}
		
		fmt.Printf("Initializing configuration file: %s\n", configPath)
		
		// Create default configuration
		cfg := config.GetDefaultConfig()
		
		// Save configuration to file
		if err := config.Save(cfg, configPath); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create configuration file: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("✓ Configuration file created successfully: %s\n", configPath)
		fmt.Println("You can now customize the configuration to match your project's needs.")
		fmt.Println("Use 'goclean scan --config " + configPath + "' to use this configuration.")
	},
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of GoClean",
	Long:  `Print the version number and build information for GoClean.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Note: Version info will be imported from the parent package when available
		fmt.Printf("GoClean v%s\n", rootCmd.Version)
		fmt.Println("A clean code analysis tool for modern development teams")
	},
}

// generateConsoleViolationsOutput outputs violations in a structured format for AI agents
func generateConsoleViolationsOutput(summary *models.ScanSummary, results []*models.ScanResult) {
	const maxDescriptionLength = 100
	
	// Print basic scan summary first
	fmt.Printf("=== GoClean Scan Results ===\n")
	fmt.Printf("Total Files: %d\n", summary.TotalFiles)
	fmt.Printf("Scanned Files: %d\n", summary.ScannedFiles)
	fmt.Printf("Total Violations: %d\n", summary.TotalViolations)
	fmt.Printf("Scan Duration: %v\n\n", summary.Duration)
	
	if summary.TotalViolations == 0 {
		fmt.Println("✅ No violations found!")
		return
	}
	
	fmt.Println("=== Violations (Structured Format) ===")
	fmt.Printf("%-60s %-8s %-15s %-6s %s\n", "FILE", "LINE", "TYPE", "LEVEL", "MESSAGE")
	fmt.Println(strings.Repeat("-", 140))
	
	// Collect all violations from all files
	var allViolations []*models.Violation
	for _, result := range results {
		for _, violation := range result.Violations {
			allViolations = append(allViolations, violation)
		}
	}
	
	// Sort violations by file and then by line number
	sort.Slice(allViolations, func(i, j int) bool {
		if allViolations[i].File != allViolations[j].File {
			return allViolations[i].File < allViolations[j].File
		}
		return allViolations[i].Line < allViolations[j].Line
	})
	
	// Output each violation in a structured format
	for _, violation := range allViolations {
		// Truncate file path to make it more readable
		displayFile := violation.File
		const maxFilePathLength = 55
		if len(displayFile) > maxFilePathLength {
			displayFile = "..." + displayFile[len(displayFile)-(maxFilePathLength-3):]
		}
		
		// Truncate message if too long
		message := violation.Message
		if len(message) > maxDescriptionLength {
			message = message[:maxDescriptionLength-3] + "..."
		}
		
		// Clean up message (remove newlines)
		message = strings.ReplaceAll(message, "\n", " ")
		message = strings.ReplaceAll(message, "\t", " ")
		
		fmt.Printf("%-60s %-8d %-15s %-6s %s\n",
			displayFile,
			violation.Line,
			string(violation.Type),
			violation.Severity.String(),
			message)
	}
	
	fmt.Printf("\n=== Summary by Type ===\n")
	// Count violations by type
	violationCounts := make(map[models.ViolationType]int)
	for _, violation := range allViolations {
		violationCounts[violation.Type]++
	}
	
	// Sort by count descending
	type typeCount struct {
		Type  models.ViolationType
		Count int
	}
	var sortedTypes []typeCount
	for vtype, count := range violationCounts {
		sortedTypes = append(sortedTypes, typeCount{Type: vtype, Count: count})
	}
	sort.Slice(sortedTypes, func(i, j int) bool {
		return sortedTypes[i].Count > sortedTypes[j].Count
	})
	
	for _, tc := range sortedTypes {
		fmt.Printf("%-30s: %d\n", tc.Type.GetDisplayName(), tc.Count)
	}
	
	fmt.Printf("\n=== Summary by Severity ===\n")
	// Count violations by severity
	severityCounts := make(map[models.Severity]int)
	for _, violation := range allViolations {
		severityCounts[violation.Severity]++
	}
	
	// Output in order of severity
	severities := []models.Severity{models.SeverityCritical, models.SeverityHigh, models.SeverityMedium, models.SeverityLow, models.SeverityInfo}
	for _, severity := range severities {
		if count, exists := severityCounts[severity]; exists && count > 0 {
			fmt.Printf("%-10s: %d\n", severity.String(), count)
		}
	}
	
	fmt.Printf("\n⚠️  Total: %d violations found\n", summary.TotalViolations)
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is goclean.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Scan command flags
	scanCmd.Flags().StringSliceVarP(&exclude, "exclude", "e", []string{}, "exclude patterns (comma-separated)")
	scanCmd.Flags().StringSliceVarP(&fileTypes, "types", "t", []string{}, "file types to scan (comma-separated, e.g., .go,.js,.py)")
	scanCmd.Flags().StringVarP(&format, "format", "f", "html", "output format (html, markdown, json)")
	scanCmd.Flags().StringVarP(&outputPath, "output", "o", "", "output file path")
	
	// Test file handling flags
	scanCmd.Flags().BoolVar(&aggressive, "aggressive", false, "Enable aggressive mode (scan test files and apply stricter rules)")
	scanCmd.Flags().BoolVar(&includeTests, "include-tests", false, "Include test files in analysis (alias for --aggressive)")
	scanCmd.Flags().StringSliceVar(&customTestPatterns, "test-patterns", []string{}, "Additional test file patterns to recognize")
	
	// Console output flags
	scanCmd.Flags().BoolVar(&consoleViolations, "console-violations", false, "Output violations directly to console in structured format for AI agents")

	// Config subcommands
	configCmd.AddCommand(configInitCmd)
	
	// Add commands to root
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}