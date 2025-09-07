# Assignment Pull Request Creator - Examples

This file contains various examples of how to use the Assignment Pull Request
Creator action in different scenarios.

## Example 1: Basic Course Setup

For a basic computer science course with numbered assignments:

```yaml
name: Course Assignment Setup
on:
    push:
        branches: [main]
        paths: ["assignments/**"]
    workflow_dispatch:

jobs:
    setup-assignments:
        runs-on: ubuntu-latest
        permissions:
            contents: write
            pull-requests: write

        steps:
            - uses: actions/checkout@v4

            - name: Create assignment pull requests
              uses: majikmate/assignment-pull-request@v1
              with:
                  assignments-folder: "assignments"
                  assignment-regex: '^assignment-\d+$'
                  github-token: ${{ secrets.GITHUB_TOKEN }}
```

**Repository Structure:**

```
course-repo/
├── assignments/
│   ├── assignment-1/
│   ├── assignment-2/
│   └── assignment-3/
└── README.md
```

## Example 2: Weekly Lab Setup

For a course with weekly labs and projects:

```yaml
name: Weekly Lab Setup
on:
    schedule:
        - cron: "0 8 * * 1" # Every Monday at 8 AM
    workflow_dispatch:

jobs:
    setup-labs:
        runs-on: ubuntu-latest
        permissions:
            contents: write
            pull-requests: write

        steps:
            - uses: actions/checkout@v4

            - name: Create lab pull requests
              uses: majikmate/assignment-pull-request@v1
              with:
                  assignments-folder: "labs"
                  assignment-regex: '^(lab|project)-week-\d+$'
                  github-token: ${{ secrets.GITHUB_TOKEN }}
```

**Repository Structure:**

```
lab-course/
├── labs/
│   ├── lab-week-1/
│   ├── lab-week-2/
│   ├── project-week-3/
│   └── lab-week-4/
└── README.md
```

## Example 3: Module-Based Course

For a course organized by modules with nested assignments:

```yaml
name: Module Assignment Setup
on:
    push:
        branches: [main]
        paths: ["modules/**"]

jobs:
    setup-modules:
        runs-on: ubuntu-latest
        permissions:
            contents: write
            pull-requests: write

        steps:
            - uses: actions/checkout@v4

            - name: Create module assignment PRs
              uses: majikmate/assignment-pull-request@v1
              with:
                  assignments-folder: "modules"
                  assignment-regex: '^(assignment|homework|quiz)-\d+$'
                  github-token: ${{ secrets.GITHUB_TOKEN }}
```

**Repository Structure:**

```
modular-course/
├── modules/
│   ├── module-1/
│   │   ├── assignment-1/
│   │   └── homework-1/
│   ├── module-2/
│   │   ├── assignment-2/
│   │   └── quiz-1/
│   └── module-3/
│       └── assignment-3/
└── README.md
```

## Example 4: Multi-Course Repository

For managing multiple courses in one repository:

```yaml
name: Multi-Course Setup
on:
    workflow_dispatch:
        inputs:
            course:
                description: "Course to set up"
                required: true
                type: choice
                options:
                    - cs101
                    - cs201
                    - cs301

jobs:
    setup-course:
        runs-on: ubuntu-latest
        permissions:
            contents: write
            pull-requests: write

        steps:
            - uses: actions/checkout@v4

            - name: Create assignment PRs for CS101
              if: github.event.inputs.course == 'cs101'
              uses: majikmate/assignment-pull-request@v1
              with:
                  assignments-folder: "courses/cs101/assignments"
                  assignment-regex: '^(assignment|lab)-\d+$'
                  github-token: ${{ secrets.GITHUB_TOKEN }}

            - name: Create assignment PRs for CS201
              if: github.event.inputs.course == 'cs201'
              uses: majikmate/assignment-pull-request@v1
              with:
                  assignments-folder: "courses/cs201/projects"
                  assignment-regex: '^project-\d+$'
                  github-token: ${{ secrets.GITHUB_TOKEN }}

            - name: Create assignment PRs for CS301
              if: github.event.inputs.course == 'cs301'
              uses: majikmate/assignment-pull-request@v1
              with:
                  assignments-folder: "courses/cs301/research"
                  assignment-regex: "^research-(proposal|paper|presentation)$"
                  github-token: ${{ secrets.GITHUB_TOKEN }}
```

