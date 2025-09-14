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
	rootFolder        string
	rootPattern       *regex.Processor
	assignmentPattern *regex.Processor
}

// NewProcessor creates a new Processor with regex pattern processors
func NewProcessor(rootFolder string, rootProcessor, assignmentProcessor *regex.Processor) (*Processor, error) {
	// Get compiled patterns to validate assignment patterns have capturing groups
	assignmentPatterns, err := assignmentProcessor.Compiled()
	if err != nil {
		return nil, fmt.Errorf("failed to compile assignment patterns: %w", err)
	}

	// Validate that assignment patterns have capturing groups
	for _, pattern := range assignmentPatterns {
		if !HasCapturingGroups(pattern) {
			return nil, fmt.Errorf("assignment regex '%s' must contain at least one capturing group (e.g., (?P<name>...) or (...)) to extract branch names", pattern.String())
		}
	}

	return &Processor{
		rootFolder:        rootFolder,
		rootPattern:       rootProcessor,
		assignmentPattern: assignmentProcessor,
	}, nil
}

// HasCapturingGroups checks if a compiled regex pattern has at least one capturing group (named or unnamed)
func HasCapturingGroups(regex *regexp.Regexp) bool {
	names := regex.SubexpNames()
	// SubexpNames() returns a slice where the first element is always an empty string
	// for the entire match. If there are more elements, there are capturing groups
	return len(names) > 1
}

// ProcessAssignments discovers all assignments and returns assignment info with unique branch names
func (ap *Processor) ProcessAssignments() ([]Info, error) {
	// Find all assignment paths
	assignments, err := ap.findAssignments()
	if err != nil {
		return nil, err
	}

	if len(assignments) == 0 {
		return []Info{}, nil
	}

	// Convert to Info with branch names
	var assignmentInfos []Info
	for _, assignmentPath := range assignments {
		branchName, matched := ap.extractBranchNameFromPath(assignmentPath)
		if !matched {
			// Skip assignments that don't match any pattern
			continue
		}

		assignmentInfos = append(assignmentInfos, Info{
			Path:       assignmentPath,
			BranchName: branchName,
		})
	}

	// Validate branch name uniqueness
	if err := ap.validateBranchNameUniqueness(assignmentInfos); err != nil {
		return nil, err
	}

	return assignmentInfos, nil
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
	var assignments []string

	// Determine the root directory to walk
	rootDir := ap.rootFolder
	if rootDir == "" {
		rootDir = "."
	}

	// Walk through the directory tree
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
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

		// Check if this directory matches any root pattern
		rootPatterns, err := ap.rootPattern.Compiled()
		if err != nil {
			return err
		}
		for _, rootPattern := range rootPatterns {
			if rootPattern.MatchString(path) {
				// Find assignments in this root directory
				assignmentsInRoot, err := ap.findAssignmentsInDirectory(path)
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

	// Sort assignments
	sort.Strings(assignments)

	return assignments, nil
}

// findAssignmentsInDirectory finds assignments within a specific directory
func (ap *Processor) findAssignmentsInDirectory(rootDir string) ([]string, error) {
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
		assignmentPatterns, err := ap.assignmentPattern.Compiled()
		if err != nil {
			return nil
		}
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

// GetRootRegexStrings returns the root regex patterns as strings
func (ap *Processor) GetRootRegexStrings() []string {
	return ap.rootPattern.Patterns()
}

// GetAssignmentRegexStrings returns the assignment regex patterns as strings
func (ap *Processor) GetAssignmentRegexStrings() []string {
	return ap.assignmentPattern.Patterns()
}

// extractBranchNameFromPath extracts a branch name from a path using the processor's compiled patterns
func (ap *Processor) extractBranchNameFromPath(assignmentPath string) (string, bool) {
	assignmentPatterns, err := ap.assignmentPattern.Compiled()
	if err != nil {
		return "", false
	}

	for _, pattern := range assignmentPatterns {
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
			branchName = ap.sanitizeBranchName(branchName)

			return branchName, true
		}
	}

	return "", false
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
