package assignment

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"assignment-pull-request/internal/regex"
)

// Info represents an assignment with its path and generated branch name
type Info struct {
	Path       string
	BranchName string
}

// Processor handles assignment discovery and processing
type Processor struct {
	repositoryRoot    string
	assignmentPattern *regex.Processor
}

// NewProcessor creates a new Processor with assignment regex patterns
func NewProcessor(repositoryRoot string, assignmentProcessor *regex.Processor) (*Processor, error) {
	// Get compiled patterns to validate assignment patterns have capturing groups
	assignmentPatterns, err := assignmentProcessor.Compiled()
	if err != nil {
		return nil, fmt.Errorf("failed to compile assignment patterns: %w", err)
	}

	// Validate that assignment patterns have capturing groups
	for _, pattern := range assignmentPatterns {
		if !hasCapturingGroups(pattern) {
			return nil, fmt.Errorf("assignment regex '%s' must contain at least one capturing group (e.g., (?P<name>...) or (...)) to extract branch names", pattern.String())
		}
	}

	return &Processor{
		repositoryRoot:    repositoryRoot,
		assignmentPattern: assignmentProcessor,
	}, nil
}

// ProcessAssignments discovers all assignments and returns assignment info with unique branch names
func (ap *Processor) ProcessAssignments() ([]Info, error) {
	fmt.Printf("🎯 Processing assignments for branch matching...\n")

	// Find all assignment paths
	assignments, err := ap.findAssignments()
	if err != nil {
		return nil, fmt.Errorf("error finding assignments: %w", err)
	}

	if len(assignments) == 0 {
		fmt.Printf("⚠️  No assignment folders found\n")
		return nil, nil
	}

	fmt.Printf("Debug: Found %d assignment folder(s), extracting branch names...\n", len(assignments))

	var results []Info
	branchCounts := make(map[string]int)

	for _, assignment := range assignments {
		fmt.Printf("  Processing assignment: %s\n", assignment)
		branchName, found := ap.extractBranchNameFromPath(assignment)
		if found {
			branchCounts[branchName]++
			uniqueBranchName := branchName
			if branchCounts[branchName] > 1 {
				uniqueBranchName = fmt.Sprintf("%s-%d", branchName, branchCounts[branchName])
			}
			fmt.Printf("    ✅ Branch name: %s\n", uniqueBranchName)
			results = append(results, Info{
				Path:       assignment,
				BranchName: uniqueBranchName,
			})
		} else {
			fmt.Printf("    ⚠️  Could not extract branch name\n")
		}
	}

	fmt.Printf("📊 Branch extraction summary:\n")
	fmt.Printf("  - Assignments processed: %d\n", len(assignments))
	fmt.Printf("  - Branch names extracted: %d\n", len(results))
	if len(results) > 0 {
		fmt.Printf("  - Available branches:\n")
		for _, result := range results {
			fmt.Printf("    * %s (from %s)\n", result.BranchName, result.Path)
		}
	}

	return results, nil
}

// validateBranchNameUniqueness checks that all assignments generate unique branch names
func (ap *Processor) validateBranchNameUniqueness(assignments []Info) error {
	branchToAssignments := make(map[string][]string)

	// Collect all branch names and track which assignments generate them
	for _, assignment := range assignments {
		branchToAssignments[assignment.BranchName] = append(branchToAssignments[assignment.BranchName], assignment.Path)
	}

	// Check for duplicates
	var conflicts []string
	for branchName, assignmentPaths := range branchToAssignments {
		if len(assignmentPaths) > 1 {
			conflicts = append(conflicts, fmt.Sprintf("Branch '%s' would be created by multiple assignments: %v", branchName, assignmentPaths))
		}
	}

	if len(conflicts) > 0 {
		return fmt.Errorf("branch name conflicts detected:\n  %s", strings.Join(conflicts, "\n  "))
	}

	return nil
}

