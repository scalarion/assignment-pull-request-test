# Development Container

This repository includes a development container configuration that provides a
consistent, cross-platform development environment for the Assignment Pull
Request Creator GitHub Action.

## What's Included

The devcontainer includes:

- **Base Image**: `ghcr.io/majikmate/classroom-codespace-image:latest`
- **Python Environment**: Conda environment with Python 3.12 and development
  tools
- **Conda Environment**: `assignment-pr-creator` with all project dependencies
- **GitHub CLI**: Pre-installed for GitHub API interactions
- **Docker CLI**: Available for container management
- **Git**: Up-to-date version built from source
- **VS Code Extensions**:
  - Python development tools (Python, Pylint, Flake8, Black formatter)
  - GitHub Actions workflow support
  - YAML and JSON editing with validation
- **Cross-Platform Support**: Normalized paths for Windows/Unix compatibility

## Getting Started

### Option 1: GitHub Codespaces (Recommended)

1. Click the "Code" button on the GitHub repository
2. Select "Codespaces" tab
3. Click "Create codespace on main"
4. Wait for the environment to build (first time may take a few minutes)

### Option 2: VS Code with Dev Containers Extension

1. Install the
   [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)
   for VS Code
2. Clone this repository locally
3. Open the repository in VS Code
4. When prompted, click "Reopen in Container" or use the command palette:
   `Dev Containers: Reopen in Container`

### Option 3: Docker CLI

```bash
# Clone the repository
git clone https://github.com/majikmate/assignment-pull-request.git
cd assignment-pull-request

# Build and run the container
docker run -it --rm -v $(pwd):/workspace -w /workspace ghcr.io/majikmate/classroom-codespace-image:latest bash

# Activate the conda environment
conda activate assignment-pr-creator

# Install/update dependencies (if needed)
conda env update -f environment.yml
```

## Development Workflow

Once the devcontainer is running:

1. **Environment Setup**: Conda environment `assignment-pr-creator` is
   automatically created and activated
2. **Run Tests**: Execute the comprehensive test suite
   ```bash
   # Run all unit tests with pytest
   python -m pytest tests/test_assignment_creator.py -v

   # Run integration tests and discovery
   cd tests && python test_local.py discover

   # Run the complete test suite
   cd tests && bash test_runner.sh all
   ```
3. **Test with Fixtures**: All tests now use the fixture-based approach in
   `tests/fixtures/`
4. **Cross-Platform Testing**: Tests include path normalization for Windows/Unix
   compatibility
5. **Format Code**: Black formatter is configured and will auto-format Python
   code
6. **Lint Code**: Pylint and Flake8 are configured for code quality checking

## Test Structure

The project uses a centralized testing approach:

```
tests/
├── test_assignment_creator.py      # Unit tests (pytest)
├── test_local.py                   # Integration & CLI tests  
├── test_runner.sh                  # Unified test execution
├── fixtures/                       # All test data
│   ├── assignments/                # Standard assignment structure
│   ├── homework/                   # Homework assignment structure
│   ├── labs/                       # Lab assignment structure
│   └── projects/                   # Project assignment structure
└── README.md                       # Comprehensive test documentation
```

## Environment Variables for Testing

When testing the GitHub Action functionality, you'll need to set these
environment variables:

```bash
export GITHUB_TOKEN="your_github_token"
export GITHUB_REPOSITORY="owner/repo"
export ASSIGNMENTS_ROOT_REGEX="^assignments$"
export ASSIGNMENT_REGEX="^assignment-\d+$"
export DEFAULT_BRANCH="main"
```

**Note**: The project has moved away from using root-level assignments folder.
All testing now uses the fixture-based approach in `tests/fixtures/` for better
isolation and cross-platform compatibility.

## Debugging

The devcontainer includes comprehensive debugging tools:

- **Python Debugger**: VS Code Python debugging is pre-configured for pytest and
  integration tests
- **GitHub CLI**: Test GitHub API interactions directly
- **Conda Environment**: Isolated Python environment with all dependencies
- **Cross-Platform Path Testing**: Tools for testing Windows/Unix path
  compatibility
- **Terminal Access**: Full bash terminal with zsh shell support for advanced
  command execution
- **Test Fixtures**: Isolated test data structures for reproducible testing

## Customization

To customize the devcontainer:

1. **Environment**: Edit `environment.yml` to modify the conda environment
2. **VS Code Settings**: Edit `.devcontainer/devcontainer.json`
   - Add additional VS Code extensions in the `extensions` array
   - Modify Python and other tool settings in the `settings` object
3. **Dependencies**: Update `environment.yml` for conda packages or
   `requirements.txt` for pip packages
4. **Post-Create Commands**: Modify the `postCreateCommand` to add setup scripts
5. **Test Configuration**: Add new fixtures in `tests/fixtures/` for additional
   test scenarios

## Troubleshooting

### Container Won't Start

- Check that Docker is running
- Verify you have access to the
  `ghcr.io/majikmate/classroom-codespace-image:latest` image
- Try rebuilding the container: `Dev Containers: Rebuild Container`

### Conda Environment Issues

- Check that the conda environment was created: `conda env list`
- Manually create the environment: `conda env create -f environment.yml`
- Activate the environment: `conda activate assignment-pr-creator`
- Check the container logs for errors during `postCreateCommand`

### Test Failures

- Ensure you're using the fixture-based tests: all test data is in
  `tests/fixtures/`
- For cross-platform issues, check path normalization in test assertions
- Run tests individually to isolate issues:
  `python -m pytest tests/test_assignment_creator.py::test_name -v`
- Check that Python 3.12 is being used: `python --version`

### Dependencies Not Installing

- Check the `environment.yml` file exists and is valid YAML
- Manually update: `conda env update -f environment.yml`
- For pip dependencies: `pip install -r requirements.txt`
- Check the container logs for errors

### GitHub API Issues

- Verify your `GITHUB_TOKEN` has the necessary permissions
- Check that the repository exists and you have access
- Ensure the token has `repo` and `pull_requests` scopes

## Performance Tips

- The devcontainer mounts the `.git` directory for better Git performance
- Files are cached for faster rebuilds
- Use GitHub Codespaces for the best performance with this image

## Recent Improvements

This devcontainer has been optimized for:

- **Centralized Testing**: All test artifacts moved to `tests/` directory for
  better organization
- **Fixture-Based Testing**: Isolated test data in `tests/fixtures/` eliminates
  dependencies on root-level folders
- **Cross-Platform Compatibility**: Path normalization ensures tests pass on
  Windows, macOS, and Linux
- **Python 3.12**: Updated to latest Python version for improved performance and
  features
- **Consolidated Documentation**: Comprehensive test documentation in
  `tests/README.md`
- **CI/CD Optimization**: GitHub Actions workflows updated for the new structure

## Examples and Documentation

- **Usage Examples**: See `/examples/` directory for workflow examples and usage
  patterns
- **Test Documentation**: Comprehensive guide in `/tests/README.md`
- **API Documentation**: Detailed docstrings in `create_assignment_prs.py`
- **Action Configuration**: See `action.yml` for input/output specifications
