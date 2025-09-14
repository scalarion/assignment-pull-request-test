package creator

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"assignment-pull-request/internal/assignment"
	"assignment-pull-request/internal/constants"
)

// cleanupEnv clears environment variables that might affect tests
func cleanupEnv() {
	clearEnvVars := []string{
		"GITHUB_TOKEN", "GITHUB_REPOSITORY", "ASSIGNMENTS_ROOT_REGEX",
		"ASSIGNMENT_REGEX", "DEFAULT_BRANCH", "DRY_RUN",
	}
	for _, key := range clearEnvVars {
		_ = os.Unsetenv(key)
	}
}

// TestNew verifies the Creator initialization with various configurations
func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid configuration",
			envVars: map[string]string{
				"GITHUB_TOKEN":           "test-token",
				"GITHUB_REPOSITORY":      "owner/repo",
				"ASSIGNMENTS_ROOT_REGEX": "^assignments$",
				"ASSIGNMENT_REGEX":       `^(?P<branch>assignment-\d+)$`,
				"DEFAULT_BRANCH":         "main",
				"DRY_RUN":                "false",
			},
			wantErr: false,
		},
		{
			name: "missing github token",
			envVars: map[string]string{
				"GITHUB_REPOSITORY": "owner/repo",
			},
			wantErr: true,
			errMsg:  "GITHUB_TOKEN environment variable is required",
		},
		{
			name: "missing repository name",
			envVars: map[string]string{
				"GITHUB_TOKEN": "test-token",
			},
			wantErr: true,
			errMsg:  "GITHUB_REPOSITORY environment variable is required",
		},
		{
			name: "invalid root regex",
			envVars: map[string]string{
				"GITHUB_TOKEN":           "test-token",
				"GITHUB_REPOSITORY":      "owner/repo",
				"ASSIGNMENTS_ROOT_REGEX": "[invalid",
			},
			wantErr: true,
			errMsg:  "failed to create assignment processor",
		},
		{
			name: "invalid assignment regex",
			envVars: map[string]string{
				"GITHUB_TOKEN":      "test-token",
				"GITHUB_REPOSITORY": "owner/repo",
				"ASSIGNMENT_REGEX":  "[invalid",
			},
			wantErr: true,
			errMsg:  "failed to create assignment processor",
		},
		{
			name: "default values",
			envVars: map[string]string{
				"GITHUB_TOKEN":      "test-token",
				"GITHUB_REPOSITORY": "owner/repo",
			},
			wantErr: false,
		},
		{
			name: "dry run enabled",
			envVars: map[string]string{
				"GITHUB_TOKEN":      "test-token",
				"GITHUB_REPOSITORY": "owner/repo",
				"DRY_RUN":           "true",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables first
			cleanupEnv()

			// Set test environment variables
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
			}

			creator, err := NewFromEnv()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error message to contain '%s', got: %s", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if creator == nil {
				t.Error("Creator should not be nil")
				return
			}

			// Verify configuration
			if creator.config.GitHubToken != tt.envVars["GITHUB_TOKEN"] {
				t.Errorf("Expected GitHubToken=%s, got=%s", tt.envVars["GITHUB_TOKEN"], creator.config.GitHubToken)
			}

			if creator.config.RepositoryName != tt.envVars["GITHUB_REPOSITORY"] {
				t.Errorf("Expected RepositoryName=%s, got=%s", tt.envVars["GITHUB_REPOSITORY"], creator.config.RepositoryName)
			}

			// Check dry run setting
			expectedDryRun := isDryRun(tt.envVars["DRY_RUN"])
			if creator.config.DryRun != expectedDryRun {
				t.Errorf("Expected DryRun=%t, got=%t", expectedDryRun, creator.config.DryRun)
			}

			// Verify assignment processor was created
			if creator.assignmentProcessor == nil {
				t.Error("Assignment processor should not be nil")
			}

			// Verify regex patterns are available through assignment processor
			rootPatterns := creator.assignmentProcessor.GetRootRegexStrings()
			assignmentPatterns := creator.assignmentProcessor.GetAssignmentRegexStrings()
			if len(rootPatterns) == 0 {
				t.Error("Root patterns should not be empty")
			}
			if len(assignmentPatterns) == 0 {
				t.Error("Assignment patterns should not be empty")
			}
		})
	}

	// Clean up environment variables
	cleanupEnv()
}

