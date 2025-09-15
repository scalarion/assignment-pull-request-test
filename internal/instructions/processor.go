package instructions

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"assignment-pull-request/internal/constants"
)

// Processor handles reading and processing instruction files for a specific assignment
type Processor struct {
	branch         string
	assignmentPath string
}

// New creates a new instructions processor for the given assignment path
func New(assignmentPath string) *Processor {
	return &Processor{
		branch:         "main", // Default fallback
		assignmentPath: assignmentPath,
	}
}

// NewWithDefaults creates a new instructions processor with branch and assignment path
func NewWithDefaults(branch, assignmentPath string) *Processor {
	return &Processor{
		branch:         branch,
		assignmentPath: assignmentPath,
	}
}

// CreatePullRequestBody creates pull request body content from the processor's assignment path
func (p *Processor) CreatePullRequestBody() (string, error) {
	// Try to find instructions.md in the assignment directory
	instructionsPath := p.findInstructionsFile()

	if instructionsPath != "" {
		content, err := p.readAndProcessInstructions(instructionsPath)
		if err != nil {
			fmt.Printf("Warning: failed to read instructions file '%s': %v\n", instructionsPath, err)
			fmt.Printf("Falling back to generic template\n")
		} else {
			return content, nil
		}
	}

	// Fall back to generic template
	return p.createGenericPullRequestBody(), nil
}

// findInstructionsFile looks for instructions.md or INSTRUCTIONS.md in the assignment directory
func (p *Processor) findInstructionsFile() string {
	candidates := []string{
		filepath.Join(p.assignmentPath, constants.InstructionsFileName),
		filepath.Join(p.assignmentPath, constants.InstructionsFileNameUpper),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

// readAndProcessInstructions reads the instructions file and processes image links
func (p *Processor) readAndProcessInstructions(instructionsPath string) (string, error) {
	content, err := os.ReadFile(instructionsPath)
	if err != nil {
		return "", fmt.Errorf("failed to read instructions file: %w", err)
	}

	processedContent := p.rewriteImageLinks(string(content))

	// Wrap the content in a nice pull request format
	wrappedContent := fmt.Sprintf(`%s

---
<sub>*This pull request was automatically created by the Assignment Pull Request Creator action.*</sub>
<sub>*Original instructions from: %s*</sub>
`, processedContent, filepath.Base(instructionsPath))

	return wrappedContent, nil
}

// rewriteImageLinks rewrites relative image links to reference the assignment path
func (p *Processor) rewriteImageLinks(content string) string {
	// Regex to match markdown image syntax: ![alt text](relative/path/to/image)
	// Note: This handles standard paths; escaped parentheses in paths are extremely rare
	imageRegex := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

	return imageRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatches := imageRegex.FindStringSubmatch(match)
		if len(submatches) != 3 {
			return match // Return original if parsing fails
		}

		altText := submatches[1]
		imagePath := submatches[2]

		// Skip if it's already an absolute URL
		if strings.HasPrefix(imagePath, "http://") || strings.HasPrefix(imagePath, "https://") {
			return match
		}

		// Skip if it's already an absolute path from repo root (Unix-style in markdown)
		if strings.HasPrefix(imagePath, "/") {
			return match
		}

		// Check if it's an absolute path (cross-platform)
		if filepath.IsAbs(imagePath) {
			return match
		}

		// Rewrite relative path for GitHub pull requests and issues
		// Join the assignment path with the relative image path
		rewrittenPath := filepath.Join(p.assignmentPath, imagePath)
		// Ensure we use forward slashes for GitHub compatibility
		rewrittenPath = filepath.ToSlash(rewrittenPath)

		// For pull requests and issues, use blob URL format with ?raw=true
		// This ensures images display correctly in PR descriptions
		rewrittenPath = fmt.Sprintf("../blob/%s/%s?raw=true", p.branch, rewrittenPath)

		return fmt.Sprintf("![%s](%s)", altText, rewrittenPath)
	})
} // createGenericPullRequestBody creates the default generic pull request body
func (p *Processor) createGenericPullRequestBody() string {
	return fmt.Sprintf(`## Assignment Pull Request

This pull request contains the setup for the assignment located at
`+"`%s`"+`.

### Changes included:
- ✅ Created `+constants.ReadmeFileName+` with assignment template
- ✅ Set up branch structure for assignment submission

### Next steps:
1. Review the assignment requirements in the `+constants.ReadmeFileName+`
2. Add any additional assignment materials
3. Students can fork this repository and work on their submissions

---

*This pull request was automatically created by the Assignment Pull*
*Request Creator action.*
`, p.assignmentPath)
}
