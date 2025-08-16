# Automatic Release Tagging

This document explains the automatic tagging system that creates version tags when pull requests are merged into the `release` branch.

## How It Works

The GitHub Action `.github/workflows/auto-tag-release.yml` automatically:

1. **Triggers** when a pull request is merged into the `release` branch
2. **Reads** the current version from the `VERSION` file
3. **Checks** if a tag for that version already exists
4. **Creates** a new git tag (e.g., `v2025.08.16.2`) if it doesn't exist
5. **Pushes** the tag to the repository
6. **Creates** a GitHub release with automatic release notes

## Workflow Process

### 1. Version Detection
- Reads the version from the `VERSION` file in the repository root
- Uses the CalVer format: `YYYY.MM.DD.PATCH`
- Example: `2025.08.16.2` becomes tag `v2025.08.16.2`

### 2. Duplicate Prevention
- Checks if a tag with that version already exists
- Skips tag creation if the tag is already present
- Prevents duplicate releases for the same version

### 3. Tag Creation
- Creates an annotated git tag with the version
- Uses the format `v{VERSION}` (e.g., `v2025.08.16.2`)
- Includes a descriptive tag message

### 4. GitHub Release
- Automatically creates a GitHub release
- Uses the pull request title and link in release notes
- Includes the PR number that triggered the release

## Usage

### Normal Workflow
1. Create a feature branch from `develop`
2. Make your changes and commit (version auto-bumps via pre-commit hook)
3. Create a pull request to `release` branch
4. When the PR is merged, the tag is automatically created

### Release Notes
The GitHub release will include:
- Version number in the title
- Pull request title as the main change description
- Link to the original pull request
- PR number reference

## Example Release Notes

```
## GoClean v2025.08.16.2

This release was automatically created when PR #45 was merged into the release branch.

### Changes
Add cross-platform pattern matching improvements

For detailed changes, see the pull request (#45).
```

## Requirements

### Repository Permissions
The workflow requires:
- `contents: write` - To create tags and releases
- `GITHUB_TOKEN` - Automatically provided by GitHub Actions

### File Dependencies
- `VERSION` file must exist in repository root
- Version must follow `YYYY.MM.DD.PATCH` format

## Troubleshooting

### Tag Already Exists
If a tag already exists for the current version:
- The workflow will skip tag creation
- No error will occur
- Check if the version was properly bumped before merging

### VERSION File Missing
If the `VERSION` file is not found:
- The workflow will fail with an error
- Ensure the file exists and contains a valid version

### Permission Issues
If tag creation fails:
- Check repository settings for GitHub Actions permissions
- Ensure the workflow has `contents: write` permission

## Integration with Version Bumping

This system works seamlessly with the automatic version bumping:
1. Pre-commit hook bumps version on each commit
2. Pull request contains the new version
3. When merged to `release`, the tag is created automatically
4. No manual version management required