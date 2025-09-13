package creator

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"assignment-pull-request/internal/git"
	"assignment-pull-request/internal/github"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Config holds configuration for the PR creator
type Config struct {
	GitHubToken          string
	AssignmentsRootRegex []string
	AssignmentRegex      []string
	RepositoryName       string
	DefaultBranch        string
	DryRun               bool
}

// Creator is the main Assignment PR Creator
type Creator struct {
	config               *Config
	gitOps               *git.Operations
	githubClient         *github.Client
	rootPatterns         []*regexp.Regexp
	assignmentPatterns   []*regexp.Regexp
	createdBranches      []string
	createdPullRequests  []string
	pendingPushes        []string
}

// New creates a new Assignment PR Creator with environment variables
func New() (*Creator, error) {
	config := &Config{
		GitHubToken:          os.Getenv("GITHUB_TOKEN"),
		RepositoryName:       os.Getenv("GITHUB_REPOSITORY"),
		AssignmentsRootRegex: parseRegexPatterns(getEnvWithDefault("ASSIGNMENTS_ROOT_REGEX", "^assignments$")),
		AssignmentRegex:      parseRegexPatterns(getEnvWithDefault("ASSIGNMENT_REGEX", `^(?P<branch>assignment-\d+)$`)),
		DefaultBranch:        getEnvWithDefault("DEFAULT_BRANCH", "main"),
		DryRun:               isDryRun(getEnvWithDefault("DRY_RUN", "false")),
	}

	if config.GitHubToken == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}
	if config.RepositoryName == "" {
		return nil, fmt.Errorf("GITHUB_REPOSITORY environment variable is required")
	}

	// Compile regex patterns
	rootPatterns := make([]*regexp.Regexp, 0, len(config.AssignmentsRootRegex))
	for _, pattern := range config.AssignmentsRootRegex {
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid assignments root regex '%s': %w", pattern, err)
		}
		rootPatterns = append(rootPatterns, compiled)
	}

	assignmentPatterns := make([]*regexp.Regexp, 0, len(config.AssignmentRegex))
	for _, pattern := range config.AssignmentRegex {
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid assignment regex '%s': %w", pattern, err)
		}
		assignmentPatterns = append(assignmentPatterns, compiled)
	}

	creator := &Creator{
		config:             config,
		gitOps:             git.NewOperations(config.DryRun),
		githubClient:       github.NewClient(config.GitHubToken, config.RepositoryName, config.DryRun),
		rootPatterns:       rootPatterns,
		assignmentPatterns: assignmentPatterns,
		createdBranches:    make([]string, 0),
		createdPullRequests: make([]string, 0),
		pendingPushes:      make([]string, 0),
	}

	return creator, nil
}

