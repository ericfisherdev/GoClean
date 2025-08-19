// Package main provides the GoClean command-line interface for clean code analysis.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/ericfisherdev/goclean/internal/config"
	"github.com/ericfisherdev/goclean/internal/models"
	"github.com/ericfisherdev/goclean/internal/scanner"
	"github.com/ericfisherdev/goclean/internal/reporters"
)

var (
	// Version can be overridden at build time with -ldflags "-X main.Version=x.y.z"
	Version = getVersionFromFile()
	
	// Global flags
	cfgFile     string
	verbose     bool
	outputPath  string
	format      string
	
	// Scan flags
	paths       []string
	exclude     []string
	fileTypes   []string
	languages   []string
	thresholds  map[string]int
	
	// Test file handling flags
	aggressive       bool
	includeTests     bool
	customTestPatterns []string
	
	// Console output flags
	consoleViolations bool
	
	// Rust-specific flags
	rustOptimizations bool
	rustCacheSize     int
	rustCacheTTL      int // in minutes

)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goclean",
	Short: "GoClean - Clean Code Analysis Tool",
	Long: `GoClean is a powerful CLI tool for analyzing codebases to identify clean code violations.
It provides real-time HTML reporting and optional markdown output for AI analysis.

Features:
- Multi-language support (Go, Rust, JavaScript, TypeScript, Python, Java, C#)
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
  goclean scan . --languages go,rust --verbose
  goclean scan . --languages rust --config ./rust-config.yaml
  goclean scan . --console-violations  # AI-friendly output`,
	Args: cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if !consoleViolations {
			fmt.Printf("GoClean v%s - Clean Code Analysis Tool\n", rootCmd.Version)
			fmt.Println("Starting code analysis...")
		}
		
		// Load configuration
		cfg, err := config.Load(cfgFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
			os.Exit(1)
		}
		
		// Override configuration with command-line flags if provided
		if format == "json" {
			// Enable JSON export when format is explicitly set to json
			cfg.Export.JSON.Enabled = true
			if outputPath != "" {
				cfg.Export.JSON.Path = outputPath
			} else if cfg.Export.JSON.Path == "" {
				// Use default path if not configured
				cfg.Export.JSON.Path = "./reports/violations.json"
			}
		} else if outputPath != "" {
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
		
		// Handle language filtering - map languages to file extensions
		if len(languages) > 0 {
			languageExtensions := getLanguageExtensions(languages)
			if len(languageExtensions) > 0 {
				fileTypesList = languageExtensions
			}
		}
		
		// Display configuration
		if verbose && !consoleViolations {
			fmt.Printf("Scan paths: %v\n", scanPaths)
			if len(excludePatterns) > 0 {
				fmt.Printf("Exclude patterns: %v\n", excludePatterns)
			}
			if len(languages) > 0 {
				fmt.Printf("Languages: %v\n", languages)
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
		
		// Configure concurrent file processing if specified
		if cfg.Scan.ConcurrentFiles > 0 {
			engine.SetMaxWorkers(cfg.Scan.ConcurrentFiles)
			if verbose && !consoleViolations {
				fmt.Printf("Concurrent file processing set to: %d workers\n", cfg.Scan.ConcurrentFiles)
			}
		}
		
		// Configure max file size if specified
		if cfg.Scan.MaxFileSize != "" {
			if err := engine.SetMaxFileSize(cfg.Scan.MaxFileSize); err != nil {
				fmt.Fprintf(os.Stderr, "Invalid max file size configuration: %v\n", err)
				os.Exit(1)
			}
			if verbose && !consoleViolations {
				fmt.Printf("Max file size limit set to: %s\n", cfg.Scan.MaxFileSize)
			}
		}
		
		// Configure Rust optimizations if Rust language is being scanned
		if containsRust(languages, fileTypesList) || rustOptimizations {
			engine.EnableRustOptimization(rustOptimizations || containsRust(languages, fileTypesList))
			
			// Configure cache based on flags or auto-estimation
			cacheSize := rustCacheSize
			if cacheSize == 0 {
				cacheSize = estimateCacheSize(scanPaths)
			}
			
			cacheTTL := time.Duration(rustCacheTTL) * time.Minute
			if rustCacheTTL == 0 {
				cacheTTL = 30 * time.Minute
			}
			
			engine.SetRustCacheConfig(cacheSize, cacheTTL)
			
			if verbose && !consoleViolations {
				fmt.Printf("Rust optimizations enabled: cache size=%d, TTL=%v\n", cacheSize, cacheTTL)
			}
		}
		
		// Set progress callback for real-time updates
		progressCallback := func(message string) {
			if verbose && !consoleViolations {
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
			// Exit immediately with appropriate code for console violations mode
			if summary.TotalViolations > 0 {
				os.Exit(1)
			}
			return
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
			
			if jsonPath := reporterManager.GetJSONOutputPath(); jsonPath != "" {
				fmt.Printf("📄 JSON report generated: %s\n", jsonPath)
			}
		}
		
		// Display Rust performance metrics if verbose and optimizations were enabled
		if verbose && !consoleViolations && (containsRust(languages, fileTypesList) || rustOptimizations) {
			if metrics := engine.GetRustPerformanceMetrics(); metrics != nil {
				fmt.Printf("\n🦀 Rust Performance Metrics:\n")
				fmt.Printf("   Cache hits: %d, misses: %d (%.1f%% hit rate)\n", 
					metrics["cache_hits"], metrics["cache_misses"], metrics["cache_hit_rate"])
				fmt.Printf("   Cache size: %d entries\n", metrics["cache_size"])
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

// getVersionFromFile reads the version from the VERSION file, falling back to a default if not found
func getVersionFromFile() string {
	const defaultVersion = "dev"
	
	// Try to find VERSION file in the current directory or parent directories
	versionPaths := []string{
		"VERSION",
		"../VERSION",
		"../../VERSION",
	}
	
	for _, versionPath := range versionPaths {
		if file, err := os.Open(versionPath); err == nil {
			defer file.Close()
			
			content, err := io.ReadAll(file)
			if err == nil {
				version := strings.TrimSpace(string(content))
				if version != "" {
					return version
				}
			}
		}
	}
	
	// If VERSION file not found, try to find it relative to the executable
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		versionFile := filepath.Join(execDir, "VERSION")
		if file, err := os.Open(versionFile); err == nil {
			defer file.Close()
			
			content, err := io.ReadAll(file)
			if err == nil {
				version := strings.TrimSpace(string(content))
				if version != "" {
					return version
				}
			}
		}
	}
	
	return defaultVersion
}

// generateConsoleViolationsOutput outputs violations in a structured format for AI agents
func generateConsoleViolationsOutput(summary *models.ScanSummary, results []*models.ScanResult) {
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
	
	// Create tab writer for structured output
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	
	// Write header
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", "FILE", "LINE", "TYPE", "LEVEL", "MESSAGE")
	
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
	
	// Output each violation in tab-delimited format
	for _, violation := range allViolations {
		// Clean up message (remove newlines and tabs but preserve content)
		message := strings.ReplaceAll(violation.Message, "\n", " ")
		message = strings.ReplaceAll(message, "\t", " ")
		
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\n",
			violation.File,
			violation.Line,
			string(violation.Type),
			violation.Severity.String(),
			message)
	}
	
	// Flush the tabwriter
	w.Flush()
	
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

// getLanguageExtensions maps language names to their file extensions
func getLanguageExtensions(languages []string) []string {
	languageMap := map[string][]string{
		"go":         {".go"},
		"rust":       {".rs"},
		"javascript": {".js", ".jsx"},
		"typescript": {".ts", ".tsx"},
		"python":     {".py"},
		"java":       {".java"},
		"c#":         {".cs"},
		"csharp":     {".cs"},
		"c":          {".c", ".h"},
		"cpp":        {".cpp", ".cc", ".cxx", ".hpp"},
		"c++":        {".cpp", ".cc", ".cxx", ".hpp"},
		"php":        {".php"},
		"ruby":       {".rb"},
		"swift":      {".swift"},
		"kotlin":     {".kt", ".kts"},
		"scala":      {".scala"},
	}
	
	var extensions []string
	for _, lang := range languages {
		if exts, exists := languageMap[strings.ToLower(lang)]; exists {
			extensions = append(extensions, exts...)
		}
	}
	
	return extensions
}

// containsRust checks if Rust is being scanned based on languages or file types
func containsRust(languages, fileTypes []string) bool {
	// Check explicit language specification
	for _, lang := range languages {
		if strings.ToLower(lang) == "rust" {
			return true
		}
	}
	
	// Check file types for Rust extensions
	for _, fileType := range fileTypes {
		if fileType == ".rs" {
			return true
		}
	}
	
	return false
}

// estimateCacheSize estimates appropriate cache size based on project scope
func estimateCacheSize(scanPaths []string) int {
	// Start with a base cache size
	baseSize := 500
	
	// Estimate based on number of scan paths
	pathMultiplier := len(scanPaths)
	if pathMultiplier == 0 {
		pathMultiplier = 1
	}
	
	// Simple heuristic: more paths likely means larger project
	estimatedSize := baseSize * pathMultiplier
	
	// Cap the cache size to reasonable limits
	if estimatedSize > 2000 {
		return 2000
	}
	if estimatedSize < 100 {
		return 100
	}
	
	return estimatedSize
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is goclean.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Scan command flags
	scanCmd.Flags().StringSliceVarP(&exclude, "exclude", "e", []string{}, "exclude patterns (comma-separated)")
	scanCmd.Flags().StringSliceVarP(&fileTypes, "types", "t", []string{}, "file types to scan (comma-separated, e.g., .go,.js,.py)")
	scanCmd.Flags().StringSliceVarP(&languages, "languages", "l", []string{}, "languages to scan (comma-separated, e.g., go,rust,javascript)")
	scanCmd.Flags().StringVarP(&format, "format", "f", "html", "output format (html, markdown, json)")
	scanCmd.Flags().StringVarP(&outputPath, "output", "o", "", "output file path")
	
	// Test file handling flags
	scanCmd.Flags().BoolVar(&aggressive, "aggressive", false, "Enable aggressive mode (scan test files and apply stricter rules)")
	scanCmd.Flags().BoolVar(&includeTests, "include-tests", false, "Include test files in analysis (alias for --aggressive)")
	scanCmd.Flags().StringSliceVar(&customTestPatterns, "test-patterns", []string{}, "Additional test file patterns to recognize")
	
	// Console output flags
	scanCmd.Flags().BoolVar(&consoleViolations, "console-violations", false, "Output violations directly to console in structured format for AI agents")
	
	// Rust-specific flags
	scanCmd.Flags().BoolVar(&rustOptimizations, "rust-opt", false, "Enable Rust performance optimizations (auto-enabled when scanning Rust)")
	scanCmd.Flags().IntVar(&rustCacheSize, "rust-cache-size", 0, "Rust AST cache size (0 = auto-estimate)")
	scanCmd.Flags().IntVar(&rustCacheTTL, "rust-cache-ttl", 0, "Rust cache TTL in minutes (0 = 30 minutes)")

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