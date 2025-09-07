# Development Container

This repository includes a development container configuration that provides a consistent development environment using the `ghcr.io/majikmate/classroom-codespace-image:latest` image.

## What's Included

The devcontainer includes:

- **Base Image**: `ghcr.io/majikmate/classroom-codespace-image:latest`
- **Python Environment**: Pre-configured Python with common development tools
- **GitHub CLI**: For interacting with GitHub repositories and APIs
- **VS Code Extensions**:
  - Python development tools (Python, Pylint, Flake8, Black formatter)
  - GitHub Actions support
  - YAML and JSON editing support

## Getting Started

### Option 1: GitHub Codespaces (Recommended)

1. Click the "Code" button on the GitHub repository
2. Select "Codespaces" tab
3. Click "Create codespace on main"
4. Wait for the environment to build (first time may take a few minutes)

### Option 2: VS Code with Dev Containers Extension

1. Install the [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) for VS Code
2. Clone this repository locally
3. Open the repository in VS Code
4. When prompted, click "Reopen in Container" or use the command palette: `Dev Containers: Reopen in Container`

### Option 3: Docker CLI

```bash
# Clone the repository
git clone https://github.com/majikmate/assignment-pull-request.git
cd assignment-pull-request

# Build and run the container
docker run -it --rm -v $(pwd):/workspace -w /workspace ghcr.io/majikmate/classroom-codespace-image:latest bash

# Install dependencies
pip install -r requirements.txt
```

## Development Workflow

Once the devcontainer is running:

1. **Install Dependencies**: Dependencies are automatically installed via the `postCreateCommand`
2. **Run Tests**: Execute the local test script
   ```bash
   python3 test_local.py
   ```
3. **Test the Action**: Use the example workflow or test locally
4. **Format Code**: Black formatter is configured and will auto-format Python code
5. **Lint Code**: Pylint and Flake8 are configured for code quality checking

## Environment Variables for Testing

When testing the GitHub Action functionality, you'll need to set these environment variables:

```bash
export GITHUB_TOKEN="your_github_token"
export GITHUB_REPOSITORY="owner/repo"
export ASSIGNMENTS_FOLDER="assignments"
export ASSIGNMENT_REGEX="^assignment-\d+$"
export GITHUB_REF_NAME="main"
```

## Debugging

The devcontainer includes tools for debugging:

- **Python Debugger**: VS Code Python debugging is pre-configured
- **GitHub CLI**: Test GitHub API interactions
- **Terminal Access**: Full bash terminal for running commands

## Customization

To customize the devcontainer:

1. Edit `.devcontainer/devcontainer.json`
2. Add additional VS Code extensions in the `extensions` array
3. Modify Python settings in the `settings` object
4. Add additional tools via the `postCreateCommand`

## Troubleshooting

### Container Won't Start
- Check that Docker is running
- Verify you have access to the `ghcr.io/majikmate/classroom-codespace-image:latest` image
- Try rebuilding the container: `Dev Containers: Rebuild Container`

### Dependencies Not Installing
- Check the `requirements.txt` file exists
- Manually run: `pip install -r requirements.txt`
- Check the container logs for errors

### GitHub API Issues
- Verify your `GITHUB_TOKEN` has the necessary permissions
- Check that the repository exists and you have access
- Ensure the token has `repo` and `pull_requests` scopes

## Performance Tips

- The devcontainer mounts the `.git` directory for better Git performance
- Files are cached for faster rebuilds
- Use GitHub Codespaces for the best performance with this image
