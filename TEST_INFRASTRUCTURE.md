# Test Infrastructure Documentation

## Overview

This document describes the comprehensive test infrastructure built for the
Assignment Pull Request Creator project. The system includes unit tests,
integration tests, test utilities, and coverage reporting.

## Project Structure

````
/workspaces/assignment-pull-request/
├─### Test Data Management

### Fixtures

- Located in `tests/fixtures/` (existing) and generated dynamically (testutil)
- Realistic assignment structures that mirror production usage
- Image files and complex directory hierarchies
- **New**: Test patterns for unnamed groups and mixed group scenarios

### Regex Helper Functions

The creator package includes helper functions for regex analysis:

```go
// hasNamedGroups checks if a regex pattern contains named capturing groups
func hasNamedGroups(pattern *regexp.Regexp) bool

// hasCapturingGroups checks if a regex pattern contains any capturing groups (named or unnamed)
func hasCapturingGroups(pattern *regexp.Regexp) bool
````

These helpers enable:

- Validation that patterns have capturing groups for branch extraction
- Testing different types of regex patterns (named vs unnamed groups)
- Proper error handling for invalid patterns

### Sample Contentl/

│ ├── creator/ # Core business logic │ │ ├── creator.go │ │ └── creator_test.go
│ ├── git/ # Git operations │ │ ├── operations.go │ │ └── operations_test.go │
├── github/ # GitHub API client │ │ ├── client.go │ │ └── client_test.go │ └──
testutil/ # Test utilities │ ├── testutil.go │ └── testutil_test.go ├── cmd/ │
└── assignment-pr-creator/ │ ├── main.go │ └── main_test.go # Integration tests
└── Makefile # Enhanced build system

````
## Test Coverage

### Unit Tests

**Creator Package** (`internal/creator/creator_test.go`)

- **Coverage**: 28.0%
- **Test Functions**: 12 test functions, 40+ test cases
- **Focus Areas**:
  - Configuration validation and environment handling
  - Regex pattern parsing and validation for capturing groups
  - Branch name extraction with **named groups** and **unnamed groups**
  - Pattern priority and ordering (specific patterns before general ones)
  - Image link rewriting for GitHub repository URLs
  - Pull request body generation from assignment paths
  - **New**: Comprehensive unnamed groups support and mixed group scenarios

**Git Package** (`internal/git/operations_test.go`)

- **Coverage**: 60.3%
- **Test Functions**: 10 test functions, 18 test cases
- **Focus Areas**:
  - Git command abstraction with dry-run support
  - Branch creation, switching, and management
  - File staging, committing, and atomic push operations
  - Integration with local git repository (when available)

**GitHub Package** (`internal/github/client_test.go`)

- **Coverage**: 37.8%
- **Test Functions**: 6 test functions, 20 test cases
- **Focus Areas**:
  - GitHub API client initialization and configuration
  - Pull request creation with proper dry-run simulation
  - Repository name parsing and validation
  - Error handling for edge cases and invalid inputs

**Testutil Package** (`internal/testutil/testutil_test.go`)

- **Coverage**: 78.7%
- **Test Functions**: 10 test functions, 16 test cases
- **Focus Areas**:
  - Temporary workspace management for test isolation
  - Assignment and file structure creation utilities
  - Environment variable setup and restoration
  - String matching and file validation helpers

### Integration Tests

**Main Package** (`cmd/assignment-pr-creator/main_test.go`)

- **Coverage**: 0.0% (expected for integration tests)
- **Test Functions**: 3 test functions, 10 test cases
- **Focus Areas**:
  - End-to-end workflow simulation with complex assignment structures
  - Multi-assignment discovery and processing
  - Image rewriting in realistic scenarios
  - Comprehensive dry-run validation

## Test Utilities

### TempWorkspace

Provides isolated temporary workspaces for testing:

```go
ws := testutil.NewTempWorkspace(t)
ws.CreateStandardStructure()  // Creates realistic assignment structure
restore := ws.ChangeToWorkspace()  // Changes working directory
defer restore()
````

### Environment Management

Manages environment variables safely across tests:

```go
env := testutil.NewEnvSetup()
defer env.Restore()
env.SetTestDefaults()  // Sets common test environment variables
```

### Standard Test Fixtures

Pre-built assignment structures for consistent testing:

- Basic assignments, homework, labs, projects
- Course structures with modules and semesters
- Bootcamp-style hierarchical assignments
- Image-rich assignments with static content

## Makefile Targets

### Test Commands

```bash
# Run all tests
make test

# Run unit tests only (internal packages)
make test-unit

# Run integration tests
make test-integration

# Run individual package tests
make test-creator
make test-git
make test-github
make test-testutil

# Run short tests (skip long-running)
make test-short

# Run benchmarks
make bench
```

### Coverage Commands

