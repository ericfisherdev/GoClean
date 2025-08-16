#!/bin/bash
set -e

# GoClean Release Script
# This script automates the release process for GoClean

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if version is provided
if [ -z "$1" ]; then
    log_error "Usage: $0 <version>"
    log_info "Example: $0 0.1.0"
    exit 1
fi

VERSION=$1
TAG="v${VERSION}"

# Validate version format (semantic versioning)
if ! echo "$VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
    log_error "Version must be in semantic versioning format (e.g., 0.1.0)"
    exit 1
fi

log_info "Starting release process for version $VERSION"

# Check if we're on the main branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$CURRENT_BRANCH" != "main" ] && [ "$CURRENT_BRANCH" != "master" ]; then
    log_warning "You are not on the main branch (current: $CURRENT_BRANCH)"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Aborted by user"
        exit 1
    fi
fi

# Check if working directory is clean
if ! git diff --quiet || ! git diff --cached --quiet; then
    log_error "Working directory is not clean. Please commit or stash your changes."
    git status --short
    exit 1
fi

# Check if tag already exists
if git tag --list | grep -q "^${TAG}$"; then
    log_error "Tag $TAG already exists"
    exit 1
fi

# Update version in files
log_info "Updating version in files..."

# Update CHANGELOG.md
if [ -f "CHANGELOG.md" ]; then
    sed -i.bak "s/## \[Unreleased\]/## [Unreleased]\n\n## [$VERSION] - $(date +%Y-%m-%d)/" CHANGELOG.md
    rm CHANGELOG.md.bak
    log_success "Updated CHANGELOG.md"
fi

# Commit version updates
if ! git diff --quiet; then
    log_info "Committing version updates..."
    git add .
    git commit -m "Prepare release $VERSION"
    log_success "Version updates committed"
fi

# Run pre-release checks
log_info "Running pre-release checks..."
if ! make release-check; then
    log_error "Pre-release checks failed"
    exit 1
fi
log_success "Pre-release checks passed"

# Create and build release
log_info "Building release..."
if ! make release-build VERSION="$VERSION"; then
    log_error "Release build failed"
    exit 1
fi
log_success "Release build completed"

# Package release
log_info "Packaging release..."
if ! make release-package VERSION="$VERSION"; then
    log_error "Release packaging failed"
    exit 1
fi
log_success "Release packaged"

# Create git tag
log_info "Creating git tag $TAG..."
git tag -a "$TAG" -m "Release $VERSION"
log_success "Git tag $TAG created"

# Push changes and tags
log_info "Pushing changes to remote..."
git push origin "$CURRENT_BRANCH"
git push origin "$TAG"
log_success "Changes and tag pushed to remote"

# Display release artifacts
log_info "Release artifacts created:"
ls -la "releases/$VERSION/"

# Display next steps
echo
log_success "Release $VERSION completed successfully!"
echo
log_info "Next steps:"
echo "  1. The GitHub Actions workflow will automatically create a GitHub release"
echo "  2. Release binaries are available in releases/$VERSION/"
echo "  3. Update documentation if needed"
echo "  4. Announce the release to users"
echo
log_info "GitHub release will be available at:"
echo "  https://github.com/ericfisherdev/GoClean/releases/tag/$TAG"