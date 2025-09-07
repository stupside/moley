// Package shared provides common utilities used across Moley.
package shared

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

type MoleyError struct {
	Op  string
	Err error
	Msg string
}

func (e *MoleyError) Error() string {
	if e.Msg != "" {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Msg, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *MoleyError) Unwrap() error {
	return e.Err
}

func WrapError(err error, msg string) error {
	if err == nil {
		return nil
	}

	pc, file, line, ok := runtime.Caller(1)
	op := "unknown"

	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			name := fn.Name()
			if lastDot := strings.LastIndexByte(name, '.'); lastDot != -1 {
				name = name[lastDot+1:]
			}
			file = filepath.Base(file)
			op = fmt.Sprintf("%s (%s:%d)", name, file, line)
		}
	}

	return &MoleyError{Op: op, Err: err, Msg: msg}
}
