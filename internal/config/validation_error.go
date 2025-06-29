package config

import (
	"fmt"
	"strings"
)

// ValidationError represents configuration validation errors
type ValidationError struct {
	Issues []string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if len(e.Issues) == 0 {
		return "configuration validation failed"
	}

	if len(e.Issues) == 1 {
		return fmt.Sprintf("configuration validation failed: %s", e.Issues[0])
	}

	return fmt.Sprintf("configuration validation failed:\n  - %s", strings.Join(e.Issues, "\n  - "))
}

// AddIssue adds a validation issue to the error
func (e *ValidationError) AddIssue(issue string) {
	e.Issues = append(e.Issues, issue)
}

// HasIssues returns true if there are validation issues
func (e *ValidationError) HasIssues() bool {
	return len(e.Issues) > 0
}

// IssuesCount returns the number of validation issues
func (e *ValidationError) IssuesCount() int {
	return len(e.Issues)
}
