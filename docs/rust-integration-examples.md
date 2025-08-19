# GoClean Rust Integration Examples

This document provides practical examples for integrating GoClean with Rust projects across different scenarios and toolchains.

## Table of Contents

1. [Project Setup Examples](#project-setup-examples)
2. [Cargo Integration](#cargo-integration)
3. [IDE Integration](#ide-integration)
4. [CI/CD Pipeline Examples](#cicd-pipeline-examples)
5. [Docker Integration](#docker-integration)
6. [Pre-commit Hooks](#pre-commit-hooks)
7. [Workspace Management](#workspace-management)
8. [Real-world Project Examples](#real-world-project-examples)

## Project Setup Examples

### Single Rust Crate

**Project Structure:**
```
my-rust-project/
├── Cargo.toml
├── src/
│   ├── main.rs
│   ├── lib.rs
│   └── modules/
├── tests/
├── benches/
├── examples/
├── goclean.yaml
└── .gitignore
```

**GoClean Configuration (goclean.yaml):**
```yaml
scan:
  paths:
    - ./src
    - ./examples
  file_types:
    - .rs
  exclude:
    - target/
    - Cargo.lock
    - "*.bak"

rust:
  enable_ownership_analysis: true
  enable_error_handling_check: true
  allow_unwrap: false
  enforce_snake_case: true

clippy:
  enabled: true
  categories: [correctness, suspicious, style, complexity, perf]

output:
  html: { enabled: true, path: ./reports/quality-report.html }
  markdown: { enabled: true, path: ./reports/violations.md }
  console: { enabled: true, format: table }

thresholds:
  function_lines: 30
  cyclomatic_complexity: 10
  parameters: 5
```

**Setup Commands:**
```bash
# Initialize GoClean in Rust project
cd my-rust-project
goclean config init --rust
goclean scan

# Add to Cargo.toml for easy access
echo -e "\n[package.metadata.goclean]\nconfig = \"goclean.yaml\"" >> Cargo.toml
```

### Rust Workspace

**Project Structure:**
```
rust-workspace/
├── Cargo.toml (workspace)
├── crates/
│   ├── core/
│   │   ├── Cargo.toml
│   │   └── src/
│   ├── api/
│   │   ├── Cargo.toml
│   │   └── src/
│   └── utils/
│       ├── Cargo.toml
│       └── src/
├── goclean.yaml
└── tools/
    └── quality-check.sh
```

**Workspace Configuration:**
```yaml
scan:
  paths:
    - ./crates/*/src
    - ./crates/*/benches
  file_types:
    - .rs
  exclude:
    - target/
    - "**/target/"
    - Cargo.lock
    - "**/Cargo.lock"

rust:
  enable_ownership_analysis: true
  enable_error_handling_check: true
  max_trait_bounds: 5
  max_impl_methods: 20

clippy:
  enabled: true
  categories: [correctness, suspicious, style, complexity, perf]

output:
  html: 
    enabled: true
    path: ./reports/workspace-quality.html
    group_by_crate: true
  markdown:
    enabled: true
    path: ./reports/workspace-violations.md

# Per-crate thresholds
crates:
  core:
    thresholds:
      function_lines: 25
      cyclomatic_complexity: 8
  api:
    thresholds:
      function_lines: 40  # API handlers can be longer
      parameters: 8
  utils:
    thresholds:
      function_lines: 20  # Utilities should be simple
```

**Quality Check Script (tools/quality-check.sh):**
```bash
#!/bin/bash
set -e

echo "🔍 Running GoClean workspace analysis..."

# Run GoClean on the entire workspace
goclean scan --config goclean.yaml

# Generate per-crate reports
for crate in crates/*/; do
    crate_name=$(basename "$crate")
    echo "📊 Analyzing crate: $crate_name"
    
    goclean scan \
        --path "$crate/src" \
        --file-types .rs \
        --format html \
        --output "reports/${crate_name}-report.html"
done

echo "✅ Quality analysis complete. Check ./reports/ for results."
```

### Library Crate

**Configuration for a Rust library:**
```yaml
scan:
  paths:
    - ./src
  file_types:
    - .rs
  exclude:
    - target/
    - examples/
    - benches/

rust:
  # Strict library standards
  enable_ownership_analysis: true
  enable_error_handling_check: true
  allow_unwrap: false
  allow_expect: false
  enforce_result_propagation: true
  
  # Documentation requirements for public APIs
  require_public_docs: true
  min_doc_coverage: 0.9

clippy:
  enabled: true
  categories: [correctness, suspicious, style, complexity, perf, pedantic]

thresholds:
  function_lines: 25
  cyclomatic_complexity: 6
  parameters: 4

output:
  markdown:
    enabled: true
    path: ./docs/QUALITY_REPORT.md
    include_examples: true
```

## Cargo Integration

### Custom Cargo Commands

**Install as cargo subcommand:**
```bash
# Create cargo-goclean wrapper
cat > ~/.cargo/bin/cargo-goclean << 'EOF'
#!/bin/bash
# Cargo subcommand for GoClean
shift  # Remove 'goclean' from args
exec goclean "$@"
EOF

chmod +x ~/.cargo/bin/cargo-goclean

# Usage
cargo goclean scan
cargo goclean config init --rust
```

### Cargo.toml Integration

**Add GoClean metadata to Cargo.toml:**
```toml
[package]
name = "my-crate"
version = "0.1.0"
edition = "2021"

# GoClean configuration
[package.metadata.goclean]
config = "goclean.yaml"
strict_mode = true
pre_publish_check = true

[package.metadata.goclean.thresholds]
function_lines = 25
cyclomatic_complexity = 8
parameters = 4

[package.metadata.goclean.clippy]
enabled = true
categories = ["correctness", "suspicious", "style", "complexity", "perf"]
```

### Build Script Integration

**build.rs:**
```rust
use std::process::Command;

fn main() {
    // Run GoClean during build if in development
    if cfg!(debug_assertions) {
        println!("cargo:rerun-if-changed=src/");
        
        let output = Command::new("goclean")
            .args(&["scan", "--format", "json", "--quiet"])
            .output();
            
        match output {
            Ok(output) if output.status.success() => {
                println!("cargo:warning=GoClean analysis passed");
            }
            Ok(_) => {
                println!("cargo:warning=GoClean found code quality issues");
            }
            Err(_) => {
                println!("cargo:warning=GoClean not available");
            }
        }
    }
}
```

### Make Integration

**Makefile:**
```makefile
.PHONY: quality check test build clean

# Quality checks
quality:
	@echo "Running GoClean analysis..."
	goclean scan --config goclean.yaml

# Pre-commit checks
check: quality test
	@echo "All checks passed!"

# Development workflow
dev-check:
	goclean scan --config configs/rust-minimal.yaml --console

# CI workflow
ci-check:
	goclean scan --config configs/rust-ci.yaml --format json --fail-on critical

# Release preparation
release-check:
	goclean scan --config configs/rust-strict.yaml --html --markdown

# Clean reports
clean-reports:
	rm -rf reports/
	mkdir -p reports

# Full development cycle
dev: clean-reports dev-check test
```

## IDE Integration

### VS Code Integration

**.vscode/settings.json:**
```json
{
    "rust-analyzer.cargo.features": "all",
    "rust-analyzer.checkOnSave.command": "clippy",
    "files.associations": {
        "goclean.yaml": "yaml",
        "*.goclean.yaml": "yaml"
    },
    "yaml.schemas": {
        "https://raw.githubusercontent.com/ericfisherdev/goclean/main/schema/goclean-schema.json": [
            "goclean.yaml",
            "*.goclean.yaml"
        ]
    }
}
```

**.vscode/tasks.json:**
```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "GoClean: Scan",
            "type": "shell",
            "command": "goclean",
            "args": ["scan", "--console"],
            "group": "test",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            },
            "problemMatcher": {
                "owner": "goclean",
                "fileLocation": ["relative", "${workspaceFolder}"],
                "pattern": {
                    "regexp": "^(.*):(\\d+):(\\d+):\\s+(warning|error|info):\\s+(.*)$",
                    "file": 1,
                    "line": 2,
                    "column": 3,
                    "severity": 4,
                    "message": 5
                }
            }
        },
        {
            "label": "GoClean: Generate Report",
            "type": "shell",
            "command": "goclean",
            "args": ["scan", "--html", "--markdown"],
            "group": "build"
        }
    ]
}
```

### IntelliJ/CLion Integration

**Run Configurations:**

1. **GoClean Scan Configuration:**
   - Program: `goclean`
   - Arguments: `scan --console`
   - Working directory: `$ProjectFileDir$`

2. **GoClean Report Configuration:**
   - Program: `goclean`
   - Arguments: `scan --html --auto-refresh`
   - Working directory: `$ProjectFileDir$`

**External Tools:**
```xml
<tool name="GoClean Scan" description="Run GoClean analysis" showInMainMenu="true" showInEditor="true" showInProject="true" showInSearchPopup="true" disabled="false" useConsole="true" showConsoleOnStdOut="false" showConsoleOnStdErr="false" synchronizeAfterRun="true">
  <exec>
    <option name="COMMAND" value="goclean" />
    <option name="PARAMETERS" value="scan --console" />
    <option name="WORKING_DIRECTORY" value="$ProjectFileDir$" />
  </exec>
</tool>
```

## CI/CD Pipeline Examples

### GitHub Actions

**.github/workflows/rust-quality.yml:**
```yaml
name: Rust Code Quality

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

permissions:
  contents: read
  checks: write

jobs:
  quality:
    name: Code Quality Analysis
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Install Rust toolchain
      uses: dtolnay/rust-toolchain@stable
      with:
        components: clippy

    - name: Install GoClean
      run: |
        curl -L https://github.com/ericfisherdev/goclean/releases/latest/download/goclean-linux-amd64 -o goclean
        chmod +x goclean
        sudo mv goclean /usr/local/bin/

    - name: Cache dependencies
      uses: actions/cache@v4
      with:
        path: |
          ~/.cargo/registry
          ~/.cargo/git
          target
        key: ${{ runner.os }}-cargo-${{ hashFiles('**/Cargo.lock') }}

    - name: Run GoClean analysis
      run: |
        goclean scan \
          --config configs/rust-ci.yaml \
          --format json \
          --output goclean-results.json

    - name: Generate quality report
      run: |
        goclean scan \
          --config configs/rust-ci.yaml \
          --html \
          --html-output quality-report.html \
          --markdown \
          --markdown-output quality-report.md

    - name: Upload quality report
      uses: actions/upload-artifact@v4
      with:
        name: quality-reports
        path: |
          quality-report.html
          quality-report.md
          goclean-results.json

    - name: Comment PR with quality report
      if: github.event_name == 'pull_request'
      uses: actions/github-script@v7
      with:
        script: |
          const fs = require('fs');
          const report = fs.readFileSync('quality-report.md', 'utf8');
          
          github.rest.issues.createComment({
            issue_number: context.issue.number,
            owner: context.repo.owner,
            repo: context.repo.repo,
            body: `## 📊 Code Quality Report\n\n${report}`
          });

    - name: Check quality gate
      run: |
        # Fail if critical violations found
        if jq -e '.violations[] | select(.severity == "critical")' goclean-results.json > /dev/null; then
          echo "❌ Critical violations found"
          exit 1
        else
          echo "✅ Quality gate passed"
        fi
```

### GitLab CI

**.gitlab-ci.yml:**
```yaml
stages:
  - quality
  - build
  - test

variables:
  CARGO_HOME: $CI_PROJECT_DIR/.cargo
  GOCLEAN_VERSION: "latest"

cache:
  paths:
    - .cargo/
    - target/

rust_quality:
  stage: quality
  image: rust:latest
  before_script:
    - rustup component add clippy
    - curl -L https://github.com/ericfisherdev/goclean/releases/latest/download/goclean-linux-amd64 -o /usr/local/bin/goclean
    - chmod +x /usr/local/bin/goclean
  script:
    - goclean scan --config configs/rust-ci.yaml --format json --export-json --json-output gl-code-quality-report.json
    - goclean scan --config configs/rust-ci.yaml --html --html-output quality-report.html
  artifacts:
    reports:
      codequality: gl-code-quality-report.json
    paths:
      - quality-report.html
    expire_in: 1 week
  rules:
    - if: $CI_COMMIT_BRANCH
    - if: $CI_MERGE_REQUEST_ID

quality_gate:
  stage: quality
  image: rust:latest
  needs: ["rust_quality"]
  script:
    - |
      if jq -e '.violations[] | select(.severity == "critical")' gl-code-quality-report.json > /dev/null; then
        echo "Critical violations found"
        exit 1
      fi
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
    - if: $CI_MERGE_REQUEST_TARGET_BRANCH_NAME == "main"
```

### Jenkins Pipeline

**Jenkinsfile:**
```groovy
pipeline {
    agent any
    
    tools {
        // Assume GoClean is available as a tool
        'goclean' 'latest'
    }
    
    environment {
        CARGO_HOME = "$WORKSPACE/.cargo"
        RUSTUP_HOME = "$WORKSPACE/.rustup"
    }
    
    stages {
        stage('Setup') {
            steps {
                sh '''
                    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
                    source $CARGO_HOME/env
                    rustup component add clippy
                '''
            }
        }
        
        stage('Quality Analysis') {
            steps {
                sh '''
                    source $CARGO_HOME/env
                    goclean scan \
                        --config configs/rust-ci.yaml \
                        --format json \
                        --export-json \
                        --json-output target/goclean-report.json \
                        --html \
                        --html-output target/quality-report.html
                '''
            }
            
            post {
                always {
                    publishHTML([
                        allowMissing: false,
                        alwaysLinkToLastBuild: true,
                        keepAll: true,
                        reportDir: 'target',
                        reportFiles: 'quality-report.html',
                        reportName: 'GoClean Quality Report'
                    ])
                    
                    archiveArtifacts artifacts: 'target/goclean-report.json', fingerprint: true
                }
            }
        }
        
        stage('Quality Gate') {
            steps {
                script {
                    def report = readJSON file: 'target/goclean-report.json'
                    def criticalViolations = report.violations.findAll { it.severity == 'critical' }
                    
                    if (criticalViolations.size() > 0) {
                        error("Found ${criticalViolations.size()} critical violations")
                    }
                }
            }
        }
    }
}
```

## Docker Integration

### Dockerfile for GoClean + Rust

```dockerfile
FROM rust:1.70 as builder

# Install GoClean
RUN curl -L https://github.com/ericfisherdev/goclean/releases/latest/download/goclean-linux-amd64 -o /usr/local/bin/goclean \
    && chmod +x /usr/local/bin/goclean

# Install clippy
RUN rustup component add clippy

# Set working directory
WORKDIR /app

# Copy configuration
COPY goclean.yaml .

# Copy source code
COPY . .

# Run quality analysis
RUN goclean scan --config goclean.yaml --format json --export-json --json-output quality-report.json

# Build the application
RUN cargo build --release

# Runtime stage
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/target/release/my-app /usr/local/bin/my-app
COPY --from=builder /app/quality-report.json /quality-report.json

CMD ["my-app"]
```

### Docker Compose for Development

**docker-compose.yml:**
```yaml
version: '3.8'

services:
  rust-dev:
    build:
      context: .
      dockerfile: Dockerfile.dev
    volumes:
      - .:/workspace
      - cargo-cache:/usr/local/cargo/registry
    working_dir: /workspace
    command: tail -f /dev/null
    environment:
      - RUST_LOG=debug

  goclean-analysis:
    image: goclean:latest
    volumes:
      - .:/workspace
    working_dir: /workspace
    command: >
      sh -c "
        goclean scan --config goclean.yaml --format html --output /workspace/reports/quality-report.html &&
        goclean scan --config goclean.yaml --format json --output /workspace/reports/quality-data.json
      "
    depends_on:
      - rust-dev

volumes:
  cargo-cache:
```

## Pre-commit Hooks

### Using pre-commit framework

**.pre-commit-config.yaml:**
```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.4.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml

  - repo: https://github.com/doublify/pre-commit-rust
    rev: v1.0
    hooks:
      - id: fmt
      - id: cargo-check
      - id: clippy

  - repo: local
    hooks:
      - id: goclean
        name: GoClean Quality Check
        entry: goclean
        args: ['scan', '--config', 'configs/rust-minimal.yaml', '--console-violations']
        language: system
        files: '\.rs$'
        pass_filenames: false
```

### Custom Git Hooks

**pre-commit hook (.git/hooks/pre-commit):**
```bash
#!/bin/bash
set -e

echo "🔍 Running GoClean pre-commit analysis..."

# Get list of staged Rust files
STAGED_RUST_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.rs$' || true)

if [ -z "$STAGED_RUST_FILES" ]; then
    echo "No Rust files to check"
    exit 0
fi

# Run GoClean on staged files
goclean scan \
    --config configs/rust-minimal.yaml \
    --console \
    --format table \
    --severity critical,high

# Check exit code
if [ $? -ne 0 ]; then
    echo "❌ GoClean found critical issues. Commit blocked."
    echo "Fix the issues above or use 'git commit --no-verify' to skip checks."
    exit 1
fi

echo "✅ GoClean pre-commit checks passed"
```

**pre-push hook (.git/hooks/pre-push):**
```bash
#!/bin/bash
set -e

echo "🚀 Running GoClean pre-push analysis..."

# Run comprehensive analysis before push
goclean scan \
    --config configs/rust-strict.yaml \
    --console-violations

if [ $? -ne 0 ]; then
    echo "❌ Critical quality issues found. Push blocked."
    exit 1
fi

echo "✅ Quality checks passed. Proceeding with push."
```

## Real-world Project Examples

### Web API Project

**Project: Rust REST API with actix-web**

```yaml
# goclean.yaml for web API
scan:
  paths:
    - ./src
    - ./migrations
  exclude:
    - target/
    - "src/generated/"

rust:
  enable_error_handling_check: true
  allow_unwrap: false
  enforce_result_propagation: true
  max_impl_methods: 25  # API handlers can have many methods

clippy:
  enabled: true
  categories: [correctness, suspicious, style, complexity, perf]

thresholds:
  function_lines: 40    # API handlers can be longer
  parameters: 8         # HTTP handlers often have many params
  cyclomatic_complexity: 12

output:
  html:
    enabled: true
    path: ./docs/api-quality-report.html
    group_by: ["severity", "module"]
```

### CLI Application

**Project: Command-line tool**

```yaml
# goclean.yaml for CLI tool
scan:
  paths:
    - ./src
  exclude:
    - target/
    - "tests/"

rust:
  enable_ownership_analysis: true
  detect_unnecessary_clones: true
  enforce_snake_case: true

clippy:
  enabled: true
  categories: [correctness, suspicious, style, complexity, perf, pedantic]

thresholds:
  function_lines: 25
  cyclomatic_complexity: 8
  parameters: 5

# CLI tools should have excellent error handling
rules:
  error_handling:
    enforce_result_propagation: true
    disallow_unwrap: true
    require_error_context: true
```

### Game Engine

**Project: High-performance game engine**

```yaml
# goclean.yaml for game engine
scan:
  paths:
    - ./engine/src
    - ./examples/src
  exclude:
    - target/
    - "assets/"
    - "generated/"

rust:
  allow_unsafe: true
  require_unsafe_comments: true
  detect_inefficient_string: true
  detect_blocking_in_async: true

# Performance-focused thresholds
thresholds:
  function_lines: 50     # Graphics code can be complex
  cyclomatic_complexity: 15
  parameters: 10

clippy:
  enabled: true
  categories: [correctness, suspicious, perf]
  # Skip style checks for performance code
```

This comprehensive integration guide provides practical examples for every major Rust development scenario. Each example can be adapted to specific project needs and development workflows.