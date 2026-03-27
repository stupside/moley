//go:build windows

package sys

import (
	"fmt"
	"os"
	"syscall"
)

// TerminateProcess terminates a process on Windows.
func TerminateProcess(p *os.Process) error {
	if err := p.Kill(); err != nil {
		return fmt.Errorf("failed to terminate process %d: %w", p.Pid, err)
	}
	return nil
}

// GetProcessAttributes returns process attributes for Windows systems.
func GetProcessAttributes() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
