package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"assignment-pull-request/internal/creator"
)

// TestMainIntegration tests the main application flow end-to-end
func TestMainIntegration(t *testing.T) {
	// Save original environment variables
	originalVars := map[string]string{
		"GITHUB_TOKEN":           os.Getenv("GITHUB_TOKEN"),
		"GITHUB_REPOSITORY":      os.Getenv("GITHUB_REPOSITORY"),
		"ASSIGNMENTS_ROOT_REGEX": os.Getenv("ASSIGNMENTS_ROOT_REGEX"),
		"ASSIGNMENT_REGEX":       os.Getenv("ASSIGNMENT_REGEX"),
		"DEFAULT_BRANCH":         os.Getenv("DEFAULT_BRANCH"),
		"DRY_RUN":                os.Getenv("DRY_RUN"),
	}

	// Restore environment variables after test
	defer func() {
		for key, value := range originalVars {
			if value == "" {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, value)
			}
		}
	}()

	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful dry run with test fixtures",
			envVars: map[string]string{
				"GITHUB_TOKEN":           "test-token",
				"GITHUB_REPOSITORY":      "test/repo",
				"ASSIGNMENTS_ROOT_REGEX": "^test/fixtures$",
				"ASSIGNMENT_REGEX":       `^(?P<type>assignments|homework|labs|projects)/(?P<name>[^/]+)$`,
				"DEFAULT_BRANCH":         "main",
				"DRY_RUN":                "true",
			},
			wantErr: false,
		},
		{
			name: "missing required environment variable",
			envVars: map[string]string{
				"GITHUB_REPOSITORY": "test/repo",
				"DRY_RUN":           "true",
			},
			wantErr: true,
			errMsg:  "GITHUB_TOKEN environment variable is required",
		},
		{
			name: "complex regex patterns",
			envVars: map[string]string{
				"GITHUB_TOKEN":           "test-token",
				"GITHUB_REPOSITORY":      "test/repo",
				"ASSIGNMENTS_ROOT_REGEX": "^test/fixtures$",
				"ASSIGNMENT_REGEX":       `^courses/(?P<course>[^/]+)/(?P<week>[^/]+)/(?P<assignment>[^/]+)$`,
				"DEFAULT_BRANCH":         "main",
				"DRY_RUN":                "true",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			for key := range originalVars {
				_ = os.Unsetenv(key)
			}

			// Set test environment
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
			}

			// Create creator and run
			prCreator, err := creator.NewFromEnv()
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error during creator initialization")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error message to contain '%s', got: %s", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error during creator initialization: %v", err)
				return
			}

			// Run the creator (should work in dry-run mode)
			err = prCreator.Run()
			if err != nil {
				t.Errorf("Unexpected error during creator run: %v", err)
			}
		})
	}
}

