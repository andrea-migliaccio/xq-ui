# Release Workflow Guide

## Overview

XQ-UI uses automated releases via GitHub Actions with semantic versioning and cross-platform builds powered by Goreleaser.

## Versioning Strategy

We follow [Semantic Versioning (SemVer)](https://semver.org/):

- **MAJOR**: Breaking changes (v2.0.0)
- **MINOR**: New features, backward compatible (v1.1.0)  
- **PATCH**: Bug fixes, backward compatible (v1.0.1)

### Version Examples
- `v1.0.0` - Initial release
- `v1.1.0` - Added new AI provider support
- `v1.0.1` - Fixed file picker bug
- `v2.0.0` - Changed CLI argument structure (breaking)

## Release Process

### 1. Prepare Release

1. **Update code and documentation:**
   ```bash
   # Make your changes
   git add .
   git commit -m "feat: add new AI provider support"
   ```

2. **Update CHANGELOG.md:**
   ```bash
   # Move items from [Unreleased] to new version section
   # Update links at bottom of file
   git add CHANGELOG.md
   git commit -m "docs: update changelog for v1.1.0"
   ```

3. **Push changes:**
   ```bash
   git push origin main
   ```

### 2. Create Release Tag

**For a new minor version (v1.1.0):**
```bash
git tag v1.1.0
git push origin v1.1.0
```

**For a patch version (v1.0.1):**
```bash
git tag v1.0.1
git push origin v1.0.1
```

**For a major version (v2.0.0):**
```bash
git tag v2.0.0
git push origin v2.0.0
```

### 3. Automated Build Process

Once the tag is pushed, GitHub Actions automatically:

1. **Triggers the release workflow** (`.github/workflows/release.yml`)
2. **Runs tests** to ensure quality
3. **Builds cross-platform binaries:**
   - Linux: amd64, arm64
   - macOS: amd64, arm64 
   - Windows: amd64, arm64
4. **Creates release archives** (tar.gz for Unix, zip for Windows)
5. **Generates checksums** for security verification
6. **Creates GitHub Release** with:
   - Auto-generated changelog
   - Release notes
   - Download links
   - Installation instructions

### 4. Release Verification

After the automated release:

1. **Check the release page:** https://github.com/andrea-migliaccio/xq-ui/releases
2. **Verify all platforms built successfully**
3. **Test the installation script:**
   ```bash
   curl -sSL https://raw.githubusercontent.com/andrea-migliaccio/xq-ui/main/install.sh | bash
   xq-ui --version
   ```
4. **Test manual download** for your platform

## Manual Release (Emergency)

If GitHub Actions fails, you can create a release manually:

```bash
# Install goreleaser
go install github.com/goreleaser/goreleaser@latest

# Create release
export GITHUB_TOKEN=your_github_token
goreleaser release --clean
```

## Pre-release Testing

For testing before official release:

```bash
# Create a pre-release tag
git tag v1.1.0-rc1
git push origin v1.1.0-rc1

# This creates a pre-release on GitHub
```

## Release Branch Strategy (Optional)

For critical hotfixes to production:

```bash
# Create release branch from tag
git checkout v1.0.0
git checkout -b hotfix/v1.0.1

# Make fixes
git commit -m "fix: critical security issue"

# Tag and push
git tag v1.0.1
git push origin v1.0.1

# Merge back to main
git checkout main
git merge hotfix/v1.0.1
git push origin main
```

## Post-Release Tasks

1. **Announce the release:**
   - Update project documentation
   - Social media announcement (optional)
   - Notify users in relevant channels

2. **Monitor for issues:**
   - Check for bug reports
   - Monitor download statistics
   - Respond to user feedback

3. **Update development version:**
   ```bash
   # In CHANGELOG.md, add new [Unreleased] section
   git add CHANGELOG.md
   git commit -m "docs: prepare for next release"
   git push origin main
   ```

## Troubleshooting

### Release Failed
1. Check the Actions tab for error logs
2. Common issues:
   - Test failures
   - Missing goreleaser configuration
   - Invalid tag format

### Wrong Version Released
1. Delete the tag and release:
   ```bash
   git tag -d v1.0.0
   git push origin --delete v1.0.0
   ```
2. Delete the release from GitHub UI
3. Create correct tag and release

### Missing Binaries
1. Check goreleaser configuration
2. Verify all target platforms in `.goreleaser.yml`
3. Check build logs for platform-specific errors

## Configuration Files

Key files for releases:
- `.goreleaser.yml` - Build and release configuration
- `.github/workflows/release.yml` - GitHub Actions workflow
- `.github/workflows/ci.yml` - Continuous integration
- `CHANGELOG.md` - Version history
- `install.sh` - Installation script

## Security Notes

- Never commit GitHub tokens or API keys
- Use GitHub secrets for sensitive data
- Verify checksums before installing releases
- Monitor for security vulnerabilities in dependencies

---

For questions or issues with releases, please open an issue or contact the maintainers.