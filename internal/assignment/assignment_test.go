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
			branch, matched := ExtractBranchNameFromPath(tt.assignmentPath, patterns)

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
		`^(?P<course>[^/]+)/(hw|lab)-(\d+)$`,
		// Multiple unnamed groups
		`^(projects)/(semester-\d+)/(week-\d+)/(assignment-[^/]+)$`,
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
			branch, matched := ExtractBranchNameFromPath(tt.assignmentPath, patterns)

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
	patterns := []string{
		// Multiple named groups to test alphabetical ordering
		`^(?P<module>[^/]+)/(?P<course>[^/]+)/(?P<assignment>[^/]+)$`,
		// Mixed named and unnamed groups
		`^(?P<year>\d+)/(?P<course>[^/]+)/(week-\d+)/(assignment-\d+)$`,
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
		{
			name:           "mixed groups with specific naming order",
			assignmentPath: "group1/group2/group3/group4",
			expectedBranch: "group4-group2-group1-group3", // 01named (group4), 02named (group2), then unnamed in order (group1, group3)
			expectedMatch:  true,
		},
	}

	// Add pattern for the specific naming order test
	patterns = append(patterns,
		// Pattern with groups: 1st unnamed, 2nd named (02prefix), 3rd unnamed, 4th named (01prefix)
		`^([^/]+)/(?P<02named>[^/]+)/([^/]+)/(?P<01named>[^/]+)$`,
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			branch, matched := ExtractBranchNameFromPath(tt.assignmentPath, patterns)

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
			result := SanitizeBranchName(tt.assignmentPath)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// BenchmarkExtractBranchName benchmarks branch name extraction
func BenchmarkExtractBranchName(b *testing.B) {
	patterns := []string{
		`^(?P<branch>assignment-\d+)$`,
		`^(?P<course>[^/]+)/(?P<type>hw)-(?P<number>\d+)$`,
		`^test/fixtures/(?P<type>assignments|homework|labs|projects)/(?P<name>[^/]+)$`,
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
			ExtractBranchNameFromPath(path, patterns)
		}
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
		{
			name:     "pattern with escaped comma",
			input:    `^(?P<options>a\,b\,c)$`,
			expected: []string{"^(?P<options>a,b,c)$"},
		},
		{
			name:     "multiple patterns with escaped commas",
			input:    `^(?P<list>a\,b)$,^(?P<items>x\,y\,z)$`,
			expected: []string{"^(?P<list>a,b)$", "^(?P<items>x,y,z)$"},
		},
		{
			name:     "mixed escaped and unescaped commas",
			input:    `^(?P<options>a\,b)$,^homework$,^(?P<choices>x\,y)$`,
			expected: []string{"^(?P<options>a,b)$", "^homework$", "^(?P<choices>x,y)$"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseRegexPatterns(tt.input)
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
