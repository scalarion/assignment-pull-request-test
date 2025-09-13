package assignment

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// FindAssignments finds all assignment folders matching the given regex patterns
func FindAssignments(rootRegexPatterns, assignmentRegexPatterns []string) ([]string, error) {
	// Compile regex patterns
	rootPatterns, err := CompilePatterns(rootRegexPatterns)
	if err != nil {
		return nil, fmt.Errorf("invalid root regex patterns: %w", err)
	}

	assignmentPatterns, err := CompilePatterns(assignmentRegexPatterns)
	if err != nil {
		return nil, fmt.Errorf("invalid assignment regex patterns: %w", err)
	}

	return findAssignments(rootPatterns, assignmentPatterns)
}

// findAssignments is the core assignment finding logic
func findAssignments(rootPatterns, assignmentPatterns []*regexp.Regexp) ([]string, error) {
	var assignments []string

	// Walk through the current directory tree
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and files (but not the current directory ".")
		baseName := filepath.Base(path)
		if strings.HasPrefix(baseName, ".") && path != "." {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process directories
		if !info.IsDir() {
			return nil
		}

		// Check if this directory matches any root pattern
		for _, rootPattern := range rootPatterns {
			if rootPattern.MatchString(path) {
				// Find assignments in this root directory
				assignmentsInRoot, err := findAssignmentsInDirectory(path, assignmentPatterns)
				if err != nil {
					return err
				}
				assignments = append(assignments, assignmentsInRoot...)
				break // Don't check other root patterns for this path
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory tree: %w", err)
	}

	// Sort
	sort.Strings(assignments)

	return assignments, nil
}

// findAssignmentsInDirectory finds assignments within a specific directory
func findAssignmentsInDirectory(rootDir string, assignmentPatterns []*regexp.Regexp) ([]string, error) {
	var assignments []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and files
		if strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process directories
		if !info.IsDir() {
			return nil
		}

		// Check if this directory matches any assignment pattern
		for _, assignmentPattern := range assignmentPatterns {
			if assignmentPattern.MatchString(path) {
				assignments = append(assignments, path)
				break // Don't check other patterns for this path
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory %s: %w", rootDir, err)
	}

	return assignments, nil
}

// ExtractBranchNameFromPath extracts a branch name from a path using regex patterns
func ExtractBranchNameFromPath(assignmentPath string, assignmentPatterns []string) (string, bool) {
	// Compile patterns
	compiledPatterns := make([]*regexp.Regexp, len(assignmentPatterns))
	for i, pattern := range assignmentPatterns {
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			// Skip invalid patterns
			continue
		}
		compiledPatterns[i] = compiled
	}

	return ExtractBranchNameFromCompiledPatterns(assignmentPath, compiledPatterns)
}

// ExtractBranchNameFromCompiledPatterns extracts a branch name from a path using compiled regex patterns
func ExtractBranchNameFromCompiledPatterns(assignmentPath string, compiledPatterns []*regexp.Regexp) (string, bool) {
	for _, pattern := range compiledPatterns {
		if pattern == nil {
			continue
		}

		matches := pattern.FindStringSubmatch(assignmentPath)
		if matches != nil {
			names := pattern.SubexpNames()
			var branchParts []string

			// Collect named groups and their values, sorted alphabetically by name
			namedGroups := make(map[string]string)
			var namedGroupNames []string

			for i, name := range names {
				if name != "" && i < len(matches) && matches[i] != "" {
					part := strings.TrimSpace(matches[i])
					if part != "" {
						namedGroups[name] = part
						namedGroupNames = append(namedGroupNames, name)
					}
				}
			}

			// Sort named group names alphabetically
			if len(namedGroupNames) > 0 {
				sort.Strings(namedGroupNames)
				// Add named groups in alphabetical order
				for _, name := range namedGroupNames {
					branchParts = append(branchParts, namedGroups[name])
				}
			}

			// Collect unnamed groups in order of appearance
			var unnamedParts []string
			for i := 1; i < len(matches); i++ { // Skip index 0 (full match)
				// Skip if this index corresponds to a named group
				isNamedGroup := false
				if i < len(names) && names[i] != "" {
					isNamedGroup = true
				}

				if !isNamedGroup && matches[i] != "" {
					part := strings.TrimSpace(matches[i])
					if part != "" {
						unnamedParts = append(unnamedParts, part)
					}
				}
			}

			// Add unnamed groups after named groups
			branchParts = append(branchParts, unnamedParts...)

			if len(branchParts) == 0 {
				continue
			}

			// Combine parts and sanitize
			branchName := strings.Join(branchParts, "-")
			branchName = SanitizeBranchName(branchName)

			return branchName, true
		}
	}

	return "", false
}

// SanitizeBranchName sanitizes a branch name to match Creator's original behavior
// Only sanitizes spaces and slashes, preserves other special characters
func SanitizeBranchName(name string) string {
	// Remove leading/trailing whitespace
	branchName := strings.TrimSpace(name)

	// Replace spaces with hyphens
	branchName = regexp.MustCompile(`\s+`).ReplaceAllString(branchName, "-")

	// Remove slashes
	branchName = strings.ReplaceAll(branchName, "/", "-")

	// Remove consecutive hyphens
	branchName = regexp.MustCompile(`-+`).ReplaceAllString(branchName, "-")

	// Convert to lowercase
	branchName = strings.ToLower(branchName)

	// Remove leading/trailing hyphens
	branchName = strings.Trim(branchName, "-")

	return branchName
}

// CompilePatterns compiles string patterns into regex patterns
func CompilePatterns(patterns []string) ([]*regexp.Regexp, error) {
	compiled := make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern '%s': %w", pattern, err)
		}
		compiled[i] = regex
	}
	return compiled, nil
}

// ParseRegexPatterns parses a comma-separated string of regex patterns into a slice
// Supports escaping commas with \, to allow commas within regex patterns
func ParseRegexPatterns(patterns string) []string {
	if patterns == "" {
		return []string{}
	}

	// Replace escaped commas with a placeholder to preserve them
	placeholder := "\x00ESCAPED_COMMA\x00"
	patterns = strings.ReplaceAll(patterns, "\\,", placeholder)

	// Split by unescaped commas and trim whitespace
	parts := strings.Split(patterns, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			// Restore escaped commas
			restored := strings.ReplaceAll(trimmed, placeholder, ",")
			result = append(result, restored)
		}
	}
	return result
}
