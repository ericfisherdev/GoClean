# Developer Guide

This guide helps contributors understand the GoClean codebase and development workflow.

## Table of Contents

1. [Development Setup](#development-setup)
2. [Architecture Overview](#architecture-overview)
3. [Code Organization](#code-organization)
4. [Adding New Features](#adding-new-features)
5. [Testing Strategy](#testing-strategy)
6. [Performance Guidelines](#performance-guidelines)
7. [Documentation Standards](#documentation-standards)
8. [Release Process](#release-process)

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git
- Make (optional, for convenience commands)
- golangci-lint (for code quality checks)

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/ericfisherdev/goclean.git
cd goclean

# Install dependencies
go mod download

# Build the application
make build

# Run tests
make test

# Run linting
make lint
```

### Development Tools

Install recommended development tools:

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install test coverage tools
go install github.com/axw/gocov/gocov@latest
go install github.com/matm/gocov-html@latest

# Install benchmarking tools
go install golang.org/x/perf/cmd/benchstat@latest
```

### IDE Configuration

#### VS Code

Recommended extensions:
- Go (official Go extension)
- golangci-lint
- Go Test Explorer
- GitLens

Workspace settings (`.vscode/settings.json`):
```json
{
    "go.lintTool": "golangci-lint",
    "go.lintFlags": ["--fast"],
    "go.testFlags": ["-v"],
    "go.testTimeout": "30s",
    "go.coverOnSave": true,
    "editor.formatOnSave": true
}
```

#### GoLand/IntelliJ

Configure file watchers for:
- gofmt on Go files
- golangci-lint on Go files
- Test runner for *_test.go files

## Architecture Overview

GoClean follows clean architecture principles with clear separation of concerns.

### Core Components

```
internal/
├── config/          # Configuration management
├── scanner/         # File scanning and parsing
├── violations/      # Violation detection logic
├── reporters/       # Report generation
├── models/          # Data structures
└── testutils/       # Testing utilities
```

### Data Flow

```
1. Configuration Loading → 2. File Discovery → 3. AST Parsing → 4. Violation Detection → 5. Report Generation
```

### Key Interfaces

#### Scanner Interface

```go
type Scanner interface {
    Scan(ctx context.Context, paths []string) (*models.ScanResult, error)
}
```

#### Detector Interface

```go
type Detector interface {
    Detect(file *models.FileInfo) ([]models.Violation, error)
    Name() string
    Severity() models.Severity
}
```

#### Reporter Interface

```go
type Reporter interface {
    Generate(result *models.ScanResult) error
    Format() string
}
```

## Code Organization

### Package Structure

#### internal/config

Handles configuration loading, validation, and defaults.

- `config.go`: Main configuration struct and loading logic
- `defaults.go`: Default configuration values
- `validation.go`: Configuration validation rules

#### internal/scanner

Core scanning functionality and file processing.

- `engine.go`: Main scanning orchestrator
- `parser.go`: Language-specific parsing logic
- `ast_analyzer.go`: AST analysis utilities
- `file_walker.go`: Directory traversal and file discovery

#### internal/violations

Violation detection implementations.

- `detector.go`: Base detector interface and registry
- `function.go`: Function-related violation detection
- `naming.go`: Naming convention checks
- `structure.go`: Code structure analysis
- `documentation.go`: Documentation-related checks

#### internal/reporters

Report generation in various formats.

- `manager.go`: Reporter coordination and management
- `html.go`: HTML report generation
- `markdown.go`: Markdown report generation
- `console.go`: Terminal output formatting

#### internal/models

Data structures and types used throughout the application.

- `violation.go`: Violation representation
- `file_info.go`: File metadata and content
- `report.go`: Report structure and statistics

### Naming Conventions

- **Packages**: Short, lowercase, descriptive names
- **Types**: PascalCase with clear, descriptive names
- **Functions**: camelCase, start with verb for actions
- **Variables**: camelCase, descriptive names
- **Constants**: UPPER_CASE with underscores
- **Interfaces**: End with "-er" when possible (Scanner, Detector)

### Error Handling

Use structured error handling with context:

```go
// Good
func (s *Scanner) scanFile(path string) error {
    if _, err := os.Stat(path); err != nil {
        return fmt.Errorf("failed to access file %s: %w", path, err)
    }
    // ...
}

// Bad
func (s *Scanner) scanFile(path string) error {
    if _, err := os.Stat(path); err != nil {
        return err
    }
    // ...
}
```

## Adding New Features

### Adding a New Violation Detector

1. **Create the detector file** in `internal/violations/`:

```go
// internal/violations/new_detector.go
package violations

import "github.com/ericfisherdev/goclean/internal/models"

type NewDetector struct {
    config Config
}

func NewNewDetector(config Config) *NewDetector {
    return &NewDetector{config: config}
}

func (d *NewDetector) Detect(file *models.FileInfo) ([]models.Violation, error) {
    var violations []models.Violation
    
    // Detection logic here
    
    return violations, nil
}

func (d *NewDetector) Name() string {
    return "new_detector"
}

func (d *NewDetector) Severity() models.Severity {
    return models.SeverityMedium
}
```

2. **Register the detector** in `detector.go`:

```go
func NewDetectorRegistry(config Config) *DetectorRegistry {
    registry := &DetectorRegistry{
        detectors: make(map[string]Detector),
    }
    
    // ... existing detectors
    registry.Register(NewNewDetector(config))
    
    return registry
}
```

3. **Add configuration** in `internal/config/config.go`:

```go
type ThresholdConfig struct {
    // ... existing thresholds
    NewDetectorThreshold int `yaml:"new_detector_threshold"`
}
```

4. **Write tests** in `internal/violations/new_detector_test.go`:

```go
func TestNewDetector_Detect(t *testing.T) {
    detector := NewNewDetector(Config{})
    
    testCases := []struct {
        name     string
        input    string
        expected int
    }{
        {
            name:     "no violations",
            input:    "// clean code",
            expected: 0,
        },
        {
            name:     "one violation",
            input:    "// problematic code",
            expected: 1,
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            file := &models.FileInfo{Content: tc.input}
            violations, err := detector.Detect(file)
            
            assert.NoError(t, err)
            assert.Len(t, violations, tc.expected)
        })
    }
}
```

### Adding a New Report Format

1. **Create the reporter** in `internal/reporters/`:

```go
// internal/reporters/new_format.go
package reporters

import "github.com/ericfisherdev/goclean/internal/models"

type NewFormatReporter struct {
    config Config
    output string
}

func NewNewFormatReporter(config Config) *NewFormatReporter {
    return &NewFormatReporter{
        config: config,
        output: config.Output.NewFormat.Path,
    }
}

func (r *NewFormatReporter) Generate(result *models.ScanResult) error {
    // Generation logic here
    return nil
}

func (r *NewFormatReporter) Format() string {
    return "new_format"
}
```

2. **Register the reporter** in `manager.go`:

```go
func (m *Manager) initReporters() {
    if m.config.Output.HTML.Enabled {
        m.reporters = append(m.reporters, NewHTMLReporter(m.config))
    }
    
    // ... other reporters
    
    if m.config.Output.NewFormat.Enabled {
        m.reporters = append(m.reporters, NewNewFormatReporter(m.config))
    }
}
```

### Adding Language Support

1. **Extend the parser** in `internal/scanner/parser.go`:

```go
func (p *Parser) parseFile(file *models.FileInfo) error {
    switch file.Language {
    case "go":
        return p.parseGoFile(file)
    case "javascript":
        return p.parseJavaScriptFile(file)
    // Add new language
    case "rust":
        return p.parseRustFile(file)
    default:
        return fmt.Errorf("unsupported language: %s", file.Language)
    }
}
```

2. **Implement language-specific parsing**:

```go
func (p *Parser) parseRustFile(file *models.FileInfo) error {
    // Rust-specific AST parsing
    return nil
}
```

3. **Update configuration** to include file extensions:

```yaml
scan:
  file_types:
    - ".rs"  # Rust files
```

## Testing Strategy

### Test Types

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test component interactions
3. **End-to-End Tests**: Test complete workflows
4. **Benchmark Tests**: Performance regression detection

### Test Organization

```
package_test.go        # Unit tests for package
integration_test.go    # Integration tests
benchmark_test.go      # Performance benchmarks
testdata/             # Test fixtures and sample files
```

### Writing Good Tests

#### Unit Test Example

```go
func TestFunctionDetector_Detect(t *testing.T) {
    t.Parallel()
    
    tests := []struct {
        name           string
        input          string
        expectedCount  int
        expectedType   string
    }{
        {
            name: "short function",
            input: `
func shortFunc() {
    fmt.Println("hello")
}`,
            expectedCount: 0,
        },
        {
            name: "long function",
            input: generateLongFunction(50), // Helper function
            expectedCount: 1,
            expectedType: "function_length",
        },
    }
    
    detector := NewFunctionDetector(Config{
        Thresholds: ThresholdConfig{
            FunctionLines: 25,
        },
    })
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            file := &models.FileInfo{
                Path:    "test.go",
                Content: tt.input,
            }
            
            violations, err := detector.Detect(file)
            
            require.NoError(t, err)
            assert.Len(t, violations, tt.expectedCount)
            
            if tt.expectedCount > 0 {
                assert.Equal(t, tt.expectedType, violations[0].Type)
            }
        })
    }
}
```

#### Integration Test Example

```go
func TestScanner_Integration(t *testing.T) {
    tempDir := t.TempDir()
    
    // Create test files
    testFiles := map[string]string{
        "good.go": `package main
func good() {
    fmt.Println("clean")
}`,
        "bad.go": generateLongFunction(100),
    }
    
    for name, content := range testFiles {
        path := filepath.Join(tempDir, name)
        require.NoError(t, os.WriteFile(path, []byte(content), 0644))
    }
    
    // Configure scanner
    config := Config{
        Scan: ScanConfig{
            Paths: []string{tempDir},
        },
        Thresholds: ThresholdConfig{
            FunctionLines: 25,
        },
    }
    
    scanner := NewScanner(config)
    result, err := scanner.Scan(context.Background())
    
    require.NoError(t, err)
    assert.Equal(t, 2, result.FilesScanned)
    assert.Equal(t, 1, len(result.Violations))
}
```

#### Benchmark Test Example

```go
func BenchmarkScanner_LargeCodebase(b *testing.B) {
    // Setup large test codebase
    tempDir := setupLargeCodebase(b, 1000) // 1000 files
    defer os.RemoveAll(tempDir)
    
    config := getDefaultConfig()
    config.Scan.Paths = []string{tempDir}
    
    scanner := NewScanner(config)
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        _, err := scanner.Scan(context.Background())
        require.NoError(b, err)
    }
}
```

### Test Utilities

Use the `testutils` package for common test functionality:

```go
// Create temporary test files
files := testutils.CreateTempFiles(t, map[string]string{
    "test.go": "package main\nfunc test() {}",
})

// Generate test violations
violation := testutils.NewViolation("test_type", "test message")

// Create mock file info
fileInfo := testutils.NewFileInfo("test.go", "package main")
```

### Coverage Requirements

- Minimum 80% code coverage for new features
- 100% coverage for critical paths (violation detection)
- Use `make coverage` to generate coverage reports

## Performance Guidelines

### General Principles

1. **Minimize Memory Allocations**: Reuse objects where possible
2. **Concurrent Processing**: Use goroutines for I/O-bound operations
3. **Streaming Processing**: Process large files incrementally
4. **Efficient Data Structures**: Choose appropriate data structures

### Memory Management

```go
// Good: Reuse slices
violations := make([]models.Violation, 0, 10) // Pre-allocate capacity
for _, item := range items {
    if needsProcessing(item) {
        violations = append(violations, processItem(item))
    }
}

// Good: Use object pools for frequently allocated objects
var violationPool = sync.Pool{
    New: func() interface{} {
        return &models.Violation{}
    },
}

func newViolation() *models.Violation {
    v := violationPool.Get().(*models.Violation)
    // Reset fields
    *v = models.Violation{}
    return v
}

func recycleViolation(v *models.Violation) {
    violationPool.Put(v)
}
```

### Concurrent Processing

```go
// Process files concurrently
func (s *Scanner) processFiles(files []string) error {
    const maxConcurrency = 10
    sem := make(chan struct{}, maxConcurrency)
    
    var wg sync.WaitGroup
    errChan := make(chan error, len(files))
    
    for _, file := range files {
        wg.Add(1)
        go func(file string) {
            defer wg.Done()
            sem <- struct{}{} // Acquire semaphore
            defer func() { <-sem }() // Release semaphore
            
            if err := s.processFile(file); err != nil {
                errChan <- err
            }
        }(file)
    }
    
    wg.Wait()
    close(errChan)
    
    for err := range errChan {
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

### Profiling and Optimization

Use Go's built-in profiling tools:

```bash
# CPU profiling
go test -cpuprofile cpu.prof -bench .

# Memory profiling
go test -memprofile mem.prof -bench .

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

Add profiling endpoints for development:

```go
// In main.go for development builds
import _ "net/http/pprof"

func init() {
    if os.Getenv("GOCLEAN_PROFILE") == "true" {
        go func() {
            log.Println(http.ListenAndServe("localhost:6060", nil))
        }()
    }
}
```

## Documentation Standards

### Code Documentation

- All public functions and types must have godoc comments
- Comments should explain "why" not just "what"
- Use examples in documentation when helpful

```go
// Scanner analyzes source code files for clean code violations.
// It processes files concurrently and generates detailed reports
// about code quality issues found during analysis.
//
// Example usage:
//   scanner := NewScanner(config)
//   result, err := scanner.Scan(ctx, []string{"./src"})
//   if err != nil {
//       log.Fatal(err)
//   }
//   fmt.Printf("Found %d violations\n", len(result.Violations))
type Scanner struct {
    config    Config
    detectors []Detector
}
```

### README Updates

Keep the main README.md updated with:
- Current feature status
- Installation instructions
- Basic usage examples
- Links to detailed documentation

### Changelog

Follow semantic versioning and maintain CHANGELOG.md:

```markdown
## [1.2.0] - 2024-03-15

### Added
- Support for TypeScript files
- Custom violation severity configuration
- Export to CSV format

### Changed
- Improved HTML report performance
- Updated Go version requirement to 1.21

### Fixed
- File path handling on Windows
- Memory leak in large file processing

### Deprecated
- Old configuration format (will be removed in 2.0.0)
```

## Release Process

### Version Management

Use semantic versioning (MAJOR.MINOR.PATCH):
- **MAJOR**: Breaking changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

### Release Checklist

1. **Update version** in relevant files
2. **Update CHANGELOG.md** with release notes
3. **Run full test suite** and benchmarks
4. **Update documentation** if needed
5. **Create release branch** from develop
6. **Tag the release** with version number
7. **Build and publish** artifacts
8. **Update package managers** (if applicable)

### Automated Release Process

GitHub Actions workflow (`.github/workflows/release.yml`):

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v3
        with:
          go-version: 1.21
      
      - name: Run tests
        run: make test
      
      - name: Build binaries
        run: make build-all
      
      - name: Create Release
        uses: actions/create-release@v1
        with:
          tag_name: ${{ github.ref }}
          release_name: Release ${{ github.ref }}
          draft: false
          prerelease: false
```

### Post-Release Tasks

1. **Merge release branch** back to main and develop
2. **Update documentation** website
3. **Announce release** on relevant channels
4. **Monitor for issues** and prepare hotfixes if needed

## Contributing Workflow

### Before Starting

1. Check existing issues for similar work
2. Create or comment on relevant issue
3. Fork the repository
4. Create feature branch from develop

### During Development

1. Follow coding standards and guidelines
2. Write tests for new functionality
3. Update documentation as needed
4. Ensure all tests pass locally

### Submitting Changes

1. **Create pull request** against develop branch
2. **Include clear description** of changes
3. **Reference related issues** with keywords (fixes #123)
4. **Ensure CI passes** all checks
5. **Respond to review feedback** promptly

### Code Review Process

All changes require review:
- **Code quality**: Follows standards and best practices
- **Testing**: Adequate test coverage
- **Documentation**: Updated and clear
- **Performance**: No significant regressions
- **Security**: No vulnerabilities introduced