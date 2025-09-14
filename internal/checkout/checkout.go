package checkout

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"assignment-pull-request/internal/assignment"
	"assignment-pull-request/internal/constants"
	"assignment-pull-request/internal/git"
)

// Processor handles Git sparse-checkout configuration based on assignment patterns
type Processor struct {
	assignmentProcessor *assignment.Processor
	gitOps              *git.Operations
}

// New creates a new checkout processor
func New(assignmentProcessor *assignment.Processor) *Processor {
	return &Processor{
		assignmentProcessor: assignmentProcessor,
		gitOps:              git.NewOperations(false), // Not in dry-run mode
	}
}

// NewWithGitOps creates a new checkout processor with custom git operations
func NewWithGitOps(assignmentProcessor *assignment.Processor, gitOps *git.Operations) *Processor {
	return &Processor{
		assignmentProcessor: assignmentProcessor,
		gitOps:              gitOps,
	}
}

// getCurrentBranch returns the name of the currently checked out branch
func (p *Processor) getCurrentBranch() (string, error) {
	return p.gitOps.GetCurrentBranch()
}

// Configure sets up Git sparse-checkout to include all root folders and only
// assignment folders that match the current branch name exactly
func (p *Processor) Configure() error {
	// Get current branch
	currentBranch, err := p.getCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Find all assignments
	allAssignments, err := p.assignmentProcessor.ProcessAssignments()
	if err != nil {
		return fmt.Errorf("failed to find assignments: %w", err)
	}

	// Filter assignments that match the current branch
	var matchingAssignments []string
	for _, assignmentInfo := range allAssignments {
		if assignmentInfo.BranchName == currentBranch {
			matchingAssignments = append(matchingAssignments, assignmentInfo.Path)
		}
	}

	if len(matchingAssignments) == 0 {
		return fmt.Errorf("no assignment folders match branch '%s'", currentBranch)
	}

	// Setup sparse-checkout with the matching assignments
	return p.setupSparseCheckout(matchingAssignments)
}

// ConfigureWithPaths sets up Git sparse-checkout with the provided assignment paths
func (p *Processor) ConfigureWithPaths(assignmentPaths []string) error {
	if len(assignmentPaths) == 0 {
		return fmt.Errorf("no assignment paths provided")
	}

	return p.setupSparseCheckout(assignmentPaths)
}

// GetMatchingAssignments returns the assignment paths that match the current branch
func (p *Processor) GetMatchingAssignments() ([]string, error) {
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

// setupSparseCheckout configures git sparse-checkout for the given assignment paths
func (p *Processor) setupSparseCheckout(assignmentPaths []string) error {
	// Enable sparse-checkout
	if err := p.gitOps.EnableSparseCheckout(); err != nil {
		return err
	}

	// Write sparse-checkout file
	sparseCheckoutPath := constants.SparseCheckoutFile

	// Always include essential files and all root folders
	content := []string{
		"/*",                                                                       // Include all files in root
		"!*/",                                                                      // Exclude all directories
		filepath.ToSlash(filepath.Join(constants.GitHubActionsWorkflowDir, "")), // Include .github directory
		constants.ReadmeFileName,                                                   // Include README
		"*" + constants.MarkdownExtension,                                          // Include all markdown files
	}

	// Add the matching assignment folders
	for _, path := range assignmentPaths {
		content = append(content, filepath.ToSlash(path)+"/")
	}

	contentStr := strings.Join(content, "\n") + "\n"

	err := os.WriteFile(sparseCheckoutPath, []byte(contentStr), 0644)
	if err != nil {
		return fmt.Errorf("failed to write sparse-checkout file: %w", err)
	}

	// Apply sparse-checkout
	return p.gitOps.ApplyCheckout()
}

// Disable disables Git sparse-checkout and checks out all files
func (p *Processor) Disable() error {
	// Disable sparse-checkout
	if err := p.gitOps.DisableSparseCheckout(); err != nil {
		return err
	}

	// Remove sparse-checkout file
	if err := os.Remove(constants.SparseCheckoutFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove sparse-checkout file: %w", err)
	}

	// Apply changes to check out all files
	return p.gitOps.ApplyCheckout()
}