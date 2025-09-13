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

| Variable                 | Required | Default            | Description                             |
| ------------------------ | -------- | ------------------ | --------------------------------------- |
| `GITHUB_TOKEN`           | ✅       | -                  | GitHub personal access token            |
| `GITHUB_REPOSITORY`      | ✅       | -                  | Repository name (`owner/repo`)          |
| `ASSIGNMENTS_ROOT_REGEX` | ❌       | `^assignments$`    | Pattern for assignment root directories |
| `ASSIGNMENT_REGEX`       | ❌       | `^assignment-\d+$` | Pattern for individual assignments      |
| `DEFAULT_BRANCH`         | ❌       | `main`             | Default branch name                     |
| `DRY_RUN`                | ❌       | `false`            | Enable simulation mode                  |

### GitHub Action Inputs

```yaml
- uses: majikmate/assignment-pull-request@v1
  with:
      github-token: ${{ secrets.GITHUB_TOKEN }}
      assignments-root-regex: "^(assignments|homework)$"
      assignment-regex: "^(assignment|hw)-\\d+$"
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
│   ├── git/                       # Git operations  
│   └── github/                    # GitHub API client
├── bin/                           # Built binaries
├── tests/                         # Test fixtures
├── examples/                      # Usage examples
├── Makefile                       # Build commands
└── go.mod                         # Go module
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
# Match multiple root directories
ASSIGNMENTS_ROOT_REGEX="^(assignments|homework|labs)$"

# Match different naming conventions
ASSIGNMENT_REGEX="^(assignment|hw|lab)-\d+$"
```

### Testing Patterns

```bash
# Safe testing with dry-run
DRY_RUN=true GITHUB_TOKEN=dummy GITHUB_REPOSITORY=test/repo make run

# Test custom patterns
DRY_RUN=true \
ASSIGNMENTS_ROOT_REGEX="^homework$" \
ASSIGNMENT_REGEX="^hw-\d+$" \
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

## Usage- **README Generation**: Template README.md files for each assignment- **README Generation**: Template README.md files for each assignment

### Basic GitHub Action- **Pull Request Creation**: Automated PRs for assignment review- **Pull Request Creation**: Automated PRs for assignment review

```yaml- **Safe Operation**: Only creates branches/PRs when they don't already exist- **Safe Operation**: Only creates branches/PRs when they don't already exist
name: Create Assignment PRs

on:- **Dry-Run Mode**: Preview operations without making actual changes- **Dry-Run Mode**: Preview operations without making actual changes

  push:

    branches: [main]- **4-Phase Processing**: Sync → Local work → Atomic push → PR creation- **4-Phase Processing**: Sync → Local work → Atomic push → PR creation



jobs:

  create-assignment-prs:

    runs-on: ubuntu-latest## Usage## Usage

    steps:

      - uses: actions/checkout@v4

      - uses: majikmate/assignment-pull-request@v1

        with:### Basic GitHub Action### Basic GitHub Action

          github-token: ${{ secrets.GITHUB_TOKEN }}

          dry-run: "false"
```

`yaml`yaml

### Local Development & Testing

name: Create Assignment PRsname: Create Assignment PRs

```````bash
# Build the applicationon:on:

make build

  push:  push:

# Test with dry-run mode (safe)

make run    branches: [main]    branches: [main]



# Test with your repository

GITHUB_TOKEN=your_token GITHUB_REPOSITORY=owner/repo make run-live

```jobs:jobs:



## Configuration  create-assignment-prs:  create-assignment-prs:



### Environment Variables    runs-on: ubuntu-latest    runs-on: ubuntu-latest



**Required:**    steps:    steps:

- `GITHUB_TOKEN`: GitHub personal access token

- `GITHUB_REPOSITORY`: Repository name in "owner/repo" format      - uses: actions/checkout@v4      - uses: actions/checkout@v4



**Optional:**      - uses: majikmate/assignment-pull-request@v1      - uses: majikmate/assignment-pull-request@v1

- `ASSIGNMENTS_ROOT_REGEX`: Pattern for assignment root directories (default: "^assignments$")

- `ASSIGNMENT_REGEX`: Pattern for individual assignments (default: "^assignment-\\d+$")        with:        with:

- `DEFAULT_BRANCH`: Default branch name (default: "main")

- `DRY_RUN`: Enable simulation mode (default: "false")          github-token: ${{ secrets.GITHUB_TOKEN }}          github-token: ${{ secrets.GITHUB_TOKEN }}



### Action Inputs          dry-run: "false"          dry-run: "false"



```yaml``````

