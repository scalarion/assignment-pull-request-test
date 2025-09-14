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

// Action represents the structure of a GitHub Actions workflow file
type Action struct {
	Jobs map[string]Job `yaml:"jobs"`
}

// Job represents a job in a GitHub Actions workflow
type Job struct {
	Uses string                 `yaml:"uses"`
	With map[string]interface{} `yaml:"with"`
}

// Processor handles workflow file parsing and pattern extraction
type Processor struct {
	rootPattern       *regex.Processor
	assignmentPattern *regex.Processor
}

// New creates a new workflow processor
func New() *Processor {
	return &Processor{
		rootPattern:       regex.New(),
		assignmentPattern: regex.New(),
	}
}

// RootPattern returns the regex processor for root patterns
func (p *Processor) RootPattern() *regex.Processor {
	return p.rootPattern
}

// AssignmentPattern returns the regex processor for assignment patterns
func (p *Processor) AssignmentPattern() *regex.Processor {
	return p.assignmentPattern
}

// ParseAllFiles finds and parses all workflow files
func (p *Processor) ParseAllFiles() error {
	workflowFiles, err := p.findFiles()
	if err != nil {
		return fmt.Errorf("error finding workflow files: %w", err)
	}

	for _, file := range workflowFiles {
		if err := p.parseFile(file); err != nil {
			// Continue with other files if one fails
			continue
		}
	}

	return nil
}

// findFiles finds all GitHub Actions workflow files in the repository
func (p *Processor) findFiles() ([]string, error) {
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

// isAssignmentAction checks if a job uses the assignment pull request action
func (p *Processor) isAssignmentAction(uses string) bool {
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

// parseFile parses a single workflow file and extracts patterns
func (p *Processor) parseFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading workflow file %s: %w", filePath, err)
	}

	var config Action
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("error parsing workflow file %s: %w", filePath, err)
	}

	// Look for jobs that use the assignment action
	for _, job := range config.Jobs {
		if p.isAssignmentAction(job.Uses) {
			if with := job.With; with != nil {
				// Extract root patterns
				if rootPatterns, ok := with[constants.WorkflowAssignmentsRootRegexKey]; ok {
					if rootStr, ok := rootPatterns.(string); ok {
						p.rootPattern.AddCommaSeparated(rootStr)
					}
				}

				// Extract assignment patterns
				if assignmentPatterns, ok := with[constants.WorkflowAssignmentRegexKey]; ok {
					if assignmentStr, ok := assignmentPatterns.(string); ok {
						p.assignmentPattern.AddCommaSeparated(assignmentStr)
					}
				}
			}
		}
	}

	return nil
}
