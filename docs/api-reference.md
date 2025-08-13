# API Reference

> **Disclaimer**: The programmatic API for GoClean is not yet stable and is subject to change. The packages documented here are in the `internal` directory, which is not intended for public use according to Go conventions. This documentation is provided for advanced users and contributors. A stable public API in a `pkg` directory is planned for a future release.

This document provides API documentation for GoClean's internal packages.

## Table of Contents

1. [Overview](#overview)
2. [Core Packages](#core-packages)
3. [Scanner Engine API (`internal/scanner`)](#scanner-engine-api-internalscanner)
4. [Configuration API (`internal/config`)](#configuration-api-internalconfig)
5. [Data Models (`internal/models`)](#data-models-internalmodels)
6. [Example Usage](#example-usage)

## Overview

Programmatic access to GoClean's functionality is available through its internal packages. The main entry point for scanning is the `scanner.Engine`.

## Core Packages

-   `github.com/ericfisherdev/goclean/internal/scanner`: Contains the core scanning engine, file walker, and parser.
-   `github.com/ericfisherdev/goclean/internal/config`: Handles loading and managing configuration.
-   `github.com/ericfisherdev/goclean/internal/models`: Defines the data structures for violations, reports, and file information.
-   `github.com/ericfisherdev/goclean/internal/violations`: Contains the logic for detecting specific code violations.
-   `github.com/ericfisherdev/goclean/internal/reporters`: Includes the report generators for HTML, Markdown, and console output.

## Scanner Engine API (`internal/scanner`)

The `Engine` is the main orchestrator for scanning files.

### `scanner.Engine`

## Configuration API

### config.Config

Main configuration structure that controls all aspects of GoClean's behavior.

```go
type Config struct {
    Scan       ScanConfig       `yaml:"scan"`
    Thresholds ThresholdConfig  `yaml:"thresholds"`
    Naming     NamingConfig     `yaml:"naming"`
    Output     OutputConfig     `yaml:"output"`
    Logging    LoggingConfig    `yaml:"logging"`
}
```

### Loading Configuration

#### config.Load

```go
func Load(path string) (*Config, error)
```

Loads configuration from a YAML file.

**Parameters:**
- `path`: Path to configuration file

**Returns:**
- `*Config`: Loaded configuration
- `error`: Error if loading fails

**Example:**
```go
cfg, err := config.Load("./goclean.yaml")
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}
```

#### config.LoadFromBytes

```go
func LoadFromBytes(data []byte) (*Config, error)
```

Loads configuration from byte slice.

**Parameters:**
- `data`: YAML configuration data

**Returns:**
- `*Config`: Loaded configuration
- `error`: Error if parsing fails

#### config.Default

```go
func Default() *Config
```

Returns default configuration with sensible values.

**Returns:**
- `*Config`: Default configuration

### Configuration Validation

#### config.Validate

```go
func (c *Config) Validate() error
```

Validates configuration values and returns errors for invalid settings.

**Returns:**
- `error`: Validation error or nil if valid

**Example:**
```go
cfg := config.Default()
cfg.Thresholds.FunctionLines = -1 // Invalid value

if err := cfg.Validate(); err != nil {
    fmt.Printf("Configuration error: %v\n", err)
}
```

### Environment Variable Override

#### config.LoadWithEnv

```go
func LoadWithEnv(path string) (*Config, error)
```

Loads configuration and applies environment variable overrides.

Environment variables use the prefix `GOCLEAN_` followed by the configuration path in uppercase with underscores.

**Example:**
```bash
export GOCLEAN_THRESHOLDS_FUNCTION_LINES=30
export GOCLEAN_OUTPUT_HTML_THEME=dark
```

## Scanner API

### goclean.Analyzer

Main analysis engine that orchestrates the scanning process.

```go
type Analyzer struct {
    config    *config.Config
    detectors []violations.Detector
    reporters []reporters.Reporter
}
```

### Creating an Analyzer

#### goclean.New

```go
func New(cfg *config.Config) *Analyzer
```

Creates a new analyzer instance with the provided configuration.

**Parameters:**
- `cfg`: Configuration for the analyzer

**Returns:**
- `*Analyzer`: New analyzer instance

#### goclean.NewWithDetectors

```go
func NewWithDetectors(cfg *config.Config, detectors []violations.Detector) *Analyzer
```

Creates analyzer with custom violation detectors.

**Parameters:**
- `cfg`: Configuration for the analyzer
- `detectors`: Custom violation detectors

**Returns:**
- `*Analyzer`: New analyzer instance

### Analysis Methods

#### Analyzer.Analyze

```go
func (a *Analyzer) Analyze(ctx context.Context, paths []string) (*models.AnalysisResult, error)
```

Performs complete analysis of specified paths.

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `paths`: Directories or files to analyze

**Returns:**
- `*models.AnalysisResult`: Analysis results
- `error`: Error if analysis fails

**Example:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

result, err := analyzer.Analyze(ctx, []string{"./src", "./internal"})
if err != nil {
    return fmt.Errorf("analysis failed: %w", err)
}
```

#### Analyzer.AnalyzeFile

```go
func (a *Analyzer) AnalyzeFile(file string) (*models.FileResult, error)
```

Analyzes a single file and returns violations found.

**Parameters:**
- `file`: Path to file to analyze

**Returns:**
- `*models.FileResult`: File analysis results
- `error`: Error if analysis fails

#### Analyzer.AnalyzeContent

```go
func (a *Analyzer) AnalyzeContent(filename, content string) (*models.FileResult, error)
```

Analyzes content directly without reading from file system.

**Parameters:**
- `filename`: Logical filename for language detection
- `content`: Source code content to analyze

**Returns:**
- `*models.FileResult`: File analysis results
- `error`: Error if analysis fails

**Example:**
```go
content := `
package main

func veryLongFunctionName() {
    // 50 lines of code...
}
`

result, err := analyzer.AnalyzeContent("example.go", content)
if err != nil {
    return fmt.Errorf("content analysis failed: %w", err)
}
```

### Progress Tracking

#### Analyzer.AnalyzeWithProgress

```go
func (a *Analyzer) AnalyzeWithProgress(
    ctx context.Context, 
    paths []string, 
    progress chan<- models.ProgressUpdate,
) (*models.AnalysisResult, error)
```

Performs analysis with progress updates sent to the provided channel.

**Parameters:**
- `ctx`: Context for cancellation
- `paths`: Paths to analyze
- `progress`: Channel for progress updates

**Returns:**
- `*models.AnalysisResult`: Analysis results
- `error`: Error if analysis fails

**Example:**
```go
progress := make(chan models.ProgressUpdate, 100)
go func() {
    for update := range progress {
        fmt.Printf("Progress: %d/%d files (%.1f%%)\n", 
            update.Current, update.Total, update.Percentage)
    }
}()

result, err := analyzer.AnalyzeWithProgress(ctx, paths, progress)
```

## Violation Detection API

### violations.Detector

Interface for implementing custom violation detectors.

```go
type Detector interface {
    Detect(file *models.FileInfo) ([]models.Violation, error)
    Name() string
    Severity() models.Severity
    Category() models.Category
}
```

### Built-in Detectors

#### violations.FunctionDetector

Detects function-related violations (length, complexity, parameters).

```go
func NewFunctionDetector(config FunctionConfig) *FunctionDetector
```

**Configuration:**
```go
type FunctionConfig struct {
    MaxLines       int
    MaxComplexity  int
    MaxParameters  int
    MaxReturnStmts int
}
```

#### violations.NamingDetector

Detects naming convention violations.

```go
func NewNamingDetector(config NamingConfig) *NamingDetector
```

**Configuration:**
```go
type NamingConfig struct {
    MinLength         int
    AllowSingleLetter bool
    EnforceCamelCase  bool
    CustomPatterns    map[string]string
}
```

#### violations.StructureDetector

Detects code structure issues (large classes, deep nesting).

```go
func NewStructureDetector(config StructureConfig) *StructureDetector
```

### Custom Detector Implementation

```go
package main

import (
    "go/ast"
    "go/token"
    
    "github.com/ericfisherdev/goclean/pkg/models"
    "github.com/ericfisherdev/goclean/pkg/violations"
)

type CustomDetector struct {
    threshold int
}

func NewCustomDetector(threshold int) *CustomDetector {
    return &CustomDetector{threshold: threshold}
}

func (d *CustomDetector) Detect(file *models.FileInfo) ([]models.Violation, error) {
    var violations []models.Violation
    
    // Implement custom detection logic
    ast.Inspect(file.AST, func(n ast.Node) bool {
        if fn, ok := n.(*ast.FuncDecl); ok {
            if d.checkFunction(fn) {
                violation := models.Violation{
                    Type:        "custom_violation",
                    Message:     "Custom rule violation detected",
                    File:        file.Path,
                    Line:        file.FileSet.Position(fn.Pos()).Line,
                    Column:      file.FileSet.Position(fn.Pos()).Column,
                    Severity:    models.SeverityMedium,
                    Category:    models.CategoryStructure,
                    Rule:        d.Name(),
                }
                violations = append(violations, violation)
            }
        }
        return true
    })
    
    return violations, nil
}

func (d *CustomDetector) Name() string {
    return "custom_detector"
}

func (d *CustomDetector) Severity() models.Severity {
    return models.SeverityMedium
}

func (d *CustomDetector) Category() models.Category {
    return models.CategoryStructure
}

func (d *CustomDetector) checkFunction(fn *ast.FuncDecl) bool {
    // Custom validation logic
    return false
}
```

### Detector Registry

#### violations.Registry

Manages collection of violation detectors.

```go
type Registry struct {
    detectors map[string]Detector
}

func NewRegistry() *Registry
func (r *Registry) Register(detector Detector) error
func (r *Registry) Get(name string) (Detector, bool)
func (r *Registry) All() []Detector
func (r *Registry) DetectAll(file *models.FileInfo) ([]models.Violation, error)
```

**Example:**
```go
registry := violations.NewRegistry()

// Register built-in detectors
registry.Register(violations.NewFunctionDetector(cfg.Function))
registry.Register(violations.NewNamingDetector(cfg.Naming))

// Register custom detector
registry.Register(NewCustomDetector(10))

// Detect violations
violations, err := registry.DetectAll(fileInfo)
```

## Reporter API

### reporters.Reporter

Interface for implementing custom report generators.

```go
type Reporter interface {
    Generate(result *models.AnalysisResult) error
    Format() string
}
```

### Built-in Reporters

#### reporters.HTMLReporter

Generates interactive HTML reports.

```go
func NewHTMLReporter(config HTMLConfig) *HTMLReporter

type HTMLConfig struct {
    OutputPath      string
    AutoRefresh     bool
    RefreshInterval int
    Theme           string
    ShowSnippets    bool
}
```

#### reporters.MarkdownReporter

Generates AI-friendly markdown reports.

```go
func NewMarkdownReporter(config MarkdownConfig) *MarkdownReporter

type MarkdownConfig struct {
    OutputPath       string
    IncludeExamples  bool
    GroupBySeverity  bool
    AIFriendly       bool
}
```

#### reporters.ConsoleReporter

Generates terminal output.

```go
func NewConsoleReporter(config ConsoleConfig) *ConsoleReporter

type ConsoleConfig struct {
    Format      string  // table, json, csv
    ShowSummary bool
    Color       bool
}
```

### Custom Reporter Implementation

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    
    "github.com/ericfisherdev/goclean/pkg/models"
)

type JSONReporter struct {
    outputPath string
    prettyPrint bool
}

func NewJSONReporter(outputPath string, prettyPrint bool) *JSONReporter {
    return &JSONReporter{
        outputPath: outputPath,
        prettyPrint: prettyPrint,
    }
}

func (r *JSONReporter) Generate(result *models.AnalysisResult) error {
    file, err := os.Create(r.outputPath)
    if err != nil {
        return fmt.Errorf("failed to create output file: %w", err)
    }
    defer file.Close()
    
    encoder := json.NewEncoder(file)
    if r.prettyPrint {
        encoder.SetIndent("", "  ")
    }
    
    return encoder.Encode(result)
}

func (r *JSONReporter) Format() string {
    return "json"
}
```

## Models and Types

### models.AnalysisResult

Contains complete analysis results.

```go
type AnalysisResult struct {
    Summary     Summary               `json:"summary"`
    Violations  []Violation          `json:"violations"`
    Files       []FileResult         `json:"files"`
    Config      *config.Config       `json:"config,omitempty"`
    Timestamp   time.Time           `json:"timestamp"`
    Duration    time.Duration       `json:"duration"`
}
```

### models.Summary

Statistical summary of analysis results.

```go
type Summary struct {
    FilesScanned    int                    `json:"files_scanned"`
    FilesWithIssues int                    `json:"files_with_issues"`
    TotalViolations int                    `json:"total_violations"`
    BySeverity      map[Severity]int       `json:"by_severity"`
    ByCategory      map[Category]int       `json:"by_category"`
    ByType          map[string]int         `json:"by_type"`
}
```

### models.Violation

Represents a single clean code violation.

```go
type Violation struct {
    Type        string         `json:"type"`
    Message     string         `json:"message"`
    Description string         `json:"description,omitempty"`
    File        string         `json:"file"`
    Line        int           `json:"line"`
    Column      int           `json:"column"`
    EndLine     int           `json:"end_line,omitempty"`
    EndColumn   int           `json:"end_column,omitempty"`
    Severity    Severity      `json:"severity"`
    Category    Category      `json:"category"`
    Rule        string        `json:"rule"`
    Context     *CodeContext  `json:"context,omitempty"`
    Suggestions []Suggestion  `json:"suggestions,omitempty"`
}
```

### models.FileResult

Analysis results for a single file.

```go
type FileResult struct {
    Path        string      `json:"path"`
    Language    string      `json:"language"`
    Size        int64       `json:"size"`
    Lines       int         `json:"lines"`
    Violations  []Violation `json:"violations"`
    Metrics     FileMetrics `json:"metrics"`
    Error       string      `json:"error,omitempty"`
}
```

### models.FileMetrics

Code metrics for a file.

```go
type FileMetrics struct {
    Functions         int     `json:"functions"`
    Classes           int     `json:"classes"`
    Interfaces        int     `json:"interfaces"`
    CyclomaticComplexity int  `json:"cyclomatic_complexity"`
    LinesOfCode       int     `json:"lines_of_code"`
    CommentLines      int     `json:"comment_lines"`
    BlankLines        int     `json:"blank_lines"`
    CommentRatio      float64 `json:"comment_ratio"`
}
```

### Enumerations

#### models.Severity

```go
type Severity string

const (
    SeverityLow      Severity = "low"
    SeverityMedium   Severity = "medium"
    SeverityHigh     Severity = "high"
    SeverityCritical Severity = "critical"
)
```

#### models.Category

```go
type Category string

const (
    CategoryFunction      Category = "function"
    CategoryNaming        Category = "naming"
    CategoryStructure     Category = "structure"
    CategoryDocumentation Category = "documentation"
    CategorySecurity      Category = "security"
    CategoryPerformance   Category = "performance"
)
```

## CLI Integration

### Running GoClean Programmatically

```go
package main

import (
    "context"
    "os"
    
    "github.com/ericfisherdev/goclean/internal/cmd"
)

func main() {
    // Create CLI application
    app := cmd.NewApp()
    
    // Set arguments programmatically
    os.Args = []string{"goclean", "scan", "--path", "./src", "--html"}
    
    // Run CLI
    if err := app.Run(context.Background()); err != nil {
        os.Exit(1)
    }
}
```

### Custom CLI Commands

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/spf13/cobra"
    "github.com/ericfisherdev/goclean/pkg/goclean"
    "github.com/ericfisherdev/goclean/pkg/config"
)

func customScanCommand() *cobra.Command {
    var configPath string
    var outputFormat string
    
    cmd := &cobra.Command{
        Use:   "custom-scan [paths...]",
        Short: "Custom scan command with special features",
        RunE: func(cmd *cobra.Command, args []string) error {
            cfg, err := config.Load(configPath)
            if err != nil {
                return err
            }
            
            analyzer := goclean.New(cfg)
            result, err := analyzer.Analyze(context.Background(), args)
            if err != nil {
                return err
            }
            
            // Custom result processing
            fmt.Printf("Custom scan complete: %d violations found\n", 
                len(result.Violations))
                
            return nil
        },
    }
    
    cmd.Flags().StringVarP(&configPath, "config", "c", "goclean.yaml", "Config file")
    cmd.Flags().StringVar(&outputFormat, "format", "table", "Output format")
    
    return cmd
}
```

## Error Handling

### Error Types

GoClean defines specific error types for different failure scenarios:

#### errors.ConfigError

Configuration-related errors.

```go
type ConfigError struct {
    Field   string
    Value   interface{}
    Message string
}

func (e ConfigError) Error() string
func IsConfigError(err error) bool
```

#### errors.ScanError

Scanning and parsing errors.

```go
type ScanError struct {
    File    string
    Line    int
    Message string
    Cause   error
}

func (e ScanError) Error() string
func (e ScanError) Unwrap() error
func IsScanError(err error) bool
```

#### errors.DetectorError

Violation detection errors.

```go
type DetectorError struct {
    Detector string
    File     string
    Message  string
    Cause    error
}
```

### Error Handling Patterns

```go
// Check for specific error types
result, err := analyzer.Analyze(ctx, paths)
if err != nil {
    var configErr *errors.ConfigError
    var scanErr *errors.ScanError
    
    switch {
    case errors.As(err, &configErr):
        fmt.Printf("Configuration error in field %s: %s\n", 
            configErr.Field, configErr.Message)
    case errors.As(err, &scanErr):
        fmt.Printf("Scan error in file %s:%d: %s\n", 
            scanErr.File, scanErr.Line, scanErr.Message)
    default:
        fmt.Printf("Unexpected error: %v\n", err)
    }
    return
}

// Handle partial failures
for _, fileResult := range result.Files {
    if fileResult.Error != "" {
        fmt.Printf("Warning: Failed to process %s: %s\n", 
            fileResult.Path, fileResult.Error)
    }
}
```

### Graceful Degradation

GoClean is designed to continue processing even when individual files fail:

```go
// Configure error handling behavior
cfg.ErrorHandling = config.ErrorHandling{
    FailOnFirstError:    false,  // Continue on errors
    MaxErrors:          10,      // Stop after 10 errors
    SkipUnreadableFiles: true,   // Skip files that can't be read
    LogErrors:          true,    // Log all errors
}
```

## Integration Examples

### CI/CD Integration

```go
package main

import (
    "context"
    "fmt"
    "os"
    
    "github.com/ericfisherdev/goclean/pkg/goclean"
    "github.com/ericfisherdev/goclean/pkg/config"
)

func main() {
    // Load CI-specific configuration
    cfg := config.Default()
    cfg.Output.Console.Format = "json"
    cfg.Output.HTML.Enabled = false
    
    // Set stricter thresholds for CI
    cfg.Thresholds.FunctionLines = 20
    cfg.Thresholds.CyclomaticComplexity = 5
    
    analyzer := goclean.New(cfg)
    result, err := analyzer.Analyze(context.Background(), []string{"."})
    if err != nil {
        fmt.Fprintf(os.Stderr, "Analysis failed: %v\n", err)
        os.Exit(1)
    }
    
    // Fail build if critical violations found
    criticalCount := result.Summary.BySeverity[models.SeverityCritical]
    if criticalCount > 0 {
        fmt.Fprintf(os.Stderr, "Build failed: %d critical violations found\n", criticalCount)
        os.Exit(1)
    }
    
    fmt.Printf("Analysis passed: %d total violations found\n", 
        result.Summary.TotalViolations)
}
```

### IDE Plugin Integration

```go
package main

import (
    "context"
    
    "github.com/ericfisherdev/goclean/pkg/goclean"
    "github.com/ericfisherdev/goclean/pkg/config"
    "github.com/ericfisherdev/goclean/pkg/models"
)

type IDEAnalyzer struct {
    analyzer *goclean.Analyzer
}

func NewIDEAnalyzer() *IDEAnalyzer {
    cfg := config.Default()
    // IDE-specific configuration
    cfg.Output.Console.Enabled = false
    cfg.Output.HTML.Enabled = false
    
    return &IDEAnalyzer{
        analyzer: goclean.New(cfg),
    }
}

func (ide *IDEAnalyzer) AnalyzeBuffer(filename, content string) []models.Violation {
    result, err := ide.analyzer.AnalyzeContent(filename, content)
    if err != nil {
        return nil
    }
    return result.Violations
}

func (ide *IDEAnalyzer) GetQuickFixes(violation models.Violation) []models.Suggestion {
    return violation.Suggestions
}
```