//go:build !darwin

package system

import (
	"context"
	"runtime"
	"slices"

	"github.com/shirou/gopsutil/v4/process"
)

func processExited(ctx context.Context, p *process.Process) (bool, error) {
	// Windows has no Unix zombie processes and gopsutil does not implement Status.
	if runtime.GOOS == "windows" {
		return false, nil
	}
	statuses, err := p.StatusWithContext(ctx)
	return slices.Contains(statuses, process.Zombie), err
}