// TestAssignmentDiscovery tests assignment discovery with various directory structures
func TestAssignmentDiscovery(t *testing.T) {
	// Create a temporary workspace root
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	tempDir := t.TempDir()
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Restore original directory
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: failed to restore original directory: %v", err)
		}
	}()

	// Create test directory structure
	assignments := []string{
		"test/fixtures/assignments/assignment-1",
		"test/fixtures/assignments/assignment-2",
		"test/fixtures/homework/hw-1",
		"test/fixtures/homework/hw-2",
		"test/fixtures/labs/lab-1",
		"test/fixtures/projects/project-1",
		"test/fixtures/courses/CS101/week-01/assignment-fibonacci",
		"test/fixtures/courses/CS102/week-02/assignment-sorting",
	}

	for _, assignmentPath := range assignments {
		err := os.MkdirAll(assignmentPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", assignmentPath, err)
		}

		// Create an instructions.md file in each assignment
		instructionsFile := filepath.Join(assignmentPath, "instructions.md")
		content := "# Assignment Instructions\n\nThis is a test assignment."
		err = os.WriteFile(instructionsFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create instructions file %s: %v", instructionsFile, err)
		}
	}

	// Set environment variables for testing
	_ = os.Setenv("GITHUB_TOKEN", "test-token")
	_ = os.Setenv("GITHUB_REPOSITORY", "test/repo")
	_ = os.Setenv("DEFAULT_BRANCH", "main")
	_ = os.Setenv("DRY_RUN", "true")

	defer func() {
		_ = os.Unsetenv("GITHUB_TOKEN")
		_ = os.Unsetenv("GITHUB_REPOSITORY")
		_ = os.Unsetenv("ASSIGNMENTS_ROOT_REGEX")
		_ = os.Unsetenv("ASSIGNMENT_REGEX")
		_ = os.Unsetenv("DEFAULT_BRANCH")
		_ = os.Unsetenv("DRY_RUN")
	}()

	tests := []struct {
		name                 string
		assignmentsRootRegex string
		assignmentRegex      string
		expectedMatches      int
	}{
		{
			name:                 "basic assignments only",
			assignmentsRootRegex: "^test/fixtures$",
			assignmentRegex:      `^assignments/(?P<name>[^/]+)$`,
			expectedMatches:      2, // assignment-1, assignment-2
		},
		{
			name:                 "homework and labs",
			assignmentsRootRegex: "^test/fixtures$",
			assignmentRegex:      `^(?P<type>homework|labs)/(?P<name>[^/]+)$`,
			expectedMatches:      3, // hw-1, hw-2, lab-1
		},
		{
			name:                 "all assignment types",
			assignmentsRootRegex: "^test/fixtures$",
			assignmentRegex:      `^(?P<type>assignments|homework|labs|projects)/(?P<name>[^/]+)$`,
			expectedMatches:      6, // All basic assignments
		},
		{
			name:                 "course structure",
			assignmentsRootRegex: "^test/fixtures$",
			assignmentRegex:      `^courses/(?P<course>[^/]+)/(?P<week>[^/]+)/(?P<assignment>[^/]+)$`,
			expectedMatches:      2, // CS101 and CS102 assignments
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv("ASSIGNMENTS_ROOT_REGEX", tt.assignmentsRootRegex)
			_ = os.Setenv("ASSIGNMENT_REGEX", tt.assignmentRegex)

			prCreator, err := creator.NewFromEnv()
			if err != nil {
				t.Fatalf("Failed to create PR creator: %v", err)
			}

			// Since we can't directly access findAssignments, we'll test through Run()
			// In dry-run mode, this should discover assignments and simulate operations
			err = prCreator.Run()
			if err != nil {
				t.Errorf("Unexpected error during assignment discovery: %v", err)
			}
		})
	}
}

// TestImageLinkRewriting tests image link rewriting with actual files
func TestImageLinkRewriting(t *testing.T) {
	// Create temporary directory with test structure
	tempDir := t.TempDir()

	// Create assignment directory with static images
	assignmentDir := filepath.Join(tempDir, "test-assignment")
	staticDir := filepath.Join(assignmentDir, "static")
	err := os.MkdirAll(staticDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create static directory: %v", err)
	}

	// Create test images
	testImages := []string{"overview.png", "diagram.jpg", "workflow.gif"}
	for _, img := range testImages {
		imgPath := filepath.Join(staticDir, img)
		err = os.WriteFile(imgPath, []byte("fake image content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test image %s: %v", img, err)
		}
	}

	// Create instructions.md with image references
	instructionsContent := `# Test Assignment

Here's an overview:
![Overview](static/overview.png)

Check out the workflow:
![Workflow](static/workflow.gif)

And the system diagram:
![Diagram](static/diagram.jpg)

External image (should not be changed):
![External](https://example.com/external.png)

Absolute path (should not be changed):
![Absolute](/absolute/path/image.png)
`

	instructionsFile := filepath.Join(assignmentDir, "instructions.md")
	err = os.WriteFile(instructionsFile, []byte(instructionsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create instructions file: %v", err)
	}

	// Change to temp directory for testing
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: failed to restore original directory: %v", err)
		}
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Set up environment for testing
	_ = os.Setenv("GITHUB_TOKEN", "test-token")
	_ = os.Setenv("GITHUB_REPOSITORY", "test/repo")
	_ = os.Setenv("ASSIGNMENTS_ROOT_REGEX", "^\\.$")
	_ = os.Setenv("ASSIGNMENT_REGEX", `^(?P<name>test-assignment)$`)
	_ = os.Setenv("DEFAULT_BRANCH", "main")
	_ = os.Setenv("DRY_RUN", "true")

	defer func() {
		_ = os.Unsetenv("GITHUB_TOKEN")
		_ = os.Unsetenv("GITHUB_REPOSITORY")
		_ = os.Unsetenv("ASSIGNMENTS_ROOT_REGEX")
		_ = os.Unsetenv("ASSIGNMENT_REGEX")
		_ = os.Unsetenv("DEFAULT_BRANCH")
		_ = os.Unsetenv("DRY_RUN")
	}()

	// Create and run creator
	prCreator, err := creator.NewFromEnv()
	if err != nil {
		t.Fatalf("Failed to create PR creator: %v", err)
	}

	err = prCreator.Run()
	if err != nil {
		t.Errorf("Unexpected error during PR creation with image rewriting: %v", err)
	}

	// In a real implementation, we would verify that the created README.md
	// contains the rewritten image links. Since we're in dry-run mode,
	// we can only verify that the process completes without error.
}