// TestRegexValidation tests that regex patterns are validated for named groups
func TestRegexValidation(t *testing.T) {
	// Clean up any existing environment variables
	cleanupEnv()

	tests := []struct {
		name             string
		assignmentRegex  string
		shouldFail       bool
		expectedErrorMsg string
	}{
		{
			name:            "valid regex with named groups",
			assignmentRegex: `^(?P<type>assignment)-(?P<number>\d+)$`,
			shouldFail:      false,
		},
		{
			name:            "valid regex with single named group",
			assignmentRegex: `^(?P<branch>hw-\d+)$`,
			shouldFail:      false,
		},
		{
			name:            "valid regex with unnamed groups",
			assignmentRegex: `^(assignment)-(\d+)$`,
			shouldFail:      false,
		},
		{
			name:            "valid regex with mixed named and unnamed groups",
			assignmentRegex: `^(?P<type>assignment)-(group-\d+)-(\d+)$`,
			shouldFail:      false,
		},
		{
			name:             "invalid regex without any capturing groups",
			assignmentRegex:  `^assignment-\d+$`,
			shouldFail:       true,
			expectedErrorMsg: "must contain at least one capturing group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			envVars := map[string]string{
				"GITHUB_TOKEN":      "test-token",
				"GITHUB_REPOSITORY": "test/repo",
				"ASSIGNMENT_REGEX":  tt.assignmentRegex,
			}

			for key, value := range envVars {
				os.Setenv(key, value)
			}
			defer cleanupEnv()

			// Try to create a new Creator
			creator, err := NewFromEnv()

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected error for regex '%s', but got none", tt.assignmentRegex)
					return
				}
				if !strings.Contains(err.Error(), tt.expectedErrorMsg) {
					t.Errorf("Expected error message to contain '%s', got: %v", tt.expectedErrorMsg, err)
				}
				if creator != nil {
					t.Error("Creator should be nil when validation fails")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for valid regex '%s': %v", tt.assignmentRegex, err)
					return
				}
				if creator == nil {
					t.Error("Creator should not be nil for valid regex")
				}
			}
		})
	}
}

// TestHasCapturingGroups tests the hasCapturingGroups helper function
func TestHasCapturingGroups(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected bool
	}{
		{
			name:     "pattern with named groups",
			pattern:  `^(?P<type>assignment)-(?P<number>\d+)$`,
			expected: true,
		},
		{
			name:     "pattern with single named group",
			pattern:  `^(?P<branch>hw-\d+)$`,
			expected: true,
		},
		{
			name:     "pattern with unnamed groups",
			pattern:  `^(assignment)-(\d+)$`,
			expected: true,
		},
		{
			name:     "pattern with mixed named and unnamed groups",
			pattern:  `^(?P<type>assignment)-(group-\d+)-(?P<number>\d+)$`,
			expected: true,
		},
		{
			name:     "pattern without any groups",
			pattern:  `^assignment-\d+$`,
			expected: false,
		},
		{
			name:     "pattern with non-capturing groups",
			pattern:  `^(?:assignment)-(?:\d+)$`,
			expected: false,
		},
		{
			name:     "empty pattern",
			pattern:  ``,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex, err := regexp.Compile(tt.pattern)
			if err != nil {
				t.Fatalf("Failed to compile regex '%s': %v", tt.pattern, err)
			}

			result := assignment.HasCapturingGroups(regex)
			if result != tt.expected {
				t.Errorf("HasCapturingGroups('%s') = %t, expected %t", tt.pattern, result, tt.expected)
			}
		})
	}
}

