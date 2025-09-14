package creator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"assignment-pull-request/internal/assignment"
	"assignment-pull-request/internal/constants"
	"assignment-pull-request/internal/git"
	"assignment-pull-request/internal/github"
	"assignment-pull-request/internal/regex"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Config holds configuration for the PR creator
type Config struct {
	gitHubToken                   string
	rootPatternProcessor          *regex.PatternProcessor
	assignmentPatternProcessor    *regex.PatternProcessor
	repositoryName                string
	defaultBranch                 string
	dryRun                        bool
}

// NewConfig creates a new Config with the given parameters
func NewConfig(gitHubToken, repositoryName, defaultBranch string, assignmentsRootRegex, assignmentRegex []string, dryRun bool) *Config {
	return &Config{
		gitHubToken:                   gitHubToken,
		repositoryName:                repositoryName,
		defaultBranch:                 defaultBranch,
		rootPatternProcessor:          regex.NewPatternProcessorWithPatterns(assignmentsRootRegex),
		assignmentPatternProcessor:    regex.NewPatternProcessorWithPatterns(assignmentRegex),
		dryRun:                        dryRun,
	}
}

// NewConfigFromEnv creates a new Config from environment variables
func NewConfigFromEnv() *Config {
	// Parse environment variables into string arrays
	rootPatterns := regex.ParseCommaSeparated(getEnvWithDefault(constants.EnvAssignmentsRootRegex, constants.DefaultAssignmentsRootRegex))
	assignmentPatterns := regex.ParseCommaSeparated(getEnvWithDefault(constants.EnvAssignmentRegex, constants.DefaultAssignmentRegex))
	
	// Use NewConfig to create the config with proper validation and initialization
	return NewConfig(
		os.Getenv(constants.EnvGitHubToken),
		os.Getenv(constants.EnvGitHubRepository),
		getEnvWithDefault(constants.EnvDefaultBranch, constants.DefaultBranch),
		rootPatterns,
		assignmentPatterns,
		isDryRun(getEnvWithDefault(constants.EnvDryRun, constants.DefaultDryRun)),
	)
}

// Creator is the main Assignment PR Creator
type Creator struct {
	config              *Config
	gitOps              *git.Operations
	githubClient        *github.Client
	assignmentProcessor *assignment.AssignmentProcessor
	createdBranches     []string
	createdPullRequests []string
	pendingPushes       []string
}

// NewWithConfig creates a new Assignment PR Creator with the given configuration
func NewWithConfig(config *Config) (*Creator, error) {
	if config.gitHubToken == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}
	if config.repositoryName == "" {
		return nil, fmt.Errorf("GITHUB_REPOSITORY environment variable is required")
	}

	// Create assignment processor with pattern processors from config
	assignmentProc, err := assignment.NewAssignmentProcessor("", config.rootPatternProcessor, config.assignmentPatternProcessor)
	if err != nil {
		return nil, fmt.Errorf("failed to create assignment processor: %w", err)
	}

	creator := &Creator{
		config:              config,
		gitOps:              git.NewOperations(config.dryRun),
		githubClient:        github.NewClient(config.gitHubToken, config.repositoryName, config.dryRun),
		assignmentProcessor: assignmentProc,
		createdBranches:     make([]string, 0),
		createdPullRequests: make([]string, 0),
	}

	return creator, nil
}

// NewFromEnv creates a new Assignment PR Creator with environment variables
func NewFromEnv() (*Creator, error) {
	config := NewConfigFromEnv()
	return NewWithConfig(config)
}

