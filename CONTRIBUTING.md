# Contributing to GoClean

Thank you for your interest in contributing to GoClean! This document provides guidelines for contributing to the project.

## Development Workflow

GoClean uses a Git workflow with two main branches:
- **`develop`** (default): Main development branch
- **`release`**: Production releases only

See [.github/WORKFLOW.md](./.github/WORKFLOW.md) for detailed workflow documentation.

## Quick Start

1. **Fork and Clone**
   ```bash
   git clone https://github.com/your-username/GoClean.git
   cd GoClean
   ```

2. **Setup Development Environment**
   ```bash
   make deps
   make test
   make build
   ```

3. **Create Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

4. **Make Changes and Test**
   ```bash
   # Make your changes
   make test        # Run tests
   make lint        # Run linter (if available)
   make build       # Build application
   ```

5. **Create Pull Request**
   ```bash
   git push -u origin feature/your-feature-name
   gh pr create --base develop
   ```

## Code Standards

### Go Code Style
- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Add comments for public functions and complex logic
- Follow the existing code structure and patterns

### Testing Requirements
- All new features must include unit tests
- Maintain or improve test coverage
- Tests should be deterministic and fast
- Use table-driven tests where appropriate

### Commit Messages
Use conventional commit format:
```
type(scope): short description

Longer description if needed

Closes #123
```

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`

## Development Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the application |
| `make test` | Run all tests |
| `make test-coverage` | Run tests with coverage report |
| `make clean` | Clean build artifacts |
| `make deps` | Install dependencies |
| `make lint` | Run linter |
| `make fmt` | Format code |
| `make run` | Build and run application |

## Project Structure

```
GoClean/
├── cmd/goclean/          # CLI application entry point
├── internal/             # Internal packages
│   ├── config/          # Configuration management
│   ├── models/          # Data models
│   ├── scanner/         # File scanning and parsing
│   └── violations/      # Violation detection (future)
├── configs/             # Default configuration files
├── .github/            # GitHub templates and workflows
└── docs/               # Documentation
```

## Types of Contributions

### 🐛 Bug Reports
- Use GitHub Issues with bug report template
- Include steps to reproduce
- Provide expected vs actual behavior
- Include system information and versions

### ✨ Feature Requests
- Use GitHub Issues with feature request template
- Describe the problem you're trying to solve
- Provide examples of desired behavior
- Consider backward compatibility

### 🔧 Code Contributions
- Start with an issue discussion for large changes
- Follow the development workflow
- Include tests for new functionality
- Update documentation as needed

### 📚 Documentation
- Improve existing documentation
- Add examples and use cases
- Fix typos and clarify confusing sections

## Review Process

1. **Automated Checks**: CI runs tests and linting
2. **Code Review**: At least one team member reviews
3. **Testing**: Verify functionality works as expected
4. **Documentation**: Ensure docs are updated if needed
5. **Merge**: Squash and merge to maintain clean history

## Getting Help

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Code Review**: Ask questions in PR comments

## Code of Conduct

Be respectful and inclusive:
- Use welcoming and inclusive language
- Be respectful of differing viewpoints
- Accept constructive criticism gracefully
- Focus on what is best for the community

## Recognition

Contributors are recognized in:
- GitHub contributor graphs
- Release notes for significant contributions
- Special thanks in documentation

Thank you for contributing to GoClean! 🎉