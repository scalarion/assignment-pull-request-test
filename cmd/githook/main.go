package main

import (
	"fmt"
	"log"
	"os"

	"assignment-pull-request/internal/assignment"
	"assignment-pull-request/internal/checkout"
	"assignment-pull-request/internal/workflow"
)

func main() {
	// Check if this is a post-checkout hook call
	if len(os.Args) < 4 {
		log.Fatal("Usage: post-checkout <old-ref> <new-ref> <branch-checkout-flag>")
	}

	branchCheckout := os.Args[3]

	// Only process branch checkouts
	if branchCheckout != "1" {
		return
	}

	// Get repository root (current working directory)
	repositoryRoot, err := os.Getwd()
	if err != nil {
		return
	}

	// Process the configuration
	err = processAssignmentBranch(repositoryRoot)
	if err != nil {
		log.Printf("Error processing assignments: %v", err)
	}
}

// processAssignmentBranch handles the assignment branch logic
func processAssignmentBranch(repositoryRoot string) error {
	// Parse workflow files to find assignment configurations
	workflowProcessor := workflow.New()
	err := workflowProcessor.ParseAllFiles()
	if err != nil {
		log.Printf("Failed to parse workflow files: %v", err)
		os.Exit(0)
	}

	// Get processors directly
	rootProcessor := workflowProcessor.RootPattern()
	assignmentProcessor := workflowProcessor.AssignmentPattern()

	// Skip operations if no patterns found
	if len(rootProcessor.Patterns()) == 0 || len(assignmentProcessor.Patterns()) == 0 {
		fmt.Println("No assignment patterns found in workflow files, skipping sparse-checkout configuration")
		return nil
	}

	// Find all assignment folders using assignment package
	assignmentProc, err := assignment.NewProcessor(repositoryRoot, rootProcessor, assignmentProcessor)
	if err != nil {
		return fmt.Errorf("failed to create assignment processor: %w", err)
	}

	// Create checkout processor and configure sparse-checkout
	checkoutProcessor := checkout.New(repositoryRoot, assignmentProc)

	// Configure sparse-checkout for the current branch
	err = checkoutProcessor.Configure()
	if err != nil {
		return fmt.Errorf("failed to configure sparse checkout: %w", err)
	}

	return nil
}
