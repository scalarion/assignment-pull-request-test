# Assignment Pull Request Creator

A GitHub Action that automatically scans for assignment folders and creates pull
requests with README files for educational repositories.

**🚀 Built with Go for high performance and reliability**

## Requirements

- **Platform**: Linux runners only (`ubuntu-latest`)
- **Runtime**: GitHub Actions environment
- **Permissions**: `contents: write` and `pull-requests: write`

## Features

- 🔍 **Smart Scanning**: Configurable regex patterns for assignment discovery
- 🌿 **Branch Management**: Automatic branch creation with sanitized names
- 📝 **README Generation**: Template README.md files for each assignment
- 🔄 **Pull Request Creation**: Automated PRs for assignment review
- 🛡️ **Safe Operation**: Only creates branches/PRs when they don't already exist
- 🏃 **Dry-Run Mode**: Preview operations without making actual changes
- ⚡ **4-Phase Processing**: Sync → Local work → Atomic push → PR creation
- ✅ **Branch Name Validation**: Prevents conflicts between assignments

## Quick Start

### GitHub Action

```yaml
name: Create Assignment PRs
on: [push]
jobs:
    assignments:
        runs-on: ubuntu-latest
        steps:
            - uses: actions/checkout@v4
            - uses: majikmate/assignment-pull-request@v1
              with:
                  github-token: ${{ secrets.GITHUB_TOKEN }}
                  dry-run: "false"
```

### Local Usage

```bash
# Build and test
make build && make run

# Live mode with your repository  
GITHUB_TOKEN=your_token GITHUB_REPOSITORY=owner/repo make run-live
```

## Configuration

### Environment Variables

| Variable               | Default                        | Description                                                               |
| ---------------------- | ------------------------------ | ------------------------------------------------------------------------- |
| `GITHUB_TOKEN` ✅      | -                              | GitHub personal access token                                              |
| `GITHUB_REPOSITORY` ✅ | -                              | Repository name (`owner/repo`)                                            |
| `ASSIGNMENT_REGEX`     | `^(?P<branch>assignment-\d+)$` | Comma-separated patterns with capturing groups for branch name extraction |
| `DEFAULT_BRANCH`       | `main`                         | Default branch name                                                       |
| `DRY_RUN`              | `false`                        | Enable simulation mode                                                    |

**Note**: Use `\,` to escape literal commas within regex patterns.

### GitHub Action Configuration

```yaml
- uses: majikmate/assignment-pull-request@v1
  with:
      github-token: ${{ secrets.GITHUB_TOKEN }}
      assignment-regex: "^assignments/(?P<branch>assignment-\\d+)$,^homework/(?P<subject>[^/]+)/(?P<number>\\d+)-assignment-(?P<id>\\d+)$"
      default-branch: "main"
      dry-run: "false"
```

## Pattern Examples

### Basic Patterns

```bash
# Single pattern - full path from repository root
ASSIGNMENT_REGEX="^assignments/(?P<branch>assignment-\d+)$"

# Multiple patterns (comma-separated) - full paths from repository root
ASSIGNMENT_REGEX="^assignments/(?P<branch>assignment-\d+)$,^homework/(hw-\d+)$,^labs/(?P<branch>lab-\d+)$"

# Escaped commas in patterns
ASSIGNMENT_REGEX="^assignments/(?P<options>red\,green\,blue)$,^homework/(?P<list>a\,b\,c)$"
```

### Advanced Patterns

```bash
# Subject-based: 20-assignments/CSS/01-assignment-01 → css-assignment-01
ASSIGNMENT_REGEX="^[^/]+/(?P<a_subject>[^/]+)/\d+-(?P<b_type>assignment)-(?P<c_number>\d+)$"

# Course structure: courses/CS101/week-01/assignment-fibonacci → cs101-week-01-assignment-fibonacci  
ASSIGNMENT_REGEX="^courses/(?P<a_course>[A-Z]+\d+)/(?P<b_period>.*?)/(?P<c_assignment>assignment-.+)$"

# Mixed groups: bootcamp/2024-fall/module-frontend/week-1/assignment-html-basics → assignment-html-basics-module-frontend-week-1-2024-fall
ASSIGNMENT_REGEX="^bootcamp/(?P<year>\d+-\w+)/(?P<module>module-\w+)/(?P<week>week-\d+)/(?P<assignment>assignment-.+)$"
```

### Branch Name Rules

- **Named Groups**: Use `(?P<name>...)` - sorted alphabetically by group
  **name** in branch name
- **Unnamed Groups**: Use `(...)` - appear after named groups in order of
  appearance
- **Pattern Priority**: First matching pattern wins - order from specific to
  general
- **Automatic Sanitization**: Lowercase, special characters become hyphens

