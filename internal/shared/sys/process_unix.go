//go:build !windows

package sys

import (
	"fmt"
	"os"
	"syscall"
)

// TerminateProcess sends SIGTERM to a process on Unix systems
func TerminateProcess(process *os.Process) error {
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM to process %d: %w", process.Pid, err)
	}
	return nil
}

// CheckProcessExists checks if a process exists on Unix systems
func CheckProcessExists(process *os.Process) error {
	return process.Signal(syscall.Signal(0))
}

// GetProcessAttributes returns process attributes for Unix systems
func GetProcessAttributes() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setsid: true, // Create new session
	}
}
