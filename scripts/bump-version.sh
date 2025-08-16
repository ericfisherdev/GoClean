#!/bin/bash

# bump-version.sh - Automatically bump version in YYYY.MM.DD.PATCH format
# This script is designed to be called automatically on each commit

set -e

# Get the current date
CURRENT_DATE=$(date +"%Y.%m.%d")

# Read the current version from VERSION file
if [ -f VERSION ]; then
    CURRENT_VERSION=$(cat VERSION)
else
    echo "VERSION file not found!"
    exit 1
fi

# Extract date and patch from current version
# Handle both old format (0.3.0) and new format (YYYY.MM.DD.PATCH)
if [[ $CURRENT_VERSION =~ ^[0-9]{4}\.[0-9]{2}\.[0-9]{2}\.[0-9]+$ ]]; then
    # New format
    VERSION_DATE=$(echo $CURRENT_VERSION | cut -d'.' -f1-3)
    VERSION_PATCH=$(echo $CURRENT_VERSION | cut -d'.' -f4)
else
    # Old format or invalid - start fresh
    VERSION_DATE=""
    VERSION_PATCH=0
fi

# Determine new version
if [ "$VERSION_DATE" == "$CURRENT_DATE" ]; then
    # Same day - increment patch
    NEW_PATCH=$((VERSION_PATCH + 1))
else
    # New day - reset patch to 0
    NEW_PATCH=0
fi

NEW_VERSION="${CURRENT_DATE}.${NEW_PATCH}"

echo "Bumping version from $CURRENT_VERSION to $NEW_VERSION"

# Update VERSION file
echo "$NEW_VERSION" > VERSION

# Update main.go
sed -i "s/Version: \".*\"/Version: \"$NEW_VERSION\"/" cmd/goclean/main.go

# Update main_test.go
sed -i "s/rootCmd.Version != \".*\"/rootCmd.Version != \"$NEW_VERSION\"/" cmd/goclean/main_test.go
sed -i "s/Expected version '.*'/Expected version '$NEW_VERSION'/" cmd/goclean/main_test.go

echo "Version bumped to $NEW_VERSION"