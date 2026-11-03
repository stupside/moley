package tunnel

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/domain"
	logger "github.com/stupside/moley/v2/internal/platform/logging"
	framework "github.com/stupside/moley/v2/internal/platform/orchestration"
)

// TunnelRunner is the process lifecycle required by reconciliation.
// Implementations keep live handles outside the persisted ProcessIdentity.
type TunnelRunner interface {
	Run(ctx context.Context, tunnel *domain.Tunnel) (ProcessIdentity, error)
	// Check reports false with no error only when the process is confirmed absent.
	Check(ctx context.Context, identity ProcessIdentity) (bool, error)
	// Stop waits for termination; observation or termination failures remain errors.
	Stop(ctx context.Context, identity ProcessIdentity) error
}

const RunHandlerName = "tunnel-run"

type RunInput struct {
	TunnelName  string `json:"tunnel_name"`
	ConfigPath  string `json:"config_path"`  // included for hash-based change detection
	ContentHash string `json:"content_hash"` // included for hash-based change detection
}

type RunOutput struct {
	TunnelName string          `json:"tunnel_name"`
	Process    ProcessIdentity `json:"process"`
}

// ProcessIdentity is the serializable identity used to observe and stop a process.
type ProcessIdentity struct {
	PID     int    `json:"pid"`
	Command string `json:"command"`
}

type runHandler struct {
	runner TunnelRunner
}

var _ framework.Lifecycle[RunInput, RunOutput] = (*runHandler)(nil)

func NewRunHandler(runner TunnelRunner) *runHandler {
	return &runHandler{
		runner: runner,
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

	identity, err := h.runner.Run(ctx, &domain.Tunnel{Name: input.TunnelName})
	if err != nil {
		return RunOutput{}, fmt.Errorf("failed to start tunnel process: %w", err)
	}

	output := RunOutput{
		TunnelName: input.TunnelName,
		Process:    identity,
	}

	logger.Infof("Tunnel process started", map[string]any{"pid": identity.PID})
	return output, nil
}

func (h *runHandler) Destroy(ctx context.Context, output RunOutput) error {
	return h.runner.Stop(ctx, output.Process)
}

func (h *runHandler) Check(ctx context.Context, output RunOutput) (framework.Status, error) {
	running, err := h.runner.Check(ctx, output.Process)
	if err != nil {
		return framework.StatusUnknown, err
	}
	if !running {
		return framework.StatusDown, nil
	}
	return framework.StatusUp, nil
}

func (h *runHandler) Recover(ctx context.Context, input RunInput) (RunOutput, framework.Status, error) {
	return RunOutput{TunnelName: input.TunnelName}, framework.StatusDown, nil
}
