package creator

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
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
			errMsg:  "invalid assignments root regex",
		},
		{
			name: "invalid assignment regex",
			envVars: map[string]string{
				"GITHUB_TOKEN":      "test-token",
				"GITHUB_REPOSITORY": "owner/repo",
				"ASSIGNMENT_REGEX":  "[invalid",
			},
			wantErr: true,
			errMsg:  "invalid assignment regex",
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

			creator, err := New()

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

			// Verify regex patterns were compiled
			if len(creator.rootPatterns) == 0 {
				t.Error("Root patterns should not be empty")
			}
			if len(creator.assignmentPatterns) == 0 {
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
			creator, err := New()

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

// TestHasNamedGroups tests the hasNamedGroups helper function
func TestHasNamedGroups(t *testing.T) {
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
			name:     "pattern without any groups",
			pattern:  `^assignment-\d+$`,
			expected: false,
		},
		{
			name:     "pattern with only unnamed groups",
			pattern:  `^(assignment)-(\d+)$`,
			expected: false,
		},
		{
			name:     "pattern with mixed named and unnamed groups",
			pattern:  `^(?P<type>assignment)-(group-\d+)-(?P<number>\d+)$`,
			expected: true,
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

			result := hasNamedGroups(regex)
			if result != tt.expected {
				t.Errorf("hasNamedGroups('%s') = %t, expected %t", tt.pattern, result, tt.expected)
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

			result := hasCapturingGroups(regex)
			if result != tt.expected {
				t.Errorf("hasCapturingGroups('%s') = %t, expected %t", tt.pattern, result, tt.expected)
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

// TestParseRegexPatterns tests regex pattern parsing
func TestParseRegexPatterns(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single pattern",
			input:    "^assignments$",
			expected: []string{"^assignments$"},
		},
		{
			name:     "multiple patterns",
			input:    "^assignments$,^homework$,^labs$",
			expected: []string{"^assignments$", "^homework$", "^labs$"},
		},
		{
			name:     "patterns with whitespace",
			input:    " ^assignments$ , ^homework$ , ^labs$ ",
			expected: []string{"^assignments$", "^homework$", "^labs$"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "empty patterns",
			input:    ",,",
			expected: []string{},
		},
		{
			name:     "mixed empty and valid",
			input:    "^assignments$,,^homework$",
			expected: []string{"^assignments$", "^homework$"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRegexPatterns(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Expected result[%d]=%s, got=%s", i, expected, result[i])
				}
			}
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

// TestExtractBranchName tests branch name extraction from assignment paths
func TestExtractBranchName(t *testing.T) {
	// Create a test creator with known patterns
	creator := &Creator{
		assignmentPatterns: []*regexp.Regexp{
			regexp.MustCompile(`^(?P<branch>assignment-\d+)$`),
			regexp.MustCompile(`^(?P<course>[^/]+)/(?P<type>hw)-(?P<number>\d+)$`),
			regexp.MustCompile(`^test/fixtures/(?P<type>assignments|homework|labs|projects)/(?P<name>[^/]+)$`),
			regexp.MustCompile(`^test/fixtures/courses/(?P<course>[^/]+)/(?P<week>[^/]+)/(?P<assignment>[^/]+)$`),
		},
	}

	tests := []struct {
		name           string
		assignmentPath string
		expectedBranch string
		expectedMatch  bool
	}{
		{
			name:           "simple assignment pattern",
			assignmentPath: "assignment-1",
			expectedBranch: "assignment-1",
			expectedMatch:  true,
		},
		{
			name:           "homework pattern",
			assignmentPath: "CS101/hw-2",
			expectedBranch: "cs101-2-hw", // Named groups alphabetically: course, number, type
			expectedMatch:  true,
		},
		{
			name:           "test fixtures assignment",
			assignmentPath: "test/fixtures/assignments/assignment-1",
			expectedBranch: "assignment-1-assignments", // Named groups alphabetically: name, type
			expectedMatch:  true,
		},
		{
			name:           "test fixtures homework",
			assignmentPath: "test/fixtures/homework/hw-1",
			expectedBranch: "hw-1-homework", // Named groups alphabetically: name, type
			expectedMatch:  true,
		},
		{
			name:           "course structure",
			assignmentPath: "test/fixtures/courses/CS101/week-02/assignment-sorting",
			expectedBranch: "assignment-sorting-cs101-week-02", // Named groups alphabetically: assignment, course, week
			expectedMatch:  true,
		},
		{
			name:           "no match",
			assignmentPath: "random/path/not/matching",
			expectedBranch: "",
			expectedMatch:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			branch, matched := creator.extractBranchName(tt.assignmentPath)

			if matched != tt.expectedMatch {
				t.Errorf("Expected match=%t, got=%t", tt.expectedMatch, matched)
			}

			if branch != tt.expectedBranch {
				t.Errorf("Expected branch=%s, got=%s", tt.expectedBranch, branch)
			}
		})
	}
}

// TestExtractBranchNameWithUnnamedGroups tests branch name extraction with unnamed capturing groups
func TestExtractBranchNameWithUnnamedGroups(t *testing.T) {
	// Create a test creator with patterns using unnamed groups
	// Note: More specific patterns should come first to avoid conflicts
	creator := &Creator{
		assignmentPatterns: []*regexp.Regexp{
			// Unnamed groups pattern: (type)-(number)
			regexp.MustCompile(`^(assignment)-(\d+)$`),
			// Single unnamed group - match only the relevant part (SPECIFIC - must come before general course pattern)
			regexp.MustCompile(`^homework/(hw-\d+)$`),
			// Mixed named and unnamed: named-unnamed-named (GENERAL - comes after specific patterns)
			regexp.MustCompile(`^(?P<course>[^/]+)/(hw|lab)-(\d+)$`),
			// Multiple unnamed groups
			regexp.MustCompile(`^(projects)/(semester-\d+)/(week-\d+)/(assignment-[^/]+)$`),
		},
	}

	tests := []struct {
		name           string
		assignmentPath string
		expectedBranch string
		expectedMatch  bool
	}{
		{
			name:           "unnamed groups: assignment-number",
			assignmentPath: "assignment-1",
			expectedBranch: "assignment-1",
			expectedMatch:  true,
		},
		{
			name:           "mixed groups: course/hw-number",
			assignmentPath: "CS101/hw-2",
			expectedBranch: "cs101-hw-2", // Named group "course" alphabetically first + unnamed groups in order
			expectedMatch:  true,
		},
		{
			name:           "multiple unnamed groups",
			assignmentPath: "projects/semester-1/week-3/assignment-variables",
			expectedBranch: "projects-semester-1-week-3-assignment-variables",
			expectedMatch:  true,
		},
		{
			name:           "single unnamed group",
			assignmentPath: "homework/hw-5",
			expectedBranch: "hw-5",
			expectedMatch:  true,
		},
		{
			name:           "no match",
			assignmentPath: "random/path/not/matching",
			expectedBranch: "",
			expectedMatch:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			branch, matched := creator.extractBranchName(tt.assignmentPath)

			if matched != tt.expectedMatch {
				t.Errorf("Expected match=%t, got=%t", tt.expectedMatch, matched)
			}

			if branch != tt.expectedBranch {
				t.Errorf("Expected branch=%s, got=%s", tt.expectedBranch, branch)
			}
		})
	}
}

// TestExtractBranchNameAlphabeticalOrdering tests alphabetical ordering of named groups
func TestExtractBranchNameAlphabeticalOrdering(t *testing.T) {
	// Test pattern with multiple named groups: module, course, assignment (alphabetically: assignment, course, module)
	creator := &Creator{
		assignmentPatterns: []*regexp.Regexp{
			// Multiple named groups to test alphabetical ordering
			regexp.MustCompile(`^(?P<module>[^/]+)/(?P<course>[^/]+)/(?P<assignment>[^/]+)$`),
			// Mixed named and unnamed groups
			regexp.MustCompile(`^(?P<year>\d+)/(?P<course>[^/]+)/(week-\d+)/(assignment-\d+)$`),
		},
	}

	tests := []struct {
		name           string
		assignmentPath string
		expectedBranch string
		expectedMatch  bool
	}{
		{
			name:           "multiple named groups alphabetical order",
			assignmentPath: "backend/CS101/variables",
			expectedBranch: "variables-cs101-backend", // assignment, course, module (alphabetical)
			expectedMatch:  true,
		},
		{
			name:           "mixed named and unnamed groups",
			assignmentPath: "2024/CS102/week-5/assignment-3",
			expectedBranch: "cs102-2024-week-5-assignment-3", // course, year (alphabetical) + unnamed groups (order)
			expectedMatch:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			branch, matched := creator.extractBranchName(tt.assignmentPath)

			if matched != tt.expectedMatch {
				t.Errorf("Expected match=%t, got=%t", tt.expectedMatch, matched)
			}

			if branch != tt.expectedBranch {
				t.Errorf("Expected branch=%s, got=%s", tt.expectedBranch, branch)
			}
		})
	}
}

// TestSanitizeBranchName tests branch name sanitization
func TestSanitizeBranchName(t *testing.T) {
	creator := &Creator{}

	tests := []struct {
		name           string
		assignmentPath string
		expected       string
	}{
		{
			name:           "simple path",
			assignmentPath: "assignment-1",
			expected:       "assignment-1",
		},
		{
			name:           "path with slashes",
			assignmentPath: "assignments/assignment-1",
			expected:       "assignments-assignment-1",
		},
		{
			name:           "complex path",
			assignmentPath: "test/fixtures/assignments/assignment-1",
			expected:       "test-fixtures-assignments-assignment-1",
		},
		{
			name:           "path with spaces",
			assignmentPath: "my assignment/project 1",
			expected:       "my-assignment-project-1",
		},
		{
			name:           "path with special characters",
			assignmentPath: "assignment_#1@project!",
			expected:       "assignment_#1@project!", // Only sanitizes spaces and slashes in the actual implementation
		},
		{
			name:           "consecutive slashes",
			assignmentPath: "assignment///path///1",
			expected:       "assignment-path-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := creator.sanitizeBranchName(tt.assignmentPath)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
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

	instructionsFile := filepath.Join(assignmentDir, "instructions.md")
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

// BenchmarkExtractBranchName benchmarks branch name extraction
func BenchmarkExtractBranchName(b *testing.B) {
	creator := &Creator{
		assignmentPatterns: []*regexp.Regexp{
			regexp.MustCompile(`^(?P<branch>assignment-\d+)$`),
			regexp.MustCompile(`^(?P<course>[^/]+)/(?P<type>hw)-(?P<number>\d+)$`),
			regexp.MustCompile(`^test/fixtures/(?P<type>assignments|homework|labs|projects)/(?P<name>[^/]+)$`),
		},
	}

	testPaths := []string{
		"assignment-1",
		"CS101/hw-2",
		"test/fixtures/assignments/assignment-sorting",
		"test/fixtures/homework/hw-advanced",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range testPaths {
			creator.extractBranchName(path)
		}
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
