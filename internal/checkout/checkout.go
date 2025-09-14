package checkout

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"assignment-pull-request/internal/assignment"
	"assignment-pull-request/internal/git"
)

// Processor handles Git sparse-checkout configuration based on assignment patterns
type Processor struct {
	assignmentProcessor *assignment.Processor
	gitOps              *git.Operations
	repositoryRoot      string
}

// New creates a new checkout processor
func New(repositoryRoot string, assignmentProcessor *assignment.Processor) *Processor {
	return &Processor{
		assignmentProcessor: assignmentProcessor,
		gitOps:              git.NewOperations(false), // Not in dry-run mode
		repositoryRoot:      repositoryRoot,
	}
}

// NewWithGitOps creates a new checkout processor with custom git operations
func NewWithGitOps(assignmentProcessor *assignment.Processor, gitOps *git.Operations, repositoryRoot string) *Processor {
	return &Processor{
		assignmentProcessor: assignmentProcessor,
		gitOps:              gitOps,
		repositoryRoot:      repositoryRoot,
	}
}

// configureWithPaths sets up Git sparse-checkout with the provided assignment paths
// If assignmentPaths is nil or empty, it will automatically get matching assignments for the current branch
func (p *Processor) Configure() error {
	// Disable sparse-checkout at the very beginning to reset state
	if err := p.gitOps.DisableSparseCheckout(); err != nil {
		// Ignore error if sparse-checkout wasn't enabled
		fmt.Printf("Warning: could not disable sparse-checkout (may not be enabled): %v\n", err)
	}

	// Get matching assignments for current branch
	assignmentPaths, err := p.getMatchingAssignments()
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
	allAssignments, err := p.assignmentProcessor.ProcessAssignments()
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

	paths := []string{}

	// Add all root folders to paths, but exclude assignment root folders
	for _, rootFolder := range rootFolders {
		if !assignmentRootFoldersMap[rootFolder] {
			paths = append(paths, rootFolder)
		}
	}
	fmt.Printf("Debug: Found root folders (excluding assignment roots): %v\n", rootFolders)

	// Add only the assignment folders that match the current branch name
	for _, path := range assignmentPaths {
		normalizedPath := filepath.ToSlash(path)
		paths = append(paths, normalizedPath)
	}
	fmt.Printf("Debug: Final paths for sparse-checkout: %v\n", paths)

	// Enable sparse-checkout with cone mode
	if err := p.gitOps.InitSparseCheckoutCone(); err != nil {
		return fmt.Errorf("failed to enable sparse-checkout with cone mode: %w", err)
	}

	// Use git sparse-checkout set command for cone mode
	err = p.gitOps.SetSparseCheckoutPaths(paths)
	if err != nil {
		return fmt.Errorf("failed to setup sparse checkout: %w", err)
	}

	fmt.Printf("Sparse checkout configured for %d assignment folder(s)\n", len(assignmentPaths))
	return nil
}

// getCurrentBranch returns the name of the currently checked out branch
func (p *Processor) getCurrentBranch() (string, error) {
	return p.gitOps.GetCurrentBranch()
}

// getMatchingAssignments returns the assignment paths that match the current branch
func (p *Processor) getMatchingAssignments() ([]string, error) {
	// Get current branch
	currentBranch, err := p.getCurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	allAssignments, err := p.assignmentProcessor.ProcessAssignments()
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
			// Skip .git
			folders = append(folders, entry.Name())
		}
	}

	return folders, nil
}