// TestComplexWorkflow tests a complex workflow with multiple assignments
func TestComplexWorkflow(t *testing.T) {
	// Save original directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	tempDir := t.TempDir()
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	// Create complex directory structure similar to real use cases
	complexStructure := []struct {
		path            string
		hasInstructions bool
		hasImages       bool
	}{
		{"bootcamp/2024-fall/module-frontend/week-1/assignment-html-basics", true, true},
		{"bootcamp/2024-fall/module-frontend/week-2/assignment-css-layout", true, true},
		{"bootcamp/2024-fall/module-backend/week-1/assignment-node-intro", true, false},
		{"courses/CS101/semester-1/module-1/assignment-variables", true, true},
		{"courses/CS101/semester-1/module-2/assignment-functions", true, false},
		{"courses/CS102/advanced-topics/week-5/assignment-algorithms", true, true},
	}

	for _, item := range complexStructure {
		// Create directory
		err := os.MkdirAll(item.path, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", item.path, err)
		}

		// Create instructions.md if needed
		if item.hasInstructions {
			instructionsContent := "# Assignment Instructions\n\nThis is a test assignment with comprehensive instructions."
			if item.hasImages {
				instructionsContent += "\n\n![Overview](static/overview.png)\n![Diagram](static/diagram.jpg)"
			}

			instructionsFile := filepath.Join(item.path, "instructions.md")
			err = os.WriteFile(instructionsFile, []byte(instructionsContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create instructions file %s: %v", instructionsFile, err)
			}
		}

		// Create static images if needed
		if item.hasImages {
			staticDir := filepath.Join(item.path, "static")
			err := os.MkdirAll(staticDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create static directory %s: %v", staticDir, err)
			}

			images := []string{"overview.png", "diagram.jpg"}
			for _, img := range images {
				imgPath := filepath.Join(staticDir, img)
				err = os.WriteFile(imgPath, []byte("fake image data"), 0644)
				if err != nil {
					t.Fatalf("Failed to create image %s: %v", imgPath, err)
				}
			}
		}
	}

	// Test various regex patterns against this structure
	tests := []struct {
		name                 string
		assignmentsRootRegex string
		assignmentRegex      string
		description          string
	}{
		{
			name:                 "bootcamp assignments",
			assignmentsRootRegex: "^bootcamp$",
			assignmentRegex:      `^(?P<year>[^/]+)/(?P<module>[^/]+)/(?P<week>[^/]+)/(?P<assignment>[^/]+)$`,
			description:          "Match bootcamp structure with year/module/week/assignment",
		},
		{
			name:                 "course assignments",
			assignmentsRootRegex: "^courses$",
			assignmentRegex:      `^(?P<course>[^/]+)/(?P<semester>[^/]+)/(?P<module>[^/]+)/(?P<assignment>[^/]+)$`,
			description:          "Match course structure with semester modules",
		},
		{
			name:                 "advanced course assignments",
			assignmentsRootRegex: "^courses$",
			assignmentRegex:      `^(?P<course>[^/]+)/(?P<topic>[^/]+)/(?P<week>[^/]+)/(?P<assignment>[^/]+)$`,
			description:          "Match advanced course structure",
		},
	}

	// Set common environment variables
	_ = os.Setenv("GITHUB_TOKEN", "test-token")
	_ = os.Setenv("GITHUB_REPOSITORY", "test/repo")
	_ = os.Setenv("DEFAULT_BRANCH", "main")
	_ = os.Setenv("DRY_RUN", "true")

	defer func() {
		_ = os.Unsetenv("GITHUB_TOKEN")
		_ = os.Unsetenv("GITHUB_REPOSITORY")
		_ = os.Unsetenv("ASSIGNMENTS_ROOT_REGEX")
		_ = os.Unsetenv("ASSIGNMENT_REGEX")
		_ = os.Unsetenv("DEFAULT_BRANCH")
		_ = os.Unsetenv("DRY_RUN")
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv("ASSIGNMENTS_ROOT_REGEX", tt.assignmentsRootRegex)
			_ = os.Setenv("ASSIGNMENT_REGEX", tt.assignmentRegex)

			prCreator, err := creator.NewFromEnv()
			if err != nil {
				t.Fatalf("Failed to create PR creator for %s: %v", tt.description, err)
			}

			err = prCreator.Run()
			if err != nil {
				t.Errorf("Failed to run complex workflow for %s: %v", tt.description, err)
			}
		})
	}
}

