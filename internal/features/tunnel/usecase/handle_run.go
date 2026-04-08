package tunnel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/stupside/moley/v2/internal/domain"
	logger "github.com/stupside/moley/v2/internal/platform/logging"
	framework "github.com/stupside/moley/v2/internal/platform/orchestration"
	sys "github.com/stupside/moley/v2/internal/platform/system"
)

type TunnelRunner interface {
	Run(ctx context.Context, tunnel *domain.Tunnel) (int, error)
}

const RunHandlerName = "tunnel-run"

type RunInput struct {
	TunnelName  string `json:"tunnel_name"`
	ConfigPath  string `json:"config_path"`  // included for hash-based change detection
	ContentHash string `json:"content_hash"` // included for hash-based change detection
}

type RunOutput struct {
	TunnelName string          `json:"tunnel_name"`
	Process    processIdentity `json:"process"`
}

// processIdentity tracks a process by both PID and command name to detect PID reuse.
type processIdentity struct {
	PID     int    `json:"pid"`
	Command string `json:"command"`
}

type runHandler struct {
	tunnelService TunnelRunner
}

var _ framework.Lifecycle[RunInput, RunOutput] = (*runHandler)(nil)

func NewRunHandler(tunnelService TunnelRunner) *runHandler {
	return &runHandler{
		tunnelService: tunnelService,
	}
}

func (h *runHandler) Name() string {
	return RunHandlerName
}

func (h *runHandler) Key(input RunInput) string {
	return input.TunnelName
}

func (h *runHandler) Create(ctx context.Context, input RunInput) (RunOutput, error) {
	logger.Debug("Starting tunnel process")

	pid, err := h.tunnelService.Run(ctx, &domain.Tunnel{Name: input.TunnelName})
	if err != nil {
		return RunOutput{}, fmt.Errorf("failed to start tunnel process: %w", err)
	}

	output := RunOutput{
		TunnelName: input.TunnelName,
		Process: processIdentity{
			PID:     pid,
			Command: sys.GetProcessCommand(pid),
		},
	}

	logger.Infof("Tunnel process started", map[string]any{"pid": pid})
	return output, nil
}

func (h *runHandler) Destroy(ctx context.Context, output RunOutput) error {
	if output.Process.PID == 0 {
		return nil // dry-run: no real process
	}

	logger.Debugf("Stopping tunnel process", map[string]any{"pid": output.Process.PID})

	process, err := os.FindProcess(output.Process.PID)
	if err != nil {
		logger.Warnf("Failed to find process, may have already exited", map[string]any{
			"pid":   output.Process.PID,
			"error": err.Error(),
		})
		return nil
	}

	if err := sys.TerminateProcess(process); err != nil {
		if errors.Is(err, os.ErrProcessDone) || isProcessNotFoundError(err) {
			logger.Info("Process has already exited, skipping termination")
			return nil
		}
		return err
	}

	logger.Infof("Tunnel process stopped", map[string]any{"pid": output.Process.PID})
	return nil
}

func isProcessNotFoundError(err error) bool {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return errno == syscall.ESRCH
	}
	return false
}

func (h *runHandler) Check(ctx context.Context, output RunOutput) (framework.Status, error) {
	if output.Process.PID == 0 {
		return framework.StatusUp, nil // dry-run: no real process
	}
	if !sys.CheckProcessIdentity(output.Process.PID, output.Process.Command) {
		return framework.StatusDown, nil
	}
	return framework.StatusUp, nil
}

func (h *runHandler) Recover(ctx context.Context, input RunInput) (RunOutput, framework.Status, error) {
	return RunOutput{TunnelName: input.TunnelName}, framework.StatusDown, nil
}
