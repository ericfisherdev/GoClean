// Package config provides configuration management for GoClean.
// It handles loading, parsing, and validating configuration files,
// as well as providing default values for scanning thresholds and output options.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Default threshold constants
const (
	DefaultFunctionLines        = 25
	DefaultCyclomaticComplexity = 8
	DefaultParameters           = 4
	DefaultNestingDepth         = 3
	DefaultClassLines           = 150
	DefaultRefreshInterval      = 10
)

// Config represents the application configuration
type Config struct {
	Scan        ScanConfig    `yaml:"scan"`
	Thresholds  Thresholds    `yaml:"thresholds"`
	Output      OutputConfig  `yaml:"output"`
	Logging     LoggingConfig `yaml:"logging"`
	Rust        RustConfig    `yaml:"rust"`
}

// ScanConfig contains scanning-related settings
type ScanConfig struct {
	Paths            []string `yaml:"paths"`
	Exclude          []string `yaml:"exclude"`
	FileTypes        []string `yaml:"file_types"`
	
	// New test file handling options - using pointers to detect explicit setting
	SkipTestFiles    *bool    `yaml:"skip_test_files"`    // Default: true
	AggressiveMode   *bool    `yaml:"aggressive_mode"`    // Default: false
	CustomTestPatterns []string `yaml:"custom_test_patterns"`
}

// Thresholds contains clean code thresholds
type Thresholds struct {
	FunctionLines        int `yaml:"function_lines"`
	CyclomaticComplexity int `yaml:"cyclomatic_complexity"`
	Parameters           int `yaml:"parameters"`
	NestingDepth         int `yaml:"nesting_depth"`
	ClassLines           int `yaml:"class_lines"`
}

// OutputConfig contains output-related settings
type OutputConfig struct {
	HTML     HTMLConfig     `yaml:"html"`
	Markdown MarkdownConfig `yaml:"markdown"`
}

// HTMLConfig contains HTML report settings
type HTMLConfig struct {
	Path            string `yaml:"path"`
	AutoRefresh     bool   `yaml:"auto_refresh"`
	RefreshInterval int    `yaml:"refresh_interval"`
	Theme           string `yaml:"theme"`
}