- uses: majikmate/assignment-pull-request@v1

  with:

    github-token: ${{ secrets.GITHUB_TOKEN }}

    assignments-root-regex: "^(assignments|homework)$"### Local Development & Testing### Local Development & Testing

    assignment-regex: "^(assignment|hw)-\\d+$"

    default-branch: "main"

    dry-run: "false"

``````bash```bash



## Examples# Build the application# Build the application



See the [`examples/`](examples/) directory for:make buildmake build

- GitHub Actions workflow configurations

- Repository structure examples

- Testing and development patterns

# Test with dry-run mode (safe)# Test with dry-run mode (safe)

## Development

make runmake run

This project uses the [Standard Go Project Layout](https://github.com/golang-standards/project-layout). For detailed development information, see [GO_README.md](GO_README.md).



### Quick Development Commands

# Test with your repository# Test with your repository

```bash

make help        # Show all available commandsGITHUB_TOKEN=your_token GITHUB_REPOSITORY=owner/repo make run-liveGITHUB_TOKEN=your_token GITHUB_REPOSITORY=owner/repo make run-live

make build       # Build the binary

make test        # Run tests``````

make lint        # Run linter

make clean       # Clean build artifacts

make check       # Run all quality checks

```## Configuration## Configuration



## Migration from Python



This action was originally implemented in Python and has been rewritten in Go for better performance and maintainability. The functionality and interface remain identical - only the implementation language has changed.### Environment Variables### Environment Variables



## Algorithm



### Directory Discovery**Required:****Required:**



The action uses a flexible two-tier regex system:- `GITHUB_TOKEN`: GitHub personal access token- `GITHUB_TOKEN`: GitHub personal access token



1. **Root Discovery**: Uses `ASSIGNMENTS_ROOT_REGEX` to find assignment container directories- `GITHUB_REPOSITORY`: Repository name in "owner/repo" format- `GITHUB_REPOSITORY`: Repository name in "owner/repo" format

2. **Assignment Discovery**: Within each root, uses `ASSIGNMENT_REGEX` to identify individual assignments



### Assignment Processing

**Optional:****Optional:**

For each discovered assignment directory:

- `ASSIGNMENTS_ROOT_REGEX`: Pattern for assignment root directories (default: "^assignments$")- `ASSIGNMENTS_ROOT_REGEX`: Pattern for assignment root directories (default: "^assignments$")

1. **Sanitization**: Converts directory name to a valid Git branch name

2. **Branch Creation**: Creates a branch like `assignment/<sanitized-name>`- `ASSIGNMENT_REGEX`: Pattern for individual assignments (default: "^assignment-\\d+$")- `ASSIGNMENT_REGEX`: Pattern for individual assignments (default: "^assignment-\\d+$")

3. **README Addition**: Adds a standardized README.md file

4. **Pull Request**: Creates a PR from the assignment branch to the default branch- `DEFAULT_BRANCH`: Default branch name (default: "main")- `DEFAULT_BRANCH`: Default branch name (default: "main")



### Git Operations- `DRY_RUN`: Enable simulation mode (default: "false")- `DRY_RUN`: Enable simulation mode (default: "false")



- **Atomic Operations**: All git operations for an assignment happen atomically

- **Safe Handling**: Skips assignments that already have branches or PRs

- **Clean Workspace**: Each assignment gets a fresh working directory### Action Inputs### Action Inputs



## Directory Structure Examples



The action works with various educational repository structures:```yaml```yaml



```- uses: majikmate/assignment-pull-request@v1- uses: majikmate/assignment-pull-request@v1

assignments/

├── assignment-1/  with:  with:

│   └── instructions.md

├── assignment-2/    github-token: ${{ secrets.GITHUB_TOKEN }}    github-token: ${{ secrets.GITHUB_TOKEN }}

│   └── instructions.md

└── assignment-3/    assignments-root-regex: "^(assignments|homework)$"    assignments-root-regex: "^(assignments|homework)$"

    └── instructions.md

```    assignment-regex: "^(assignment|hw)-\\d+$"    assignment-regex: "^(assignment|hw)-\\d+$"



Or more complex structures:    default-branch: "main"    default-branch: "main"



```    dry-run: "false"    dry-run: "false"

