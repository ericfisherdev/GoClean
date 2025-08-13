# GoClean

A powerful Go CLI tool that scans codebases to identify clean code violations with real-time HTML reporting and AI-friendly markdown output.

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-success.svg)](#)

## Features

### 🔍 Comprehensive Analysis
- **Function Quality**: Detect long functions, high complexity, excessive parameters
- **Naming Conventions**: Identify non-descriptive names and inconsistent patterns  
- **Code Structure**: Find large classes, deep nesting, and code duplication
- **Documentation**: Track missing docs, outdated comments, and technical debt

### 📊 Rich Reporting
- **Interactive HTML Dashboard**: Real-time auto-refreshing reports with modern UI
- **AI-Friendly Markdown**: Structured output optimized for AI analysis tools
- **Multiple Formats**: Console tables, JSON, CSV exports
- **Detailed Insights**: Code snippets, severity indicators, improvement suggestions

### ⚙️ Highly Configurable  
- **Custom Thresholds**: Adjust limits for functions, complexity, naming rules
- **Language Support**: Go, JavaScript, TypeScript, Python, Java, C#
- **Flexible Output**: Choose report formats and customize styling
- **Team Standards**: Version-controlled configuration for consistent rules

## Installation

### Pre-built Binaries
Download from [GitHub Releases](https://github.com/ericfisherdev/goclean/releases):

```bash
# Linux/macOS
curl -L https://github.com/ericfisherdev/goclean/releases/latest/download/goclean-linux-amd64 -o goclean
chmod +x goclean && sudo mv goclean /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/ericfisherdev/goclean/releases/latest/download/goclean-windows-amd64.exe" -OutFile "goclean.exe"
```

### Go Install
```bash
go install github.com/ericfisherdev/goclean/cmd/goclean@latest
```

### Build from Source
```bash
git clone https://github.com/ericfisherdev/goclean.git
cd goclean && make build
```

## Quick Start

```bash
# Scan current directory with defaults
goclean scan

# Scan specific paths
goclean scan --path ./src --path ./internal

# Generate configuration file
goclean config init

# Custom scan with configuration  
goclean scan --config ./goclean.yaml --html --markdown

# Filter by severity
goclean scan --severity high --console

# CI/CD friendly output
goclean scan --format json --no-color --quiet
```

## Project Status

✅ **Production Ready** - All core features implemented and tested

### Development Progress

- [x] **Phase 1**: Foundation (Project structure, CLI framework, configuration)
- [x] **Phase 2**: Core Detection (AST parsing, violation detection, severity classification)
- [x] **Phase 3**: Reporting System (HTML templates, real-time updates, file navigation)
- [x] **Phase 4**: Enhanced Features (Markdown output, advanced detection, performance optimization)
- [x] **Phase 5**: Polish and Documentation (UI improvements, comprehensive docs, testing)

## Documentation

Comprehensive documentation is available in the `docs/` directory:

| Document | Description |
|----------|-------------|
| 📖 [User Guide](docs/user-guide.md) | Complete installation, configuration, and usage guide |
| ⚙️ [Configuration Reference](docs/configuration.md) | Detailed configuration options and examples |
| 🛠️ [Developer Guide](docs/developer-guide.md) | Contributing guidelines and architecture overview |
| 📚 [API Reference](docs/api-reference.md) | Programmatic API documentation for integrations |

### Quick Links

- [Installation Instructions](docs/user-guide.md#installation)
- [Configuration Examples](docs/configuration.md#configuration-examples)
- [Adding Custom Detectors](docs/developer-guide.md#adding-new-features)
- [CI/CD Integration](docs/api-reference.md#integration-examples)

## Architecture

```
GoClean/
├── cmd/goclean/          # CLI entry point
├── internal/             # Core application logic
│   ├── config/           # Configuration management
│   ├── scanner/          # File scanning engine
│   ├── violations/       # Violation detection
│   ├── reporters/        # Report generation
│   └── models/           # Data structures
├── web/                  # HTML templates and assets
├── configs/              # Default configuration
└── docs/                 # Comprehensive documentation
```

## Examples

### Basic Usage
```bash
# Initialize configuration
goclean config init

# Run analysis
goclean scan

# View HTML report
open reports/clean-code-report.html
```

### Advanced Configuration
```yaml
# goclean.yaml
scan:
  paths: ["./src", "./internal"]
  exclude: ["vendor/", "*.test.go"]

thresholds:
  function_lines: 20
  cyclomatic_complexity: 5
  parameters: 3

output:
  html:
    theme: "dark"
    auto_refresh: true
  markdown:
    enabled: true
    ai_friendly: true
```

### Programmatic Usage
```go
package main

import (
    "context"
    "github.com/ericfisherdev/goclean/pkg/goclean"
    "github.com/ericfisherdev/goclean/pkg/config"
)

func main() {
    cfg, _ := config.Load("goclean.yaml")
    analyzer := goclean.New(cfg)
    result, _ := analyzer.Analyze(context.Background(), []string{"./src"})
    
    fmt.Printf("Found %d violations\n", len(result.Violations))
}
```

## Use Cases

- **Code Reviews**: Identify issues before merging
- **CI/CD Pipelines**: Enforce quality gates
- **Refactoring**: Find improvement opportunities  
- **Team Standards**: Maintain consistent code quality
- **Technical Debt**: Track and prioritize fixes
- **Learning**: Understand clean code principles

## Contributing

We welcome contributions! Please see our [Developer Guide](docs/developer-guide.md) for:

- Development setup and workflow
- Architecture overview and coding standards  
- Adding new features and detectors
- Testing strategies and requirements
- Documentation standards

### Quick Start for Contributors

```bash
# Fork and clone the repository
git clone https://github.com/your-username/goclean.git
cd goclean

# Set up development environment
go mod download
make build
make test

# Create feature branch
git checkout -b feature/your-feature

# Make changes and test
make test
make lint

# Submit pull request
```

## Support

- 📝 [Documentation](docs/)
- 🐛 [Bug Reports](https://github.com/ericfisherdev/goclean/issues)
- 💬 [Discussions](https://github.com/ericfisherdev/goclean/discussions)
- 📧 [Email Support](mailto:ericfisherdev@example.com)

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

<div align="center">
Made with ❤️ by <a href="https://github.com/ericfisherdev">Eric Fisher</a>
</div>