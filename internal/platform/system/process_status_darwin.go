package system

import (
	"context"

	"github.com/shirou/gopsutil/v4/process"
	"golang.org/x/sys/unix"
)

func processExited(ctx context.Context, p *process.Process) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	// gopsutil's Darwin Status launches ps. Query the kernel directly so polling
	// does not spawn a subprocess on every tick. An empty result means it exited.
	entries, err := unix.SysctlKinfoProcSlice("kern.proc.pid", int(p.Pid))
	if err != nil {
		return false, err
	}
	const zombie = 5 // SZOMB in Darwin's sys/proc.h: awaiting collection by parent.
	return len(entries) == 0 || entries[0].Proc.P_stat == zombie, nil
}
