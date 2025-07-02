package steps

import (
	"context"

	"github.com/stupside/moley/internal/domain"
	"github.com/stupside/moley/internal/errors"
	"github.com/stupside/moley/internal/logger"
	"github.com/stupside/moley/internal/services"
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
	logger.Debugf("Creating tunnel", map[string]interface{}{
		"tunnel": t.tunnel.GetName(),
	})
	tunnelId, err := t.tunnelService.CreateTunnel(ctx, t.tunnel)
	if err != nil {
		logger.Warnf("Tunnel creation failed", map[string]interface{}{
			"tunnel": t.tunnel.GetName(),
			"error":  err.Error(),
		})
		return errors.NewExecutionError(errors.ErrCodeCommandFailed, "cloudflared tunnel create failed", err)
	}
	logger.Debugf("Tunnel created successfully", map[string]interface{}{
		"id":     tunnelId,
		"tunnel": t.tunnel.GetName(),
	})
	return nil
}

// Down deletes the tunnel
func (t *TunnelStep) Down(ctx context.Context) error {
	logger.Debugf("Deleting tunnel", map[string]interface{}{
		"tunnel": t.tunnel.GetName(),
	})
	if err := t.tunnelService.DeleteTunnel(ctx, t.tunnel); err != nil {
		logger.Warnf("Tunnel deletion failed", map[string]interface{}{
			"tunnel": t.tunnel.GetName(),
			"error":  err.Error(),
		})
		return errors.NewExecutionError(errors.ErrCodeCommandFailed, "cloudflared tunnel delete failed", err)
	}
	logger.Infof("Tunnel deleted successfully", map[string]interface{}{
		"tunnel": t.tunnel.GetName(),
	})
	return nil
}
