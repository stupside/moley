package tunnel

import (
	"context"

	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/core/ports"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/shared"
)

type TunnelCreateHandler struct {
	tunnelService ports.TunnelService
}

type TunnelCreatePayload struct {
	Tunnel *domain.Tunnel `json:"tunnel"`
}

func NewTunnelCreateHandler(tunnelService ports.TunnelService) *TunnelCreateHandler {
	return &TunnelCreateHandler{
		tunnelService: tunnelService,
	}
}

func (h *TunnelCreateHandler) Name(ctx context.Context) string {
	return "tunnel-create"
}

func (h *TunnelCreateHandler) Up(ctx context.Context, payload any) error {
	tunnelPayload, err := castPayload[TunnelCreatePayload](payload)
	if err != nil {
		return err
	}

	logger.Debug("Creating tunnel")

	if _, err := h.tunnelService.CreateTunnel(ctx, tunnelPayload.Tunnel); err != nil {
		return shared.WrapError(err, "cloudflared tunnel create failed")
	}

	logger.Info("Tunnel created")
	return nil
}

func (h *TunnelCreateHandler) Down(ctx context.Context, payload any) error {
	tunnelPayload, err := castPayload[TunnelCreatePayload](payload)
	if err != nil {
		return err
	}

	logger.Debug("Deleting tunnel")

	if err := h.tunnelService.DeleteTunnel(ctx, tunnelPayload.Tunnel); err != nil {
		return shared.WrapError(err, "cloudflared tunnel delete failed")
	}

	logger.Info("Tunnel deleted")
	return nil
}

func (h *TunnelCreateHandler) Status(ctx context.Context, payload any) (domain.State, error) {
	tunnelPayload, err := castPayload[TunnelCreatePayload](payload)
	if err != nil {
		return "", err
	}
	exists, err := h.tunnelService.TunnelExists(ctx, tunnelPayload.Tunnel)
	if err != nil {
		return domain.StateDown, nil
	}
	if !exists {
		return domain.StateDown, nil
	}
	return domain.StateUp, nil
}