// getEnvWithDefault returns the environment variable value or a default value if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseRegexPatterns parses a comma-separated string of regex patterns into a slice
func parseRegexPatterns(patterns string) []string {
	if patterns == "" {
		return []string{}
	}
	
	// Split by comma and trim whitespace
	parts := strings.Split(patterns, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// isDryRun checks if dry run mode is enabled
func isDryRun(dryRunStr string) bool {
	return strings.ToLower(dryRunStr) == "true" || dryRunStr == "1" || strings.ToLower(dryRunStr) == "yes"
}

// extractBranchName tries to match assignment path against patterns and extract branch name
// Returns the extracted branch name and true if matched, empty string and false if no match
func (c *Creator) extractBranchName(assignmentPath string) (string, bool) {
	for _, pattern := range c.assignmentPatterns {
		matches := pattern.FindStringSubmatch(assignmentPath)
		if matches != nil {
			names := pattern.SubexpNames()
			var branchParts []string
			
			// Look for named groups and collect them
			for i, name := range names {
				if name != "" && i < len(matches) && matches[i] != "" {
					part := strings.TrimSpace(matches[i])
					if part != "" {
						branchParts = append(branchParts, part)
					}
				}
			}
			
			// If we found named groups, combine them
			if len(branchParts) > 0 {
				branchName := strings.Join(branchParts, "-")
				// Sanitize the extracted branch name
				branchName = strings.ToLower(branchName)
				branchName = regexp.MustCompile(`[^a-z0-9\-]`).ReplaceAllString(branchName, "-")
				branchName = regexp.MustCompile(`-+`).ReplaceAllString(branchName, "-")
				branchName = strings.Trim(branchName, "-")
				return branchName, true
			}
			
			// If no named groups found, look for "branch" specifically
			for i, name := range names {
				if name == "branch" && i < len(matches) {
					branchName := strings.TrimSpace(matches[i])
					if branchName != "" {
						branchName = strings.ToLower(branchName)
						branchName = regexp.MustCompile(`[^a-z0-9\-]`).ReplaceAllString(branchName, "-")
						branchName = regexp.MustCompile(`-+`).ReplaceAllString(branchName, "-")
						branchName = strings.Trim(branchName, "-")
						return branchName, true
					}
				}
			}
			
			// Fall back to using the entire match
			if len(matches) > 0 {
				branchName := strings.TrimSpace(matches[0])
				if branchName != "" {
					branchName = c.sanitizeBranchName(branchName)
					return branchName, true
				}
			}
		}
	}
	return "", false
}

// sanitizeBranchName sanitizes assignment path to create a valid branch name
func (c *Creator) sanitizeBranchName(assignmentPath string) string {
	// Remove leading/trailing whitespace
	branchName := strings.TrimSpace(assignmentPath)

	// Replace spaces with hyphens
	branchName = regexp.MustCompile(`\s+`).ReplaceAllString(branchName, "-")

	// Remove slashes
	branchName = strings.ReplaceAll(branchName, "/", "-")

	// Remove consecutive hyphens
	branchName = regexp.MustCompile(`-+`).ReplaceAllString(branchName, "-")

	// Convert to lowercase
	branchName = strings.ToLower(branchName)

	// Remove leading/trailing hyphens
	branchName = strings.Trim(branchName, "-")

	return branchName
}

// createBranch creates a new branch from the default branch locally
func (c *Creator) createBranch(branchName string) error {
	// First, ensure we're on the default branch
	if err := c.gitOps.SwitchToBranch(c.config.DefaultBranch); err != nil {
		return err
	}

	// Create and switch to new branch
	if err := c.gitOps.CreateAndSwitchToBranch(branchName); err != nil {
		return err
	}

	fmt.Printf("✅ Created branch: %s (local)\n", branchName)
	c.createdBranches = append(c.createdBranches, branchName)
	c.pendingPushes = append(c.pendingPushes, branchName)
	return nil
}

// createReadme creates or augments README.md file in the assignment folder locally
func (c *Creator) createReadme(assignmentPath, branchName string) error {
	readmePath := filepath.Join(assignmentPath, "README.md")
	caser := cases.Title(language.English)
	assignmentTitle := caser.String(strings.ReplaceAll(assignmentPath, "/", " - "))

	// Create assignment directory if it doesn't exist
	if !c.config.DryRun {
		if err := os.MkdirAll(assignmentPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", assignmentPath, err)
		}
	} else {
		fmt.Printf("[DRY RUN] Would create directory: mkdir -p %s\n", assignmentPath)
	}

	// Check if README already exists
	var readmeContent string
	if _, err := os.Stat(readmePath); err == nil {
		fmt.Printf("README already exists at %s, augmenting...\n", readmePath)

		// Read existing content
		existingBytes, err := os.ReadFile(readmePath)
		if err != nil {
			return fmt.Errorf("failed to read existing README: %w", err)
		}
		existingContent := string(existingBytes)

		// Add workflow augmentation comment
		augmentationComment := `

---

*This README was augmented by the Assignment Pull Request Creator action.*
`

		readmeContent = strings.TrimSpace(existingContent) + augmentationComment

		if !c.config.DryRun {
			if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
				return fmt.Errorf("failed to write augmented README: %w", err)
			}
			fmt.Printf("✅ Augmented README.md at %s (local)\n", readmePath)
		} else {
			fmt.Printf("[DRY RUN] Would augment README at %s\n", readmePath)
			fmt.Printf("[DRY RUN] Augmentation content:\n%s\n", augmentationComment)
		}
	} else {
		// Create new README content
		readmeContent = fmt.Sprintf(`# %s

This is the README for the assignment located at `+"`%s`"+`.

## Instructions

Please add your assignment instructions and requirements here.

## Submission

Please add your submission guidelines here.

---

*This README was automatically generated by the Assignment Pull Request*
*Creator action.*
`, assignmentTitle, assignmentPath)

		// Write new README
		if !c.config.DryRun {
			if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
				return fmt.Errorf("failed to write new README: %w", err)
			}
			fmt.Printf("✅ Created README.md at %s (local)\n", readmePath)
		} else {
			fmt.Printf("[DRY RUN] Would create README at %s\n", readmePath)
			fmt.Printf("[DRY RUN] README content:\n%s\n", readmeContent)
		}
	}

	// Add and commit the README
	if err := c.gitOps.AddFile(readmePath); err != nil {
		return err
	}

	commitMessage := fmt.Sprintf("Add README for assignment %s", assignmentPath)
	if _, err := os.Stat(readmePath); err == nil && !c.config.DryRun {
		commitMessage = fmt.Sprintf("Augment README for assignment %s", assignmentPath)
	}

	return c.gitOps.Commit(commitMessage)
}

