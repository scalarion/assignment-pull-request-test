package main

import (
	"fmt"
	"log"
	"os"

	"assignment-pull-request/internal/assignment"
	"assignment-pull-request/internal/checkout"
	"assignment-pull-request/internal/git"
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
		return
	}

	// Get current branch name for display
	gitOps := git.NewOperations(false)
	currentBranch, err := gitOps.GetCurrentBranch()
	if err != nil {
		log.Printf("Failed to get current branch: %v", err)
		return
	}

	fmt.Printf("Post-checkout hook: switched from %s to %s (branch: %s)\n", oldRef[:8], newRef[:8], currentBranch)

	// Parse workflow files to find assignment configurations
	workflowProcessor := workflow.New()
	err = workflowProcessor.ParseAllFiles()
	if err != nil {
		log.Printf("Failed to parse workflow files: %v", err)
		os.Exit(0)
	}

	// Process the configuration
	err = processAssignmentBranch(workflowProcessor)
	if err != nil {
		log.Printf("Error processing assignments: %v", err)
	}
}

// processAssignmentBranch handles the assignment branch logic
func processAssignmentBranch(workflowProcessor *workflow.Processor) error {
	// Get processors directly
	rootProcessor := workflowProcessor.RootProcessor()
	assignmentProcessor := workflowProcessor.AssignmentProcessor()

	// Skip operations if no patterns found
	if len(rootProcessor.Patterns()) == 0 || len(assignmentProcessor.Patterns()) == 0 {
		fmt.Println("No assignment patterns found in workflow files, skipping sparse-checkout configuration")
		return nil
	}

	// Get repository root (current working directory)
	repositoryRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Find all assignment folders using assignment package
	assignmentProc, err := assignment.NewProcessor("", rootProcessor, assignmentProcessor)
	if err != nil {
		return fmt.Errorf("failed to create assignment processor: %w", err)
	}

	// Create checkout processor and configure sparse-checkout
	checkoutProcessor := checkout.New(assignmentProc, repositoryRoot)

	// Configure sparse-checkout for the current branch
	err = checkoutProcessor.Configure()
	if err != nil {
		return fmt.Errorf("failed to configure sparse checkout: %w", err)
	}

	return nil
}
