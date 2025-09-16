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

// Ensure TunnelRunHandler implements the required interfaces
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

	process, err := os.FindProcess(state.PID)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", state.PID, err)
	}

	// Try graceful termination first
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM to process %d: %w", state.PID, err)
	}

	logger.Infof("Tunnel process stopped", map[string]any{
		"pid": state.PID,
	})

	return nil
}

func (h *TunnelRunHandler) CheckFromState(ctx context.Context, state TunnelRunState) (domain.State, error) {
	process, err := os.FindProcess(state.PID)
	if err != nil {
		return domain.StateDown, nil
	}

	// Send signal 0 to check if process exists and is accessible
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return domain.StateDown, nil
	}

	return domain.StateUp, nil
}

// CheckFromConfig finds existing tunnel process from config and returns state + status
func (h *TunnelRunHandler) CheckFromConfig(ctx context.Context, config TunnelRunConfig) (TunnelRunState, domain.State, error) {
	pid := 0 // TODO: Implement a way to retrieve the PID of the running tunnel process if needed

	state := TunnelRunState{
		PID:    pid,
		Tunnel: config.Tunnel,
	}

	return state, domain.StateUp, nil
}

func (h *TunnelRunHandler) Equals(a, b TunnelRunConfig) bool {
	return a.Tunnel.ID == b.Tunnel.ID
}
