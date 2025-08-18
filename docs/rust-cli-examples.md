# GoClean Rust CLI Usage Examples

This document provides comprehensive command-line examples for using GoClean with Rust projects.

**Note: This documentation has been updated to reflect only the actually implemented CLI flags. Many flags shown in previous versions were not implemented in the codebase.**

## Table of Contents

1. [Basic Commands](#basic-commands)
2. [Configuration Management](#configuration-management)
3. [Scanning Strategies](#scanning-strategies)
4. [Output Formats](#output-formats)
5. [Mixed Language Projects](#mixed-language-projects)
6. [CI/CD Integration](#cicd-integration)
7. [Advanced Usage](#advanced-usage)
8. [Troubleshooting Commands](#troubleshooting-commands)

## Basic Commands

### Simple Rust Project Scanning

```bash
# Scan current directory for Rust files
goclean scan --types .rs

# Scan specific directory
goclean scan ./src --types .rs

# Scan with structured console output for AI processing
goclean scan ./src --types .rs --console-violations

# Scan with HTML report (default format)
goclean scan ./src --types .rs --format html
```

### Quick Start Examples

```bash
# First-time setup for Rust project
goclean config init
goclean scan --languages rust

# Quick validation scan with structured output
goclean scan --types .rs --console-violations

# Generate HTML report to specific location
goclean scan --types .rs --format html --output ./reports/rust-report.html
```

## Configuration Management

### Using Predefined Configurations

```bash
# Use minimal Rust configuration
goclean scan --config configs/rust-minimal.yaml

# Use strict configuration for production code
goclean scan --config configs/rust-strict.yaml

# Use performance-focused configuration for large codebases
goclean scan --config configs/rust-performance-focused.yaml

# Use mixed project configuration
goclean scan --config configs/rust-mixed-project.yaml
```

### Configuration Generation

```bash
# Generate basic configuration file
goclean config init

# The generated config can be customized for Rust-specific settings
# Edit goclean.yaml to adjust thresholds and enable Rust features
```

## Scanning Strategies

### Basic Scanning

```bash
# Simple Rust scan
goclean scan --types .rs

# Scan with configuration file
goclean scan --config configs/rust-minimal.yaml
```

### Targeted Scanning

```bash
# Scan specific files by type
goclean scan --types .rs

# Scan with exclusions
goclean scan ./src --exclude "tests/" --exclude "examples/" --types .rs

# Scan specific paths only
goclean scan ./src/main.rs ./src/lib.rs

# Include test files in analysis
goclean scan --types .rs --include-tests
```

### Advanced Scanning Options

```bash
# Verbose output for debugging
goclean scan --types .rs --verbose

# Enable Rust optimizations explicitly
goclean scan --types .rs --rust-opt

# Configure Rust cache settings
goclean scan --types .rs --rust-cache-size 1000 --rust-cache-ttl 60
```

## Output Formats

### Console Output

```bash
# Structured console output for AI processing
goclean scan --types .rs --console-violations

# Verbose console output
goclean scan --types .rs --verbose
```

### HTML Reports

```bash
# Generate HTML report (default)
goclean scan --types .rs --format html

# HTML report to specific file
goclean scan --types .rs --format html --output ./reports/rust-analysis.html
```

### Markdown Reports

```bash
# Generate Markdown report
goclean scan --types .rs --format markdown

# Markdown to specific file
goclean scan --types .rs --format markdown --output ./docs/violations.md
```

### JSON Export

```bash
# Generate JSON output
goclean scan --types .rs --format json

# JSON to specific file
goclean scan --types .rs --format json --output results.json
```

## Mixed Language Projects

### Multi-Language Scanning

```bash
# Scan both Go and Rust files
goclean scan --types .go,.rs

# Use configuration for mixed projects
goclean scan --config configs/rust-mixed-project.yaml

# Language-specific reports
goclean scan --languages go --format html --output go-report.html
goclean scan --languages rust --format html --output rust-report.html
```

### Microservices Architecture

```bash
# Scan specific service directories
goclean scan ./services/api --types .go --format html --output go-services-report.html
goclean scan ./services/processor --types .rs --format html --output rust-services-report.html

# Combined analysis
goclean scan ./services --types .go,.rs --exclude "vendor/" --exclude "target/"
```

## CI/CD Integration

### Basic CI Usage

```bash
# CI-friendly JSON output
goclean scan --config configs/rust-minimal.yaml --format json --verbose > goclean-results.json

# Exit with non-zero code on violations
goclean scan --config configs/rust-strict.yaml --console-violations
```

### Advanced CI Integration

```bash
# Generate multiple report formats for CI artifacts
goclean scan --config configs/rust-minimal.yaml --format html --output ./artifacts/code-quality-report.html

# Performance-optimized scanning for large repositories
goclean scan --types .rs --rust-opt --verbose
```

### Platform-Specific Examples

#### Jenkins
```bash
goclean scan --config jenkins-rust.yaml --format json --output target/goclean-report.json
```

#### GitLab CI
```bash
goclean scan --config .goclean-ci.yaml --format json --output gl-code-quality-report.json
```

#### GitHub Actions
```bash
goclean scan --config configs/rust-strict.yaml --console-violations
```

## Advanced Usage

### Performance Optimization

```bash
# Large codebase scanning with optimizations
goclean scan --types .rs --rust-opt --rust-cache-size 2000

# Verbose performance information
goclean scan --types .rs --rust-opt --verbose
```

### Custom Thresholds

These are configured in the YAML configuration file, not via CLI flags:

```yaml
# In goclean.yaml
thresholds:
  function_lines: 20
  cyclomatic_complexity: 6
  parameters: 3
```

### Debug and Development

```bash
# Verbose output for debugging
goclean scan --types .rs --verbose

# Aggressive mode (includes test files)
goclean scan --types .rs --aggressive
```

## Troubleshooting Commands

### Configuration Issues

```bash
# Check configuration loading
goclean scan --config configs/rust-strict.yaml --verbose

# Test with minimal configuration
goclean config init
goclean scan --types .rs --verbose
```

### Performance Issues

```bash
# Enable Rust optimizations
goclean scan --types .rs --rust-opt --verbose

# Adjust cache settings
goclean scan --types .rs --rust-cache-size 500 --rust-cache-ttl 30
```

### File Detection Issues

```bash
# Verify file type detection
goclean scan --types .rs --verbose

# Check exclusion patterns
goclean scan ./src --exclude "target/" --types .rs --verbose
```

## Common Workflows

### Development Workflow

```bash
# Quick check during development
goclean scan --types .rs --console-violations

# Detailed analysis for code review
goclean scan --types .rs --format html --output ./reports/review.html
```

### Team Collaboration

```bash
# Generate team report
goclean scan --config team-standards.yaml --format html --output shared/team-quality-report.html

# Console summary for standups
goclean scan --config configs/rust-minimal.yaml --console-violations
```

### Release Preparation

```bash
# Comprehensive pre-release scan
goclean scan --config configs/rust-strict.yaml --format html --format json

# Quality metrics export
goclean scan --config configs/rust-minimal.yaml --format json --output "releases/quality-report.json"
```

## Notes

- Many flags shown in earlier versions of this documentation were not implemented
- Use `goclean scan --help` to see all available flags
- Configuration files provide more flexibility than CLI flags for complex setups
- The `--console-violations` flag provides structured output suitable for AI processing
- Rust optimizations are automatically enabled when scanning Rust files