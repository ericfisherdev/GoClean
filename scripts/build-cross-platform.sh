#!/bin/bash

# Cross-platform build script for GoClean with Rust support
# This script handles building both Go and Rust components across different platforms

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
RUST_PARSER_PATH="$PROJECT_ROOT/../playground/goclean-rust-parser"
BUILD_DIR="$PROJECT_ROOT/bin"
LIB_DIR="$PROJECT_ROOT/lib"

# Platform detection
OS_NAME=$(uname -s)
ARCH_NAME=$(uname -m)

echo -e "${BLUE}GoClean Cross-Platform Build Script${NC}"
echo -e "${BLUE}====================================${NC}"

# Function to print colored output
print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check Rust installation
check_rust() {
    print_info "Checking Rust installation..."
    
    if ! command_exists cargo; then
        print_error "Rust/Cargo not found!"
        echo "Please install Rust from https://rustup.rs/"
        echo "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
        return 1
    fi
    
    print_status "Rust toolchain found: $(rustc --version)"
    print_status "Cargo found: $(cargo --version)"
    return 0
}

# Function to setup Rust targets
setup_rust_targets() {
    print_info "Setting up Rust cross-compilation targets..."
    
    local targets=(
        "x86_64-unknown-linux-gnu"
        "aarch64-unknown-linux-gnu"
        "x86_64-apple-darwin"
        "aarch64-apple-darwin"
        "x86_64-pc-windows-gnu"
    )
    
    for target in "${targets[@]}"; do
        if rustup target list --installed | grep -q "$target"; then
            print_status "Target $target already installed"
        else
            print_info "Installing target $target..."
            if rustup target add "$target" 2>/dev/null; then
                print_status "Target $target installed successfully"
            else
                print_warning "Failed to install target $target (may require additional setup)"
            fi
        fi
    done
}

# Function to build Rust library for specific target
build_rust_target() {
    local target=$1
    local lib_name=$2
    
    print_info "Building Rust library for $target..."
    
    cd "$RUST_PARSER_PATH"
    
    if cargo build --release --target "$target" 2>/dev/null; then
        print_status "Successfully built Rust library for $target"
        
        # Copy the built library to lib directory
        local source_file=""
        local dest_file=""
        
        case "$target" in
            *linux*)
                source_file="target/$target/release/libgoclean_rust_parser.so"
                dest_file="$LIB_DIR/libgoclean_rust_parser-${target/unknown-/}.so"
                ;;
            *darwin*)
                source_file="target/$target/release/libgoclean_rust_parser.dylib"
                dest_file="$LIB_DIR/libgoclean_rust_parser-${target/apple-/}.dylib"
                ;;
            *windows*)
                source_file="target/$target/release/goclean_rust_parser.dll"
                dest_file="$LIB_DIR/libgoclean_rust_parser-${target/pc-/}.dll"
                ;;
        esac
        
        if [[ -f "$source_file" ]]; then
            mkdir -p "$LIB_DIR"
            cp "$source_file" "$dest_file"
            print_status "Copied library to $dest_file"
        else
            print_warning "Built library not found at $source_file"
        fi
    else
        print_warning "Failed to build Rust library for $target"
        return 1
    fi
    
    cd "$PROJECT_ROOT"
}

# Function to build all Rust libraries
build_rust_libraries() {
    print_info "Building Rust libraries for all supported platforms..."
    
    if [[ ! -d "$RUST_PARSER_PATH" ]]; then
        print_error "Rust parser directory not found: $RUST_PARSER_PATH"
        return 1
    fi
    
    mkdir -p "$LIB_DIR"
    
    # Build for each target
    build_rust_target "x86_64-unknown-linux-gnu" "libgoclean_rust_parser"
    build_rust_target "aarch64-unknown-linux-gnu" "libgoclean_rust_parser"
    build_rust_target "x86_64-apple-darwin" "libgoclean_rust_parser"
    build_rust_target "aarch64-apple-darwin" "libgoclean_rust_parser"
    build_rust_target "x86_64-pc-windows-gnu" "libgoclean_rust_parser"
    
    print_status "Rust library builds completed"
    echo ""
    echo "Built libraries:"
    ls -la "$LIB_DIR"/libgoclean_rust_parser* 2>/dev/null || print_warning "No Rust libraries found"
}

