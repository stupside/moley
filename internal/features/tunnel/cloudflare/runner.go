package cloudflare

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/stupside/moley/v2/internal/domain"
	tunnel "github.com/stupside/moley/v2/internal/features/tunnel/usecase"
	logger "github.com/stupside/moley/v2/internal/platform/logging"
	sys "github.com/stupside/moley/v2/internal/platform/system"
)

// tunnelConfiguration supplies only the metadata needed to launch cloudflared.
type tunnelConfiguration interface {
	GetConfigurationPath(context.Context, *domain.Tunnel) (string, error)
	GetID(context.Context, *domain.Tunnel) (string, error)
}

// Runner adapts cloudflared's command and process lifecycle to the tunnel use case.
// Cloudflare provisioning and configuration persistence belong to TunnelService.
type Runner struct {
	configuration tunnelConfiguration
	processes     sys.Processes
	dryRun        bool
}

var _ tunnel.TunnelRunner = (*Runner)(nil)

func NewRunner(configuration tunnelConfiguration, dryRun bool) *Runner {
	return &Runner{configuration: configuration, dryRun: dryRun}
}

func (r *Runner) Run(ctx context.Context, t *domain.Tunnel) (tunnel.ProcessIdentity, error) {
	if err := ctx.Err(); err != nil {
		return tunnel.ProcessIdentity{}, err
	}
	if r.dryRun {
		return tunnel.ProcessIdentity{}, nil
	}
	path, err := r.configuration.GetConfigurationPath(ctx, t)
	if err != nil {
		return tunnel.ProcessIdentity{}, fmt.Errorf("get tunnel configuration path: %w", err)
	}
	uuid, err := r.configuration.GetID(ctx, t)
	if err != nil {
		return tunnel.ProcessIdentity{}, fmt.Errorf("resolve tunnel UUID: %w", err)
	}
	// Passing the UUID avoids cloudflared doing a name lookup that needs cert.pem.
	args := []string{"tunnel", "--config", path, "run", uuid}
	// The session owns termination. Canceling startup must not kill a detached child.
	cmd := exec.Command("cloudflared", args...)
	cmd.SysProcAttr = sys.GetProcessAttributes()
	logger.Debugf("Starting cloudflared", map[string]any{"args": strings.Join(args, " ")})
	pid, err := r.processes.Start(ctx, cmd)
	if err != nil {
		return tunnel.ProcessIdentity{}, fmt.Errorf("start cloudflared: %w", err)
	}
	return tunnel.ProcessIdentity{PID: pid, Command: sys.GetProcessCommand(pid)}, nil
}

func (r *Runner) Check(ctx context.Context, identity tunnel.ProcessIdentity) (bool, error) {
	if r.dryRun {
		return true, nil
	}
	// A snapshot left by a dry run is not a live process in a real session.
	if identity.PID == 0 {
		return false, nil
	}
	return sys.CheckProcessIdentity(ctx, identity.PID, identity.Command)
}

func (r *Runner) Stop(ctx context.Context, identity tunnel.ProcessIdentity) error {
	if r.dryRun || identity.PID == 0 {
		return nil
	}
	return r.processes.Stop(ctx, identity.PID, identity.Command)
}

func (r *Runner) Wait(ctx context.Context, identity tunnel.ProcessIdentity) error {
	if r.dryRun {
		<-ctx.Done()
		return ctx.Err()
	}
	return r.processes.Wait(ctx, identity.PID, identity.Command)
}

func (r *Runner) StopOwned(ctx context.Context) error {
	return r.processes.StopOwned(ctx)
}
