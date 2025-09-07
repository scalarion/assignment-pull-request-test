# Assignment Pull Request Creator

A reusable GitHub Action that automatically scans for assignment folders and
creates pull requests with README files for educational repositories.

## Features

- 🔍 **Smart Scanning**: Recursively scans a specified folder for assignments
  matching a regex pattern
- 🌿 **Branch Management**: Creates branches automatically with sanitized names
- 📝 **README Generation**: Creates template README.md files for each assignment
- 🔄 **Pull Request Creation**: Opens pull requests for assignment review and
  collaboration
- ⚙️ **Configurable**: Customizable folder paths and regex patterns
- 🛡️ **Safe**: Only creates branches/PRs when they don't already exist

## Usage

### Basic Usage

```yaml
name: Create Assignment Pull Requests
on:
    push:
        branches: [main]
    workflow_dispatch:

jobs:
    create-assignments:
        runs-on: ubuntu-latest
        steps:
            - name: Checkout repository
              uses: actions/checkout@v4

            - name: Create assignment pull requests
              uses: majikmate/assignment-pull-request@v1
              with:
                  github-token: ${{ secrets.GITHUB_TOKEN }}
```

### Advanced Usage

```yaml
name: Create Assignment Pull Requests
on:
    push:
        branches: [main]
    workflow_dispatch:

jobs:
    create-assignments:
        runs-on: ubuntu-latest
        steps:
            - name: Checkout repository
              uses: actions/checkout@v4

            - name: Create assignment pull requests
              uses: majikmate/assignment-pull-request@v1
              with:
                  assignments-folder: "coursework"
                  assignment-regex: '^(assignment|homework|lab)-\d+$'
                  github-token: ${{ secrets.GITHUB_TOKEN }}
```

## Inputs

| Input                | Description                                    | Required | Default               |
| -------------------- | ---------------------------------------------- | -------- | --------------------- |
| `assignments-folder` | Root folder containing assignments             | No       | `assignments`         |
| `assignment-regex`   | Regular expression to match assignment folders | No       | `^assignment-\d+$`    |
| `github-token`       | GitHub token for API access                    | Yes      | `${{ github.token }}` |

## Outputs

| Output                  | Description                                          |
| ----------------------- | ---------------------------------------------------- |
| `created-branches`      | JSON array of branch names that were created         |
| `created-pull-requests` | JSON array of pull request numbers that were created |

## How It Works

1. **Scan**: The action recursively scans the specified `assignments-folder` for
   subdirectories
2. **Match**: Each subdirectory name is checked against the `assignment-regex`
   pattern
3. **Sanitize**: Assignment paths are converted to valid branch names by:
   - Removing leading/trailing whitespace
   - Replacing spaces with hyphens
   - Removing slashes
   - Converting to lowercase
   - Removing consecutive/leading/trailing hyphens
4. **Create Branch**: If a branch doesn't exist, create it from the default
   branch
5. **Generate README**: Create a template README.md in the assignment folder
6. **Open PR**: Create a pull request if one doesn't already exist for the
   branch

## Repository Structure

The action expects your repository to have a structure like:

```
your-repo/
├── assignments/              # Default assignments folder
│   ├── assignment-1/         # Matches default regex
│   ├── assignment-2/         # Matches default regex
│   └── week-3/
│       └── assignment-3/     # Nested assignment
├── src/                      # Your other code
└── README.md
```

## Branch Naming

Assignment paths are converted to branch names using these rules:

| Assignment Path           | Branch Name               |
| ------------------------- | ------------------------- |
| `assignment-1`            | `assignment-1`            |
| `week 2/assignment-2`     | `week-2-assignment-2`     |
| `Module 3/Lab Assignment` | `module-3-lab-assignment` |

## Example Workflows

### Weekly Assignment Creation

```yaml
name: Weekly Assignment Setup
on:
    schedule:
        - cron: "0 9 * * 1" # Every Monday at 9 AM
    workflow_dispatch:

jobs:
    setup-assignments:
        runs-on: ubuntu-latest
        steps:
            - uses: actions/checkout@v4
            - uses: majikmate/assignment-pull-request@v1
              with:
                  assignments-folder: "weekly-assignments"
                  assignment-regex: '^week-\d+$'
                  github-token: ${{ secrets.GITHUB_TOKEN }}
```

### Course Module Setup

```yaml
name: Course Module Setup
on:
    push:
        paths:
            - "modules/**"
    workflow_dispatch:

jobs:
    setup-modules:
        runs-on: ubuntu-latest
        steps:
            - uses: actions/checkout@v4
            - uses: majikmate/assignment-pull-request@v1
              with:
                  assignments-folder: "modules"
                  assignment-regex: '^module-\d+-(assignment|lab|project)$'
                  github-token: ${{ secrets.GITHUB_TOKEN }}
```

## Permissions

The action requires the following permissions:

```yaml
permissions:
    contents: write
    pull-requests: write
```

## Error Handling

The action handles various scenarios gracefully:

- ✅ Missing assignments folder (logs warning, continues)
- ✅ No matching assignments found (logs info, exits successfully)
- ✅ Branch already exists (skips branch creation)
- ✅ Pull request already exists (skips PR creation)
- ✅ README already exists (skips file creation)
- ❌ Invalid GitHub token (fails with error)
- ❌ Repository access issues (fails with error)

## Development

### Using Dev Containers (Recommended)

This repository includes a development container configuration for a consistent development environment:

1. **GitHub Codespaces**: Click "Code" → "Codespaces" → "Create codespace on main"
2. **VS Code Dev Containers**: Install the Dev Containers extension, open the repo, and select "Reopen in Container"
3. **Docker CLI**: Run `docker run -it --rm -v $(pwd):/workspace -w /workspace ghcr.io/majikmate/classroom-codespace-image:latest bash`

The devcontainer automatically installs dependencies and configures the development environment. See [.devcontainer/README.md](.devcontainer/README.md) for details.

### Local Testing

1. Clone the repository
2. Set environment variables:
   ```bash
   export GITHUB_TOKEN="your_token"
   export GITHUB_REPOSITORY="owner/repo"
   export ASSIGNMENTS_FOLDER="assignments"
   export ASSIGNMENT_REGEX="^assignment-\d+$"
   ```
3. Run the script:
   ```bash
   python create_assignment_prs.py
   ```

### Dependencies

- Python 3.9+
- PyGithub
- requests

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file
for details.

## Support

If you encounter any issues or have questions:

1. Check the
   [Issues](https://github.com/majikmate/assignment-pull-request/issues) page
2. Create a new issue with detailed information
3. Include your workflow configuration and error logs