// findAssignments finds all assignment folders matching the processor's regex patterns
func (ap *Processor) findAssignments() ([]string, error) {
	fmt.Printf("📁 Searching for assignment folders...\n")
	var assignments []string

	// Determine the root directory to walk
	rootDir := ap.repositoryRoot
	if rootDir == "" {
		rootDir = "."
	}
	fmt.Printf("Debug: Walking directory tree from: %s\n", rootDir)

	// Get compiled patterns for debugging
	assignmentPatterns, err := ap.assignmentPattern.Compiled()
	if err != nil {
		return nil, fmt.Errorf("failed to compile assignment patterns: %w", err)
	}
	fmt.Printf("Debug: Using %d compiled assignment patterns\n", len(assignmentPatterns))

	checkedDirs := 0
	matchedDirs := 0

	// Walk the entire directory tree and check each directory against assignment patterns
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and files (but not the current directory ".")
		baseName := filepath.Base(path)
		if strings.HasPrefix(baseName, ".") && path != "." && path != rootDir {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process directories
		if !info.IsDir() {
			return nil
		}

		// Skip the root directory itself
		if path == rootDir {
			return nil
		}

		checkedDirs++

		// Normalize path to use forward slashes for pattern matching
		normalizedPath := filepath.ToSlash(path)
		fmt.Printf("  Checking directory: %s\n", normalizedPath)

		for i, assignmentPattern := range assignmentPatterns {
			fmt.Printf("    Testing pattern %d: %s\n", i+1, assignmentPattern.String())
			if assignmentPattern.MatchString(normalizedPath) {
				assignments = append(assignments, path)
				matchedDirs++
				fmt.Printf("    ✅ MATCH! Added: %s\n", path)
				break // Don't check other patterns for this path
			} else {
				fmt.Printf("    - No match\n")
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error finding assignments: %w", err)
	}

	// Sort assignments
	sort.Strings(assignments)

	fmt.Printf("📊 Assignment discovery summary:\n")
	fmt.Printf("  - Directories checked: %d\n", checkedDirs)
	fmt.Printf("  - Assignments found: %d\n", matchedDirs)
	fmt.Printf("  - Assignment paths: %v\n", assignments)

	return assignments, nil
}

// GetAssignmentRegexStrings returns the assignment regex patterns as strings
func (ap *Processor) GetAssignmentRegexStrings() []string {
	return ap.assignmentPattern.Patterns()
}

// extractBranchNameFromPath extracts a branch name from a path using the processor's compiled patterns
func (ap *Processor) extractBranchNameFromPath(assignmentPath string) (string, bool) {
	fmt.Printf("    Debug: Extracting branch name from: %s\n", assignmentPath)

	assignmentPatterns, err := ap.assignmentPattern.Compiled()
	if err != nil {
		fmt.Printf("    Error: Failed to compile patterns: %v\n", err)
		return "", false
	}

	// Normalize path to use forward slashes for pattern matching
	normalizedPath := filepath.ToSlash(assignmentPath)
	fmt.Printf("    Debug: Normalized path: %s\n", normalizedPath)

	for i, pattern := range assignmentPatterns {
		if pattern == nil {
			continue
		}

		fmt.Printf("    Debug: Testing pattern %d: %s\n", i+1, pattern.String())
		matches := pattern.FindStringSubmatch(normalizedPath)
		if matches != nil {
			fmt.Printf("    Debug: Pattern matched! Found %d groups: %v\n", len(matches), matches)
			names := pattern.SubexpNames()
			fmt.Printf("    Debug: Group names: %v\n", names)
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
						fmt.Printf("    Debug: Named group '%s' = '%s'\n", name, part)
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
						fmt.Printf("    Debug: Unnamed group %d = '%s'\n", i, part)
					}
				}
			}

			// Add unnamed groups after named groups
			branchParts = append(branchParts, unnamedParts...)
			fmt.Printf("    Debug: All branch parts: %v\n", branchParts)

			if len(branchParts) == 0 {
				fmt.Printf("    Debug: No branch parts found, continuing to next pattern\n")
				continue
			}

			// Combine parts and sanitize
			branchName := strings.Join(branchParts, "-")
			branchName = ap.sanitizeBranchName(branchName)
			fmt.Printf("    Debug: Final branch name: '%s'\n", branchName)

			return branchName, true
		} else {
			fmt.Printf("    Debug: Pattern did not match\n")
		}
	}

	fmt.Printf("    Debug: No patterns matched\n")
	return "", false
}

// hasCapturingGroups checks if a compiled regex pattern has at least one capturing group (named or unnamed)
func hasCapturingGroups(regex *regexp.Regexp) bool {
	names := regex.SubexpNames()
	// SubexpNames() returns a slice where the first element is always an empty string
	// for the entire match. If there are more elements, there are capturing groups
	return len(names) > 1
}

// sanitizeBranchName sanitizes a branch name to match Creator's original behavior
// Only sanitizes spaces and slashes, preserves other special characters
func (ap *Processor) sanitizeBranchName(name string) string {
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
