package assignment

import (
	"testing"
)

// TestExtractBranchName tests branch name extraction from assignment paths
func TestExtractBranchName(t *testing.T) {
	// Test patterns for branch name extraction
	patterns := []string{
		`^(?P<branch>assignment-\d+)$`,
		`^(?P<course>[^/]+)/(?P<type>hw)-(?P<number>\d+)$`,
		`^test/fixtures/(?P<type>assignments|homework|labs|projects)/(?P<name>[^/]+)$`,
		`^test/fixtures/courses/(?P<course>[^/]+)/(?P<week>[^/]+)/(?P<assignment>[^/]+)$`,
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
			// Create an AssignmentProcessor to test the branch name extraction
			processor, err := NewAssignmentProcessor("", []string{}, patterns)
			if err != nil {
				t.Fatalf("Failed to create AssignmentProcessor: %v", err)
			}

			branch, matched := processor.extractBranchNameFromPath(tt.assignmentPath)

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
	// Test patterns using unnamed groups
	// Note: More specific patterns should come first to avoid conflicts
	patterns := []string{
		// Unnamed groups pattern: (type)-(number)
		`^(assignment)-(\d+)$`,
		// Single unnamed group - match only the relevant part (SPECIFIC - must come before general course pattern)
		`^homework/(hw-\d+)$`,
		// Mixed named and unnamed: named-unnamed-named (GENERAL - comes after specific patterns)
		`^(?P<course>[^/]+)/(week-\d+)/assignment-(\d+)$`,
	}

	tests := []struct {
		name           string
		assignmentPath string
		expectedBranch string
		expectedMatch  bool
	}{
		{
			name:           "unnamed groups assignment",
			assignmentPath: "assignment-1",
			expectedBranch: "assignment-1",
			expectedMatch:  true,
		},
		{
			name:           "single unnamed group homework",
			assignmentPath: "homework/hw-3",
			expectedBranch: "hw-3",
			expectedMatch:  true,
		},
		{
			name:           "mixed named and unnamed groups",
			assignmentPath: "CS101/week-05/assignment-2",
			expectedBranch: "cs101-week-05-2", // Named groups first (alphabetically), then unnamed
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
			// Create an AssignmentProcessor to test the branch name extraction
			processor, err := NewAssignmentProcessor("", []string{}, patterns)
			if err != nil {
				t.Fatalf("Failed to create AssignmentProcessor: %v", err)
			}

			branch, matched := processor.extractBranchNameFromPath(tt.assignmentPath)

			if matched != tt.expectedMatch {
				t.Errorf("Expected match=%t, got=%t", tt.expectedMatch, matched)
			}

			if branch != tt.expectedBranch {
				t.Errorf("Expected branch=%s, got=%s", tt.expectedBranch, branch)
			}
		})
	}
}

// TestExtractBranchNameAlphabeticalOrdering tests that named groups are ordered alphabetically
func TestExtractBranchNameAlphabeticalOrdering(t *testing.T) {
	// Pattern with named groups that should be ordered alphabetically
	patterns := []string{
		`^(?P<zebra>[^/]+)/(?P<alpha>[^/]+)/(?P<beta>[^/]+)$`,
	}

	tests := []struct {
		name           string
		assignmentPath string
		expectedBranch string
		expectedMatch  bool
	}{
		{
			name:           "alphabetical ordering test",
			assignmentPath: "first/second/third",
			expectedBranch: "second-third-first", // Should be alpha-beta-zebra
			expectedMatch:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an AssignmentProcessor to test the branch name extraction
			processor, err := NewAssignmentProcessor("", []string{}, patterns)
			if err != nil {
				t.Fatalf("Failed to create AssignmentProcessor: %v", err)
			}

			branch, matched := processor.extractBranchNameFromPath(tt.assignmentPath)

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
	// Create an AssignmentProcessor to test the sanitization method
	processor, err := NewAssignmentProcessor("", []string{}, []string{})
	if err != nil {
		t.Fatalf("Failed to create AssignmentProcessor: %v", err)
	}

	tests := []struct {
		assignmentPath string
		expected       string
	}{
		{"Simple Test", "simple-test"},
		{"Test/With/Slashes", "test-with-slashes"},
		{"test   with   spaces", "test-with-spaces"},
		{"Test-With-Hyphens", "test-with-hyphens"},
		{"--leading-trailing--", "leading-trailing"},
		{"   trim   whitespace   ", "trim-whitespace"},
		{"UPPERCASE", "uppercase"},
		{"Mixed-Case_With_Underscores", "mixed-case_with_underscores"},
		{"Test@With#Special$Chars", "test@with#special$chars"},
		{"Test123Numbers456", "test123numbers456"},
		{"", ""},
		{"   ", ""},
		{"---", ""},
	}

	for _, tt := range tests {
		t.Run(tt.assignmentPath, func(t *testing.T) {
			result := processor.SanitizeBranchName(tt.assignmentPath)

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestParseRegexPatterns tests that regex patterns trigger proper branch name extraction
func TestParseRegexPatterns(t *testing.T) {
	// Test pattern that should fail regex compilation for error handling
	invalidPatterns := []string{
		"[", // Invalid regex pattern
	}

	// Create an AssignmentProcessor with invalid patterns
	_, err := NewAssignmentProcessor("", []string{}, invalidPatterns)
	if err == nil {
		t.Error("Expected error for invalid regex patterns, but got none")
	}

	// Test with valid patterns
	validPatterns := []string{
		`^assignment-(\d+)$`,
	}

	processor, err := NewAssignmentProcessor("", []string{}, validPatterns)
	if err != nil {
		t.Fatalf("Expected no error for valid regex patterns, but got: %v", err)
	}

	// Test that the processor can extract branch names correctly
	branch, matched := processor.extractBranchNameFromPath("assignment-1")
	if !matched {
		t.Error("Expected match for valid assignment path")
	}
	if branch != "1" {
		t.Errorf("Expected branch '1', got '%s'", branch)
	}
}