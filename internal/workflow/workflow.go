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
	assignmentPattern *regex.Processor
}

// New creates a new workflow processor
func New() *Processor {
	return &Processor{
		assignmentPattern: regex.New(),
	}
}

// AssignmentPattern returns the regex processor for assignment patterns
func (p *Processor) AssignmentPattern() *regex.Processor {
	return p.assignmentPattern
}

// ParseAllFiles finds and parses all workflow files
func (p *Processor) ParseAllFiles() error {
	fmt.Printf("📄 Starting workflow file parsing...\n")

	workflowFiles, err := p.findFiles()
	if err != nil {
		return fmt.Errorf("error finding workflow files: %w", err)
	}

	fmt.Printf("Debug: Found %d workflow files to process\n", len(workflowFiles))
	for i, file := range workflowFiles {
		fmt.Printf("  %d. %s\n", i+1, file)
	}

	if len(workflowFiles) == 0 {
		fmt.Printf("Warning: No workflow files found in common directories\n")
		return nil
	}

	parsedCount := 0
	skippedCount := 0

	for _, file := range workflowFiles {
		fmt.Printf("Debug: Parsing workflow file: %s\n", file)
		if err := p.parseFile(file); err != nil {
			fmt.Printf("  ⚠️  Skipped %s (parse error: %v)\n", file, err)
			skippedCount++
			// Continue with other files if one fails
			continue
		}
		fmt.Printf("  ✅ Successfully parsed %s\n", file)
		parsedCount++
	}

	fmt.Printf("📊 Workflow parsing summary:\n")
	fmt.Printf("  - Parsed: %d files\n", parsedCount)
	fmt.Printf("  - Skipped: %d files\n", skippedCount)
	fmt.Printf("  - Total patterns found: %d\n", len(p.assignmentPattern.Patterns()))

	if len(p.assignmentPattern.Patterns()) > 0 {
		fmt.Printf("  - Assignment regex patterns:\n")
		for i, pattern := range p.assignmentPattern.Patterns() {
			fmt.Printf("    %d. %s\n", i+1, pattern)
		}
	}

	return nil
}

// findFiles finds all GitHub Actions workflow files in the repository
func (p *Processor) findFiles() ([]string, error) {
	fmt.Printf("Debug: Searching for workflow files...\n")
	var workflowFiles []string

	// Check common workflow directories
	workflowDirs := []string{
		constants.GitHubActionsWorkflowDir,
		constants.GitHubWorkflowTemplatesDir,
	}

	fmt.Printf("Debug: Checking workflow directories: %v\n", workflowDirs)

	for _, dir := range workflowDirs {
		fmt.Printf("Debug: Checking directory: %s\n", dir)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			fmt.Printf("  - Directory does not exist: %s\n", dir)
			continue
		}
		fmt.Printf("  - Directory exists: %s\n", dir)

		fileCount := 0
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				fmt.Printf("  - Error walking path %s: %v\n", path, err)
				return err
			}

			if d.IsDir() {
				return nil
			}

			// Check for YAML/YML files
			ext := strings.ToLower(filepath.Ext(path))
			if ext == constants.YamlExtension || ext == constants.YamlAltExtension {
				workflowFiles = append(workflowFiles, path)
				fileCount++
				fmt.Printf("    + Found workflow file: %s\n", path)
			} else {
				fmt.Printf("    - Skipped non-YAML file: %s\n", path)
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("error walking workflow directory %s: %w", dir, err)
		}

		fmt.Printf("  - Found %d workflow files in %s\n", fileCount, dir)
	}

	fmt.Printf("Debug: Total workflow files found: %d\n", len(workflowFiles))
	return workflowFiles, nil
}

// isAssignmentAction checks if a job uses the assignment pull request action
func (p *Processor) isAssignmentAction(uses string) bool {
	if uses == "" {
		return false
	}

	// Check for local action reference
	if uses == "./" || uses == "." {
		fmt.Printf("          Debug: Detected local action reference\n")
		return true
	}

	// Check for GitHub repository references that might be this action
	// This is a heuristic - in practice, you might want to be more specific
	isMatch := strings.Contains(uses, constants.ActionName)
	if isMatch {
		fmt.Printf("          Debug: Detected assignment action by name match\n")
	}
	return isMatch
}

// parseFile parses a single workflow file and extracts patterns
func (p *Processor) parseFile(filePath string) error {
	fmt.Printf("    Debug: Reading file content from %s\n", filePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading workflow file %s: %w", filePath, err)
	}

	fmt.Printf("    Debug: File size: %d bytes\n", len(data))

	var config Action
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("error parsing workflow file %s: %w", filePath, err)
	}

	fmt.Printf("    Debug: Found %d jobs in workflow\n", len(config.Jobs))

	jobsWithAssignmentAction := 0
	patternsExtracted := 0

	// Look for jobs that use the assignment action
	for jobName, job := range config.Jobs {
		fmt.Printf("      Job '%s': uses='%s'\n", jobName, job.Uses)

		if p.isAssignmentAction(job.Uses) {
			jobsWithAssignmentAction++
			fmt.Printf("        ✅ Job uses assignment action\n")

			if with := job.With; with != nil {
				fmt.Printf("        Debug: Job has %d 'with' parameters\n", len(with))
				for key, value := range with {
					fmt.Printf("          - %s: %v\n", key, value)
				}

				// Extract assignment patterns
				if assignmentPatterns, ok := with[constants.WorkflowAssignmentRegexKey]; ok {
					fmt.Printf("        ✅ Found assignment patterns parameter\n")
					if assignmentStr, ok := assignmentPatterns.(string); ok {
						fmt.Printf("        Debug: Raw assignment patterns: %s\n", assignmentStr)
						beforeCount := len(p.assignmentPattern.Patterns())
						p.assignmentPattern.AddCommaSeparated(assignmentStr)
						afterCount := len(p.assignmentPattern.Patterns())
						newPatterns := afterCount - beforeCount
						patternsExtracted += newPatterns
						fmt.Printf("        ✅ Added %d new patterns (total now: %d)\n", newPatterns, afterCount)
					} else {
						fmt.Printf("        ⚠️  Assignment patterns value is not a string: %T\n", assignmentPatterns)
					}
				} else {
					fmt.Printf("        ⚠️  No '%s' parameter found in job\n", constants.WorkflowAssignmentRegexKey)
				}
			} else {
				fmt.Printf("        ⚠️  Job has no 'with' parameters\n")
			}
		} else {
			fmt.Printf("        - Job does not use assignment action\n")
		}
	}

	fmt.Printf("    📊 File parsing summary:\n")
	fmt.Printf("      - Jobs with assignment action: %d/%d\n", jobsWithAssignmentAction, len(config.Jobs))
	fmt.Printf("      - New patterns extracted: %d\n", patternsExtracted)

	return nil
}