// TestGetEnvWithDefault tests environment variable parsing with defaults
func TestGetEnvWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "environment variable set",
			envKey:       "TEST_VAR",
			envValue:     "test-value",
			defaultValue: "default-value",
			expected:     "test-value",
		},
		{
			name:         "environment variable empty",
			envKey:       "TEST_VAR",
			envValue:     "",
			defaultValue: "default-value",
			expected:     "default-value",
		},
		{
			name:         "environment variable not set",
			envKey:       "NONEXISTENT_VAR",
			envValue:     "",
			defaultValue: "default-value",
			expected:     "default-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up first
			_ = os.Unsetenv(tt.envKey)

			// Set environment variable if specified
			if tt.envValue != "" {
				_ = os.Setenv(tt.envKey, tt.envValue)
			}

			result := getEnvWithDefault(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}

			// Clean up
			_ = os.Unsetenv(tt.envKey)
		})
	}
}

// TestIsDryRun tests dry run detection
func TestIsDryRun(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"true lowercase", "true", true},
		{"TRUE uppercase", "TRUE", true},
		{"True mixed case", "True", true},
		{"1 numeric", "1", true},
		{"yes lowercase", "yes", true},
		{"YES uppercase", "YES", true},
		{"false", "false", false},
		{"0", "0", false},
		{"no", "no", false},
		{"empty", "", false},
		{"random", "random", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDryRun(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

// TestFindInstructionsFile tests instructions file discovery
func TestFindInstructionsFile(t *testing.T) {
	// Create temporary test directory structure
	tempDir := t.TempDir()

	// Create test assignment with instructions.md
	assignmentDir := filepath.Join(tempDir, "assignment-1")
	err := os.MkdirAll(assignmentDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	instructionsFile := filepath.Join(assignmentDir, constants.InstructionsFileName)
	err = os.WriteFile(instructionsFile, []byte("# Test Assignment"), 0644)
	if err != nil {
		t.Fatalf("Failed to create instructions file: %v", err)
	}

	creator := &Creator{}

	// Test finding existing instructions file
	result := creator.findInstructionsFile(assignmentDir)
	expected := instructionsFile
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// Test with non-existent directory
	result = creator.findInstructionsFile(filepath.Join(tempDir, "nonexistent"))
	if result != "" {
		t.Errorf("Expected empty string for non-existent directory, got %s", result)
	}
}

// TestRewriteImageLinks tests image link rewriting functionality
func TestRewriteImageLinks(t *testing.T) {
	creator := &Creator{}

	tests := []struct {
		name           string
		content        string
		assignmentPath string
		expected       string
	}{
		{
			name:           "markdown image",
			content:        "![diagram](static/diagram.png)",
			assignmentPath: "test/fixtures/assignments/assignment-1",
			expected:       "![diagram](test/fixtures/assignments/assignment-1/static/diagram.png)",
		},
		{
			name:           "html image - not processed by current implementation",
			content:        `<img src="static/overview.png" alt="overview">`,
			assignmentPath: "test/fixtures/homework/hw-1",
			expected:       `<img src="static/overview.png" alt="overview">`, // HTML images not processed
		},
		{
			name:           "multiple images",
			content:        "![img1](static/img1.png) and ![img2](static/img2.jpg)",
			assignmentPath: "assignment-1",
			expected:       "![img1](assignment-1/static/img1.png) and ![img2](assignment-1/static/img2.jpg)",
		},
		{
			name:           "mixed markdown and html",
			content:        `![md](static/md.png) and <img src="static/html.gif">`,
			assignmentPath: "project-1",
			expected:       `![md](project-1/static/md.png) and <img src="static/html.gif">`, // Only markdown processed
		},
		{
			name:           "no images",
			content:        "This is just text with no images.",
			assignmentPath: "assignment-1",
			expected:       "This is just text with no images.",
		},
		{
			name:           "absolute paths ignored",
			content:        "![abs](/absolute/path.png) ![rel](static/relative.png)",
			assignmentPath: "assignment-1",
			expected:       "![abs](/absolute/path.png) ![rel](assignment-1/static/relative.png)",
		},
		{
			name:           "http urls ignored",
			content:        "![url](https://example.com/image.png) ![rel](static/relative.png)",
			assignmentPath: "assignment-1",
			expected:       "![url](https://example.com/image.png) ![rel](assignment-1/static/relative.png)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := creator.rewriteImageLinks(tt.content, tt.assignmentPath)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

// TestCreateGenericPullRequestBody tests generic PR body creation
func TestCreateGenericPullRequestBody(t *testing.T) {
	creator := &Creator{}

	tests := []struct {
		name           string
		assignmentPath string
		expectedPath   string
	}{
		{
			name:           "simple assignment",
			assignmentPath: "assignment-1",
			expectedPath:   "assignment-1",
		},
		{
			name:           "homework assignment",
			assignmentPath: "homework/hw-2",
			expectedPath:   "homework/hw-2",
		},
		{
			name:           "complex path",
			assignmentPath: "test/fixtures/assignments/assignment-sorting",
			expectedPath:   "test/fixtures/assignments/assignment-sorting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := creator.createGenericPullRequestBody(tt.assignmentPath)

			if !strings.Contains(body, tt.expectedPath) {
				t.Errorf("Expected body to contain path '%s', got:\n%s", tt.expectedPath, body)
			}

			// Check that body contains required sections from actual implementation
			requiredSections := []string{
				"## Assignment Pull Request",
				"This pull request contains the setup for the assignment located at",
				"### Changes included:",
				"### Next steps:",
			}

			for _, section := range requiredSections {
				if !strings.Contains(body, section) {
					t.Errorf("Expected body to contain section '%s'", section)
				}
			}
		})
	}
}

// TestValidateBranchNameUniqueness tests the branch name uniqueness validation through AssignmentProcessor
func TestValidateBranchNameUniqueness(t *testing.T) {
	tests := []struct {
		name          string
		rootRegex     []string
		assignRegex   []string
		expectError   bool
		errorContains string
	}{
		{
			name:      "no conflicts - valid patterns",
			rootRegex: []string{"."},
			assignRegex: []string{
				`^(?P<assignment>assignment-\d+)$`,
				`^(?P<type>homework)/(?P<name>hw-\d+)$`,
			},
			expectError: false,
		},
		{
			name:      "conflict - patterns could create same branch names",
			rootRegex: []string{"."},
			assignRegex: []string{
				`^(?P<branch>assignment-\d+)$`,           // creates assignment-N
				`^[^/]+/(?P<assignment>assignment-\d+)$`, // also creates assignment-N
			},
			expectError: false, // Would only error if actual conflicting directories exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that we can create an AssignmentProcessor with these patterns
			processor, err := assignment.NewAssignmentProcessor("", tt.rootRegex, tt.assignRegex)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', but got: %s", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %s", err.Error())
					return
				}

				// Verify processor was created successfully
				if processor == nil {
					t.Error("Expected processor to be created")
				}
			}
		})
	}
}

// BenchmarkRewriteImageLinks benchmarks image link rewriting
func BenchmarkRewriteImageLinks(b *testing.B) {
	creator := &Creator{}

	content := `
# Assignment Instructions

Here's an overview diagram:
![overview](static/overview.png)

And here's the detailed workflow:
<img src="static/workflow.png" alt="workflow">

Additional resources:
![resource1](static/res1.jpg)
![resource2](static/res2.gif)
<img src="static/final.png">
`

	assignmentPath := "test/fixtures/assignments/assignment-complex"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		creator.rewriteImageLinks(content, assignmentPath)
	}
}