## Example 5: Conditional Assignment Release

For releasing assignments based on schedule or conditions:

```yaml
name: Scheduled Assignment Release
on:
    schedule:
        # Release assignments every Wednesday at 9 AM
        - cron: "0 9 * * 3"
    workflow_dispatch:

jobs:
    release-assignments:
        runs-on: ubuntu-latest
        permissions:
            contents: write
            pull-requests: write

        steps:
            - uses: actions/checkout@v4

            - name: Get current week
              id: week
              run: |
                  week=$(date +%U)
                  echo "current_week=$week" >> $GITHUB_OUTPUT

            - name: Create assignment PRs
              uses: majikmate/assignment-pull-request@v1
              with:
                  assignments-folder: "weekly-assignments"
                  assignment-regex: "^week-([1-9]|1[0-6])$" # Weeks 1-16
                  github-token: ${{ secrets.GITHUB_TOKEN }}
```

## Example 6: Integration with Issue Creation

Combine assignment PR creation with issue tracking:

```yaml
name: Assignment and Issue Setup
on:
    push:
        branches: [main]
        paths: ["assignments/**"]

jobs:
    setup-assignments:
        runs-on: ubuntu-latest
        permissions:
            contents: write
            pull-requests: write
            issues: write

        steps:
            - uses: actions/checkout@v4

            - name: Create assignment pull requests
              id: create-prs
              uses: majikmate/assignment-pull-request@v1
              with:
                  assignments-folder: "assignments"
                  assignment-regex: '^assignment-\d+$'
                  github-token: ${{ secrets.GITHUB_TOKEN }}

            - name: Create tracking issues
              if: steps.create-prs.outputs.created-pull-requests != '[]'
              run: |
                  echo "Created PRs: ${{ steps.create-prs.outputs.created-pull-requests }}"
                  # Add logic to create issues for tracking student submissions
```

## Regular Expression Examples

Here are common regex patterns for different assignment naming conventions:

| Pattern               | Regex                              | Matches                                   |
| --------------------- | ---------------------------------- | ----------------------------------------- |
| Numbered assignments  | `^assignment-\d+$`                 | assignment-1, assignment-2, assignment-10 |
| Weekly labs           | `^lab-week-\d+$`                   | lab-week-1, lab-week-2                    |
| Homework with numbers | `^hw-\d+$`                         | hw-1, hw-2, hw-15                         |
| Projects by semester  | `^project-(fall\|spring)-\d+$`     | project-fall-1, project-spring-2          |
| Mixed assignments     | `^(assignment\|lab\|project)-\d+$` | assignment-1, lab-2, project-3            |
| Date-based            | `^assignment-\d{4}-\d{2}-\d{2}$`   | assignment-2024-01-15                     |
| Module assignments    | `^module-\d+-assignment-\d+$`      | module-1-assignment-1                     |

## Troubleshooting

### Common Issues

1. **No assignments found**: Check that your regex pattern matches your folder
   names exactly
2. **Permission denied**: Ensure your workflow has the correct permissions
3. **Branch already exists**: The action will skip creating branches that
   already exist
4. **API rate limits**: For large repositories, consider running less frequently

### Debug Mode

Add debug output to your workflow:

```yaml
- name: Debug assignment scanning
  run: |
      echo "Looking for assignments in: ${{ inputs.assignments-folder }}"
      echo "Using regex pattern: ${{ inputs.assignment-regex }}"
      find ${{ inputs.assignments-folder }} -type d -name "*" | head -20
```
