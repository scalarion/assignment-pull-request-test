# Usage Examples

This directory contains example files showing how to use the Assignment Pull
Request Creator action in various scenarios.

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
# Local dry-run test
DRY_RUN=true GITHUB_TOKEN=fake_token GITHUB_REPOSITORY=owner/repo python create_assignment_prs.py

# Test with your specific patterns
DRY_RUN=true GITHUB_TOKEN=fake_token GITHUB_REPOSITORY=owner/repo \
ASSIGNMENTS_ROOT_REGEX="^your-pattern$" \
ASSIGNMENT_REGEX="^your-assignment-pattern$" \
python create_assignment_prs.py
```