```bash
# Run tests with coverage reporting
make coverage

# Generate HTML coverage report
make coverage-html

# Show detailed coverage in terminal
make coverage-show
```

### Quality Commands

```bash
# Format code
make fmt

# Run linter
make lint

# Run go vet
make vet

# Run all checks (format, vet, lint, tests, coverage)
make check

# Run quick checks (format, vet, short tests)
make check-quick

# CI pipeline (for GitHub Actions)
make ci
```

## Test Patterns and Best Practices

### Regex Pattern Validation

All regex patterns undergo comprehensive validation:

```go
func TestRegexValidation(t *testing.T) {
    tests := []struct {
        name        string
        pattern     string
        shouldError bool
    }{
        {"valid_regex_with_named_groups", `^(?P<branch>assignment-\d+)$`, false},
        {"valid_regex_with_unnamed_groups", `^(assignment)-(\d+)$`, false},
        {"invalid_regex_without_any_capturing_groups", `^assignment-\d+$`, true},
    }
    // ...
}
```

### Branch Name Extraction Testing

Supports both named and unnamed capturing groups:

```go
func TestExtractBranchNameWithUnnamedGroups(t *testing.T) {
    // Test cases include:
    // - Single unnamed group: homework/hw-5 → hw-5
    // - Multiple unnamed groups: projects/semester-1/week-3/assignment-variables → projects-semester-1-week-3-assignment-variables
    // - Mixed named/unnamed groups (named groups take priority)
    // - Pattern ordering considerations (specific before general)
}
```

### Dry-Run Testing

All tests respect the dry-run mode to avoid side effects:

```go
func TestCreatePullRequest(t *testing.T) {
    client := github.NewClient("test-token", "owner/repo", true) // dry-run=true
    _, err := client.CreatePullRequest(ctx, "title", "head", "base", "body")
    assert.NoError(t, err) // Should succeed in dry-run mode
}
```

### Table-Driven Tests

Complex scenarios use table-driven patterns:

```go
tests := []struct {
    name     string
    input    string
    expected string
}{
    {"simple_path", "assignment-1", "assignment-1"},
    {"path_with_slashes", "course/week-1/assignment", "course-week-1-assignment"},
}
```

### Test Isolation

Each test creates isolated environments:

```go
func TestAssignmentDiscovery(t *testing.T) {
    ws := testutil.NewTempWorkspace(t)  // Isolated temporary directory
    ws.CreateStandardStructure()       // Consistent test data
    // Test logic here
    // Cleanup is automatic via t.TempDir()
}
```

### Subtests for Organization

Related test cases are grouped using subtests:

```go
func TestBranchNameExtraction(t *testing.T) {
    t.Run("simple_assignment_pattern", func(t *testing.T) { /* ... */ })
    t.Run("homework_pattern", func(t *testing.T) { /* ... */ })
    t.Run("course_structure", func(t *testing.T) { /* ... */ })
}
```

## Coverage Goals and Analysis

### Current Coverage

- **Overall**: 39.8%
- **Creator**: 28.0% (core business logic)
- **Git**: 60.3% (command abstraction)
- **GitHub**: 37.8% (API client)
- **Testutil**: 77.6% (utility functions)

### Coverage Analysis

- **Integration tests** don't contribute to package coverage but validate
  end-to-end functionality
- **Dry-run mode** allows comprehensive testing without external dependencies
- **Error paths** are well-covered for robustness, including regex validation
  failures
- **Edge cases** include empty inputs, invalid patterns, malformed data, and
  pattern ordering conflicts
- **New**: Unnamed groups extraction, mixed group scenarios, and pattern
  priority testing

## Running Tests

### Local Development

```bash
# Quick feedback loop
make test-short

# Full test suite
make test

# With coverage
make coverage

# Check everything before commit
make check
```

### Continuous Integration

```bash
# CI pipeline
make ci
```

This runs:

1. Dependency verification
2. Code formatting check
3. Static analysis (vet)
4. Linting
5. Full test suite with coverage

## Test Data Management

### Fixtures

- Located in `tests/fixtures/` (existing) and generated dynamically (testutil)
- Realistic assignment structures that mirror production usage
- Image files and complex directory hierarchies

### Sample Content

- Pre-defined instruction templates for different assignment types
- Sample image content for testing link rewriting
- Comprehensive test data covering edge cases

## Future Enhancements

### Planned Improvements

1. **Increase unit test coverage** to 80%+ for core packages
2. **Add performance benchmarks** for large assignment sets
3. **Implement property-based testing** for regex patterns
4. **Add mutation testing** to validate test quality
5. **Create visual coverage reports** for CI

### Test Infrastructure

1. **Test parallelization** for faster execution
2. **Test result caching** for repeated runs
3. **Flaky test detection** and retries
4. **Test environment validation** checks

This comprehensive test infrastructure ensures the Assignment Pull Request
Creator is robust, reliable, and maintainable across various assignment
structures and usage patterns.
