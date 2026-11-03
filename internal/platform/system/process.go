package system

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"syscall"

	"github.com/shirou/gopsutil/v4/process"
)

// GetProcessCommand returns the command name for a given PID.
// Returns empty string if the process cannot be found.
func GetProcessCommand(pid int) string {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return ""
	}
	name, err := p.Name()
	if err != nil {
		return ""
	}
	return name
}

// CheckProcessIdentity distinguishes a confirmed exit from an observation error.
// An unknown process must not be treated as successfully stopped.
func CheckProcessIdentity(ctx context.Context, pid int, expectedCommand string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if pid <= 0 || int64(pid) > math.MaxInt32 {
		return false, fmt.Errorf("invalid process ID %d", pid)
	}
	p, err := process.NewProcessWithContext(ctx, int32(pid))
	if err != nil {
		return processObservation(ctx, pid, err)
	}
	running, err := p.IsRunningWithContext(ctx)
	if err != nil {
		return processObservation(ctx, pid, err)
	}
	if !running {
		return false, nil
	}
	exited, err := processExited(ctx, p)
	if err != nil {
		return processObservation(ctx, pid, err)
	}
	if exited {
		return false, nil
	}

	if expectedCommand == "" {
		return false, fmt.Errorf("cannot verify identity of process %d without its command name", pid)
	}
	name, err := p.NameWithContext(ctx)
	if err != nil {
		return processObservation(ctx, pid, err)
	}
	if name == "" {
		return false, fmt.Errorf("cannot read command name of process %d", pid)
	}
	return name == expectedCommand, nil
}

func processObservation(ctx context.Context, pid int, err error) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if errors.Is(err, process.ErrorProcessNotRunning) || errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ESRCH) {
		return false, nil
	}
	// A process can disappear between existence, status, and name queries.
	// Confirm absence instead of translating arbitrary observation failures to exit.
	if exists, probeErr := process.PidExistsWithContext(ctx, int32(pid)); probeErr == nil && !exists {
		return false, nil
	}
	return false, err
}
