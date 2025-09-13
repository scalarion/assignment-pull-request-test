# Assignment Pull Request Creator

A GitHub Action that automatically scans for assignment folders and creates pull
requests with README files for educational repositories.

**🚀 Built with Go for high performance and reliability**

## Overview

This tool helps educators automate assignment repository setup by:

- 🔍 **Smart Scanning**: Configurable regex patterns for assignment discovery
- 🌿 **Branch Management**: Automatic branch creation with sanitized names
- 📝 **README Generation**: Template README.md files for each assignment
- 🔄 **Pull Request Creation**: Automated PRs for assignment review
- 🛡️ **Safe Operation**: Only creates branches/PRs when they don't already exist
- 🏃 **Dry-Run Mode**: Preview operations without making actual changes
- ⚡ **4-Phase Processing**: Sync → Local work → Atomic push → PR creation

## Quick Start

### As a GitHub Action

```yaml
name: Create Assignment PRs
on:
    push:
        branches: [main]

jobs:
    create-assignment-prs:
        runs-on: ubuntu-latest
        steps:
            - uses: actions/checkout@v4
            - uses: majikmate/assignment-pull-request@v1
              with:
                  github-token: ${{ secrets.GITHUB_TOKEN }}
                  dry-run: "false"
```

### Local Development

```bash
# Build and test
make build && make run

# Live mode with your repository  
GITHUB_TOKEN=your_token GITHUB_REPOSITORY=owner/repo make run-live
```

## Configuration

### Environment Variables

| Variable                 | Required | Default                        | Description                                                                                                   |
| ------------------------ | -------- | ------------------------------ | ------------------------------------------------------------------------------------------------------------- |
| `GITHUB_TOKEN`           | ✅       | -                              | GitHub personal access token                                                                                  |
| `GITHUB_REPOSITORY`      | ✅       | -                              | Repository name (`owner/repo`)                                                                                |
| `ASSIGNMENTS_ROOT_REGEX` | ❌       | `^assignments$`                | Comma-separated patterns for assignment root directories (use `\,` to escape literal commas)                  |
| `ASSIGNMENT_REGEX`       | ❌       | `^(?P<branch>assignment-\d+)$` | Comma-separated patterns with capturing groups for branch name extraction (use `\,` to escape literal commas) |
| `DEFAULT_BRANCH`         | ❌       | `main`                         | Default branch name                                                                                           |
| `DRY_RUN`                | ❌       | `false`                        | Enable simulation mode                                                                                        |

### GitHub Action Inputs

```yaml
- uses: majikmate/assignment-pull-request@v1
  with:
      github-token: ${{ secrets.GITHUB_TOKEN }}
      assignments-root-regex: "^assignments$,^homework$,^labs$"
      assignment-regex: "^(?P<branch>assignment-\\d+)$,^(?P<subject>[^/]+)/(?P<number>\\d+)-assignment-(?P<id>\\d+)$"
      default-branch: "main"
      dry-run: "false"
```

## Project Structure

