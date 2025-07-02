package errors

import (
	"fmt"
)

// Error types for different categories of errors
const (
	ErrTypeValidation = "validation"
	ErrTypeConfig     = "configuration"
	ErrTypeExecution  = "execution"
)

// Error codes for specific error conditions
const (
	ErrCodeInvalidConfig    = "INVALID_CONFIG"
	ErrCodeCommandFailed    = "COMMAND_FAILED"
	ErrCodePermissionDenied = "PERMISSION_DENIED"
)

// MoleyError represents a structured error with additional context
type MoleyError struct {
	Type       string            `json:"type"`
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    map[string]string `json:"details,omitempty"`
	Underlying error             `json:"-"`
	Context    map[string]string `json:"context,omitempty"`
}

// Error implements the error interface
func (e *MoleyError) Error() string {
	if e.Underlying != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Underlying)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(code, message string, underlying error) *MoleyError {
	return &MoleyError{
		Type:       ErrTypeValidation,
		Code:       code,
		Message:    message,
		Underlying: underlying,
	}
}

// NewConfigError creates a new configuration error
func NewConfigError(code, message string, underlying error) *MoleyError {
	return &MoleyError{
		Type:       ErrTypeConfig,
		Code:       code,
		Message:    message,
		Underlying: underlying,
	}
}

// NewExecutionError creates a new execution error
func NewExecutionError(code, message string, underlying error) *MoleyError {
	return &MoleyError{
		Type:       ErrTypeExecution,
		Code:       code,
		Message:    message,
		Underlying: underlying,
	}
}