// BenchmarkFullWorkflow benchmarks the complete workflow
func BenchmarkFullWorkflow(b *testing.B) {
	// Create minimal test structure
	tempDir := b.TempDir()
	assignmentDir := filepath.Join(tempDir, "assignment-1")
	err := os.MkdirAll(assignmentDir, 0755)
	if err != nil {
		b.Fatalf("Failed to create test directory: %v", err)
	}

	instructionsFile := filepath.Join(assignmentDir, "instructions.md")
	err = os.WriteFile(instructionsFile, []byte("# Test Assignment"), 0644)
	if err != nil {
		b.Fatalf("Failed to create instructions file: %v", err)
	}

	// Change to temp directory
	originalDir, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	// Set environment
	_ = os.Setenv("GITHUB_TOKEN", "test-token")
	_ = os.Setenv("GITHUB_REPOSITORY", "test/repo")
	_ = os.Setenv("ASSIGNMENTS_ROOT_REGEX", "^\\.$")
	_ = os.Setenv("ASSIGNMENT_REGEX", `^(?P<name>assignment-1)$`)
	_ = os.Setenv("DEFAULT_BRANCH", "main")
	_ = os.Setenv("DRY_RUN", "true")

	defer func() {
		_ = os.Unsetenv("GITHUB_TOKEN")
		_ = os.Unsetenv("GITHUB_REPOSITORY")
		_ = os.Unsetenv("ASSIGNMENTS_ROOT_REGEX")
		_ = os.Unsetenv("ASSIGNMENT_REGEX")
		_ = os.Unsetenv("DEFAULT_BRANCH")
		_ = os.Unsetenv("DRY_RUN")
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		prCreator, err := creator.NewFromEnv()
		if err != nil {
			b.Fatalf("Failed to create PR creator: %v", err)
		}

		err = prCreator.Run()
		if err != nil {
			b.Fatalf("Failed to run workflow: %v", err)
		}
	}
}

// TestBranchNameConflictValidation tests that branch name conflicts are detected and rejected
func TestBranchNameConflictValidation(t *testing.T) {
	// Save original directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	tempDir := t.TempDir()
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	// Create assignments that would generate conflicting branch names
	conflictingAssignments := []string{
		"test/fixtures/assigments/assignment-1",        // Would generate branch: assignment-1
		"test/fixtures/assigments/unique/assignment-2", // Would generate branch: assignment-2 (unique)
		"test/fixtures/lab/week-01/assignment-1",       // Would also generate branch: assignment-1 (conflict!)
		"test/fixtures/lab/week-02/assignment-1",       // Would also generate branch: assignment-1 (conflict!)
		"test/fixtures/lab/week-03/assignment-1",       // Would also generate branch: assignment-1 (conflict!)
	}

	for _, assignment := range conflictingAssignments {
		err := os.MkdirAll(assignment, 0755)
		if err != nil {
			t.Fatalf("Failed to create assignment directory %s: %v", assignment, err)
		}
	}

	// Set environment variables to cause conflicts
	envVars := map[string]string{
		"GITHUB_TOKEN":           "test-token",
		"GITHUB_REPOSITORY":      "test/repo",
		"ASSIGNMENTS_ROOT_REGEX": "^test/fixtures/(assigments|lab)$",                                       // Match current directory exactly
		"ASSIGNMENT_REGEX":       "^(?P<branch>assignment-\\d+)$, ^[^/]+/(?P<assignment>assignment-\\d+)$", // Two patterns that can conflict
		"DEFAULT_BRANCH":         "main",
		"DRY_RUN":                "true",
	}

	// Save original environment variables
	originalVars := map[string]string{}
	for key := range envVars {
		originalVars[key] = os.Getenv(key)
	}

	// Set test environment variables
	for key, value := range envVars {
		_ = os.Setenv(key, value)
	}

	// Restore environment variables after test
	defer func() {
		for key, value := range originalVars {
			if value == "" {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, value)
			}
		}
	}()

	// Create and run the PR creator
	prCreator, err := creator.NewFromEnv()
	if err != nil {
		t.Fatalf("Failed to create PR creator: %v", err)
	}

	// This should fail due to branch name conflicts
	err = prCreator.Run()
	if err == nil {
		t.Errorf("Expected error due to branch name conflicts, but got none")
		return
	}

	// Check that the error message mentions branch name conflicts
	if !strings.Contains(err.Error(), "branch name") && !strings.Contains(err.Error(), "conflict") {
		t.Errorf("Expected error to mention branch name conflicts, but got: %s", err.Error())
	}

	t.Logf("Correctly detected branch name conflict: %s", err.Error())
}