// createPullRequest creates a pull request for the assignment branch using GitHub API
func (c *Creator) createPullRequest(assignmentPath, branchName string) error {
	caser := cases.Title(language.English)
	title := fmt.Sprintf("Assignment: %s", caser.String(strings.ReplaceAll(assignmentPath, "/", " - ")))
	body := fmt.Sprintf(`## Assignment Pull Request

This pull request contains the setup for the assignment located at
`+"`%s`"+`.

### Changes included:
- ✅ Created README.md with assignment template
- ✅ Set up branch structure for assignment submission

### Next steps:
1. Review the assignment requirements in the README.md
2. Add any additional assignment materials
3. Students can fork this repository and work on their submissions

---

*This pull request was automatically created by the Assignment Pull*
*Request Creator action.*
`, assignmentPath)

	prNumber, err := c.githubClient.CreatePullRequest(title, body, branchName, c.config.DefaultBranch)
	if err != nil {
		return fmt.Errorf("error creating pull request for '%s': %w", assignmentPath, err)
	}

	c.createdPullRequests = append(c.createdPullRequests, prNumber)
	return nil
}

// findAssignments finds all assignment folders that match the regex patterns
func (c *Creator) findAssignments() ([]string, error) {
	var assignments []string
	workspaceRoot, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	fmt.Printf("Scanning workspace for assignment roots matching '%s'\n", c.config.AssignmentsRootRegex)
	fmt.Printf("Looking for assignments matching '%s'\n", c.config.AssignmentRegex)

	// Walk through the directory tree
	err = filepath.WalkDir(workspaceRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			return nil
		}

		dirName := d.Name()

		// Check if this directory matches any of the root patterns
		matchesRootPattern := false
		for _, pattern := range c.rootPatterns {
			if pattern.MatchString(dirName) {
				matchesRootPattern = true
				break
			}
		}
		
		if matchesRootPattern {
			fmt.Printf("Found assignment root: %s\n", path)

			// Now scan for individual assignments within this root
			return filepath.WalkDir(path, func(assignmentPath string, assignmentD fs.DirEntry, assignmentErr error) error {
				if assignmentErr != nil {
					return assignmentErr
				}

				if !assignmentD.IsDir() {
					return nil
				}

				// Skip the root assignments directory itself
				if assignmentPath == path {
					return nil
				}

				// Get relative path from workspace root for pattern matching
				relativePath, err := filepath.Rel(workspaceRoot, assignmentPath)
				if err != nil {
					return err
				}

				// Try to extract branch name from the full path
				if branchName, matched := c.extractBranchName(relativePath); matched {
					assignments = append(assignments, relativePath)
					fmt.Printf("Found assignment: %s (branch: %s)\n", relativePath, branchName)
				}

				return nil
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory tree: %w", err)
	}

	return assignments, nil
}

// BranchToProcess represents a branch that needs processing
type BranchToProcess struct {
	AssignmentPath string
	BranchName     string
}

// processAssignments processes all found assignments and creates branches/PRs as needed
func (c *Creator) processAssignments() error {
	assignments, err := c.findAssignments()
	if err != nil {
		return err
	}

	if len(assignments) == 0 {
		fmt.Println("No assignments found matching the criteria")
		return nil
	}

	// Phase 0: Fetch all remote branches to ensure complete local state
	fmt.Println("\n=== Phase 0: Syncing with remote ===")
	if err := c.gitOps.FetchAll(); err != nil {
		fmt.Println("❌ Failed to fetch remote branches, aborting")
		return err
	}

	if err := c.gitOps.GetRemoteBranches(c.config.DefaultBranch); err != nil {
		fmt.Println("❌ Failed to setup remote tracking branches")
		return err
	}

	// Return to default branch
	if err := c.gitOps.SwitchToBranch(c.config.DefaultBranch); err != nil {
		return err
	}

	// Phase 1: Get current state after sync
	existingBranches, err := c.gitOps.GetLocalBranches()
	if err != nil {
		return err
	}

	existingPRs, err := c.githubClient.GetExistingPullRequests()
	if err != nil {
		return err
	}

	fmt.Printf("Found %d assignments to process\n", len(assignments))
	fmt.Printf("Existing local branches: %d\n", len(existingBranches))
	fmt.Printf("Existing PRs: %d\n", len(existingPRs))

	// Phase 2: Process all assignments locally
	fmt.Println("\n=== Phase 2: Local processing ===")
	var branchesToProcess []BranchToProcess

	for _, assignmentPath := range assignments {
		branchName, matched := c.extractBranchName(assignmentPath)
		if !matched {
			fmt.Printf("Skipping assignment %s: no regex pattern matched\n", assignmentPath)
			continue
		}

		fmt.Printf("\nProcessing assignment: %s\n", assignmentPath)
		fmt.Printf("Branch name: %s\n", branchName)

		// Check if branch exists locally and if PR exists (or has ever existed)
		branchExists := existingBranches[branchName]
		_, prHasExisted := existingPRs[branchName]

		// Only create branch if:
		// 1. Branch doesn't exist locally AND
		// 2. No PR has ever existed for this branch name
		if !branchExists && !prHasExisted {
			fmt.Printf("Branch '%s' does not exist locally and no PR has ever existed, creating locally...\n", branchName)

			// Create branch locally
			if err := c.createBranch(branchName); err != nil {
				fmt.Printf("❌ Failed to create branch '%s', skipping: %v\n", branchName, err)
				continue
			}

			// Create README content locally
			fmt.Printf("Creating README content for assignment '%s'...\n", assignmentPath)
			if err := c.createReadme(assignmentPath, branchName); err != nil {
				fmt.Printf("❌ Failed to create README for '%s', skipping: %v\n", assignmentPath, err)
				continue
			}

			branchesToProcess = append(branchesToProcess, BranchToProcess{
				AssignmentPath: assignmentPath,
				BranchName:     branchName,
			})

		} else if !branchExists && prHasExisted {
			fmt.Printf("Branch '%s' does not exist but PR has existed before (likely merged and branch deleted), skipping\n", branchName)
			continue
		} else if branchExists && !prHasExisted {
			fmt.Printf("Branch '%s' already exists locally but no PR has ever existed, will create PR\n", branchName)
			branchesToProcess = append(branchesToProcess, BranchToProcess{
				AssignmentPath: assignmentPath,
				BranchName:     branchName,
			})
		} else if branchExists && prHasExisted {
			fmt.Printf("Branch '%s' already exists locally and PR has existed before, skipping\n", branchName)
		}
	}

	// Phase 3: Push all changes atomically to remote
	if len(branchesToProcess) > 0 {
		fmt.Printf("\n=== Phase 3: Atomic push to remote ===\n")
		fmt.Printf("Pushing %d branches to remote...\n", len(branchesToProcess))

		if len(c.pendingPushes) > 0 {
			fmt.Printf("Pushing all local branches (including %d new branches) to remote atomically...\n", len(c.pendingPushes))

			if err := c.gitOps.PushAllBranches(); err != nil {
				fmt.Println("❌ Failed to push branches to remote, aborting PR creation")
				return err
			}

			fmt.Printf("✅ Successfully pushed all local branches to remote atomically\n")
			c.pendingPushes = c.pendingPushes[:0] // Clear the slice
		} else {
			fmt.Println("No branches to push to remote")
		}
	}

	// Phase 4: Create pull requests
	if len(branchesToProcess) > 0 {
		fmt.Printf("\n=== Phase 4: Pull request creation ===\n")

		for _, branch := range branchesToProcess {
			_, prHasExisted := existingPRs[branch.BranchName]

			// Double-check PR status (should still be false)
			if !prHasExisted {
				fmt.Printf("Creating pull request for branch '%s'...\n", branch.BranchName)
				if err := c.createPullRequest(branch.AssignmentPath, branch.BranchName); err != nil {
					fmt.Printf("❌ Failed to create PR for '%s': %v\n", branch.BranchName, err)
					continue
				}
			} else {
				fmt.Printf("PR has existed for branch '%s', skipping PR creation\n", branch.BranchName)
			}
		}
	}

	if len(branchesToProcess) == 0 {
		fmt.Println("\n=== No new assignments to process ===")
		fmt.Println("All assignments either already have branches or have had PRs created previously")
	}

	return nil
}

// setOutputs sets GitHub Actions outputs
func (c *Creator) setOutputs() error {
	// Set outputs for GitHub Actions
	if githubOutput := os.Getenv("GITHUB_OUTPUT"); githubOutput != "" {
		file, err := os.OpenFile(githubOutput, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open GITHUB_OUTPUT file: %w", err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				fmt.Printf("Warning: failed to close output file: %v\n", err)
			}
		}()

		branchesJSON, err := json.Marshal(c.createdBranches)
		if err != nil {
			return fmt.Errorf("failed to marshal created branches: %w", err)
		}

		prsJSON, err := json.Marshal(c.createdPullRequests)
		if err != nil {
			return fmt.Errorf("failed to marshal created pull requests: %w", err)
		}

		if _, err := fmt.Fprintf(file, "created-branches=%s\n", branchesJSON); err != nil {
			return fmt.Errorf("failed to write created-branches output: %w", err)
		}

		if _, err := fmt.Fprintf(file, "created-pull-requests=%s\n", prsJSON); err != nil {
			return fmt.Errorf("failed to write created-pull-requests output: %w", err)
		}
	}

	fmt.Println("\nSummary:")
	fmt.Printf("Created branches: %v\n", c.createdBranches)
	fmt.Printf("Created pull requests: %v\n", c.createdPullRequests)

	return nil
}

// Run is the main execution method using local git with atomic remote operations
func (c *Creator) Run() error {
	fmt.Println("Starting Assignment Pull Request Creator")
	if c.config.DryRun {
		fmt.Println("🏃 DRY RUN MODE: Simulating local git operations without making actual changes")
	} else {
		fmt.Println("🔄 LIVE MODE: Using local git operations with atomic remote push")
	}
	fmt.Printf("Repository: %s\n", c.config.RepositoryName)
	fmt.Printf("Assignments root regex: %s\n", c.config.AssignmentsRootRegex)
	fmt.Printf("Assignment regex: %s\n", c.config.AssignmentRegex)
	fmt.Printf("Default branch: %s\n", c.config.DefaultBranch)
	fmt.Printf("Dry run mode: %t\n", c.config.DryRun)

	if err := c.processAssignments(); err != nil {
		return err
	}

	if err := c.setOutputs(); err != nil {
		return err
	}

	if c.config.DryRun {
		fmt.Println("\n🏃 DRY RUN MODE: Assignment Pull Request Creator simulation completed")
		fmt.Println("In real mode, all local changes would be pushed atomically to remote")
	} else {
		fmt.Println("\nAssignment Pull Request Creator completed successfully")
		fmt.Println("All changes have been pushed to remote repository")
	}

	return nil
}
