# Assignment Pull Request Creator

A GitHub Action that automatically scans for assignment folders and creates pull
requests with README files for educational repositories.

## Features

- 🔍 **Smart Scanning**: Configurable regex patterns for assignment discovery
- 🌿 **Branch Management**: Automatic branch creation with sanitized names
- 📝 **README Generation**: Template README.md files for each assignment
- 🔄 **Pull Request Creation**: Automated PRs for assignment review
- 🛡️ **Safe Operation**: Only creates branches/PRs when they don't already exist
- 🏃 **Dry-Run Mode**: Preview operations without making actual changes

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
