# Assignment Pull Request Creator - Test Suite

This directory contains the comprehensive test suite for the Assignment Pull
Request Creator GitHub Action.

## Quick Start

```bash
# From project root - run all tests
python -m pytest tests/test_assignment_creator.py -v

# From tests directory - run test suite  
cd tests && bash test_runner.sh all

# Discovery test with fixtures
cd tests && python test_local.py discover

# Test dry-run functionality
python -m pytest tests/test_assignment_creator.py -k "dry_run" -v

# Local dry-run simulation
DRY_RUN=true GITHUB_TOKEN=fake_token GITHUB_REPOSITORY=test/repo python create_assignment_prs.py
```

## Test Structure

```
tests/
├── test_assignment_creator.py      # Unit tests (pytest)
├── test_local.py                   # Integration & CLI tests
├── test_runner.sh                  # Unified test execution script
├── fixtures/                       # Test data and structures
│   ├── assignments/                # Standard assignment structure
│   ├── homework/                   # Homework assignment structure  
│   ├── labs/                       # Lab assignment structure
│   └── projects/                   # Project assignment structure
└── README.md                       # This documentation
```

## Test Categories

### 1. Unit Tests (`test_assignment_creator.py`)

- **Framework**: pytest with comprehensive fixtures
- **Coverage**: Assignment discovery, PR creation, GitHub API mocking, **dry-run
  functionality**
- **Scenarios**:
  - Multiple regex patterns (assignments, homework, labs, projects)
  - Edge cases (empty repos, invalid patterns, API failures)
  - Cross-platform path handling (Windows/Unix)
  - Environment variable validation
  - **Dry-run mode simulation and validation**
  - GitHub API interaction patterns

### 2. Integration Tests (`test_local.py`)

- **Framework**: Direct script execution with test fixtures
- **Coverage**: End-to-end workflow simulation without GitHub API
- **Scenarios**:
  - Assignment discovery across different folder structures
  - CLI argument parsing and validation
  - Local file system operations
  - Environment variable handling

### 3. Quality Assurance (`test_runner.sh`)

- **Framework**: Unified script with multiple test types
- **Coverage**: Code quality, security, and comprehensive testing
- **Test Types**:
  - `unit`: pytest unit tests
  - `sanitize`: Input validation and security checks
  - `quality`: Code style and best practices
  - `all`: Complete test suite

## Test Configuration

### Environment Variables

```bash
# GitHub Configuration
GITHUB_TOKEN="your_token"
GITHUB_REPOSITORY="owner/repo"
DEFAULT_BRANCH="main"

# Assignment Discovery
ASSIGNMENTS_ROOT_REGEX="^assignments$"           # Root folder pattern
ASSIGNMENT_REGEX="^assignment-\d+$"             # Assignment folder pattern

# Conda Environment (optional)
CONDA_ENVIRONMENT_NAME="myenv"
CONDA_ENVIRONMENT_FILE="environment.yml"
```

### Dry-Run Testing

The test suite includes comprehensive dry-run functionality testing:

```bash
# Test dry-run mode functionality
python -m pytest tests/test_assignment_creator.py::TestDryRunFunctionality -v

# Local dry-run simulation (safe testing)
DRY_RUN=true GITHUB_TOKEN=fake_token GITHUB_REPOSITORY=test/repo python create_assignment_prs.py

# Test with specific patterns in dry-run mode
DRY_RUN=true GITHUB_TOKEN=fake_token GITHUB_REPOSITORY=test/repo \
ASSIGNMENTS_ROOT_REGEX="^(assignments|tests/fixtures)$" \
python create_assignment_prs.py

# Test different assignment types safely
DRY_RUN=true GITHUB_TOKEN=fake_token GITHUB_REPOSITORY=test/repo \
ASSIGNMENTS_ROOT_REGEX="^tests/fixtures/(homework|labs)$" \
ASSIGNMENT_REGEX="^(hw|lab)-\d+$" \
python create_assignment_prs.py
```

**Dry-run tests verify**:

- ✅ Command output simulation (git, gh commands)
- ✅ README content generation preview
- ✅ Branch name sanitization
- ✅ No actual GitHub API calls made
- ✅ Proper tracking of simulated operations

### Test Fixtures

The `fixtures/` directory contains realistic assignment structures used for all
testing scenarios:

- **`assignments/`**: Standard structure (`assignment-1`, `assignment-2`,
  `week-3/assignment-3`)
