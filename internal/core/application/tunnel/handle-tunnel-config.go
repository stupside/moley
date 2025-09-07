package tunnel

import (
	"context"
	"os"

	"github.com/stupside/moley/internal/core/domain"
	"github.com/stupside/moley/internal/core/ports"
	"github.com/stupside/moley/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/internal/shared"
)

type TunnelConfigHandler struct {
	tunnelService ports.TunnelService
}

type TunnelConfigPayload struct {
	Tunnel  *domain.Tunnel  `json:"tunnel"`
	Ingress *domain.Ingress `json:"ingress"`
}

func NewTunnelConfigHandler(tunnelService ports.TunnelService) *TunnelConfigHandler {
	return &TunnelConfigHandler{
		tunnelService: tunnelService,
	}
}

func (h *TunnelConfigHandler) Name(ctx context.Context) string {
	return "tunnel-config"
}

func (h *TunnelConfigHandler) Up(ctx context.Context, payload any) error {
	tunnelPayload, err := castPayload[TunnelConfigPayload](payload)
	if err != nil {
		return err
	}

	logger.Debug("Configuring tunnel")

	if err := h.tunnelService.SaveConfiguration(ctx, tunnelPayload.Tunnel, tunnelPayload.Ingress); err != nil {
		return shared.WrapError(err, "failed to save tunnel configuration")
	}

	logger.Infof("Tunnel configured", map[string]any{
		"zone": tunnelPayload.Ingress.Zone,
		"apps": len(tunnelPayload.Ingress.Apps),
	})
	return nil
}

func (h *TunnelConfigHandler) Down(ctx context.Context, payload any) error {
	tunnelPayload, err := castPayload[TunnelConfigPayload](payload)
	if err != nil {
		return err
	}

	logger.Debug("Removing tunnel configuration")

	if err := h.tunnelService.DeleteConfiguration(ctx, tunnelPayload.Tunnel); err != nil {
		return shared.WrapError(err, "failed to delete tunnel configuration")
	}

	logger.Info("Tunnel configuration removed")
	return nil
}

func (h *TunnelConfigHandler) Status(ctx context.Context, payload any) (domain.State, error) {
	tunnelPayload, err := castPayload[TunnelConfigPayload](payload)
	if err != nil {
		return "", err
	}

	path, err := h.tunnelService.GetConfigurationPath(ctx, tunnelPayload.Tunnel)
	if err != nil {
		return domain.StateDown, nil
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return domain.StateDown, nil
		}
		return domain.StateDown, nil
	}

	return domain.StateUp, nil
}
