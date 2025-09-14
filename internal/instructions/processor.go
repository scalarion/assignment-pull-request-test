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
	assignmentPath string
}

// New creates a new instructions processor for the given assignment path
func New(assignmentPath string) *Processor {
	return &Processor{
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
	wrappedContent := fmt.Sprintf(`## Assignment Instructions

%s

---

*This pull request was automatically created by the Assignment Pull Request Creator action.*
*Original instructions from: %s*
`, processedContent, filepath.Base(instructionsPath))

	return wrappedContent, nil
}

// rewriteImageLinks rewrites relative image links to reference the assignment path
func (p *Processor) rewriteImageLinks(content string) string {
	// Regex to match markdown image syntax: ![alt text](relative/path/to/image)
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

		// Skip if it's already an absolute path from repo root
		if strings.HasPrefix(imagePath, "/") {
			return match
		}

		// Rewrite relative path to be relative to repo root
		rewrittenPath := filepath.Join(p.assignmentPath, imagePath)
		// Convert to forward slashes for web compatibility (Git/GitHub always uses forward slashes)
		rewrittenPath = filepath.ToSlash(rewrittenPath)

		return fmt.Sprintf("![%s](%s)", altText, rewrittenPath)
	})
}

// createGenericPullRequestBody creates the default generic pull request body
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