# Function to build Go binaries
build_go_binaries() {
    print_info "Building Go binaries for all supported platforms..."
    
    mkdir -p "$BUILD_DIR"
    
    local binary_name="goclean"
    local ldflags="-ldflags \"-X main.Version=${VERSION:-dev} -X main.GitCommit=${GIT_COMMIT:-unknown} -X main.BuildDate=$(date -u +"%Y-%m-%dT%H:%M:%SZ")\""
    
    # Build for each platform
    local platforms=(
        "linux:amd64"
        "darwin:amd64"
        "darwin:arm64"
        "windows:amd64"
    )
    
    for platform in "${platforms[@]}"; do
        local os="${platform%:*}"
        local arch="${platform#*:}"
        local output_name="$binary_name-$os-$arch"
        
        if [[ "$os" == "windows" ]]; then
            output_name="$output_name.exe"
        fi
        
        print_info "Building $output_name..."
        
        # Try with CGO first, fall back to Go-only if it fails
        if CGO_ENABLED=1 GOOS="$os" GOARCH="$arch" go build $ldflags -o "$BUILD_DIR/$output_name" ./cmd/goclean 2>/dev/null; then
            print_status "Built $output_name with CGO support"
        elif CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build $ldflags -o "$BUILD_DIR/$output_name" ./cmd/goclean 2>/dev/null; then
            print_warning "Built $output_name without CGO support (Go-only mode)"
        else
            print_error "Failed to build $output_name"
        fi
    done
    
    print_status "Go binary builds completed"
    echo ""
    echo "Built binaries:"
    ls -la "$BUILD_DIR"/goclean-* 2>/dev/null || print_warning "No Go binaries found"
}

# Function to package releases
package_releases() {
    print_info "Packaging releases..."
    
    local version="${VERSION:-dev}"
    local release_dir="$PROJECT_ROOT/releases/$version"
    
    mkdir -p "$release_dir"
    
    cd "$BUILD_DIR"
    
    # Package each platform
    local platforms=("linux-amd64" "darwin-amd64" "darwin-arm64" "windows-amd64")
    
    for platform in "${platforms[@]}"; do
        local binary_name="goclean-$platform"
        local package_name="goclean-$version-$platform"
        
        if [[ "$platform" == "windows-amd64" ]]; then
            binary_name="$binary_name.exe"
        fi
        
        if [[ -f "$binary_name" ]]; then
            print_info "Packaging $package_name..."
            
            # Check for corresponding Rust library
            local rust_lib=""
            case "$platform" in
                "linux-amd64")
                    rust_lib="$LIB_DIR/libgoclean_rust_parser-linux-gnu.so"
                    ;;
                "darwin-amd64")
                    rust_lib="$LIB_DIR/libgoclean_rust_parser-darwin.dylib"
                    ;;
                "darwin-arm64")
                    rust_lib="$LIB_DIR/libgoclean_rust_parser-darwin.dylib"
                    ;;
                "windows-amd64")
                    rust_lib="$LIB_DIR/libgoclean_rust_parser-windows-gnu.dll"
                    ;;
            esac
            
            if [[ "$platform" == "windows-amd64" ]]; then
                # Use zip for Windows
                if command_exists zip; then
                    if [[ -f "$rust_lib" ]]; then
                        zip -q "$release_dir/$package_name.zip" "$binary_name" "$rust_lib"
                        print_status "Created $package_name.zip with Rust support"
                    else
                        zip -q "$release_dir/$package_name.zip" "$binary_name"
                        print_warning "Created $package_name.zip without Rust library"
                    fi
                else
                    if [[ -f "$rust_lib" ]]; then
                        tar -czf "$release_dir/$package_name.tar.gz" "$binary_name" "$rust_lib"
                        print_status "Created $package_name.tar.gz with Rust support"
                    else
                        tar -czf "$release_dir/$package_name.tar.gz" "$binary_name"
                        print_warning "Created $package_name.tar.gz without Rust library"
                    fi
                fi
            else
                # Use tar.gz for Unix-like systems
                if [[ -f "$rust_lib" ]]; then
                    tar -czf "$release_dir/$package_name.tar.gz" "$binary_name" -C "$LIB_DIR" "$(basename "$rust_lib")"
                    print_status "Created $package_name.tar.gz with Rust support"
                else
                    tar -czf "$release_dir/$package_name.tar.gz" "$binary_name"
                    print_warning "Created $package_name.tar.gz without Rust library"
                fi
            fi
        else
            print_warning "Binary $binary_name not found, skipping packaging"
        fi
    done
    
    cd "$PROJECT_ROOT"
    
    print_status "Release packaging completed"
    echo ""
    echo "Release packages:"
    ls -la "$release_dir"/ 2>/dev/null || print_warning "No release packages found"
}

# Main execution
main() {
    local mode="${1:-all}"
    
    print_info "Platform: $OS_NAME $ARCH_NAME"
    print_info "Project root: $PROJECT_ROOT"
    print_info "Build mode: $mode"
    echo ""
    
    case "$mode" in
        "rust")
            if check_rust; then
                setup_rust_targets
                build_rust_libraries
            else
                print_error "Rust build failed - toolchain not available"
                exit 1
            fi
            ;;
        "go")
            build_go_binaries
            ;;
        "package")
            package_releases
            ;;
        "all"|*)
            if check_rust; then
                setup_rust_targets
                build_rust_libraries
            else
                print_warning "Rust not available - building Go-only binaries"
            fi
            build_go_binaries
            package_releases
            ;;
    esac
    
    print_status "Cross-platform build completed!"
}

# Run with command line argument
main "$@"