semester-1/

├── module-1/``````

│   └── assignment-1/

│       └── instructions.md

└── module-2/

    └── assignment-2/## Examples## Examples

        └── instructions.md
```````

## TestingSee the [`examples/`](examples/) directory for:See the [`examples/`](examples/) directory for:

### Build and Test- GitHub Actions workflow configurations- GitHub Actions workflow configurations

```bash- Repository structure examples- Repository structure examples
# Build and run all checks

make check- Testing and development patterns- Testing and development patterns



# Run specific tests

make test

## Development## Development

# Build the binary

make build
```

This project uses the
[Standard Go Project Layout](https://github.com/golang-standards/project-layout).
For detailed development information, see [GO_README.md](GO_README.md).This
project uses the
[Standard Go Project Layout](https://github.com/golang-standards/project-layout).
For detailed development information, see [GO_README.md](GO_README.md).

### Dry Run Testing

````bash
# Test locally with dry-run mode### Quick Development Commands### Quick Development Commands

make run



# Test with specific repository

DRY_RUN=true GITHUB_TOKEN=fake_token GITHUB_REPOSITORY=owner/repo ./bin/assignment-pr-creator```bash```bash
````

make help # Show all available commandsmake help # Show all available commands

## License

make build # Build the binarymake build # Build the binary

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file
for details.

make test # Run testsmake test # Run tests

## Contributing

make lint # Run lintermake lint # Run linter

1. Fork the repository

2. Create your feature branch (`git checkout -b feature/amazing-feature`)make
   clean # Clean build artifactsmake clean # Clean build artifacts

3. Commit your changes (`git commit -m 'Add some amazing feature'`)

4. Push to the branch (`git push origin feature/amazing-feature`)make check #
   Run all quality checksmake check # Run all quality checks

5. Open a Pull Request

`````
For development setup and guidelines, see [GO_README.md](GO_README.md).


## Migration from Python## Migration from Python



This action was originally implemented in Python and has been rewritten in Go for better performance and maintainability. The functionality and interface remain identical - only the implementation language has changed.This action was originally implemented in Python and has been rewritten in Go for better performance and maintainability. The functionality and interface remain identical - only the implementation language has changed.



## License## License



This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.



## Support## Support



- 📚 [Detailed Go Documentation](GO_README.md)- 📚 [Detailed Go Documentation](GO_README.md)

- 🔧 [Examples and Patterns](examples/)- 🔧 [Examples and Patterns](examples/)

