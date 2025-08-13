# GoClean User Guide

## Table of Contents

1. [Introduction](#introduction)
2. [Installation](#installation)
3. [Quick Start](#quick-start)
4. [Command Reference](#command-reference)
5. [Configuration](#configuration)
6. [Understanding Reports](#understanding-reports)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)

## Introduction

GoClean is a powerful static analysis tool designed to help developers maintain clean, readable, and maintainable codebases. It scans your code for violations of clean code principles and provides detailed reports through multiple formats including HTML dashboards and AI-friendly markdown.

### What GoClean Detects

- **Function Quality Issues**: Long functions, high complexity, too many parameters
- **Naming Problems**: Non-descriptive names, inconsistent conventions
- **Code Structure Issues**: Large classes, deep nesting, code duplication
- **Documentation Issues**: Missing comments, outdated documentation, TODO tracking

## Installation

### Prerequisites

- Go 1.21 or later
- Git (for installation from source)

### Method 1: Install from GitHub Releases

Download the latest binary for your platform from the [releases page](https://github.com/ericfisherdev/goclean/releases):

```bash
# Linux/macOS
curl -L https://github.com/ericfisherdev/goclean/releases/latest/download/goclean-linux-amd64 -o goclean
chmod +x goclean
sudo mv goclean /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/ericfisherdev/goclean/releases/latest/download/goclean-windows-amd64.exe" -OutFile "goclean.exe"
```

### Method 2: Install with Go

```bash
go install github.com/ericfisherdev/goclean/cmd/goclean@latest
```

### Method 3: Build from Source

```bash
git clone https://github.com/ericfisherdev/goclean.git
cd goclean
make build
```

### Verify Installation

```bash
goclean version
```

## Quick Start

### Basic Scan

Scan the current directory with default settings:

```bash
goclean scan
```

This will:
- Scan all supported files in the current directory
- Generate an HTML report at `./reports/clean-code-report.html`
- Display a summary in the terminal

### Scan Specific Directory

```bash
goclean scan --path ./src
```

### Generate Multiple Report Formats

```bash
goclean scan --html --markdown --console
```

### Use Custom Configuration

```bash
goclean scan --config ./my-goclean.yaml
```

## Command Reference

### Global Flags

- `--config, -c`: Path to configuration file (default: `./goclean.yaml`)
- `--verbose, -v`: Enable verbose logging
- `--quiet, -q`: Suppress output except errors
- `--help, -h`: Show help information

### scan command

Scans source code for clean code violations.

```bash
goclean scan [flags]
```

#### Flags

- `--path, -p`: Path to scan (default: current directory)
- `--recursive, -r`: Scan directories recursively (default: true)
- `--exclude`: Comma-separated list of patterns to exclude
- `--include`: Comma-separated list of patterns to include
- `--html`: Generate HTML report (default: true)
- `--markdown`: Generate markdown report (default: false)
- `--console`: Display results in console (default: true)
- `--output, -o`: Output directory for reports (default: `./reports`)
- `--format`: Output format for console (table, json, csv)
- `--severity`: Minimum severity level to report (low, medium, high, critical)

#### Examples

```bash
# Scan with custom exclusions
goclean scan --exclude "vendor/,*.test.go,testdata/"

# Generate only markdown report
goclean scan --markdown --no-html --no-console

# Scan with high severity filter
goclean scan --severity high

# Custom output directory
goclean scan --output ./code-analysis
```

### config command

Manage configuration settings.

```bash
# Generate default configuration file
goclean config init

# Validate configuration file
goclean config validate --config ./goclean.yaml

# Show current configuration
goclean config show
```

### version command

Display version information.

```bash
goclean version
```

## Configuration

GoClean uses YAML configuration files to customize scanning behavior and thresholds.

### Default Configuration

Generate a default configuration file:

```bash
goclean config init
```

This creates `goclean.yaml` with default settings.

### Configuration Structure

```yaml
# Scanning configuration
scan:
  paths:
    - "./src"
    - "./internal"
  exclude:
    - "vendor/"
    - "node_modules/"
    - "*.test.go"
    - "testdata/"
  file_types:
    - ".go"
    - ".js"
    - ".ts"
    - ".py"
    - ".java"
    - ".cs"
  max_file_size: "1MB"
  follow_symlinks: false

# Violation detection thresholds
thresholds:
  function_lines: 25
  cyclomatic_complexity: 8
  parameters: 4
  nesting_depth: 3
  class_lines: 150
  line_length: 120
  duplicate_lines: 6

# Naming convention rules
naming:
  min_name_length: 3
  allow_single_letter: false
  enforce_camel_case: true
  enforce_constants_upper: true

# Output configuration
output:
  html:
    enabled: true
    path: "./reports/clean-code-report.html"
    auto_refresh: true
    refresh_interval: 10
    theme: "auto"
    show_code_snippets: true
  markdown:
    enabled: false
    path: "./reports/violations.md"
    include_examples: true
    group_by_severity: true
  console:
    enabled: true
    format: "table"
    show_summary: true
    color: true

# Logging configuration
logging:
  level: "info"
  format: "structured"
  file: "./logs/goclean.log"
```

### Environment Variables

Override configuration values using environment variables:

```bash
export GOCLEAN_LOG_LEVEL=debug
export GOCLEAN_OUTPUT_HTML_THEME=dark
export GOCLEAN_SCAN_RECURSIVE=true
```

## Understanding Reports

### HTML Report

The HTML report provides an interactive dashboard with:

- **Overview Dashboard**: Summary statistics and trends
- **Violation Details**: Categorized list of all violations
- **File Explorer**: Navigate through scanned files
- **Code Snippets**: View problematic code with syntax highlighting
- **Filtering Options**: Filter by severity, type, or file

#### Features

- **Auto-refresh**: Updates every 10 seconds during active scans
- **Dark/Light Theme**: Automatic or manual theme switching
- **Responsive Design**: Works on desktop and mobile devices
- **Export Options**: Download data as CSV or JSON

### Markdown Report

The markdown report is optimized for AI analysis tools:

```markdown
# Code Analysis Report

## Executive Summary
- Total Files Scanned: 156
- Total Violations: 89
- Critical Issues: 5
- High Priority: 23

## Violations by Category

### Function Quality (32 issues)
- Long Functions: 18
- High Complexity: 14

### Naming Issues (25 issues)
- Non-descriptive Names: 15
- Inconsistent Naming: 10

## Detailed Findings

### Critical Issues

#### File: src/scanner/engine.go:45
**Violation**: Function too long (78 lines)
**Severity**: Critical
**Recommendation**: Extract smaller functions

```go
func (e *Engine) ScanDirectory(path string) error {
    // Long function implementation...
}
```
```

### Console Output

Terminal output provides quick feedback:

```
GoClean Analysis Results
========================

📊 Summary:
   Files Scanned: 156
   Violations: 89
   Critical: 5
   High: 23
   Medium: 34
   Low: 27

🔍 Top Issues:
   • Function length violations: 18
   • Naming convention issues: 15
   • High complexity functions: 14

📁 Most Problematic Files:
   • src/scanner/engine.go (12 violations)
   • internal/parser/ast.go (8 violations)
   • cmd/main.go (6 violations)

💡 Next Steps:
   • Focus on critical and high priority issues
   • Review function length in scanner package
   • Improve naming consistency across codebase

View detailed report: ./reports/clean-code-report.html
```

## Best Practices

### Project Setup

1. **Add to Git Hooks**: Run GoClean in pre-commit hooks
   ```bash
   # .git/hooks/pre-commit
   #!/bin/sh
   goclean scan --severity high --no-html --quiet
   ```

2. **CI/CD Integration**: Include in build pipelines
   ```yaml
   # GitHub Actions example
   - name: Run GoClean
     run: |
       go install github.com/ericfisherdev/goclean/cmd/goclean@latest
       goclean scan --markdown --output ./artifacts
   ```

3. **Team Configuration**: Commit configuration file
   ```bash
   git add goclean.yaml
   git commit -m "Add GoClean configuration"
   ```

### Effective Usage

1. **Start with High Severity**: Focus on critical and high issues first
2. **Gradual Improvement**: Lower thresholds over time as code quality improves
3. **Custom Rules**: Adjust thresholds to match team standards
4. **Regular Scans**: Run regularly to prevent regression

### Ignoring Violations

For legitimate exceptions, use code comments:

```go
// goclean:ignore function-length "Legacy function, refactoring scheduled"
func legacyComplexFunction() {
    // Complex implementation...
}
```

## Troubleshooting

### Common Issues

#### "No files found to scan"
- Check file extensions in configuration
- Verify paths are correct
- Ensure files aren't excluded by patterns

#### "Permission denied" errors
- Run with appropriate file permissions
- Check directory access rights
- Use sudo if scanning system directories

#### "Out of memory" errors
- Reduce scan scope
- Increase system memory
- Use `--max-file-size` to limit large files

#### HTML report not refreshing
- Check browser cache
- Verify `auto_refresh` is enabled
- Ensure file permissions allow updates

### Performance Optimization

1. **Exclude Unnecessary Files**:
   ```yaml
   scan:
     exclude:
       - "vendor/"
       - "node_modules/"
       - "*.min.js"
       - "dist/"
   ```

2. **Limit File Size**:
   ```yaml
   scan:
     max_file_size: "1MB"
   ```

3. **Use Specific Paths**:
   ```bash
   goclean scan --path ./src --path ./internal
   ```

### Debug Mode

Enable debug logging for troubleshooting:

```bash
goclean scan --verbose --log-level debug
```

### Getting Help

- **Documentation**: Check the [docs](../docs/) directory
- **Issues**: Report bugs on [GitHub Issues](https://github.com/ericfisherdev/goclean/issues)
- **Discussions**: Ask questions in [GitHub Discussions](https://github.com/ericfisherdev/goclean/discussions)

### Supported File Types

Currently supported:
- Go (`.go`)
- JavaScript (`.js`)
- TypeScript (`.ts`)
- Python (`.py`)
- Java (`.java`)
- C# (`.cs`)

Planned support:
- Rust (`.rs`)
- C++ (`.cpp`, `.hpp`)
- Ruby (`.rb`)