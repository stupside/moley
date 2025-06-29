package tunnel

import (
	"context"
	"fmt"
	"moley/internal/logger"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

const (
	gracefulShutdownTimeout = 10 * time.Second
)

// Runner handles tunnel execution and lifecycle
type Runner struct {
	tunnelName string
	config     string
}

// NewRunner creates a new tunnel runner
func NewRunner(tunnelName, config string) *Runner {
	return &Runner{
		tunnelName: tunnelName,
		config:     config,
	}
}

// Run starts the tunnel and handles graceful shutdown
func (r *Runner) Run(ctx context.Context) error {
	logger.Infof("Starting tunnel", map[string]interface{}{
		"tunnel": r.tunnelName,
		"config": r.config,
	})

	// Create the cloudflared command using the config file
	cmd := exec.CommandContext(ctx, "cloudflared", "tunnel", "--config", r.config, "run", r.tunnelName)

	// Capture stdout and stderr for debugging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Start the tunnel
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start tunnel: %w", err)
	}

	logger.Infof("Tunnel started successfully", map[string]interface{}{
		"tunnel": r.tunnelName,
		"pid":    cmd.Process.Pid,
	})

	// Wait for either context cancellation or signal
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		logger.Info("Context cancelled, shutting down tunnel")
		return r.shutdownTunnel(cmd)
	case sig := <-sigChan:
		logger.Infof("Received signal, shutting down tunnel", map[string]interface{}{
			"signal": sig.String(),
		})
		return r.shutdownTunnel(cmd)
	case err := <-done:
		return r.handleProcessExit(err)
	}
}

// handleProcessExit handles the tunnel process exit
func (r *Runner) handleProcessExit(err error) error {
	if err == nil {
		logger.Info("Tunnel process completed successfully")
		return nil
	}

	// Check if it's a signal kill (graceful shutdown)
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() == -1 {
			logger.Info("Tunnel stopped by signal (graceful shutdown)")
			return nil
		}
	}

	return fmt.Errorf("tunnel process failed: %w", err)
}

// shutdownTunnel gracefully shuts down the tunnel
func (r *Runner) shutdownTunnel(cmd *exec.Cmd) error {
	logger.Infof("Shutting down tunnel", map[string]interface{}{
		"tunnel": r.tunnelName,
	})

	if cmd.Process == nil {
		logger.Warn("Tunnel process is nil, nothing to shutdown")
		return nil
	}

	// Send SIGTERM first for graceful shutdown
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		logger.Warnf("Failed to send SIGTERM to tunnel process", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Wait for graceful shutdown with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// Wait up to gracefulShutdownTimeout for graceful shutdown
	select {
	case err := <-done:
		return r.handleProcessExit(err)
	case <-time.After(gracefulShutdownTimeout):
		// Force kill if graceful shutdown times out
		logger.Warn("Graceful shutdown timed out, force killing tunnel process")
		if err := cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill tunnel process: %w", err)
		}
		return cmd.Wait()
	}
}
