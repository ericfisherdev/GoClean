// Package main provides the GoClean command-line interface for clean code analysis.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ericfisherdev/goclean/internal/config"
	"github.com/ericfisherdev/goclean/internal/scanner"
	"github.com/ericfisherdev/goclean/internal/reporters"
)

var (
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
	Version: "2025.08.16.6",
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
  goclean scan . --verbose --config ./custom-config.yaml`,
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
		fmt.Println()
		err = reporterManager.GenerateConsoleReport(summary, results, verbose, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate console report: %v\n", err)
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