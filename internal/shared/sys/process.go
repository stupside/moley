package sys

import (
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

// CheckProcessIdentity verifies that a process with the given PID is still alive
// and optionally matches the expected command name (to detect PID reuse).
func CheckProcessIdentity(pid int, expectedCommand string) bool {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return false
	}

	running, err := p.IsRunning()
	if err != nil || !running {
		return false
	}

	if expectedCommand != "" {
		name, err := p.Name()
		if err == nil && name != "" && name != expectedCommand {
			return false
		}
	}

	return true
}
