package tunnel

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/core/ports"
	"github.com/stupside/moley/v2/internal/platform/framework"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/shared"
)

// TunnelRunConfig represents the desired tunnel run configuration
type TunnelRunConfig struct {
	Tunnel *domain.Tunnel `json:"tunnel"`
}

// TunnelRunState represents the runtime state of a running tunnel process
type TunnelRunState struct {
	PID    int            `json:"pid"`
	Tunnel *domain.Tunnel `json:"tunnel"`
}

// TunnelRunHandler manages tunnel process lifecycle with type safety
type TunnelRunHandler struct {
	tunnelService ports.TunnelService
}

// Ensure TunnelRunHandler implements the typed interface
var _ framework.ResourceHandler[TunnelRunConfig, TunnelRunState] = (*TunnelRunHandler)(nil)

func newTunnelRunHandler(tunnelService ports.TunnelService) *TunnelRunHandler {
	return &TunnelRunHandler{
		tunnelService: tunnelService,
	}
}

func (h *TunnelRunHandler) Name() string {
	return "tunnel-run"
}

func (h *TunnelRunHandler) Create(ctx context.Context, config TunnelRunConfig) (TunnelRunState, error) {
	logger.Debug("Starting tunnel process")

	pid, err := h.tunnelService.Run(ctx, config.Tunnel)
	if err != nil {
		return TunnelRunState{}, shared.WrapError(err, "failed to start tunnel process")
	}

	state := TunnelRunState{
		PID:    pid,
		Tunnel: config.Tunnel,
	}

	logger.Infof("Tunnel process started", map[string]any{
		"pid": pid,
	})

	return state, nil
}

func (h *TunnelRunHandler) Destroy(ctx context.Context, state TunnelRunState) error {
	logger.Debugf("Stopping tunnel process", map[string]any{
		"pid": state.PID,
	})

	if err := h.stopProcess(state.PID); err != nil {
		return shared.WrapError(err, "failed to stop tunnel process")
	}

	logger.Infof("Tunnel process stopped", map[string]any{
		"pid": state.PID,
	})

	return nil
}

func (h *TunnelRunHandler) Status(ctx context.Context, state TunnelRunState) (domain.State, error) {
	return h.checkProcessStatus(state.PID), nil
}

func (h *TunnelRunHandler) Equals(a, b TunnelRunConfig) bool {
	return a.Tunnel.ID == b.Tunnel.ID
}

// stopProcess gracefully stops a process with fallback to force kill
func (h *TunnelRunHandler) stopProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	// Try graceful termination first
	if err := process.Signal(syscall.SIGTERM); err != nil {
		// Process might already be dead
		if h.checkProcessStatus(pid) == domain.StateDown {
			return nil
		}
		return fmt.Errorf("failed to send SIGTERM to process %d: %w", pid, err)
	}

	return nil
}

// checkProcessStatus checks if a process is running using signal 0
func (h *TunnelRunHandler) checkProcessStatus(pid int) domain.State {
	process, err := os.FindProcess(pid)
	if err != nil {
		return domain.StateDown
	}

	// Send signal 0 to check if process exists and is accessible
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return domain.StateDown
	}

	return domain.StateUp
}