// createBranch creates a new branch from the default branch locally
func (c *Creator) createBranch(branchName string) error {
	// First, ensure we're on the default branch
	if err := c.gitOps.SwitchToBranch(c.config.defaultBranch); err != nil {
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
func (c *Creator) createReadme(assignmentPath string) error {
	readmePath := filepath.Join(assignmentPath, constants.ReadmeFileName)
	caser := cases.Title(language.English)
	assignmentTitle := caser.String(strings.ReplaceAll(assignmentPath, string(filepath.Separator), " - "))

	// Create assignment directory if it doesn't exist
	if !c.config.dryRun {
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

		if !c.config.dryRun {
			if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
				return fmt.Errorf("failed to write augmented README: %w", err)
			}
			fmt.Printf("✅ Augmented %s at %s (local)\n", constants.ReadmeFileName, readmePath)
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
		if !c.config.dryRun {
			if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
				return fmt.Errorf("failed to write new README: %w", err)
			}
			fmt.Printf("✅ Created %s at %s (local)\n", constants.ReadmeFileName, readmePath)
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
	if _, err := os.Stat(readmePath); err == nil && !c.config.dryRun {
		commitMessage = fmt.Sprintf("Augment README for assignment %s", assignmentPath)
	}

	return c.gitOps.Commit(commitMessage)
}

// createPullRequest creates a pull request for the assignment branch using GitHub API
func (c *Creator) createPullRequest(assignmentPath, branchName string) error {
	caser := cases.Title(language.English)
	title := fmt.Sprintf("Assignment: %s", caser.String(strings.ReplaceAll(assignmentPath, string(filepath.Separator), " - ")))

	// Try to read instructions.md file for PR body content
	body, err := c.createPullRequestBody(assignmentPath)
	if err != nil {
		return fmt.Errorf("error creating pull request body for '%s': %w", assignmentPath, err)
	}

	prNumber, err := c.githubClient.CreatePullRequest(title, body, branchName, c.config.defaultBranch)
	if err != nil {
		return fmt.Errorf("error creating pull request for '%s': %w", assignmentPath, err)
	}

	c.createdPullRequests = append(c.createdPullRequests, prNumber)
	return nil
}

// createPullRequestBody creates the pull request body content, preferring instructions.md if available
func (c *Creator) createPullRequestBody(assignmentPath string) (string, error) {
	// Try to find instructions.md in the assignment directory
	instructionsPath := c.findInstructionsFile(assignmentPath)

	if instructionsPath != "" {
		content, err := c.readAndProcessInstructions(instructionsPath, assignmentPath)
		if err != nil {
			fmt.Printf("Warning: failed to read instructions file '%s': %v\n", instructionsPath, err)
			fmt.Printf("Falling back to generic template\n")
		} else {
			return content, nil
		}
	}

	// Fall back to generic template
	return c.createGenericPullRequestBody(assignmentPath), nil
}

// findInstructionsFile looks for instructions.md or INSTRUCTIONS.md in the assignment directory
func (c *Creator) findInstructionsFile(assignmentPath string) string {
	candidates := []string{
		filepath.Join(assignmentPath, constants.InstructionsFileName),
		filepath.Join(assignmentPath, constants.InstructionsFileNameUpper),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

// readAndProcessInstructions reads the instructions file and processes image links
func (c *Creator) readAndProcessInstructions(instructionsPath, assignmentPath string) (string, error) {
	content, err := os.ReadFile(instructionsPath)
	if err != nil {
		return "", fmt.Errorf("failed to read instructions file: %w", err)
	}

	processedContent := c.rewriteImageLinks(string(content), assignmentPath)

	// Wrap the content in a nice pull request format
	wrappedContent := fmt.Sprintf(`## Assignment Instructions

%s

---

*This pull request was automatically created by the Assignment Pull Request Creator action.*
*Original instructions from: %s*
`, processedContent, filepath.Base(instructionsPath))

	return wrappedContent, nil
}

// rewriteImageLinks rewrites relative image links to reference the assignment path
func (c *Creator) rewriteImageLinks(content, assignmentPath string) string {
	// Regex to match markdown image syntax: ![alt text](relative/path/to/image)
	imageRegex := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

	return imageRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatches := imageRegex.FindStringSubmatch(match)
		if len(submatches) != 3 {
			return match // Return original if parsing fails
		}

		altText := submatches[1]
		imagePath := submatches[2]

		// Skip if it's already an absolute URL
		if strings.HasPrefix(imagePath, "http://") || strings.HasPrefix(imagePath, "https://") {
			return match
		}

		// Skip if it's already an absolute path from repo root
		if strings.HasPrefix(imagePath, "/") {
			return match
		}

		// Rewrite relative path to be relative to repo root
		rewrittenPath := filepath.Join(assignmentPath, imagePath)
		// Convert to forward slashes for web compatibility (Git/GitHub always uses forward slashes)
		rewrittenPath = filepath.ToSlash(rewrittenPath)

		return fmt.Sprintf("![%s](%s)", altText, rewrittenPath)
	})
}

// createGenericPullRequestBody creates the default generic pull request body
func (c *Creator) createGenericPullRequestBody(assignmentPath string) string {
	return fmt.Sprintf(`## Assignment Pull Request

This pull request contains the setup for the assignment located at
`+"`%s`"+`.

### Changes included:
- ✅ Created `+constants.ReadmeFileName+` with assignment template
- ✅ Set up branch structure for assignment submission

### Next steps:
1. Review the assignment requirements in the `+constants.ReadmeFileName+`
2. Add any additional assignment materials
3. Students can fork this repository and work on their submissions

---

*This pull request was automatically created by the Assignment Pull*
*Request Creator action.*
`, assignmentPath)
}

// BranchToProcess represents a branch that needs processing
type BranchToProcess struct {
	AssignmentPath string
	BranchName     string
}

// processAssignments processes all found assignments and creates branches/PRs as needed
func (c *Creator) processAssignments() error {
	fmt.Printf("Scanning workspace for assignment roots matching '%s'\n", c.assignmentProcessor.GetRootRegexStrings())
	fmt.Printf("Looking for assignments matching '%s'\n", c.assignmentProcessor.GetAssignmentRegexStrings())

	// Use assignment processor to discover and validate assignments
	assignments, err := c.assignmentProcessor.ProcessAssignments()
	if err != nil {
		return err
	}

	if len(assignments) == 0 {
		fmt.Println("No assignments found matching the criteria")
		return nil
	} // Phase 0: Fetch all remote branches to ensure complete local state
	fmt.Println("\n=== Phase 0: Syncing with remote ===")
	if err := c.gitOps.FetchAll(); err != nil {
		fmt.Println("❌ Failed to fetch remote branches, aborting")
		return err
	}

	if err := c.gitOps.GetRemoteBranches(c.config.defaultBranch); err != nil {
		fmt.Println("❌ Failed to setup remote tracking branches")
		return err
	}

	// Return to default branch
	if err := c.gitOps.SwitchToBranch(c.config.defaultBranch); err != nil {
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

	for _, assignmentInfo := range assignments {
		assignmentPath := assignmentInfo.Path
		branchName := assignmentInfo.BranchName

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
			if err := c.createReadme(assignmentPath); err != nil {
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
	if c.config.dryRun {
		fmt.Println("🏃 DRY RUN MODE: Simulating local git operations without making actual changes")
	} else {
		fmt.Println("🔄 LIVE MODE: Using local git operations with atomic remote push")
	}
	fmt.Printf("Repository: %s\n", c.config.repositoryName)
	fmt.Printf("Assignments root regex: %s\n", c.config.rootPatternProcessor.GetPatterns())
	fmt.Printf("Assignment regex: %s\n", c.config.assignmentPatternProcessor.GetPatterns())
	fmt.Printf("Default branch: %s\n", c.config.defaultBranch)
	fmt.Printf("Dry run mode: %t\n", c.config.dryRun)

	if err := c.processAssignments(); err != nil {
		return err
	}

	if err := c.setOutputs(); err != nil {
		return err
	}

	if c.config.dryRun {
		fmt.Println("\n🏃 DRY RUN MODE: Assignment Pull Request Creator simulation completed")
		fmt.Println("In real mode, all local changes would be pushed atomically to remote")
	} else {
		fmt.Println("\nAssignment Pull Request Creator completed successfully")
		fmt.Println("All changes have been pushed to remote repository")
	}

	return nil
}

// getEnvWithDefault returns the environment variable value or a default value if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// isDryRun checks if dry run mode is enabled
func isDryRun(dryRunStr string) bool {
	return strings.ToLower(dryRunStr) == "true" || dryRunStr == "1" || strings.ToLower(dryRunStr) == "yes"
}