This project follows the
[Standard Go Project Layout](https://github.com/golang-standards/project-layout):

```
assignment-pull-request/
├── cmd/assignment-pr-creator/      # Main application
├── internal/                       # Private packages
│   ├── creator/                    # Business logic
│   ├── git/                        # Git operations
│   └── github/                     # GitHub API client
├── bin/                            # Built binaries
├── test/                           # Test fixtures
├── examples/                       # Usage examples
├── Makefile                        # Build commands
└── go.mod                          # Go module
```

## Development

### Prerequisites

- Go 1.24+
- Git configured
- GitHub token with repo permissions

### Build Commands

```bash
make help        # Show all commands
make build       # Build binary
make run         # Build and run (dry-run)
make run-live    # Build and run (live mode)
make test        # Run tests
make lint        # Run linter
make fmt         # Format code
make clean       # Clean artifacts
make check       # All quality checks
make install     # Install to GOPATH/bin
```

### Architecture

**`cmd/assignment-pr-creator`**: Main entry point

- Minimal initialization and error handling

**`internal/creator`**: Core business logic

- Configuration management
- Assignment discovery and processing
- Workflow orchestration

**`internal/git`**: Git operations

- Command execution with dry-run support
- Branch and commit management
- Remote synchronization

**`internal/github`**: GitHub API client

- Authentication and API calls
- Pull request management
- Repository state checking

## Examples

### Repository Structure

```
my-course/
├── assignments/
│   ├── assignment-1/          # ← Creates PR
│   ├── assignment-2/          # ← Creates PR  
│   └── semester-1/
│       └── module-1/
│           └── assignment-3/  # ← Creates PR
├── lectures/
└── resources/
```

### Custom Patterns

```bash
# Single pattern (backward compatible)
ASSIGNMENTS_ROOT_REGEX="^assignments$"

# Multiple patterns using comma separation
ASSIGNMENTS_ROOT_REGEX="^assignments$,^homework$,^labs$"

# Complex patterns with alternation (single regex)
ASSIGNMENTS_ROOT_REGEX="^(assignments|homework|labs)$"

# Patterns with literal commas (escaped with \,)
ASSIGNMENT_REGEX="^(?P<options>red\,green\,blue)$,^(?P<list>a\,b\,c)$"
```

### Assignment Extraction Patterns

The assignment regex now supports **named groups** and **unnamed groups** for
extracting custom branch names from paths:

#### Named Groups (Preferred)

Use `(?P<name>...)` syntax for clear, readable patterns:

```bash
# Simple extraction - path: assignment-01 → branch: assignment-01
ASSIGNMENT_REGEX="^(?P<branch>assignment-\d+)$"

# Subject-based extraction - path: 20-assignments/CSS/01-assignment-01 → branch: css-assignment-01  
ASSIGNMENT_REGEX="^[^/]+/(?P<subject>[^/]+)/\d+-(?P<type>assignment)-(?P<number>\d+)$"

# Complex extraction - path: course/module-1/hw-03 → branch: module-1-hw-03
ASSIGNMENT_REGEX="^[^/]+/(?P<module>[^/]+)/(?P<assignment>[^/]+)$"
```

#### Unnamed Groups (Fallback)

Use standard `(...)` syntax when named groups aren't needed:

```bash
# Single unnamed group - path: homework/hw-5 → branch: hw-5
ASSIGNMENT_REGEX="^homework/(hw-\d+)$"

# Multiple unnamed groups - path: projects/semester-1/week-3/assignment-variables → branch: projects-semester-1-week-3-assignment-variables
ASSIGNMENT_REGEX="^(projects)/(semester-\d+)/(week-\d+)/(assignment-[^/]+)$"

# Mixed patterns for different structures
ASSIGNMENT_REGEX="^(?P<branch>assignment-\d+)$,^(assignment)-(\d+)$"
```

#### Pattern Priority and Ordering

⚠️ **Important**: Pattern order matters! More specific patterns should come
first:

```bash
# ❌ WRONG: General pattern matches before specific one
ASSIGNMENT_REGEX="^(?P<course>[^/]+)/(hw|lab)-(\d+)$,^homework/(hw-\d+)$"

# ✅ CORRECT: Specific pattern matches first
ASSIGNMENT_REGEX="^homework/(hw-\d+)$,^(?P<course>[^/]+)/(hw|lab)-(\d+)$"
```

**How it works:**

- Use `(?P<name>...)` for named capture groups (preferred for readability)
- Use `(...)` for unnamed capture groups (simpler for basic patterns)
- **Named groups take priority** when both are present in a pattern
- All capturing groups are joined with hyphens to create the branch name
- Branch names are automatically sanitized (lowercase, special chars → hyphens)
- **First matching pattern wins** - order patterns from specific to general

### Testing Patterns

```bash
# Safe testing with dry-run
DRY_RUN=true GITHUB_TOKEN=dummy GITHUB_REPOSITORY=test/repo make run

# Test named groups extraction
DRY_RUN=true \
ASSIGNMENTS_ROOT_REGEX="^assignments$" \
ASSIGNMENT_REGEX="^(?P<branch>assignment-\d+)$" \
make run

# Test unnamed groups extraction  
DRY_RUN=true \
ASSIGNMENTS_ROOT_REGEX="^homework$" \
ASSIGNMENT_REGEX="^homework/(hw-\d+)$" \
make run

# Test complex path extraction (e.g., 20-assignments/CSS/01-assignment-01 → css-assignment-01)
DRY_RUN=true \
ASSIGNMENTS_ROOT_REGEX="^20-assignments$" \
ASSIGNMENT_REGEX="^20-assignments/\d+-(?P<subject>[^/]+)/\d+-(?P<assignment>[^/]+)$" \
make run

# Test mixed named and unnamed groups
DRY_RUN=true \
ASSIGNMENTS_ROOT_REGEX="^courses$" \
ASSIGNMENT_REGEX="^(?P<course>[^/]+)/(semester-\d+)/(?P<assignment>assignment-[^/]+)$" \
make run

# Test multiple patterns with comma separation (specific first!)
DRY_RUN=true \
ASSIGNMENTS_ROOT_REGEX="^assignments$,^homework$,^labs$" \
ASSIGNMENT_REGEX="^homework/(hw-\d+)$,^(?P<branch>assignment-\d+)$,^(?P<course>[^/]+)/(?P<type>hw)-(?P<number>\d+)$" \
make run
```

## How It Works

### 4-Phase Processing

1. **Sync Phase**: Fetch all remote branches to ensure complete local state
2. **Local Phase**: Process assignments locally (create branches, add READMEs)
3. **Push Phase**: Atomically push all changes to remote repository
4. **PR Phase**: Create pull requests via GitHub API

### Smart Logic

**Branch Creation**:

- ✅ Creates if no branch exists AND no PR has ever existed
- ❌ Skips if PR existed before (prevents recreating completed work)

**Pull Request Creation**:

- ✅ Creates if branch exists but no PR has ever existed
- ❌ Skips if any PR has ever existed (open, closed, or merged)

## Troubleshooting

### Common Issues

**Permission Errors**: Ensure `GITHUB_TOKEN` has `repo` scope\
**Pattern Mismatches**: Test regex patterns with `DRY_RUN=true`\
**Build Failures**: Check Go version (requires 1.24+)

### Debug Commands

```bash
# Verbose dry-run output
DRY_RUN=true make run

# Check patterns match your structure
find . -name "assignment-*" -type d

# Test Go installation
go version && make check
```

## Migration Notes

This action was originally implemented in Python and rewritten in Go for:

- **Better Performance**: Faster execution, lower memory usage
- **Single Binary**: No dependency management needed
- **Type Safety**: Compile-time error detection
- **Better Tooling**: Superior development ecosystem

The API and functionality remain identical for seamless migration.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- 📂 [Examples Directory](examples/)
- 🐛
  [Issue Tracker](https://github.com/majikmate/assignment-pull-request/issues)
- 💬
  [Discussions](https://github.com/majikmate/assignment-pull-request/discussions)

## Advanced Examples

### Complex Regex Patterns

The system supports sophisticated pattern matching for diverse educational
structures:

#### Multi-Subject Assignments

```bash
# Pattern: 20-assignments/21-JavaScript/02-assignment-01
ASSIGNMENTS_ROOT_REGEX="^20-assignments$"
ASSIGNMENT_REGEX="^test/fixtures/20-assignments/(?P<subject>\d+-\w+)/(?P<number>\d+-assignment-\d+)$"
# Results: 20-css-01-assignment-01, 21-javascript-02-assignment-01, 22-python-01-assignment-01
```

#### Course-Based Structure

```bash
# Pattern: courses/CS101/week-01/assignment-fibonacci
ASSIGNMENTS_ROOT_REGEX="^courses$"
ASSIGNMENT_REGEX="^test/fixtures/courses/(?P<course>[A-Z]+\d+)/(?P<period>.*?)/(?P<assignment>assignment-.+)$"
# Results: cs101-week-01-assignment-fibonacci, math201-chapter-3-assignment-calculus
```

#### Bootcamp Programs

```bash
# Pattern: bootcamp/2024-fall/module-frontend/week-1/assignment-html-basics
ASSIGNMENTS_ROOT_REGEX="^bootcamp$"
ASSIGNMENT_REGEX="^test/fixtures/bootcamp/(?P<year>\d+-\w+)/(?P<module>module-\w+)/(?P<week>week-\d+)/(?P<assignment>assignment-.+)$"
# Results: 2024-fall-module-frontend-week-1-assignment-html-basics
```

### Test Commands

```bash
# Quick dry-run test with default patterns
make run-dry

# Test with specific assignment root and pattern
DRY_RUN=true ASSIGNMENTS_ROOT_REGEX="^20-assignments$" ASSIGNMENT_REGEX="^test/fixtures/20-assignments/(?P<subject>\d+-\w+)/(?P<number>\d+-assignment-\d+)$" make run

# Test legacy fixture structure
DRY_RUN=true ASSIGNMENTS_ROOT_REGEX="^assignments$" ASSIGNMENT_REGEX="^test/fixtures/assignments/(?P<path>.*?assignment-\d+)$" make run

# Build and test the binary directly
make build
./bin/assignment-pr-creator --help
```

### Test Suite

The repository includes comprehensive testing using Go's built-in testing
framework and dry-run validation:

- **Dry-run Testing**: Built-in dry-run mode for safe testing
  - Simulates all operations without making changes
  - Validates regex patterns and assignment discovery
  - Tests branch name extraction and sanitization
  - GitHub API simulation

- **Integration Testing**: Real directory structure validation
  - End-to-end assignment discovery using realistic test fixtures
  - Branch name extraction with real paths
  - Environment variable configuration
  - Cross-platform path handling

- **Test Fixtures**: `test/fixtures/`
  - Multiple assignment structures (assignments, homework, labs, projects)
  - Complex educational hierarchies (20-assignments, courses, bootcamp)
  - Realistic directory hierarchies and naming patterns
  - Edge cases and nested structures

- **GitHub Actions Integration**: `.github/workflows/test-suite.yml`
  - Cross-platform testing (Ubuntu, Windows, macOS)
  - Code quality checks and validation
  - Build verification and testing

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run `make check` to verify quality
6. Submit a pull request

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and updates.
