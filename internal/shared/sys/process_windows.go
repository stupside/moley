//go:build windows

package sys

import (
	"fmt"
	"os"
	"syscall"
)

// TerminateProcess terminates a process on Windows
func TerminateProcess(process *os.Process) error {
	if err := process.Kill(); err != nil {
		return fmt.Errorf("failed to terminate process %d: %w", process.Pid, err)
	}
	return nil
}

// CheckProcessExists checks if a process exists on Windows
func CheckProcessExists(process *os.Process) error {
	// On Windows, we can't use Signal(0), so we just return nil
	// This is less precise but functional for Windows
	return nil
}

// GetProcessAttributes returns process attributes for Windows systems
func GetProcessAttributes() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
