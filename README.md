# GoClean

A Go CLI tool that scans codebases to identify clean code violations with real-time HTML reporting and AI-friendly markdown output.

## Features

- **Clean Code Analysis**: Detect violations including long functions, complex code, naming issues, and documentation problems
- **Real-time HTML Reports**: Auto-refreshing dashboard with modern UI
- **AI-Friendly Output**: Structured markdown reports for AI analysis tools like Claude Code
- **Multi-language Support**: Go, JavaScript, TypeScript, Python, Java, C# (planned)
- **Configurable Rules**: Customizable thresholds and violation detection

## Installation

```bash
go install github.com/ericfisherdev/goclean/cmd/goclean@latest
```

## Quick Start

```bash
# Scan current directory
goclean scan

# Scan specific path with custom config
goclean scan --path ./src --config ./goclean.yaml

# Generate both HTML and markdown reports
goclean scan --html --markdown
```

## Project Status

🚧 **Under Development** - Currently in Phase 1 of development

### Development Progress

- [x] Project structure setup with Go modules
- [ ] CLI framework implementation
- [ ] Basic file scanning and parsing
- [ ] Configuration system
- [ ] Unit test framework

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
└── docs/                 # Documentation
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.