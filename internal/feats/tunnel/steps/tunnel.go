package steps

import (
	"context"
	"fmt"

	"github.com/stupside/moley/internal/domain"
	"github.com/stupside/moley/internal/logger"
	"github.com/stupside/moley/internal/services"
	"github.com/stupside/moley/internal/shared"
)

// TunnelStep handles tunnel operations
type TunnelStep struct {
	// tunnel holds the tunnel information for which the step is executed
	tunnel *domain.Tunnel
	// tunnelService is the service responsible for tunnel operations
	tunnelService services.TunnelService
}

// NewTunnelStep creates a new tunnel step
func NewTunnelStep(tunnelService services.TunnelService, tunnel *domain.Tunnel) *TunnelStep {
	return &TunnelStep{
		tunnel:        tunnel,
		tunnelService: tunnelService,
	}
}

// Name returns the step name
func (t *TunnelStep) Name() string {
	return "Tunnel"
}

// Up creates the tunnel and stores the tunnel ID
func (t *TunnelStep) Up(ctx context.Context) error {
	logger.Info(fmt.Sprintf("Creating tunnel: %s", t.tunnel.GetName()))

	tunnelId, err := t.tunnelService.CreateTunnel(ctx, t.tunnel)
	if err != nil {
		return shared.WrapError(err, "cloudflared tunnel create failed")
	}

	logger.Info(fmt.Sprintf("Tunnel created successfully with ID: %s", tunnelId))
	return nil
}

// Down deletes the tunnel
func (t *TunnelStep) Down(ctx context.Context) error {
	logger.Debug(fmt.Sprintf("Deleting tunnel: %s", t.tunnel.GetName()))

	if err := t.tunnelService.DeleteTunnel(ctx, t.tunnel); err != nil {
		return shared.WrapError(err, "cloudflared tunnel delete failed")
	}

	logger.Debug("Tunnel deleted successfully")
	return nil
}