// MarkdownConfig contains markdown report settings
type MarkdownConfig struct {
	Enabled         bool   `yaml:"enabled"`
	Path            string `yaml:"path"`
	IncludeExamples bool   `yaml:"include_examples"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// RustConfig contains Rust-specific analysis settings
type RustConfig struct {
	// Ownership and borrowing analysis
	EnableOwnershipAnalysis bool `yaml:"enable_ownership_analysis"`
	MaxLifetimeParams       int  `yaml:"max_lifetime_params"`
	DetectUnnecessaryClones bool `yaml:"detect_unnecessary_clones"`
	
	// Error handling analysis
	EnableErrorHandlingCheck bool `yaml:"enable_error_handling_check"`
	AllowUnwrap             bool `yaml:"allow_unwrap"`
	AllowExpect             bool `yaml:"allow_expect"`
	EnforceResultPropagation bool `yaml:"enforce_result_propagation"`
	
	// Pattern matching analysis
	EnablePatternMatchCheck  bool `yaml:"enable_pattern_match_check"`
	RequireExhaustiveMatch   bool `yaml:"require_exhaustive_match"`
	MaxNestedMatchDepth      int  `yaml:"max_nested_match_depth"`
	
	// Trait and impl analysis
	MaxTraitBounds          int  `yaml:"max_trait_bounds"`
	MaxImplMethods          int  `yaml:"max_impl_methods"`
	DetectOrphanInstances   bool `yaml:"detect_orphan_instances"`
	
	// Unsafe code analysis
	AllowUnsafe             bool `yaml:"allow_unsafe"`
	RequireUnsafeComments   bool `yaml:"require_unsafe_comments"`
	DetectTransmuteUsage    bool `yaml:"detect_transmute_usage"`
	
	// Performance analysis
	DetectIneffcientString  bool `yaml:"detect_inefficient_string"`
	DetectBoxedPrimitives   bool `yaml:"detect_boxed_primitives"`
	DetectBlockingInAsync   bool `yaml:"detect_blocking_in_async"`
	
	// Macro analysis
	MaxMacroComplexity      int  `yaml:"max_macro_complexity"`
	AllowRecursiveMacros    bool `yaml:"allow_recursive_macros"`
	
	// Module structure
	MaxModuleDepth          int  `yaml:"max_module_depth"`
	MaxFileLines            int  `yaml:"max_file_lines"`
	
	// Naming conventions (Rust-specific)
	EnforceSnakeCase        bool `yaml:"enforce_snake_case"`
	EnforcePascalCase       bool `yaml:"enforce_pascal_case"`
	EnforceScreamingSnake   bool `yaml:"enforce_screaming_snake"`
}

// Load loads configuration from a file
func Load(configPath string) (*Config, error) {
	// If no config file specified, try to find one
	if configPath == "" {
		var err error
		configPath, err = findConfigFile()
		if err != nil {
			// If no config file found, return default config
			return GetDefaultConfig(), nil
		}
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge with defaults for any missing values
	mergeWithDefaults(&config)

	return &config, nil
}

// Save saves configuration to a file
func Save(config *Config, configPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDefaultConfig returns the default configuration
func GetDefaultConfig() *Config {
	return &Config{
		Scan: ScanConfig{
			Paths: []string{
				"./src",
				"./internal",
				"./cmd",
			},
			Exclude: []string{
				"vendor/",
				"node_modules/",
				".git/",
				"testdata/",
				"build/",
				"dist/",
			},
			FileTypes: []string{
				".go",
				".js",
				".ts",
				".py",
				".java",
				".cs",
				".cpp",
				".c",
				".h",
				".hpp",
				".rs",
			},
			SkipTestFiles:  boolPtr(true),  // Skip test files by default
			AggressiveMode: boolPtr(false), // Normal mode by default
			CustomTestPatterns: []string{},
		},
		Thresholds: Thresholds{
			FunctionLines:        DefaultFunctionLines,
			CyclomaticComplexity: DefaultCyclomaticComplexity,
			Parameters:           DefaultParameters,
			NestingDepth:         DefaultNestingDepth,
			ClassLines:           DefaultClassLines,
		},
		Output: OutputConfig{
			HTML: HTMLConfig{
				Path:            "./reports/clean-code-report.html",
				AutoRefresh:     false,
				RefreshInterval: DefaultRefreshInterval,
				Theme:           "auto",
			},
			Markdown: MarkdownConfig{
				Enabled:         true,
				Path:            "./reports/violations.md",
				IncludeExamples: true,
			},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "structured",
		},
		Rust: GetDefaultRustConfig(),
	}
}

// GetDefaultRustConfig returns the default Rust configuration
func GetDefaultRustConfig() RustConfig {
	return RustConfig{
		// Ownership and borrowing
		EnableOwnershipAnalysis: true,
		MaxLifetimeParams:       3,
		DetectUnnecessaryClones: true,
		
		// Error handling
		EnableErrorHandlingCheck: true,
		AllowUnwrap:             false,
		AllowExpect:             false,
		EnforceResultPropagation: true,
		
		// Pattern matching
		EnablePatternMatchCheck:  true,
		RequireExhaustiveMatch:   true,
		MaxNestedMatchDepth:      3,
		
		// Trait and impl
		MaxTraitBounds:          5,
		MaxImplMethods:          20,
		DetectOrphanInstances:   true,
		
		// Unsafe code
		AllowUnsafe:             true,  // Allow but track
		RequireUnsafeComments:   true,
		DetectTransmuteUsage:    true,
		
		// Performance
		DetectIneffcientString:  true,
		DetectBoxedPrimitives:   true,
		DetectBlockingInAsync:   true,
		
		// Macro analysis
		MaxMacroComplexity:      10,
		AllowRecursiveMacros:    false,
		
		// Module structure
		MaxModuleDepth:          5,
		MaxFileLines:            500,
		
		// Naming conventions
		EnforceSnakeCase:        true,
		EnforcePascalCase:       true,
		EnforceScreamingSnake:   true,
	}
}

// findConfigFile looks for config files in standard locations
func findConfigFile() (string, error) {
	configNames := []string{
		"goclean.yaml",
		"goclean.yml",
		".goclean.yaml",
		".goclean.yml",
	}

	// Check current directory first
	for _, name := range configNames {
		if _, err := os.Stat(name); err == nil {
			return name, nil
		}
	}

	// Check configs directory
	for _, name := range configNames {
		configPath := filepath.Join("configs", name)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	return "", fmt.Errorf("no config file found")
}

// mergeWithDefaults fills in missing configuration values with defaults
func mergeWithDefaults(config *Config) {
	defaults := GetDefaultConfig()

	// Merge scan config
	if len(config.Scan.Paths) == 0 {
		config.Scan.Paths = defaults.Scan.Paths
	}
	if len(config.Scan.Exclude) == 0 {
		config.Scan.Exclude = defaults.Scan.Exclude
	}
	if len(config.Scan.FileTypes) == 0 {
		config.Scan.FileTypes = defaults.Scan.FileTypes
	}
	
	// Handle boolean pointers - only set defaults when nil (not explicitly set)
	if config.Scan.SkipTestFiles == nil {
		config.Scan.SkipTestFiles = defaults.Scan.SkipTestFiles
	}
	if config.Scan.AggressiveMode == nil {
		config.Scan.AggressiveMode = defaults.Scan.AggressiveMode
	}

	// Merge thresholds
	if config.Thresholds.FunctionLines == 0 {
		config.Thresholds.FunctionLines = defaults.Thresholds.FunctionLines
	}
	if config.Thresholds.CyclomaticComplexity == 0 {
		config.Thresholds.CyclomaticComplexity = defaults.Thresholds.CyclomaticComplexity
	}
	if config.Thresholds.Parameters == 0 {
		config.Thresholds.Parameters = defaults.Thresholds.Parameters
	}
	if config.Thresholds.NestingDepth == 0 {
		config.Thresholds.NestingDepth = defaults.Thresholds.NestingDepth
	}
	if config.Thresholds.ClassLines == 0 {
		config.Thresholds.ClassLines = defaults.Thresholds.ClassLines
	}

	// Merge output config
	if config.Output.HTML.Path == "" {
		config.Output.HTML.Path = defaults.Output.HTML.Path
	}
	if config.Output.HTML.RefreshInterval == 0 {
		config.Output.HTML.RefreshInterval = defaults.Output.HTML.RefreshInterval
	}
	if config.Output.HTML.Theme == "" {
		config.Output.HTML.Theme = defaults.Output.HTML.Theme
	}
	if config.Output.Markdown.Path == "" {
		config.Output.Markdown.Path = defaults.Output.Markdown.Path
	}

	// Merge logging config
	if config.Logging.Level == "" {
		config.Logging.Level = defaults.Logging.Level
	}
	if config.Logging.Format == "" {
		config.Logging.Format = defaults.Logging.Format
	}
	
	// Merge Rust config - use defaults if not explicitly set
	mergeRustConfig(&config.Rust, &defaults.Rust)
}

// mergeRustConfig merges Rust configuration with defaults
func mergeRustConfig(config *RustConfig, defaults *RustConfig) {
	// Since booleans default to false, we need a different approach
	// We'll only merge if the entire Rust config section appears to be unset
	
	// Check if config is completely empty (all defaults)
	emptyConfig := RustConfig{}
	if *config == emptyConfig {
		*config = *defaults
		return
	}
	
	// For integer fields, use defaults if zero
	if config.MaxLifetimeParams == 0 {
		config.MaxLifetimeParams = defaults.MaxLifetimeParams
	}
	if config.MaxNestedMatchDepth == 0 {
		config.MaxNestedMatchDepth = defaults.MaxNestedMatchDepth
	}
	if config.MaxTraitBounds == 0 {
		config.MaxTraitBounds = defaults.MaxTraitBounds
	}
	if config.MaxImplMethods == 0 {
		config.MaxImplMethods = defaults.MaxImplMethods
	}
	if config.MaxMacroComplexity == 0 {
		config.MaxMacroComplexity = defaults.MaxMacroComplexity
	}
	if config.MaxModuleDepth == 0 {
		config.MaxModuleDepth = defaults.MaxModuleDepth
	}
	if config.MaxFileLines == 0 {
		config.MaxFileLines = defaults.MaxFileLines
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Check that at least one scan path is specified
	if len(c.Scan.Paths) == 0 {
		return fmt.Errorf("at least one scan path must be specified")
	}

	// Check that paths exist
	for _, path := range c.Scan.Paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("scan path does not exist: %s", path)
		}
	}

	// Validate thresholds are positive
	if c.Thresholds.FunctionLines <= 0 {
		return fmt.Errorf("function_lines threshold must be positive")
	}
	if c.Thresholds.CyclomaticComplexity <= 0 {
		return fmt.Errorf("cyclomatic_complexity threshold must be positive")
	}
	if c.Thresholds.Parameters <= 0 {
		return fmt.Errorf("parameters threshold must be positive")
	}
	if c.Thresholds.NestingDepth <= 0 {
		return fmt.Errorf("nesting_depth threshold must be positive")
	}
	if c.Thresholds.ClassLines <= 0 {
		return fmt.Errorf("class_lines threshold must be positive")
	}

	// Validate output paths
	if c.Output.HTML.Path != "" {
		if err := validateOutputPath(c.Output.HTML.Path); err != nil {
			return fmt.Errorf("invalid HTML output path: %w", err)
		}
	}
	if c.Output.Markdown.Enabled && c.Output.Markdown.Path != "" {
		if err := validateOutputPath(c.Output.Markdown.Path); err != nil {
			return fmt.Errorf("invalid Markdown output path: %w", err)
		}
	}

	// Validate logging level
	validLevels := []string{"debug", "info", "warn", "error"}
	levelValid := false
	for _, level := range validLevels {
		if c.Logging.Level == level {
			levelValid = true
			break
		}
	}
	if !levelValid {
		return fmt.Errorf("invalid logging level: %s (must be one of: debug, info, warn, error)", c.Logging.Level)
	}

	return nil
}

// validateOutputPath checks if the output path is valid
func validateOutputPath(path string) error {
	// Check if directory exists or can be created
	dir := filepath.Dir(path)
	if dir != "." {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			// Try to create directory
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("cannot create output directory: %w", err)
			}
		}
	}

	// Check if file is writable (try to create/touch it)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("cannot write to output file: %w", err)
	}
	file.Close()

	return nil
}

// boolPtr returns a pointer to the given boolean value
func boolPtr(b bool) *bool {
	return &b
}

// GetSkipTestFiles safely returns the SkipTestFiles value with default fallback
func (s *ScanConfig) GetSkipTestFiles() bool {
	if s.SkipTestFiles == nil {
		return true // default value
	}
	return *s.SkipTestFiles
}

// GetAggressiveMode safely returns the AggressiveMode value with default fallback
func (s *ScanConfig) GetAggressiveMode() bool {
	if s.AggressiveMode == nil {
		return false // default value
	}
	return *s.AggressiveMode
}

// GetConfigPaths returns standard configuration file paths
func GetConfigPaths() []string {
	return []string{
		"goclean.yaml",
		"goclean.yml",
		".goclean.yaml",
		".goclean.yml",
		"configs/goclean.yaml",
		"configs/goclean.yml",
	}
}