- **`homework/`**: Homework pattern (`hw-1`, `hw-2`, `hw-3`)
- **`labs/`**: Lab pattern (`lab-1`, `lab-2`)
- **`projects/`**: Project pattern (`project-1`, `project-2`)

All integration tests use these fixtures exclusively, providing comprehensive
test coverage without requiring external dependencies or cluttering the
repository root.

### Pattern Testing Examples

```bash
# Test standard assignments only
ASSIGNMENTS_ROOT_REGEX='^assignments$' ASSIGNMENT_REGEX='^assignment-\d+$' python test_local.py

# Test multiple assignment types
ASSIGNMENTS_ROOT_REGEX='^(assignments|homework|labs)$' ASSIGNMENT_REGEX='^(assignment|hw|lab)-\d+$' python test_local.py

# Test homework only
ASSIGNMENTS_ROOT_REGEX='^homework$' ASSIGNMENT_REGEX='^hw-\d+$' python test_local.py
```

## Test Execution

### Individual Test Commands

```bash
# Unit tests with verbose output
python -m pytest tests/test_assignment_creator.py -v

# Specific test function
python -m pytest tests/test_assignment_creator.py::test_discover_assignments -v

# Integration test with discovery
cd tests && python test_local.py discover

# Local simulation (dry run)
cd tests && python test_local.py --dry-run
```

### Test Runner Script

```bash
cd tests

# Run specific test type
bash test_runner.sh unit          # Unit tests only
bash test_runner.sh sanitize      # Security/validation tests
bash test_runner.sh quality       # Code quality checks
bash test_runner.sh all           # Complete test suite

# With verbose output
bash test_runner.sh all --verbose
```

### GitHub Actions Integration

The test suite runs automatically on:

- **`test-suite.yml`**: Complete test matrix (Python 3.8-3.12,
  Ubuntu/Windows/macOS)
- **`test-action.yml`**: Action-specific testing with real GitHub integration

## Test Scenarios Coverage

### 1. Assignment Discovery

- ✅ Multiple folder patterns and regex combinations
- ✅ Nested directory structures (`week-3/assignment-3`)
- ✅ Empty repositories and missing folders
- ✅ Invalid regex patterns and error handling
- ✅ Case sensitivity and path normalization

### 2. GitHub Integration

- ✅ PR creation with mocked GitHub API
- ✅ Branch management and conflict handling
- ✅ Authentication and permission validation
- ✅ Rate limiting and API error scenarios
- ✅ Repository metadata and environment setup

### 3. Environment Setup

- ✅ Conda environment creation and activation
- ✅ Dependencies installation from requirements files
- ✅ Environment variable validation and defaults
- ✅ Cross-platform compatibility testing
- ✅ Error recovery and cleanup procedures

### 4. Edge Cases

- ✅ Special characters in folder names
- ✅ Very large numbers of assignments
- ✅ Network failures and timeout handling
- ✅ Insufficient permissions scenarios
- ✅ Malformed configuration files

## Development Workflow

### Adding New Tests

1. **Unit Tests**: Add test functions to `test_assignment_creator.py`
2. **Integration**: Extend scenarios in `test_local.py`
3. **Fixtures**: Create new structures in `fixtures/` as needed
4. **Documentation**: Update this README with new patterns

### Test Data Management

- Keep fixtures minimal but representative
- Use realistic folder names and structures
- Document new patterns in fixture README
- Maintain backwards compatibility with existing tests

### Debugging Tests

```bash
# Run with maximum verbosity
python -m pytest tests/test_assignment_creator.py -v -s

# Run single test with debug output
python -m pytest tests/test_assignment_creator.py::test_name -v -s --tb=long

# Local debugging with print statements
cd tests && python test_local.py --debug
```

## Requirements

- **Python**: 3.8+ (tested on 3.8-3.12)
- **Dependencies**: `requirements.txt` (requests, PyGithub)
- **Dev Dependencies**: `requirements-dev.txt` (pytest, pytest-mock)
- **Optional**: Conda for environment management

## Continuous Integration

Tests run automatically on:

- All pull requests
- Pushes to main branch
- Multiple Python versions (3.8, 3.9, 3.10, 3.11, 3.12)
- Multiple operating systems (Ubuntu, Windows, macOS)
- Both with and without conda environments

The test suite ensures reliability across different environments and use cases,
providing confidence in the GitHub Action's functionality.
