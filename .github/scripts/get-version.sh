#!/bin/bash
#
# get-version.sh - Extract version number for GoClean releases
#
# This script implements a single source of truth for version numbers.
# It reads from the VERSION file in the repository root.
#
# Usage: ./get-version.sh
# Output: Version string (e.g., "0.2.0")
#
# Exit codes:
#   0 - Success
#   1 - VERSION file not found
#   2 - Invalid version format

set -euo pipefail

# Navigate to repository root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

VERSION_FILE="${REPO_ROOT}/VERSION"

# Check if VERSION file exists
if [[ ! -f "${VERSION_FILE}" ]]; then
    echo "ERROR: VERSION file not found at ${VERSION_FILE}" >&2
    echo "Please create a VERSION file with the current version number" >&2
    exit 1
fi

# Read version from file
VERSION=$(cat "${VERSION_FILE}" | tr -d '[:space:]')

# Validate version format (semantic versioning)
if [[ ! "${VERSION}" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9\.\-]+)?(\+[a-zA-Z0-9\.\-]+)?$ ]]; then
    echo "ERROR: Invalid version format: ${VERSION}" >&2
    echo "Version must follow semantic versioning (e.g., 1.2.3, 1.2.3-beta.1)" >&2
    exit 2
fi

# Output the version
echo "${VERSION}"