# Configuration Reference

This document provides comprehensive information about configuring GoClean for your project needs.

## Table of Contents

1. [Configuration File Structure](#configuration-file-structure)
2. [Scanning Configuration](#scanning-configuration)
3. [Violation Thresholds](#violation-thresholds)
4. [Naming Rules](#naming-rules)
5. [Rust-Specific Configuration](#rust-specific-configuration)
6. [Clippy Integration](#clippy-integration)
7. [Output Configuration](#output-configuration)
8. [Logging Configuration](#logging-configuration)
9. [Environment Variables](#environment-variables)
10. [Configuration Examples](#configuration-examples)
11. [Migration Guide](#migration-guide)

## Configuration File Structure

GoClean uses YAML configuration files. The default configuration file is `goclean.yaml` in the project root.

### Creating Configuration

```bash
# Generate default configuration
goclean config init

# Generate with custom name
goclean config init --output my-config.yaml

# Validate existing configuration
goclean config validate
```

### Basic Structure

```yaml
scan:          # What and how to scan
thresholds:    # Violation detection limits
naming:        # Naming convention rules
output:        # Report generation settings
logging:       # Logging configuration
```

## Scanning Configuration

Controls which files are scanned and how the scanning process behaves.

### scan.paths

**Type**: `[]string`
**Default**: `["."]`

Directories to scan for source files.

```yaml
scan:
  paths:
    - "./src"
    - "./internal"
    - "./pkg"
```

### scan.exclude

**Type**: `[]string`
**Default**: `["vendor/", "node_modules/", "*.test.go", "testdata/"]`

Patterns of files and directories to exclude from scanning.

```yaml
scan:
  exclude:
    - "vendor/"
    - "node_modules/"
    - "*.test.go"
    - "testdata/"
    - "*.generated.go"
    - ".git/"
    - "dist/"
    - "target/"     # Rust build artifacts
    - "Cargo.lock"  # Rust lock file
```

**Pattern Syntax**:
- `*`: Matches any sequence of characters
- `?`: Matches any single character
- `[abc]`: Matches any character in brackets
- `**`: Matches directories recursively

### scan.file_types

**Type**: `[]string`
**Default**: `[".go", ".js", ".ts", ".py", ".java", ".cs"]`

File extensions to include in scanning.

```yaml
scan:
  file_types:
    - ".go"
    - ".js"
    - ".ts"
    - ".py"
    - ".java"
    - ".cs"
    - ".rs"    # Rust (full support)
    - ".cpp"   # C++ (experimental)
```

### scan.max_file_size

**Type**: `string`
**Default**: `"1MB"`

Maximum file size to scan. Files larger than this limit are skipped.

```yaml
scan:
  max_file_size: "2MB"  # or "2048KB" or "2097152B"
```

### scan.follow_symlinks

**Type**: `bool`
**Default**: `false`

Whether to follow symbolic links during directory traversal.

```yaml
scan:
  follow_symlinks: true
```

### scan.concurrent_files

**Type**: `int`
**Default**: `10`

Number of files to process concurrently for better performance.

```yaml
scan:
  concurrent_files: 20
```

## Violation Thresholds

Configure the limits that trigger violation detection.

### Function Quality Thresholds

```yaml
thresholds:
  # Function length violations
  function_lines: 25
  
  # Cyclomatic complexity violations
  cyclomatic_complexity: 8
  
  # Parameter count violations
  parameters: 4
  
  # Nesting depth violations
  nesting_depth: 3
  
  # Return statement count
  return_statements: 5
```

### Code Structure Thresholds

```yaml
thresholds:
  # Class/struct size violations
  class_lines: 150
  
  # Line length violations
  line_length: 120
  
  # Duplicate code detection
  duplicate_lines: 6
  
  # File size violations
  file_lines: 500
```

### Comment and Documentation

```yaml
thresholds:
  # Missing documentation violations
  undocumented_public_functions: true
  
  # Comment density (comments/code ratio)
  min_comment_density: 0.1  # 10%
  max_comment_density: 0.5  # 50%
  
  # TODO/FIXME tracking
  track_technical_debt: true
```

## Naming Rules

Configure naming convention enforcement.

### Basic Naming Rules

```yaml
naming:
  # Minimum identifier length
  min_name_length: 3
  
  # Allow single-letter variables in specific contexts
  allow_single_letter: false
  
  # Allow abbreviations and acronyms
  allow_abbreviations: false
  
  # Allow underscores in names
  allow_underscores: true
```

### Language-Specific Rules

```yaml
naming:
  go:
    enforce_camel_case: true
    enforce_constants_upper: true
    allow_package_name_underscores: false
  
  javascript:
    enforce_camel_case: true
    allow_snake_case_files: true
  
  python:
    enforce_snake_case: true
    enforce_constants_upper: true
  
  java:
    enforce_camel_case: true
    enforce_class_pascal_case: true
```

### Custom Naming Patterns

```yaml
naming:
  custom_patterns:
    functions:
      pattern: "^[a-z][a-zA-Z0-9]*$"
      message: "Function names should be camelCase"
    
    constants:
      pattern: "^[A-Z][A-Z0-9_]*$"
      message: "Constants should be UPPER_CASE"
    
    variables:
      pattern: "^[a-z][a-zA-Z0-9]*$"
      message: "Variables should be camelCase"
```

## Rust-Specific Configuration

GoClean provides comprehensive support for Rust projects with specialized analysis capabilities.

### Basic Rust Configuration

```yaml
rust:
  # Enable Rust-specific analysis
  enable_ownership_analysis: true
  enable_error_handling_check: true
  enable_pattern_match_check: true
  
  # Naming convention enforcement
  enforce_snake_case: true      # Functions, variables, modules
  enforce_pascal_case: true     # Types (structs, enums, traits)
  enforce_screaming_snake: true # Constants
```

### Ownership and Borrowing Analysis

```yaml
rust:
  enable_ownership_analysis: true
  max_lifetime_params: 3
  detect_unnecessary_clones: true
```

**Configuration Options**:
- `enable_ownership_analysis`: Enable ownership pattern analysis
- `max_lifetime_params`: Maximum lifetime parameters before flagging as complex
- `detect_unnecessary_clones`: Flag unnecessary `.clone()` calls

**Detected Violations**:
- `RUST_UNNECESSARY_CLONE` - Unnecessary clone operations
- `RUST_INEFFICIENT_BORROWING` - Suboptimal borrowing patterns
- `RUST_COMPLEX_LIFETIME` - Overly complex lifetime parameters
- `RUST_MOVE_SEMANTICS_VIOLATION` - Incorrect move semantics usage

### Error Handling Analysis

```yaml
rust:
  enable_error_handling_check: true
  allow_unwrap: false
  allow_expect: false
  enforce_result_propagation: true
```

**Configuration Options**:
- `enable_error_handling_check`: Enable error handling pattern analysis
- `allow_unwrap`: Allow `.unwrap()` calls (not recommended for production)
- `allow_expect`: Allow `.expect()` calls (not recommended for production)
- `enforce_result_propagation`: Require proper error propagation with `?`

**Detected Violations**:
- `RUST_OVERUSE_UNWRAP` - Usage of `.unwrap()` method
- `RUST_MISSING_ERROR_PROPAGATION` - Missing `?` operator usage
- `RUST_INCONSISTENT_ERROR_TYPE` - Inconsistent error types
- `RUST_PANIC_PRONE_CODE` - Code patterns that may cause panics

### Pattern Matching Analysis

```yaml
rust:
  enable_pattern_match_check: true
  require_exhaustive_match: true
  max_nested_match_depth: 3
```

**Configuration Options**:
- `enable_pattern_match_check`: Enable pattern matching analysis
- `require_exhaustive_match`: Require exhaustive pattern matching
- `max_nested_match_depth`: Maximum nesting depth for match expressions

**Detected Violations**:
- `RUST_NON_EXHAUSTIVE_MATCH` - Non-exhaustive pattern matches
- `RUST_NESTED_PATTERN_MATCHING` - Overly nested match expressions
- `RUST_INEFFICIENT_DESTRUCTURING` - Inefficient destructuring patterns

### Trait and Implementation Analysis

```yaml
rust:
  max_trait_bounds: 5
  max_impl_methods: 20
  detect_orphan_instances: true
```

**Configuration Options**:
- `max_trait_bounds`: Maximum trait bounds before flagging as complex
- `max_impl_methods`: Maximum methods per impl block
- `detect_orphan_instances`: Detect potential orphan trait implementations

### Safety Analysis

```yaml
rust:
  allow_unsafe: true
  require_unsafe_comments: true
  detect_transmute_usage: true
```

**Configuration Options**:
- `allow_unsafe`: Allow unsafe code blocks
- `require_unsafe_comments`: Require documentation for unsafe code
- `detect_transmute_usage`: Flag potentially dangerous transmute usage

**Detected Violations**:
- `RUST_UNNECESSARY_UNSAFE` - Unnecessary unsafe blocks
- `RUST_UNSAFE_WITHOUT_COMMENT` - Unsafe code without documentation
- `RUST_TRANSMUTE_ABUSE` - Dangerous transmute usage

### Performance Analysis

```yaml
rust:
  detect_inefficient_string: true
  detect_boxed_primitives: true
  detect_blocking_in_async: true
```

**Configuration Options**:
- `detect_inefficient_string`: Flag inefficient string operations
- `detect_boxed_primitives`: Detect unnecessary boxing of primitives
- `detect_blocking_in_async`: Find blocking calls in async functions

**Detected Violations**:
- `RUST_INEFFICIENT_STRING_CONCAT` - Inefficient string concatenation
- `RUST_UNNECESSARY_ALLOCATION` - Unnecessary heap allocations
- `RUST_BLOCKING_IN_ASYNC` - Blocking calls in async functions

### Module and Structure Analysis

```yaml
rust:
  max_module_depth: 5
  max_file_lines: 500
  enforce_module_privacy: true
```

**Configuration Options**:
- `max_module_depth`: Maximum module nesting depth
- `max_file_lines`: Maximum lines per file
- `enforce_module_privacy`: Enforce proper module privacy patterns

### Macro Analysis

```yaml
rust:
  max_macro_complexity: 10
  allow_recursive_macros: false
  detect_macro_abuse: true
```

**Configuration Options**:
- `max_macro_complexity`: Maximum macro complexity score
- `allow_recursive_macros`: Allow recursive macro definitions
- `detect_macro_abuse`: Flag overly complex macro usage

## Clippy Integration

GoClean integrates seamlessly with rust-clippy to provide comprehensive analysis.

### Basic Clippy Configuration

```yaml
clippy:
  enabled: true
  categories:
    - correctness    # Critical correctness issues
    - suspicious     # Suspicious code patterns
    - style         # Style and idiom violations
    - complexity    # Code complexity issues
    - perf          # Performance improvements
```

### Advanced Clippy Configuration

```yaml
clippy:
  enabled: true
  categories:
    - correctness
    - suspicious
    - style
    - complexity
    - perf
    - pedantic      # Extra strict lints
    - nursery       # Experimental lints
  
  severity_mapping:
    error: critical
    warn: high
    info: medium
    note: low
  
  additional_lints:
    - clippy::all
    - clippy::pedantic
    - clippy::cargo
```

### Clippy Categories

| Category | Description | Example Lints |
|----------|-------------|---------------|
| `correctness` | Code that is definitely wrong | `clippy::absurd_extreme_comparisons` |
| `suspicious` | Code that is probably wrong | `clippy::empty_loop` |
| `style` | Code style violations | `clippy::redundant_field_names` |
| `complexity` | Complex code patterns | `clippy::cognitive_complexity` |
| `perf` | Performance improvements | `clippy::inefficient_to_string` |
| `pedantic` | Extra strict style checks | `clippy::missing_docs_in_private_items` |
| `nursery` | Experimental lints | `clippy::future_not_send` |

### Severity Mapping

Map clippy lint levels to GoClean severity levels:

```yaml
clippy:
  severity_mapping:
    error: critical    # Must fix before production
    warn: high        # Should fix soon
    info: medium      # Improvement opportunity
    note: low         # Minor suggestion
```

### Custom Clippy Configuration

```yaml
clippy:
  enabled: true
  
  # Run clippy with specific configuration
  config_file: ".clippy.toml"
  
  # Additional command-line arguments
  args:
    - "--all-features"
    - "--all-targets"
  
  # Override default categories
  categories:
    - correctness
    - suspicious
    - perf
  
  # Disable specific lints
  disabled_lints:
    - clippy::too_many_arguments
    - clippy::module_name_repetitions
```

### Rust Configuration Examples

#### Minimal Rust Configuration

```yaml
scan:
  file_types: [".rs"]
  exclude: ["target/", "Cargo.lock"]

rust:
  enable_ownership_analysis: true
  enable_error_handling_check: true
  allow_unwrap: false

clippy:
  enabled: true
  categories: [correctness, suspicious]

thresholds:
  function_lines: 30
  cyclomatic_complexity: 10
```

#### Strict Rust Configuration

```yaml
scan:
  file_types: [".rs"]
  exclude: ["target/", "Cargo.lock"]

rust:
  enable_ownership_analysis: true
  enable_error_handling_check: true
  enable_pattern_match_check: true
  allow_unwrap: false
  allow_expect: false
  enforce_result_propagation: true
  max_lifetime_params: 2
  require_exhaustive_match: true

clippy:
  enabled: true
  categories: [correctness, suspicious, style, complexity, perf, pedantic]

thresholds:
  function_lines: 20
  cyclomatic_complexity: 6
  parameters: 3
```

#### Mixed Go/Rust Configuration

```yaml
scan:
  file_types: [".go", ".rs"]
  exclude: ["vendor/", "target/", "*.test.go"]

rust:
  enable_ownership_analysis: true
  enable_error_handling_check: true
  enforce_snake_case: true

clippy:
  enabled: true
  categories: [correctness, suspicious, style, complexity, perf]

thresholds:
  function_lines: 25
  cyclomatic_complexity: 8
  parameters: 4
```

## Output Configuration

Configure how reports are generated and formatted.

### HTML Output

```yaml
output:
  html:
    enabled: true
    path: "./reports/clean-code-report.html"
    
    # Auto-refresh during active scanning
    auto_refresh: true
    refresh_interval: 10  # seconds
    
    # Theme selection
    theme: "auto"  # auto, light, dark
    
    # Include code snippets in report
    show_code_snippets: true
    
    # Maximum lines to show in snippets
    snippet_lines: 10
    
    # Enable interactive features
    enable_filtering: true
    enable_sorting: true
    
    # Custom CSS file
    custom_css: "./custom-styles.css"
```

### Markdown Output

```yaml
output:
  markdown:
    enabled: false
    path: "./reports/violations.md"
    
    # Include code examples
    include_examples: true
    
    # Group violations by severity
    group_by_severity: true
    
    # Group violations by file
    group_by_file: false
    
    # Include summary statistics
    include_summary: true
    
    # AI-friendly formatting
    ai_friendly: true
    
    # Template file for custom formatting
    template: "./templates/markdown-template.md"
```

### Console Output

```yaml
output:
  console:
    enabled: true
    
    # Output format: table, json, csv, summary
    format: "table"
    
    # Show summary statistics
    show_summary: true
    
    # Enable colored output
    color: true
    
    # Show progress during scanning
    show_progress: true
    
    # Truncate long file paths
    truncate_paths: true
    max_path_length: 60
```

### Export Options

```yaml
output:
  export:
    # Export raw data as JSON
    json:
      enabled: false
      path: "./reports/data.json"
      pretty_print: true
    
    # Export as CSV for spreadsheet analysis
    csv:
      enabled: false
      path: "./reports/violations.csv"
      include_headers: true
    
    # Export as XML
    xml:
      enabled: false
      path: "./reports/violations.xml"
```

## Logging Configuration

Control logging behavior and output.

```yaml
logging:
  # Log level: trace, debug, info, warn, error
  level: "info"
  
  # Log format: text, json, structured
  format: "structured"
  
  # Log to file
  file: "./logs/goclean.log"
  
  # Rotate log files
  rotate: true
  max_size: "100MB"
  max_age: 30  # days
  max_backups: 5
  
  # Include timestamps
  timestamps: true
  
  # Include caller information
  caller: false
  
  # Include stack traces for errors
  stack_trace: true
```

## Environment Variables

Override configuration values using environment variables with the prefix `GOCLEAN_`.

### Variable Naming Convention

Configuration path `scan.paths` becomes `GOCLEAN_SCAN_PATHS`
Configuration path `output.html.theme` becomes `GOCLEAN_OUTPUT_HTML_THEME`

### Common Environment Variables

```bash
# Scanning configuration
export GOCLEAN_SCAN_PATHS="./src,./internal"
export GOCLEAN_SCAN_EXCLUDE="vendor/,*.test.go"
export GOCLEAN_SCAN_RECURSIVE=true

# Thresholds
export GOCLEAN_THRESHOLDS_FUNCTION_LINES=30
export GOCLEAN_THRESHOLDS_CYCLOMATIC_COMPLEXITY=10

# Output configuration
export GOCLEAN_OUTPUT_HTML_ENABLED=true
export GOCLEAN_OUTPUT_HTML_THEME=dark
export GOCLEAN_OUTPUT_MARKDOWN_ENABLED=true

# Logging
export GOCLEAN_LOG_LEVEL=debug
export GOCLEAN_LOG_FORMAT=json
```

### Loading Environment Variables

```bash
# Load from .env file
source .env
goclean scan

# Or use with docker
docker run --env-file .env goclean:latest scan
```

## Configuration Examples

### Strict Configuration

High standards for clean code:

```yaml
thresholds:
  function_lines: 15
  cyclomatic_complexity: 5
  parameters: 3
  nesting_depth: 2
  class_lines: 100
  line_length: 100

naming:
  min_name_length: 4
  allow_single_letter: false
  allow_abbreviations: false

output:
  console:
    format: "table"
    show_summary: true
```

### Legacy Codebase Configuration

More lenient settings for existing projects:

```yaml
thresholds:
  function_lines: 50
  cyclomatic_complexity: 15
  parameters: 8
  nesting_depth: 5
  class_lines: 300

naming:
  allow_abbreviations: true
  allow_single_letter: true

scan:
  exclude:
    - "legacy/"
    - "deprecated/"
    - "third-party/"
```

### Performance-Focused Configuration

Optimized for large codebases:

```yaml
scan:
  concurrent_files: 50
  max_file_size: "500KB"
  exclude:
    - "vendor/"
    - "node_modules/"
    - "*.min.js"
    - "dist/"
    - "build/"

output:
  html:
    show_code_snippets: false
  console:
    format: "summary"

logging:
  level: "warn"
```

### CI/CD Configuration

Optimized for continuous integration:

```yaml
scan:
  paths: ["./src"]
  
output:
  html:
    enabled: false
  markdown:
    enabled: true
    ai_friendly: true
  console:
    format: "json"
    color: false

logging:
  level: "error"
  format: "json"
```

## Migration Guide

### Upgrading from v1.x to v2.x

Key changes in configuration format:

1. **Threshold restructuring**:
   ```yaml
   # v1.x
   function_max_lines: 25
   
   # v2.x
   thresholds:
     function_lines: 25
   ```

2. **Output format changes**:
   ```yaml
   # v1.x
   html_output: true
   html_path: "./report.html"
   
   # v2.x
   output:
     html:
       enabled: true
       path: "./report.html"
   ```

3. **Scanning configuration**:
   ```yaml
   # v1.x
   scan_paths: ["./src"]
   exclude_patterns: ["vendor/"]
   
   # v2.x
   scan:
     paths: ["./src"]
     exclude: ["vendor/"]
   ```

### Migration Tool

Use the built-in migration tool:

```bash
# Migrate v1.x configuration to v2.x
goclean config migrate --input old-config.yaml --output new-config.yaml

# Validate migrated configuration
goclean config validate --config new-config.yaml
```

## Advanced Configuration

### Custom Violation Detectors

```yaml
custom_detectors:
  - name: "company_naming_convention"
    type: "naming"
    pattern: "^(get|set|is|has)[A-Z].*"
    message: "Use company naming convention"
    severity: "medium"
  
  - name: "no_fmt_print"
    type: "pattern"
    pattern: "fmt\\.Print"
    message: "Use structured logging instead of fmt.Print"
    severity: "high"
```

### Rule Exceptions

```yaml
exceptions:
  files:
    - path: "internal/legacy/*"
      rules: ["function_length", "complexity"]
    - path: "*/generated.go"
      rules: ["*"]  # Ignore all rules
  
  functions:
    - pattern: "Test*"
      rules: ["function_length"]
    - pattern: "*Handler"
      rules: ["parameters"]
```

### Integration Settings

```yaml
integrations:
  github:
    enabled: true
    token_env: "GITHUB_TOKEN"
    create_issues: false
    label_prefix: "goclean:"
  
  jira:
    enabled: false
    url: "https://company.atlassian.net"
    project: "TECH"
    issue_type: "Task"
  
  slack:
    enabled: false
    webhook_url_env: "SLACK_WEBHOOK"
    channel: "#code-quality"
```