// Package testutil provides common utilities for testing the assignment-pull-request system
package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// TempWorkspace creates a temporary workspace with a common directory structure for testing
type TempWorkspace struct {
	RootDir string
	t       *testing.T
}

// NewTempWorkspace creates a new temporary workspace for testing
func NewTempWorkspace(t *testing.T) *TempWorkspace {
	tempDir := t.TempDir()
	return &TempWorkspace{
		RootDir: tempDir,
		t:       t,
	}
}

// CreateAssignment creates an assignment directory with instructions.md
func (tw *TempWorkspace) CreateAssignment(path, instructionsContent string) {
	assignmentDir := filepath.Join(tw.RootDir, path)
	err := os.MkdirAll(assignmentDir, 0755)
	if err != nil {
		tw.t.Fatalf("Failed to create assignment directory %s: %v", assignmentDir, err)
	}

	if instructionsContent != "" {
		instructionsFile := filepath.Join(assignmentDir, "instructions.md")
		err = os.WriteFile(instructionsFile, []byte(instructionsContent), 0644)
		if err != nil {
			tw.t.Fatalf("Failed to create instructions file %s: %v", instructionsFile, err)
		}
	}
}

// CreateAssignmentWithImages creates an assignment with static images
func (tw *TempWorkspace) CreateAssignmentWithImages(path, instructionsContent string, images []string) {
	tw.CreateAssignment(path, instructionsContent)

	staticDir := filepath.Join(tw.RootDir, path, "static")
	err := os.MkdirAll(staticDir, 0755)
	if err != nil {
		tw.t.Fatalf("Failed to create static directory %s: %v", staticDir, err)
	}

	for _, img := range images {
		imgPath := filepath.Join(staticDir, img)
		err = os.WriteFile(imgPath, []byte("fake image content"), 0644)
		if err != nil {
			tw.t.Fatalf("Failed to create image %s: %v", imgPath, err)
		}
	}
}

// CreateStandardStructure creates a standard test structure with various assignment types
func (tw *TempWorkspace) CreateStandardStructure() {
	assignments := []struct {
		path         string
		instructions string
		images       []string
	}{
		{
			path: "test/fixtures/assignments/assignment-1",
			instructions: `# Assignment 1
This is a basic assignment.
![Overview](static/overview.png)`,
			images: []string{"overview.png"},
		},
		{
			path: "test/fixtures/assignments/assignment-2",
			instructions: `# Assignment 2
This is another assignment.
![Diagram](static/diagram.jpg)`,
			images: []string{"diagram.jpg"},
		},
		{
			path: "test/fixtures/homework/hw-1",
			instructions: `# Homework 1
Complete the following tasks.
![Guide](static/homework-guide.png)`,
			images: []string{"homework-guide.png"},
		},
		{
			path: "test/fixtures/homework/hw-2",
			instructions: `# Homework 2
Advanced homework assignment.`,
			images: []string{},
		},
		{
			path: "test/fixtures/labs/lab-1",
			instructions: `# Lab 1
Laboratory assignment.
![Setup](static/lab-setup.png)`,
			images: []string{"lab-setup.png"},
		},
		{
			path: "test/fixtures/projects/project-1",
			instructions: `# Project 1
Final project assignment.
![Architecture](static/architecture.png)
![Flowchart](static/flowchart.png)`,
			images: []string{"architecture.png", "flowchart.png"},
		},
		{
			path: "test/fixtures/courses/CS101/week-01/assignment-fibonacci",
			instructions: `# CS101 - Fibonacci Assignment
Implement the Fibonacci sequence.`,
			images: []string{},
		},
		{
			path: "test/fixtures/courses/CS102/week-02/assignment-sorting",
			instructions: `# CS102 - Sorting Assignment
Implement various sorting algorithms.
![Sorting](static/sorting-visualization.png)`,
			images: []string{"sorting-visualization.png"},
		},
	}

	for _, assignment := range assignments {
		if len(assignment.images) > 0 {
			tw.CreateAssignmentWithImages(assignment.path, assignment.instructions, assignment.images)
		} else {
			tw.CreateAssignment(assignment.path, assignment.instructions)
		}
	}
}

// ChangeToWorkspace changes the current directory to the workspace root
func (tw *TempWorkspace) ChangeToWorkspace() (restore func()) {
	originalDir, err := os.Getwd()
	if err != nil {
		tw.t.Fatalf("Failed to get current directory: %v", err)
	}

	err = os.Chdir(tw.RootDir)
	if err != nil {
		tw.t.Fatalf("Failed to change to workspace directory: %v", err)
	}

	return func() {
		if err := os.Chdir(originalDir); err != nil {
			// Log the error but don't fail the test during cleanup
			tw.t.Logf("Warning: failed to restore original directory: %v", err)
		}
	}
}

