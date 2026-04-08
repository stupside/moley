//go:build !windows

package system

import (
	"fmt"
	"os"
	"syscall"
)

// TerminateProcess sends SIGTERM to a process on Unix systems.
func TerminateProcess(p *os.Process) error {
	if err := p.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM to process %d: %w", p.Pid, err)
	}
	return nil
}

// GetProcessAttributes returns process attributes for Unix systems.
func GetProcessAttributes() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setsid: true,
	}
}