- 🐛 [Issue Tracker](https://github.com/majikmate/assignment-pull-request/issues)- 🐛 [Issue Tracker](https://github.com/majikmate/assignment-pull-request/issues)

- 💡 [Discussions](https://github.com/majikmate/assignment-pull-request/discussions)- 💡 [Discussions](https://github.com/majikmate/assignment-pull-request/discussions)

- ✅ Creates README.md content first, then creates PR if NO pull request has
  ever existed for that branch name
- ❌ Skips if ANY pull request has ever existed (open, closed, or merged)
- ❌ Skips if README creation doesn't result in changes compared to the default
  branch
- ℹ️ Ensures PRs always have meaningful content and prevents duplicates

**Content Creation Process**:

1. 📝 Creates README.md template in the assignment directory
2. 🔍 Validates that the content creation resulted in changes
3. 🔄 Creates pull request only if changes exist

**Common Scenarios**:

- 🆕 **New assignment**: Creates branch, README content, and PR
- 🔄 **Existing branch, no PR history**: Creates README content and PR (if
  changes)
- ✅ **Completed assignment** (PR merged, branch deleted): Takes no action
- 🔁 **Existing branch with PR history**: Takes no action (no new content/PR
  created)
- ⚠️ **README already exists**: May skip PR creation if no new changes detected

## Quick Start

```yaml
name: Create Assignment Pull Requests
on:
    push:
        branches: [main]
        paths: ["assignments/**"]

jobs:
    create-assignments:
        runs-on: ubuntu-latest
        permissions:
            contents: write
            pull-requests: write
        steps:
            - uses: actions/checkout@v4
            - uses: majikmate/assignment-pull-request@v1
              with:
                  github-token: ${{ secrets.GITHUB_TOKEN }}
```

## Configuration

### Input Parameters

| Parameter                | Description                                          | Required | Default                       |
| ------------------------ | ---------------------------------------------------- | -------- | ----------------------------- |
| `assignments-root-regex` | Regex pattern to match assignment root directories   | No       | `^assignments$`               |
| `assignment-regex`       | Regex pattern to match individual assignment folders | No       | `^assignment-\\d+$`           |
| `default-branch`         | Default branch to create pull requests against       | No       | `main`                        |
| `github-token`           | GitHub token for API access                          | Yes      | `${{ secrets.GITHUB_TOKEN }}` |
| `dry-run`                | Simulate operations without making actual changes    | No       | `false`                       |

### Output Parameters

| Parameter               | Description                                |
| ----------------------- | ------------------------------------------ |
| `created-branches`      | JSON array of branch names created         |
| `created-pull-requests` | JSON array of pull request numbers created |

## Dry-Run Mode

The action supports a dry-run mode that simulates all operations without making
actual changes to your repository. This is perfect for:

- 🧪 **Testing configurations** before applying them
- 🔍 **Previewing what changes** would be made
- 📚 **Learning git commands** the action would execute
- 🐛 **Debugging regex patterns** and directory structures

### Basic Dry-Run Example

```yaml
name: Preview Assignment Setup
on:
    workflow_dispatch:
        inputs:
            dry-run:
                description: "Enable dry-run mode"
                type: boolean
                default: true

jobs:
    preview-assignments:
        runs-on: ubuntu-latest
        permissions:
            contents: read # Reduced permissions for dry-run
        steps:
            - uses: actions/checkout@v4
            - uses: majikmate/assignment-pull-request@v1
              with:
                  github-token: ${{ secrets.GITHUB_TOKEN }}
                  dry-run: ${{ inputs.dry-run }}
```

### Dry-Run Output

When dry-run mode is enabled, the action outputs:

```bash
🏃 DRY RUN MODE: Simulating operations without making actual changes

[DRY RUN] Would create branch with command:
  git checkout -b assignments-assignment-1 main
  git push -u origin assignments-assignment-1

[DRY RUN] Would create README.md at assignments/assignment-1/README.md
[DRY RUN] Would commit with commands:
  git checkout assignments-assignment-1
  mkdir -p assignments/assignment-1
  echo '[content]' > assignments/assignment-1/README.md
  git add assignments/assignment-1/README.md
  git commit -m 'Add README for assignment assignments/assignment-1'
  git push origin assignments-assignment-1

[DRY RUN] Would create pull request with command:
  gh pr create \
    --title 'Assignment: Assignments - Assignment-1' \
    --body '[PR description]' \
    --head assignments-assignment-1 \
    --base main

[DRY RUN] Simulated pull request #1: Assignment: Assignments - Assignment-1
```

### Local Dry-Run Testing

```bash
# Test with dry-run mode enabled
DRY_RUN=true GITHUB_TOKEN=fake_token GITHUB_REPOSITORY=owner/repo python create_assignment_prs.py

# Test different patterns
DRY_RUN=true GITHUB_TOKEN=fake_token GITHUB_REPOSITORY=owner/repo \
ASSIGNMENTS_ROOT_REGEX="^(assignments|homework)$" \
ASSIGNMENT_REGEX="^(assignment|hw)-\d+$" \
python create_assignment_prs.py
```

## Error Handling

The action implements robust error handling for GitHub API operations:

**⚠️ Fail-Fast Behavior**: If any GitHub API operation fails (branch creation,
pull request operations, etc.), the action will immediately exit with a failure
status rather than continuing. This ensures that:

- **Workflow failures are immediate and clear** when GitHub operations encounter
  issues
- **No partial operations** are left in an inconsistent state
- **Clear error messages** are displayed showing which operation failed
- **GitHub Actions workflows fail appropriately** for proper CI/CD feedback

**Common error scenarios that cause immediate failure**:

- Authentication issues (invalid GitHub token)
- Permission problems (insufficient repository access)
- API rate limits exceeded
- Network connectivity issues
- Repository access restrictions

**💡 Tip**: Always test with dry-run mode first to validate your configuration
before running actual operations.

## Complete Configuration Example

```yaml
name: Assignment Management
on:
    push:
        branches: [main, develop]
        paths:
            - "assignments/**"
            - "homework/**"
            - "labs/**"
    workflow_dispatch:
        inputs:
            assignments-root-regex:
                description: "Regex pattern for assignment root folders"
                required: false
                default: "^(assignments|homework|labs)$"
            assignment-regex:
                description: "Regex pattern for assignment folders"
                required: false
                default: '^(assignment|hw|lab)-\d+$'
            default-branch:
                description: "Default branch for pull requests"
                required: false
                default: "main"
            dry-run:
                description: "Enable dry-run mode (preview only)"
                type: boolean
                required: false
                default: false

jobs:
    create-assignment-prs:
        runs-on: ubuntu-latest
        permissions:
            contents: write
            pull-requests: write
            issues: write

        steps:
            - name: Checkout repository
              uses: actions/checkout@v4
              with:
                  fetch-depth: 0

            - name: Create assignment pull requests
              id: create-prs
              uses: majikmate/assignment-pull-request@v1
              with:
                  assignments-root-regex: ${{ github.event.inputs.assignments-root-regex || '^(assignments|homework|labs)$' }}
                  assignment-regex: ${{ github.event.inputs.assignment-regex || '^(assignment|hw|lab)-\d+$' }}
                  default-branch: ${{ github.event.inputs.default-branch || 'main' }}
                  dry-run: ${{ github.event.inputs.dry-run || false }}
                  github-token: ${{ secrets.GITHUB_TOKEN }}

            - name: Display results
              run: |
                  echo "Created branches: ${{ steps.create-prs.outputs.created-branches }}"
                  echo "Created PRs: ${{ steps.create-prs.outputs.created-pull-requests }}"

                  # Count results
                  BRANCH_COUNT=$(echo '${{ steps.create-prs.outputs.created-branches }}' | jq 'length')
                  PR_COUNT=$(echo '${{ steps.create-prs.outputs.created-pull-requests }}' | jq 'length')

                  echo "Summary: Created $BRANCH_COUNT branches and $PR_COUNT pull requests"

            - name: Notify on failure
              if: failure()
              run: |
                  echo "::error::Assignment PR creation failed. Check the logs above for details."
```

## Common Use Cases

### 1. Standard Course Structure

```
repo/
├── assignments/
│   ├── assignment-1/
│   ├── assignment-2/
│   └── assignment-3/
```

**Configuration:**

```yaml
assignments-root-regex: "^assignments$"
assignment-regex: '^assignment-\d+$'
```

### 2. Multiple Assignment Types

```
repo/
├── assignments/
│   ├── assignment-1/
│   └── assignment-2/
├── homework/
│   ├── hw-1/
│   └── hw-2/
├── labs/
│   ├── lab-1/
│   └── lab-2/
```

**Configuration:**

```yaml
assignments-root-regex: "^(assignments|homework|labs)$"
assignment-regex: '^(assignment|hw|lab)-\d+$'
```

### 3. Nested Weekly Structure

```
repo/
├── assignments/
│   ├── week-1/
│   │   └── assignment-1/
│   ├── week-2/
│   │   └── assignment-2/
```

**Configuration:**

```yaml
assignments-root-regex: "^assignments$"
assignment-regex: '^assignment-\d+$'
```

### 4. Course-Specific Naming

```
repo/
├── cs101-assignments/
│   ├── assignment-1/
│   └── assignment-2/
├── math202-homework/
│   ├── hw-1/
│   └── hw-2/
```

**Configuration:**

```yaml
assignments-root-regex: "^(cs101-assignments|math202-homework)$"
assignment-regex: '^(assignment|hw)-\d+$'
```

## Multiple Assignment Types Example

```yaml
name: Process All Assignment Types
on:
    push:
        branches: [main]

jobs:
    process-assignments:
        runs-on: ubuntu-latest
        strategy:
            matrix:
                include:
                    - name: "Regular Assignments"
                      root-pattern: "^assignments$"
                      assignment-pattern: '^assignment-\d+$'
                    - name: "Homework"
                      root-pattern: "^homework$"
                      assignment-pattern: '^hw-\d+$'
                    - name: "Labs"
                      root-pattern: "^labs$"
                      assignment-pattern: '^lab-\d+$'
                    - name: "Projects"
                      root-pattern: "^projects$"
                      assignment-pattern: '^project-\d+$'

        steps:
            - uses: actions/checkout@v4

            - name: Process ${{ matrix.name }}
              uses: majikmate/assignment-pull-request@v1
              with:
                  assignments-root-regex: ${{ matrix.root-pattern }}
                  assignment-regex: ${{ matrix.assignment-pattern }}
                  github-token: ${{ secrets.GITHUB_TOKEN }}
```

## Development

### Local Testing

```bash
# Clone and setup
git clone https://github.com/majikmate/assignment-pull-request.git
cd assignment-pull-request

# Test dry-run mode (recommended for initial testing)
DRY_RUN=true GITHUB_TOKEN=fake_token GITHUB_REPOSITORY=owner/repo python create_assignment_prs.py

# Quick test - discover assignments using test fixtures
cd tests && python test_local.py discover

# Quick test - test branch sanitization
cd tests && python test_local.py sanitize "week-1/assignment-1"

# Run full local integration test
cd tests && python test_local.py

# Run comprehensive unit tests (includes dry-run tests)
python -m pytest tests/test_assignment_creator.py -v

# Run all tests with the test runner
cd tests && bash test_runner.sh all
```

### Test Suite

The repository includes a comprehensive test suite covering:

- **Unit Tests**: `tests/test_assignment_creator.py`
  - Assignment discovery logic with mocked file systems
  - Branch name sanitization
  - Regex pattern validation
  - Environment configuration
  - **Dry-run functionality testing**
  - GitHub API interaction patterns

- **Integration Tests**: `tests/test_local.py`
  - End-to-end assignment discovery using realistic test fixtures
  - Branch name sanitization with real paths
  - Environment variable configuration
  - Cross-platform path handling

- **Test Fixtures**: `tests/fixtures/`
  - Multiple assignment structures (assignments, homework, labs, projects)
  - Realistic directory hierarchies and naming patterns
  - Edge cases and nested structures

- **GitHub Actions Integration**: `.github/workflows/test-suite.yml`
  - Cross-platform testing (Ubuntu, Windows, macOS)
  - Code quality checks (Black, Flake8, MyPy)
  - Security and performance validation

### Test Commands

````bash
### Test Commands
```bash
# Use the test runner script (recommended)
cd tests && bash test_runner.sh help                       # Show all available commands
cd tests && bash test_runner.sh discovery                  # Discovery only
cd tests && bash test_runner.sh sanitize                   # Sanitization only
cd tests && bash test_runner.sh unit                       # Unit tests
cd tests && bash test_runner.sh all                        # Run all tests

# Direct pytest commands
python -m pytest tests/ -v                     # All unit tests
python -m pytest tests/ -k "sanitiz"          # Specific test pattern

# Custom environment testing
cd tests && ASSIGNMENT_REGEX='^hw-\d+$' bash test_runner.sh discovery
````

````
```

## Examples

Complete usage examples are available in the `examples/` directory:

- **`workflow-example.yml`**: Ready-to-use GitHub Actions workflow
- **`README.md`**: Instructions for implementing the examples

Copy the workflow example to `.github/workflows/` in your repository and customize the parameters to match your assignment structure.

## Security

### Required Permissions

```yaml
permissions:
    contents: write # To create branches and files
    pull-requests: write # To create pull requests
````

## Troubleshooting

**No assignments found**: Check your regex patterns match your folder structure

```bash
python -c "import re; print(re.match(r'^assignment-\d+$', 'assignment-1'))"
```

**Permission errors**: Ensure your workflow has the required permissions listed
above

**Pattern issues**: Test patterns with the manual workflow dispatch to debug

**Testing configurations**: Use dry-run mode to preview operations:

```bash
# Test your configuration safely
DRY_RUN=true GITHUB_TOKEN=fake_token GITHUB_REPOSITORY=owner/repo \
ASSIGNMENTS_ROOT_REGEX="your-pattern" \
ASSIGNMENT_REGEX="your-assignment-pattern" \
python create_assignment_prs.py
```

**Validating regex patterns**: Test with dry-run and check the discovered
assignments:

```yaml
# Add this to your workflow for testing
- name: Test Configuration (Dry Run)
  uses: majikmate/assignment-pull-request@v1
  with:
      dry-run: true
      assignments-root-regex: "^your-pattern$"
      assignment-regex: "^your-assignment-pattern$"
      github-token: ${{ secrets.GITHUB_TOKEN }}
```

---

For more examples and advanced usage, see the
[GitHub repository](https://github.com/majikmate/assignment-pull-request).
`````
