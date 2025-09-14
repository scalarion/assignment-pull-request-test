package regex

import (
	"fmt"
	"regexp"
	"strings"
)

// PatternProcessor handles regex pattern parsing, compilation, and automatic deduplication
type PatternProcessor struct {
	patterns         []string
	patternSet       map[string]bool // For efficient deduplication
	compiledPatterns []*regexp.Regexp
	compiled         bool // Track if patterns have been compiled
}

// NewPatternProcessor creates a new PatternProcessor
func NewPatternProcessor() *PatternProcessor {
	return &PatternProcessor{
		patterns:         make([]string, 0),
		patternSet:       make(map[string]bool),
		compiledPatterns: make([]*regexp.Regexp, 0),
		compiled:         false,
	}
}

// NewPatternProcessorWithPatterns creates a new PatternProcessor with the given patterns
func NewPatternProcessorWithPatterns(patterns []string) *PatternProcessor {
	pp := NewPatternProcessor()
	pp.AddPatterns(patterns)
	return pp
}

// NewPatternProcessorWithCommaSeparated creates a new PatternProcessor with comma-separated patterns
func NewPatternProcessorWithCommaSeparated(patterns string) *PatternProcessor {
	pp := NewPatternProcessor()
	pp.AddCommaSeparatedPatterns(patterns)
	return pp
}


// AddPatterns adds string patterns to the processor with automatic deduplication
func (pp *PatternProcessor) AddPatterns(patterns []string) {
	for _, pattern := range patterns {
		pp.addPattern(pattern)
	}
}

// AddCommaSeparatedPatterns adds comma-separated patterns to the processor with automatic deduplication
func (pp *PatternProcessor) AddCommaSeparatedPatterns(patterns string) {
	parsed := ParseCommaSeparated(patterns)
	pp.AddPatterns(parsed)
}

// GetPatterns returns the string patterns (already deduplicated)
func (pp *PatternProcessor) GetPatterns() []string {
	return pp.patterns
}

// GetCompiledPatterns returns the compiled regex patterns, compiling them automatically if needed
func (pp *PatternProcessor) GetCompiledPatterns() ([]*regexp.Regexp, error) {
	if !pp.compiled {
		if err := pp.compilePatterns(); err != nil {
			return nil, err
		}
		pp.compiled = true
	}
	return pp.compiledPatterns, nil
}

// addPattern adds a single pattern with automatic deduplication
func (pp *PatternProcessor) addPattern(pattern string) {
	// Only add if pattern is non-empty and not already present
	if pattern != "" && !pp.patternSet[pattern] {
		pp.patterns = append(pp.patterns, pattern)
		pp.patternSet[pattern] = true
		pp.compiled = false // Mark as needing recompilation
	}
}

// parseCommaSeparated parses a comma-separated string of regex patterns into a slice
// Supports escaping commas with \, to allow commas within regex patterns
func ParseCommaSeparated(patterns string) []string {
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

// compilePatterns compiles all string patterns into regex patterns
func (pp *PatternProcessor) compilePatterns() error {
	compiled := make([]*regexp.Regexp, len(pp.patterns))
	for i, pattern := range pp.patterns {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid regex pattern '%s': %w", pattern, err)
		}
		compiled[i] = regex
	}
	pp.compiledPatterns = compiled
	return nil
}

