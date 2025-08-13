# GoClean Installation Guide

This guide provides comprehensive installation instructions for GoClean across different platforms and deployment scenarios.

## Table of Contents

- [System Requirements](#system-requirements)
- [Installation Methods](#installation-methods)
  - [Binary Releases (Recommended)](#binary-releases-recommended)
  - [From Source](#from-source)
  - [Package Managers](#package-managers)
- [Verification](#verification)
- [Configuration](#configuration)
- [Integration](#integration)
- [Troubleshooting](#troubleshooting)

## System Requirements

### Minimum Requirements
- **OS**: Linux, macOS, or Windows
- **Architecture**: amd64 (Intel/AMD 64-bit) or arm64 (Apple Silicon)
- **Memory**: 100MB available RAM
- **Disk Space**: 50MB for installation, additional space for reports

### Recommended Requirements
- **Memory**: 1GB+ for large codebases (>10k files)
- **CPU**: Multi-core processor for concurrent processing
- **Disk Space**: 500MB+ for comprehensive analysis reports

### Supported File Types
- Go (`.go`)
- JavaScript (`.js`, `.mjs`)
- TypeScript (`.ts`, `.tsx`)
- Python (`.py`)
- Java (`.java`)
- C# (`.cs`)

## Installation Methods

### Binary Releases (Recommended)

Download pre-built binaries from the [GitHub releases page](https://github.com/ericfisherdev/GoClean/releases).

#### Linux (x86_64)
```bash
# Download and install
wget https://github.com/ericfisherdev/GoClean/releases/download/v0.1.0/goclean-0.1.0-linux-amd64.tar.gz
tar -xzf goclean-0.1.0-linux-amd64.tar.gz
sudo mv goclean-linux-amd64 /usr/local/bin/goclean
sudo chmod +x /usr/local/bin/goclean

# Verify installation
goclean version
```

#### macOS (Intel)
```bash
# Download and install
wget https://github.com/ericfisherdev/GoClean/releases/download/v0.1.0/goclean-0.1.0-darwin-amd64.tar.gz
tar -xzf goclean-0.1.0-darwin-amd64.tar.gz
sudo mv goclean-darwin-amd64 /usr/local/bin/goclean
sudo chmod +x /usr/local/bin/goclean

# Verify installation
goclean version
```

#### macOS (Apple Silicon)
```bash
# Download and install
wget https://github.com/ericfisherdev/GoClean/releases/download/v0.1.0/goclean-0.1.0-darwin-arm64.tar.gz
tar -xzf goclean-0.1.0-darwin-arm64.tar.gz
sudo mv goclean-darwin-arm64 /usr/local/bin/goclean
sudo chmod +x /usr/local/bin/goclean

# Verify installation
goclean version
```

#### Windows
1. Download `goclean-0.1.0-windows-amd64.zip`
2. Extract the ZIP file
3. Rename `goclean-windows-amd64.exe` to `goclean.exe`
4. Add the directory containing `goclean.exe` to your PATH environment variable
5. Open a new command prompt and verify: `goclean version`

### From Source

Building from source requires Go 1.21 or later.

```bash
# Clone repository
git clone https://github.com/ericfisherdev/GoClean.git
cd GoClean

# Build and install
make build
sudo cp bin/goclean /usr/local/bin/

# Or install directly with Go
go install ./cmd/goclean
```

### Package Managers

#### Homebrew (macOS/Linux)
```bash
# Coming soon - not yet available
# brew install ericfisherdev/tap/goclean
```

#### Snap (Linux)
```bash
# Coming soon - not yet available
# sudo snap install goclean
```

#### Chocolatey (Windows)
```powershell
# Coming soon - not yet available
# choco install goclean
```

#### Go Install
```bash
go install github.com/ericfisherdev/goclean/cmd/goclean@latest
```

## Verification

After installation, verify GoClean is working correctly:

```bash
# Check version
goclean version

# Display help
goclean --help

# Test with a quick scan
goclean scan --help
```

Expected output should show version information and available commands.

## Configuration

### Initialize Default Configuration
Create a default configuration file in your project:

```bash
cd /path/to/your/project
goclean config init
```

This creates a `goclean.yaml` file with default settings that you can customize.

### Configuration File Location
GoClean looks for configuration files in this order:
1. `--config` flag value
2. `goclean.yaml` in current directory
3. `$HOME/.goclean/config.yaml`
4. Built-in defaults

### Environment Variables
- `GOCLEAN_CONFIG` - Path to configuration file
- `GOCLEAN_VERBOSE` - Enable verbose output (true/false)
- `GOCLEAN_OUTPUT_PATH` - Default output path for reports

## Integration

### CI/CD Pipeline Integration

#### GitHub Actions
```yaml
name: Code Quality Check
on: [push, pull_request]
jobs:
  goclean:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - name: Download GoClean
      run: |
        wget https://github.com/ericfisherdev/GoClean/releases/download/v0.1.0/goclean-0.1.0-linux-amd64.tar.gz
        tar -xzf goclean-0.1.0-linux-amd64.tar.gz
        chmod +x goclean-linux-amd64
        sudo mv goclean-linux-amd64 /usr/local/bin/goclean
    - name: Run GoClean
      run: goclean scan . --format markdown --output goclean-report.md
    - name: Upload Report
      uses: actions/upload-artifact@v3
      with:
        name: goclean-report
        path: goclean-report.md
```

#### Jenkins
```groovy
pipeline {
    agent any
    stages {
        stage('Code Quality') {
            steps {
                sh 'goclean scan . --format html --output reports/code-quality.html'
                publishHTML([
                    allowMissing: false,
                    alwaysLinkToLastBuild: true,
                    keepAll: true,
                    reportDir: 'reports',
                    reportFiles: 'code-quality.html',
                    reportName: 'GoClean Report'
                ])
            }
        }
    }
}
```

#### GitLab CI
```yaml
goclean:
  stage: quality
  script:
    - wget https://github.com/ericfisherdev/GoClean/releases/download/v0.1.0/goclean-0.1.0-linux-amd64.tar.gz
    - tar -xzf goclean-0.1.0-linux-amd64.tar.gz
    - chmod +x goclean-linux-amd64
    - ./goclean-linux-amd64 scan . --format html --output goclean-report.html
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: goclean-report.html
    paths:
      - goclean-report.html
```

### Pre-commit Hooks
```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: goclean
        name: GoClean
        entry: goclean scan
        language: system
        args: ['.', '--format', 'console']
        types: [text]
```

### IDE Integration

#### VS Code
Add to your project's `.vscode/tasks.json`:
```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "GoClean Analysis",
            "type": "shell",
            "command": "goclean",
            "args": ["scan", ".", "--format", "html", "--output", "goclean-report.html"],
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            }
        }
    ]
}
```

## Troubleshooting

### Common Issues

#### Permission Denied
```bash
# On macOS/Linux, ensure executable permissions
chmod +x /usr/local/bin/goclean

# On macOS, you might need to allow the app in Security & Privacy
sudo xattr -d com.apple.quarantine /usr/local/bin/goclean
```

#### Command Not Found
```bash
# Check if the binary is in your PATH
echo $PATH

# Add to PATH if needed (bash/zsh)
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# For Windows, add to system PATH through Environment Variables
```

#### Scan Fails
```bash
# Enable verbose output for debugging
goclean scan . --verbose

# Check file permissions
ls -la /path/to/your/code

# Ensure GoClean has read access to files
```

#### Out of Memory
```bash
# For large codebases, increase available memory
export GOMEMLIMIT=2GiB
goclean scan .

# Or scan smaller directories at a time
goclean scan ./src --exclude vendor/,node_modules/
```

#### Report Generation Fails
```bash
# Check output directory permissions
mkdir -p reports
chmod 755 reports

# Use absolute paths
goclean scan . --output /full/path/to/report.html
```

### Getting Help

- **Documentation**: [User Guide](docs/user-guide.md)
- **Issues**: [GitHub Issues](https://github.com/ericfisherdev/GoClean/issues)
- **Discussions**: [GitHub Discussions](https://github.com/ericfisherdev/GoClean/discussions)

### Support Information

When reporting issues, please include:
- GoClean version (`goclean version`)
- Operating system and architecture
- Go version (if building from source)
- Command that failed
- Full error output with `--verbose` flag

## Next Steps

After installation:
1. Initialize configuration: `goclean config init`
2. Run your first scan: `goclean scan .`
3. Review the generated report
4. Customize configuration for your needs
5. Integrate into your development workflow

For detailed usage instructions, see the [User Guide](docs/user-guide.md).