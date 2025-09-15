package checkout

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"assignment-pull-request/internal/assignment"
	"assignment-pull-request/internal/git"
	"assignment-pull-request/internal/workflow"
)

// Processor handles Git sparse-checkout configuration based on assignment patterns
type Processor struct {
	repositoryRoot string
	gitOps         *git.Operations
}

// New creates a new sparse checkout processor
func New(repositoryRoot string) *Processor {
	return &Processor{
		repositoryRoot: repositoryRoot,
		gitOps:         git.NewOperations(false), // Not in dry-run mode
	}
}

// NewWithGitOps creates a new sparse checkout processor with custom git operations
func NewWithGitOps(repositoryRoot string, gitOps *git.Operations) *Processor {
	return &Processor{
		repositoryRoot: repositoryRoot,
		gitOps:         gitOps,
	}
}

// SparseCheckout configures Git sparse-checkout for assignments matching the current branch
// Automatically discovers workflow patterns, finds matching assignments, and sets up sparse-checkout
// to include all non-assignment root folders plus only the assignment folders that match the current branch
func (p *Processor) SparseCheckout() error {
	// Disable sparse-checkout at the very beginning to reset state
	if err := p.gitOps.DisableSparseCheckout(); err != nil {
		// Ignore error if sparse-checkout wasn't enabled
		fmt.Printf("Warning: could not disable sparse-checkout (may not be enabled): %v\n", err)
	}

	// Parse workflow files to find assignment configurations
	workflowProcessor := workflow.New()
	err := workflowProcessor.ParseAllFiles()
	if err != nil {
		fmt.Printf("Failed to parse workflow files: %v\n", err)
		return nil // Don't fail, just skip sparse-checkout configuration
	}

	// Get pattern processors from workflow
	assignmentPattern := workflowProcessor.AssignmentPattern()

	// Skip operations if no patterns found
	if len(assignmentPattern.Patterns()) == 0 {
		fmt.Println("No assignment patterns found in workflow files, skipping sparse-checkout configuration")
		return nil
	}

	// Create assignment processor
	assignmentProcessor, err := assignment.NewProcessor(p.repositoryRoot, assignmentPattern)
	if err != nil {
		return fmt.Errorf("failed to create assignment processor: %w", err)
	}

	// Get matching assignments for current branch
	assignmentPaths, err := p.getMatchingAssignments(assignmentProcessor)
	if err != nil {
		return fmt.Errorf("failed to get matching assignments: %w", err)
	}

	if len(assignmentPaths) == 0 {
		fmt.Printf("No assignment folders match current branch\n")
		return nil
	}

	fmt.Printf("Found %d matching assignment folder(s) for current branch\n", len(assignmentPaths))
	for _, assignmentFolder := range assignmentPaths {
		fmt.Printf("  - %s\n", assignmentFolder)
	}

	// Scan repository root folders
	rootFolders, err := p.scanRepositoryRootFolders()
	if err != nil {
		return fmt.Errorf("failed to scan repository root folders: %w", err)
	}

	// Get all unique assignment root folders by extracting the first directory component from assignment paths
	allAssignments, err := assignmentProcessor.ProcessAssignments()
	if err != nil {
		return fmt.Errorf("failed to process assignments: %w", err)
	}

	assignmentRootFoldersMap := make(map[string]bool)
	for _, assignment := range allAssignments {
		if assignment.Path != "" {
			// Extract root folder from assignment path (e.g., "assignments/hw-1" -> "assignments")
			normalizedPath := filepath.ToSlash(assignment.Path)
			pathParts := strings.Split(normalizedPath, "/")
			if len(pathParts) > 0 {
				rootFolder := pathParts[0]
				if rootFolder != "" {
					assignmentRootFoldersMap[rootFolder] = true
				}
			}
		}
	}

	// Create initial paths list (empty, will be populated with root folders and matching assignments)
	paths := []string{}

	// Add all non-assignment root folders to the sparse-checkout paths
	for _, rootFolder := range rootFolders {
		if !assignmentRootFoldersMap[rootFolder] {
			paths = append(paths, rootFolder)
		}
	}
	fmt.Printf("Debug: Found root folders (excluding assignment roots): %v\n", rootFolders)

	// Add only the assignment folders that match the current branch
	for _, path := range assignmentPaths {
		normalizedPath := filepath.ToSlash(path)
		paths = append(paths, normalizedPath)
	}
	fmt.Printf("Debug: Final paths for sparse-checkout: %v\n", paths)

	// Enable sparse-checkout with cone mode for better performance
	if err := p.gitOps.InitSparseCheckoutCone(); err != nil {
		return fmt.Errorf("failed to enable sparse-checkout with cone mode: %w", err)
	}

	// Configure sparse-checkout with the computed paths
	err = p.gitOps.SetSparseCheckoutPaths(paths)
	if err != nil {
		return fmt.Errorf("failed to configure sparse checkout: %w", err)
	}

	fmt.Printf("Sparse checkout configured for %d assignment folder(s)\n", len(assignmentPaths))
	return nil
}

// getCurrentBranch returns the name of the currently checked out branch
func (p *Processor) getCurrentBranch() (string, error) {
	return p.gitOps.GetCurrentBranch()
}

// getMatchingAssignments returns the assignment paths that match the current branch
func (p *Processor) getMatchingAssignments(assignmentProcessor *assignment.Processor) ([]string, error) {
	// Get current branch
	currentBranch, err := p.getCurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	allAssignments, err := assignmentProcessor.ProcessAssignments()
	if err != nil {
		return nil, fmt.Errorf("failed to find assignments: %w", err)
	}

	var matchingAssignments []string
	for _, assignmentInfo := range allAssignments {
		if assignmentInfo.BranchName == currentBranch {
			matchingAssignments = append(matchingAssignments, assignmentInfo.Path)
		}
	}

	return matchingAssignments, nil
}

// scanRepositoryRootFolders scans the repository root directory and returns all folder names
func (p *Processor) scanRepositoryRootFolders() ([]string, error) {
	entries, err := os.ReadDir(p.repositoryRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to read repository root directory: %w", err)
	}

	var folders []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".git") {
			// Skip .git directory but include other hidden directories like .github
			folders = append(folders, entry.Name())
		}
	}

	return folders, nil
}
