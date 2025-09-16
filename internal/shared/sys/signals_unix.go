//go:build !windows

package sys

import (
	"os"
	"syscall"
)

// GetShutdownSignals returns the signals to listen for graceful shutdown on Unix systems
func GetShutdownSignals() []os.Signal {
	return []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP}
}
