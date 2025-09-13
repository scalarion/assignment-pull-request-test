package workflow

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"assignment-pull-request/internal/assignment"
	"assignment-pull-request/internal/constants"

	"gopkg.in/yaml.v3"
)

// WorkflowConfig represents the structure of a GitHub Actions workflow file
type WorkflowConfig struct {
	Jobs map[string]Job `yaml:"jobs"`
}

// Job represents a job in a GitHub Actions workflow
type Job struct {
	Uses string                 `yaml:"uses"`
	With map[string]interface{} `yaml:"with"`
}

// WorkflowPatterns contains the regex patterns extracted from workflow files
type WorkflowPatterns struct {
	RootPatterns       []string
	AssignmentPatterns []string
}

// FindWorkflowFiles finds all GitHub Actions workflow files in the repository
func FindWorkflowFiles() ([]string, error) {
	var workflowFiles []string

	// Check common workflow directories
	workflowDirs := []string{
		constants.GitHubActionsWorkflowDir,
		constants.GitHubWorkflowTemplatesDir,
	}

	for _, dir := range workflowDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			// Check for YAML/YML files
			ext := strings.ToLower(filepath.Ext(path))
			if ext == constants.YamlExtension || ext == constants.YamlAltExtension {
				workflowFiles = append(workflowFiles, path)
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("error walking workflow directory %s: %w", dir, err)
		}
	}

	return workflowFiles, nil
}

// ParseWorkflowFile parses a workflow file and extracts assignment patterns
func ParseWorkflowFile(filePath string) (*WorkflowPatterns, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading workflow file %s: %w", filePath, err)
	}

	var workflow WorkflowConfig
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return nil, fmt.Errorf("error parsing workflow file %s: %w", filePath, err)
	}

	patterns := &WorkflowPatterns{
		RootPatterns:       []string{},
		AssignmentPatterns: []string{},
	}

	// Look for jobs that use the assignment action
	for _, job := range workflow.Jobs {
		if isAssignmentAction(job.Uses) {
			if with := job.With; with != nil {
				// Extract root patterns
				if rootPatterns, ok := with["assignments-root-regex"]; ok {
					if rootStr, ok := rootPatterns.(string); ok {
						// Parse comma-separated patterns
						parsedPatterns := assignment.ParseRegexPatterns(rootStr)
						patterns.RootPatterns = append(patterns.RootPatterns, parsedPatterns...)
					}
				}

				// Extract assignment patterns
				if assignmentPatterns, ok := with["assignment-regex"]; ok {
					if assignmentStr, ok := assignmentPatterns.(string); ok {
						// Parse comma-separated patterns
						parsedPatterns := assignment.ParseRegexPatterns(assignmentStr)
						patterns.AssignmentPatterns = append(patterns.AssignmentPatterns, parsedPatterns...)
					}
				}
			}
		}
	}

	return patterns, nil
}

// ParseAllWorkflows finds and parses all workflow files, returning combined patterns
func ParseAllWorkflows() (*WorkflowPatterns, error) {
	workflowFiles, err := FindWorkflowFiles()
	if err != nil {
		return nil, fmt.Errorf("error finding workflow files: %w", err)
	}

	combined := &WorkflowPatterns{
		RootPatterns:       []string{},
		AssignmentPatterns: []string{},
	}

	for _, file := range workflowFiles {
		patterns, err := ParseWorkflowFile(file)
		if err != nil {
			// Continue with other files if one fails
			continue
		}

		combined.RootPatterns = append(combined.RootPatterns, patterns.RootPatterns...)
		combined.AssignmentPatterns = append(combined.AssignmentPatterns, patterns.AssignmentPatterns...)
	}

	// Remove duplicates
	combined.RootPatterns = removeDuplicateStrings(combined.RootPatterns)
	combined.AssignmentPatterns = removeDuplicateStrings(combined.AssignmentPatterns)

	return combined, nil
}

// isAssignmentAction checks if a job uses the assignment pull request action
func isAssignmentAction(uses string) bool {
	if uses == "" {
		return false
	}

	// Check for local action reference
	if uses == "./" || uses == "." {
		return true
	}

	// Check for GitHub repository references that might be this action
	// This is a heuristic - in practice, you might want to be more specific
	return strings.Contains(uses, constants.ActionName)
}

// removeDuplicateStrings removes duplicate strings from a slice
func removeDuplicateStrings(slice []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if item != "" && !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