// EnvSetup manages environment variable setup for tests
type EnvSetup struct {
	originalVars map[string]string
}

// NewEnvSetup creates a new environment setup manager
func NewEnvSetup() *EnvSetup {
	return &EnvSetup{
		originalVars: make(map[string]string),
	}
}

// Set sets an environment variable and remembers the original value
func (e *EnvSetup) Set(key, value string) {
	if _, exists := e.originalVars[key]; !exists {
		e.originalVars[key] = os.Getenv(key)
	}
	_ = os.Setenv(key, value)
}

// SetTestDefaults sets common test environment variables
func (e *EnvSetup) SetTestDefaults() {
	e.Set("GITHUB_TOKEN", "test-token")
	e.Set("GITHUB_REPOSITORY", "test/repo")
	e.Set("DEFAULT_BRANCH", "main")
	e.Set("DRY_RUN", "true")
}

// Restore restores all environment variables to their original values
func (e *EnvSetup) Restore() {
	for key, originalValue := range e.originalVars {
		if originalValue == "" {
			_ = os.Unsetenv(key)
		} else {
			_ = os.Setenv(key, originalValue)
		}
	}
}

// TestImageContent provides sample image content for testing
var TestImageContent = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90wS\xde\x00\x00\x00\tpHYs\x00\x00\x0b\x13\x00\x00\x0b\x13\x01\x00\x9a\x9c\x18\x00\x00\x00\x0bIDATx\x9cc```\x00\x00\x00\x04\x00\x01\xf6\x178U\x00\x00\x00\x00IEND\xaeB`\x82")

// SampleInstructions provides sample instruction content for testing
var SampleInstructions = map[string]string{
	"basic": `# Basic Assignment

This is a simple assignment for testing purposes.

## Requirements
- Complete task A
- Submit task B

## Resources
- Link to documentation
`,
	"with-images": `# Assignment with Images

This assignment includes visual components.

## Overview
![Overview diagram](static/overview.png)

## Workflow
![Process workflow](static/workflow.png)

## Implementation
Follow the steps shown in the diagrams above.
`,
	"complex": `# Advanced Assignment

This is a comprehensive assignment with multiple components.

## Learning Objectives
- Understand core concepts
- Apply practical skills
- Analyze complex problems

## Tasks

### Task 1: Analysis
![Analysis diagram](static/analysis.png)

Complete the analysis phase using the provided framework.

### Task 2: Implementation
![Implementation flow](static/implementation.png)

Implement the solution following the workflow diagram.

### Task 3: Testing
- Unit tests required
- Integration tests recommended
- Performance benchmarks optional

## Submission Guidelines
1. Code repository
2. Documentation
3. Test results

## Grading Rubric
- Functionality: 40%
- Code Quality: 30%
- Documentation: 20%
- Testing: 10%
`,
}

// AssertFileExists checks if a file exists and fails the test if it doesn't
func AssertFileExists(t *testing.T, filePath string) {
	t.Helper()
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist, but it doesn't", filePath)
	}
}

// AssertDirExists checks if a directory exists and fails the test if it doesn't
func AssertDirExists(t *testing.T, dirPath string) {
	t.Helper()
	if stat, err := os.Stat(dirPath); os.IsNotExist(err) || !stat.IsDir() {
		t.Errorf("Expected directory %s to exist, but it doesn't", dirPath)
	}
}

// ReadFileContent reads file content and returns it, failing the test if it can't be read
func ReadFileContent(t *testing.T, filePath string) string {
	t.Helper()
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filePath, err)
	}
	return string(content)
}

// ContainsString checks if content contains the expected string
func ContainsString(t *testing.T, content, expected, context string) {
	t.Helper()
	if content == "" {
		t.Errorf("Content is empty in context: %s", context)
		return
	}
	if expected == "" {
		t.Errorf("Expected string is empty in context: %s", context)
		return
	}
	if !Contains(content, expected) {
		t.Errorf("Expected %s to contain '%s', but it doesn't", context, expected)
	}
}

// Contains checks if a string contains a substring (case-insensitive)
func Contains(content, substr string) bool {
	return len(content) > 0 && len(substr) > 0 &&
		(content == substr ||
			len(content) >= len(substr) &&
				findSubstring(content, substr))
}

// Simple substring search helper
func findSubstring(content, substr string) bool {
	contentLen := len(content)
	substrLen := len(substr)

	if substrLen > contentLen {
		return false
	}

	for i := 0; i <= contentLen-substrLen; i++ {
		if content[i:i+substrLen] == substr {
			return true
		}
	}
	return false
}
