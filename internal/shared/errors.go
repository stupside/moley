package shared

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

// MoleyError is a custom error type for wrapping errors in Moley
type MoleyError struct {
	Op  string // operation or context
	Err error  // wrapped error
	Msg string // additional message
}

// NewMoleyError creates a new MoleyError instance
func (e *MoleyError) Error() string {
	if e.Msg != "" {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Msg, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Is checks if the error matches a specific type or message
func (e *MoleyError) Unwrap() error {
	return e.Err
}

// WrapError automatically captures the caller information and wraps an error
// This is the preferred way to wrap errors as it reduces boilerplate
func WrapError(err error, msg string) error {
	if err == nil {
		return nil
	}

	// Get caller information
	pc, file, line, ok := runtime.Caller(1)
	op := "unknown"

	if ok {
		// Get the function name
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			name := fn.Name()
			// Remove the package path prefix
			if lastDot := strings.LastIndexByte(name, '.'); lastDot != -1 {
				name = name[lastDot+1:]
			}
			// Get just the filename without the path
			file = filepath.Base(file)
			op = fmt.Sprintf("%s (%s:%d)", name, file, line)
		}
	}

	return &MoleyError{Op: op, Err: err, Msg: msg}
}

// Predefined base errors (for wrapping)
var (
	ErrConfigNil                      = errors.New("configuration cannot be nil")
	ErrConfigRead                     = errors.New("failed to read configuration file")
	ErrConfigSave                     = errors.New("failed to save configuration file")
	ErrConfigWrite                    = errors.New("failed to write configuration file")
	ErrConfigNotFound                 = errors.New("configuration file not found at the specified path")
	ErrConfigMarshal                  = errors.New("failed to marshal configuration")
	ErrConfigUnmarshal                = errors.New("failed to unmarshal configuration")
	ErrConfigValidation               = errors.New("configuration validation failed")
	ErrConfigAlreadyLoaded            = errors.New("configuration already loaded at this path")
	ErrConfigAlreadyLoadedInvalidType = errors.New("configuration already loading at this path has an invalid type")
)
