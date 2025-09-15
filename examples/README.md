# Usage Examples

This directory contains example files showing how to use the Assignment Pull
Request Creator action in various scenarios.

## Requirements

- **Platform**: All examples use `ubuntu-latest` as the action requires Linux
  runners
- **Permissions**: `contents: write` and `pull-requests: write`

## Files

- **`workflow-example.yml`** - Example GitHub Actions workflow showing how to
  use the action in your repository
- **Examples include dry-run mode** for safe testing and configuration
  validation

## Using the Examples

To use any of these examples:

1. Copy the relevant file to your repository
2. Modify the parameters to match your project structure
3. Test with dry-run mode first: set `dry-run: true`
4. Ensure you have the necessary permissions and tokens configured

For the workflow example, copy `workflow-example.yml` to `.github/workflows/` in
your repository and customize as needed.

## Testing Your Configuration

Before applying any configuration to your repository:

1. **Enable dry-run mode** in the workflow
2. **Run the workflow** to see what changes would be made
3. **Review the output** for branch names, file paths, and commands
4. **Disable dry-run mode** once you're satisfied with the configuration

### Quick Test Commands

```bash
# Local dry-run test with default pattern (named groups)
DRY_RUN=true GITHUB_TOKEN=fake_token GITHUB_REPOSITORY=owner/repo ./bin/assignment-pr-creator

# Test with named groups pattern
DRY_RUN=true GITHUB_TOKEN=fake_token GITHUB_REPOSITORY=owner/repo \
ASSIGNMENT_REGEX="^assignments/(?P<branch>assignment-\d+)$" \
./bin/assignment-pr-creator

# Test with unnamed groups pattern  
DRY_RUN=true GITHUB_TOKEN=fake_token GITHUB_REPOSITORY=owner/repo \
ASSIGNMENT_REGEX="^homework/(hw-\d+)$" \
./bin/assignment-pr-creator

# Test with multiple patterns (specific before general)
DRY_RUN=true GITHUB_TOKEN=fake_token GITHUB_REPOSITORY=owner/repo \
ASSIGNMENT_REGEX="^homework/(hw-\d+)$,^assignments/(?P<branch>assignment-\d+)$" \
./bin/assignment-pr-creator
```

## Pattern Examples

### Named Groups (Recommended)

Use `(?P<name>...)` for clear, readable patterns:

- `^(?P<branch>assignment-\d+)$` - Simple assignments
- `^(?P<course>[^/]+)/(?P<week>week-\d+)/(?P<type>hw-\d+)$` - Course structure

### Unnamed Groups

Use `(...)` for simpler patterns:

- `^homework/(hw-\d+)$` - Extract homework number only
- `^(projects)/(semester-\d+)/(assignment-[^/]+)$` - Multiple groups

### Pattern Ordering

⚠️ **Important**: Order patterns from specific to general to avoid conflicts!
