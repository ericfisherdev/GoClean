# GoClean Versioning System

## Version Format

GoClean uses Calendar Versioning (CalVer) with the format: **YYYY.MM.DD.PATCH**

- **YYYY**: 4-digit year (e.g., 2025)
- **MM**: 2-digit month (01-12)
- **DD**: 2-digit day (01-31)
- **PATCH**: Incremental number for same-day releases (starts at 0)

### Examples
- `2025.08.16.0` - First release on August 16, 2025
- `2025.08.16.1` - Second release on August 16, 2025
- `2025.08.17.0` - First release on August 17, 2025

## Automatic Version Bumping

The version is automatically bumped on every commit through a git pre-commit hook.

### How It Works
1. When you make a commit, the pre-commit hook runs `scripts/bump-version.sh`
2. If it's the first commit of the day, the patch resets to 0
3. If it's a subsequent commit on the same day, the patch increments
4. The VERSION file and source files are automatically updated
5. These changes are staged and included in your commit

### Manual Version Bump
If needed, you can manually bump the version:
```bash
./scripts/bump-version.sh
```

### Check Current Version
To see the current version and its components:
```bash
./scripts/version-info.sh
```

Or check the application:
```bash
./bin/goclean version
```

## Files Affected

When the version is bumped, the following files are updated:
- `VERSION` - The source of truth for the current version
- `cmd/goclean/main.go` - The Version field in rootCmd
- `cmd/goclean/main_test.go` - Version assertions in tests

## Benefits of CalVer

1. **Temporal Clarity**: Immediately know when a version was released
2. **No Ambiguity**: Each version is unique and meaningful
3. **Automatic**: No need to decide on version numbers
4. **Multiple Daily Releases**: The patch number handles same-day releases
5. **Natural Sorting**: Versions sort correctly by date

## Migration from SemVer

GoClean migrated from Semantic Versioning (0.x.x) to Calendar Versioning on August 16, 2025.
The last SemVer release was 0.2.1.