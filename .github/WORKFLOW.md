# GoClean Git Workflow

This document describes the Git workflow for the GoClean project.

## Branch Structure

### `develop` (Default Branch)
- **Purpose**: Main development branch where all feature work gets integrated
- **Default branch**: Yes - all new clones start here
- **Source for**: All feature branches and pull requests
- **Deployment**: Not deployed, development only

### `release` (Production Branch)  
- **Purpose**: Production-ready releases
- **Protected**: Yes - requires pull requests and reviews
- **Source**: Only accepts pull requests from `develop` branch
- **Deployment**: Production deployments come from this branch

## Workflow Process

### 1. Feature Development
```bash
# Start from develop branch
git checkout develop
git pull origin develop

# Create feature branch
git checkout -b feature/your-feature-name

# Work on your feature
# ... make changes, commits ...

# Push feature branch
git push -u origin feature/your-feature-name

# Create pull request to develop branch
gh pr create --base develop --title "Your Feature" --body "Description"
```

### 2. Integration to Develop
- All feature branches merge into `develop` via pull request
- Code review and CI checks must pass
- Delete feature branch after merge

### 3. Release Process
```bash
# When develop is ready for release
git checkout develop
git pull origin develop

# Create pull request from develop to release
gh pr create --base release --head develop --title "Release v1.x.x" --body "Release notes"

# After review and approval, merge to release
# Tag the release
git checkout release
git pull origin release
git tag v1.x.x
git push origin v1.x.x
```

## Branch Protection Rules

### `release` Branch Protection
- ✅ Require pull request reviews (1 reviewer)
- ✅ Dismiss stale reviews when new commits are pushed
- ✅ Require status checks to pass before merging
- ✅ Prevent force pushes
- ✅ Prevent branch deletion

### Allowed Merge Sources
- `release` branch only accepts PRs from `develop`
- `develop` branch accepts PRs from any feature branches

## Best Practices

### Commit Messages
Follow conventional commits format:
```
type(scope): description

feat(scanner): add multi-language support
fix(config): resolve YAML parsing issue  
docs(readme): update installation instructions
test(parser): add comprehensive test coverage
```

### Feature Branch Naming
Use descriptive branch names with prefixes:
- `feature/` - New features
- `bugfix/` - Bug fixes
- `hotfix/` - Critical fixes (can branch from release)
- `docs/` - Documentation updates
- `test/` - Test improvements

### Pull Request Process
1. Create descriptive PR title and description
2. Link to related issues if applicable
3. Ensure all tests pass
4. Request review from team members
5. Address review feedback
6. Merge using "Squash and merge" for clean history

### Release Versioning
Follow semantic versioning (SemVer):
- `MAJOR.MINOR.PATCH`
- `MAJOR`: Breaking changes
- `MINOR`: New features (backward compatible)
- `PATCH`: Bug fixes (backward compatible)

## Common Commands

### Setup New Development Environment
```bash
git clone https://github.com/ericfisherdev/GoClean.git
cd GoClean
# develop branch is checked out by default
make deps
make test
make build
```

### Sync with Latest Changes
```bash
# Update develop branch
git checkout develop
git pull origin develop

# Update your feature branch
git checkout feature/your-branch
git merge develop  # or git rebase develop
```

### Clean Up Branches
```bash
# Delete merged feature branch locally
git branch -d feature/merged-branch

# Delete remote feature branch
git push origin --delete feature/merged-branch

# Clean up tracking branches
git remote prune origin
```

## CI/CD Integration

The workflow supports continuous integration:
- All PRs trigger automated testing
- Merges to `develop` can trigger staging deployments
- Merges to `release` trigger production deployments
- Tags on `release` create GitHub releases

## Emergency Hotfixes

For critical fixes that need to bypass normal workflow:

```bash
# Create hotfix branch from release
git checkout release
git pull origin release
git checkout -b hotfix/critical-fix

# Make minimal changes
# ... fix the issue ...

# Create PR directly to release
gh pr create --base release --title "Hotfix: Critical Issue" --body "Emergency fix"

# After merge, also merge to develop
git checkout develop
git pull origin develop
git merge release
git push origin develop
```

This ensures the hotfix is preserved in both branches and future development includes the fix.