**Example**: Pattern
`^(?P<course>[^/]+)/(?P<assignment>[^/]+)/(?P<module>[^/]+)$`

- Path: `CS101/sorting/algorithms`
- Named groups: `assignment="sorting"`, `course="CS101"`, `module="algorithms"`
- Groups sorted alphabetically: `assignment` → `course` → `module`
- Branch name: `sorting-cs101-algorithms`

## How It Works

### Processing Phases

1. **Sync**: Fetch remote branches for complete state
2. **Local**: Create branches and README files locally
3. **Push**: Atomically push all changes to remote
4. **PR**: Create pull requests via GitHub API

### Smart Logic

- ✅ **Creates branch/PR**: When neither exists
- ❌ **Skips**: If PR ever existed (prevents recreating completed work)
- 🔍 **Validates**: All branch names are unique before processing

## Development

### Prerequisites

- Go 1.24+
- Git configured
- GitHub token with repo permissions

### Commands

```bash
make help        # Show all commands
make build       # Build binary
make run         # Build and run (dry-run)
make run-live    # Build and run (live mode)  
make test        # Run tests
make check       # All quality checks
```

### Testing

```bash
# Safe testing with dry-run
DRY_RUN=true make run

# Test specific patterns
DRY_RUN=true \
ASSIGNMENT_REGEX="^assignments/(?P<branch>assignment-\d+)$" \
make run
```

## Troubleshooting

### Common Issues

- **Permission Errors**: Ensure `GITHUB_TOKEN` has `repo` scope
- **Pattern Mismatches**: Test regex patterns with `DRY_RUN=true`
- **Build Failures**: Check Go version (requires 1.24+)
- **Branch Name Conflicts**: Multiple assignments generating identical branch
  names

### Branch Name Validation

The system validates that all assignments generate **unique branch names**:

```bash
# ❌ This fails - both create "assignment-1":
ASSIGNMENT_REGEX="^(?P<branch>assignment-\d+)$,^[^/]+/(?P<assignment>assignment-\d+)$"

# ✅ This works - creates "assignment-1" and "cs101-assignment-1":
ASSIGNMENT_REGEX="^(?P<branch>assignment-\d+)$,^(?P<course>[^/]+)/(?P<assignment>assignment-\d+)$"
```

**Benefits**: Early detection, clear error messages, prevents data loss

### Debug Commands

```bash
DRY_RUN=true make run                    # Test patterns safely
find . -name "assignment-*" -type d      # Check directory structure  
go version && make check                 # Verify environment
```

## Git Post-Checkout Hook

The repository includes a Git post-checkout hook that automatically configures
sparse-checkout when switching to assignment branches.

### Installation

**Quick Install**:

```bash
curl -sSL https://raw.githubusercontent.com/majikmate/assignment-pull-request/main/install-hook.sh | bash
```

**Manual Setup**:

```bash
# Configure global hooks directory
mkdir -p ~/.githooks
git config --global core.hooksPath ~/.githooks

# Build and install the hook
git clone https://github.com/majikmate/assignment-pull-request.git
cd assignment-pull-request
go build -o ~/.githooks/post-checkout ./cmd/githook
chmod +x ~/.githooks/post-checkout
```

### How It Works

1. **Workflow Detection**: Scans `.github/workflows/` for
   assignment-pull-request action usage
2. **Pattern Extraction**: Extracts `assignment-regex` from workflow
   configurations
3. **Branch Matching**: When you checkout a branch, checks if it matches any
   assignment folder patterns
4. **Sparse Checkout**: If matched, configures Git to show only the relevant
   assignment folders

### Example Workflow

```bash
# You have assignments/assignment-1/, assignments/assignment-2/ folders
# Workflow uses: assignment-regex: "^assignments/(?P<branch>assignment-\d+)$"

git checkout assignment-1  # Hook runs automatically
# Git sparse-checkout now shows:
# - Root files (README.md, etc.)
# - .github/ directory  
# - assignments/assignment-1/ folder only
```

### Benefits

- **Focus**: See only relevant assignment files
- **Performance**: Faster git operations with fewer files
- **Organization**: Cleaner workspace when working on specific assignments
- **Automation**: No manual sparse-checkout configuration needed

## Architecture

**Standard Go Project Layout**:

- `cmd/assignment-pr-creator/`: Main application entry point
- `internal/creator/`: Core business logic and workflow orchestration
- `internal/git/`: Git operations with dry-run support
- `internal/github/`: GitHub API client and pull request management

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- 📂 [Examples Directory](examples/)
- 🐛
  [Issue Tracker](https://github.com/majikmate/assignment-pull-request/issues)
- 💬
  [Discussions](https://github.com/majikmate/assignment-pull-request/discussions)
