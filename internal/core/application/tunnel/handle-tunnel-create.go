package tunnel

import (
	"context"

	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/core/ports"
	"github.com/stupside/moley/v2/internal/platform/framework"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/shared"
)

// TunnelCreateConfig represents the desired tunnel configuration
type TunnelCreateConfig struct {
	Tunnel *domain.Tunnel `json:"tunnel"`
}

// TunnelCreateState represents the runtime state of a created tunnel
type TunnelCreateState struct {
	Tunnel *domain.Tunnel `json:"tunnel"`
}

// TunnelCreateHandler manages tunnel creation lifecycle with type safety
type TunnelCreateHandler struct {
	tunnelService ports.TunnelService
}

// Ensure TunnelCreateHandler implements the required interfaces
var _ framework.ResourceHandler[TunnelCreateConfig, TunnelCreateState] = (*TunnelCreateHandler)(nil)

func newTunnelCreateHandler(tunnelService ports.TunnelService) *TunnelCreateHandler {
	return &TunnelCreateHandler{
		tunnelService: tunnelService,
	}
}

func (h *TunnelCreateHandler) Name() string {
	return "tunnel-create"
}

func (h *TunnelCreateHandler) Create(ctx context.Context, config TunnelCreateConfig) (TunnelCreateState, error) {
	logger.Debug("Creating tunnel")

	_, err := h.tunnelService.CreateTunnel(ctx, config.Tunnel)
	if err != nil {
		return TunnelCreateState{}, shared.WrapError(err, "cloudflared tunnel create failed")
	}

	state := TunnelCreateState{
		Tunnel: config.Tunnel,
	}

	logger.Info("Tunnel created")
	return state, nil
}

func (h *TunnelCreateHandler) Destroy(ctx context.Context, state TunnelCreateState) error {
	logger.Debug("Deleting tunnel")

	if err := h.tunnelService.DeleteTunnel(ctx, state.Tunnel); err != nil {
		return shared.WrapError(err, "cloudflared tunnel delete failed")
	}

	logger.Info("Tunnel deleted")
	return nil
}

func (h *TunnelCreateHandler) CheckFromState(ctx context.Context, state TunnelCreateState) (domain.State, error) {
	exists, err := h.tunnelService.TunnelExists(ctx, state.Tunnel)
	if err != nil {
		return domain.StateDown, shared.WrapError(err, "failed to check tunnel existence")
	}

	if exists {
		return domain.StateUp, nil
	}
	return domain.StateDown, nil
}

func (h *TunnelCreateHandler) Equals(a, b TunnelCreateConfig) bool {
	return a.Tunnel.ID == b.Tunnel.ID
}

// CheckFromConfig finds existing tunnel from config and returns state + status
func (h *TunnelCreateHandler) CheckFromConfig(ctx context.Context, config TunnelCreateConfig) (TunnelCreateState, domain.State, error) {
	exists, err := h.tunnelService.TunnelExists(ctx, config.Tunnel)
	if err != nil {
		return TunnelCreateState{}, domain.StateDown, shared.WrapError(err, "failed to check tunnel existence")
	}

	state := TunnelCreateState{
		Tunnel: config.Tunnel,
	}

	if exists {
		return state, domain.StateUp, nil
	}
	return state, domain.StateDown, nil
}
