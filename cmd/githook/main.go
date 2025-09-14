package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"assignment-pull-request/internal/assignment"
	"assignment-pull-request/internal/constants"
	"assignment-pull-request/internal/regex"
	"assignment-pull-request/internal/workflow"
)

func main() {
	// Check if this is a post-checkout hook call
	if len(os.Args) < 4 {
		log.Fatal("Usage: post-checkout <old-ref> <new-ref> <branch-checkout-flag>")
	}

	oldRef := os.Args[1]
	newRef := os.Args[2]
	branchCheckout := os.Args[3]

	// Only process branch checkouts
	if branchCheckout != "1" {
		os.Exit(0)
	}

	// Get current branch name
	currentBranch, err := getCurrentBranch()
	if err != nil {
		log.Printf("Failed to get current branch: %v", err)
		os.Exit(0)
	}

	fmt.Printf("Post-checkout hook: switched from %s to %s (branch: %s)\n", oldRef[:8], newRef[:8], currentBranch)

	// Parse workflow files to find assignment configurations
	patterns, err := workflow.ParseAllWorkflows()
	if err != nil {
		log.Printf("Failed to parse workflow files: %v", err)
		os.Exit(0)
	}

	if len(patterns.RootPatterns) == 0 || len(patterns.AssignmentPatterns) == 0 {
		fmt.Println("No assignment-pull-request action configurations found")
		os.Exit(0)
	}

	// Process the configuration
	err = processAssignmentBranch(currentBranch, patterns)
	if err != nil {
		log.Printf("Error processing assignments: %v", err)
	}
}

// getCurrentBranch returns the name of the currently checked out branch
func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// processAssignmentBranch handles the assignment branch logic
func processAssignmentBranch(currentBranch string, patterns *workflow.WorkflowPatterns) error {
	// Use default patterns if none found
	rootPatterns := patterns.RootPatterns
	assignmentPatterns := patterns.AssignmentPatterns

	if len(rootPatterns) == 0 {
		rootPatterns = []string{constants.DefaultAssignmentsRootRegex}
	}
	if len(assignmentPatterns) == 0 {
		assignmentPatterns = []string{constants.DefaultAssignmentRegex}
	}

		// Compile regex patterns using the regex processor
	rootProcessor := regex.NewPatternProcessorWithPatterns(rootPatterns)
	assignmentProcessor := regex.NewPatternProcessorWithPatterns(assignmentPatterns)

	// Find all assignment folders using assignment package
	processor, err := assignment.NewAssignmentProcessor("", rootProcessor, assignmentProcessor)
	if err != nil {
		return fmt.Errorf("failed to create assignment processor: %w", err)
	}

	allAssignments, err := processor.ProcessAssignments()
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
		fmt.Printf("No assignment folders match branch '%s'\n", currentBranch)
		return nil
	}

	fmt.Printf("Found %d matching assignment folder(s) for branch '%s'\n", len(matchingAssignments), currentBranch)
	for _, assignmentFolder := range matchingAssignments {
		fmt.Printf("  - %s\n", assignmentFolder)
	}

	// Setup sparse-checkout for the matching assignments
	err = setupSparseCheckout(matchingAssignments)
	if err != nil {
		return fmt.Errorf("failed to setup sparse checkout: %w", err)
	}

	fmt.Printf("Sparse checkout configured for %d assignment folder(s)\n", len(matchingAssignments))
	return nil
}

// setupSparseCheckout configures git sparse-checkout for the given paths
func setupSparseCheckout(paths []string) error {
	// Enable sparse-checkout
	cmd := exec.Command("git", "config", "core.sparseCheckout", "true")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable sparse-checkout: %w", err)
	}

	// Write sparse-checkout file
	sparseCheckoutPath := constants.SparseCheckoutFile

	// Always include essential files
	content := []string{
		"/*",                                                                       // Include all files in root
		"!*/",                                                                      // Exclude all directories
		filepath.ToSlash(filepath.Join(constants.GitHubActionsWorkflowDir, "")), // Include .github directory
		constants.ReadmeFileName,                                                   // Include README
		"*" + constants.MarkdownExtension,                                          // Include all markdown files
	}

	// Add the matching assignment folders
	for _, path := range paths {
		content = append(content, filepath.ToSlash(path)+"/")
	}

	contentStr := strings.Join(content, "\n") + "\n"

	err := os.WriteFile(sparseCheckoutPath, []byte(contentStr), 0644)
	if err != nil {
		return fmt.Errorf("failed to write sparse-checkout file: %w", err)
	}

	// Apply sparse-checkout
	cmd = exec.Command("git", "read-tree", "-m", "-u", "HEAD")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply sparse-checkout: %w", err)
	}

	return nil
}
