package workflow

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"assignment-pull-request/internal/constants"
	"assignment-pull-request/internal/regex"

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

	// Use regex processor for pattern parsing (deduplication is automatic)
	rootProcessor := regex.New()
	assignmentProcessor := regex.New()

	// Look for jobs that use the assignment action
	for _, job := range workflow.Jobs {
		if isAssignmentAction(job.Uses) {
			if with := job.With; with != nil {
				// Extract root patterns
				if rootPatterns, ok := with[constants.WorkflowAssignmentsRootRegexKey]; ok {
					if rootStr, ok := rootPatterns.(string); ok {
						rootProcessor.AddCommaSeparated(rootStr)
					}
				}

				// Extract assignment patterns
				if assignmentPatterns, ok := with[constants.WorkflowAssignmentRegexKey]; ok {
					if assignmentStr, ok := assignmentPatterns.(string); ok {
						assignmentProcessor.AddCommaSeparated(assignmentStr)
					}
				}
			}
		}
	}

	return &WorkflowPatterns{
		RootPatterns:       rootProcessor.Patterns(),
		AssignmentPatterns: assignmentProcessor.Patterns(),
	}, nil
}

// ParseAllWorkflows finds and parses all workflow files, returning combined patterns
func ParseAllWorkflows() (*WorkflowPatterns, error) {
	workflowFiles, err := FindWorkflowFiles()
	if err != nil {
		return nil, fmt.Errorf("error finding workflow files: %w", err)
	}

	// Use regex processors for combined pattern handling (deduplication is automatic)
	rootProcessor := regex.New()
	assignmentProcessor := regex.New()

	for _, file := range workflowFiles {
		patterns, err := ParseWorkflowFile(file)
		if err != nil {
			// Continue with other files if one fails
			continue
		}

		rootProcessor.Add(patterns.RootPatterns...)
		assignmentProcessor.Add(patterns.AssignmentPatterns...)
	}

	return &WorkflowPatterns{
		RootPatterns:       rootProcessor.Patterns(),
		AssignmentPatterns: assignmentProcessor.Patterns(),
	}, nil
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
