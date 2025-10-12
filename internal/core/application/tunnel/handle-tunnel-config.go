package tunnel

import (
	"context"
	"os"
	"reflect"

	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/core/ports"
	"github.com/stupside/moley/v2/internal/platform/framework"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/shared"
)

// TunnelConfigConfig represents the desired tunnel configuration
type TunnelConfigConfig struct {
	Tunnel  *domain.Tunnel  `json:"tunnel"`
	Ingress *domain.Ingress `json:"ingress"`
}

// TunnelConfigState represents the runtime state of a configured tunnel
type TunnelConfigState struct {
	Tunnel     *domain.Tunnel  `json:"tunnel"`
	Ingress    *domain.Ingress `json:"ingress"`
	ConfigPath string          `json:"config_path"`
}

// TunnelConfigHandler manages tunnel configuration lifecycle with type safety
type TunnelConfigHandler struct {
	tunnelService ports.TunnelService
}

// Ensure TunnelConfigHandler implements the required interfaces
var _ framework.ResourceHandler[TunnelConfigConfig, TunnelConfigState] = (*TunnelConfigHandler)(nil)

func newTunnelConfigHandler(tunnelService ports.TunnelService) *TunnelConfigHandler {
	return &TunnelConfigHandler{
		tunnelService: tunnelService,
	}
}

func (h *TunnelConfigHandler) Name() string {
	return "tunnel-config"
}

func (h *TunnelConfigHandler) Create(ctx context.Context, config TunnelConfigConfig) (TunnelConfigState, error) {
	logger.Debug("Configuring tunnel")

	if err := h.tunnelService.SaveConfiguration(ctx, config.Tunnel, config.Ingress); err != nil {
		return TunnelConfigState{}, shared.WrapError(err, "failed to save tunnel configuration")
	}

	configPath, err := h.tunnelService.GetConfigurationPath(ctx, config.Tunnel)
	if err != nil {
		return TunnelConfigState{}, shared.WrapError(err, "failed to get configuration path")
	}

	state := TunnelConfigState{
		Tunnel:     config.Tunnel,
		Ingress:    config.Ingress,
		ConfigPath: configPath,
	}

	logger.Infof("Tunnel configured", map[string]any{
		"zone": config.Ingress.Zone,
		"apps": len(config.Ingress.Apps),
	})

	return state, nil
}

func (h *TunnelConfigHandler) Destroy(ctx context.Context, state TunnelConfigState) error {
	logger.Debug("Removing tunnel configuration")

	if err := h.tunnelService.DeleteConfiguration(ctx, state.Tunnel); err != nil {
		return shared.WrapError(err, "failed to delete tunnel configuration")
	}

	logger.Info("Tunnel configuration removed")
	return nil
}

func (h *TunnelConfigHandler) CheckFromState(ctx context.Context, state TunnelConfigState) (domain.State, error) {
	if _, err := os.Stat(state.ConfigPath); err != nil {
		if os.IsNotExist(err) {
			return domain.StateDown, nil
		}
		return domain.StateUnknown, shared.WrapError(err, "failed to check configuration file")
	}

	return domain.StateUp, nil
}

func (h *TunnelConfigHandler) Equals(a, b TunnelConfigConfig) bool {
	return a.Tunnel.ID == b.Tunnel.ID &&
		reflect.DeepEqual(a.Ingress, b.Ingress)
}

// CheckFromConfig finds existing tunnel configuration from config and returns state + status
func (h *TunnelConfigHandler) CheckFromConfig(ctx context.Context, config TunnelConfigConfig) (TunnelConfigState, domain.State, error) {
	configPath, err := h.tunnelService.GetConfigurationPath(ctx, config.Tunnel)
	if err != nil {
		return TunnelConfigState{}, domain.StateUnknown, shared.WrapError(err, "failed to get configuration path")
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return TunnelConfigState{}, domain.StateDown, nil
		}
		return TunnelConfigState{}, domain.StateUnknown, shared.WrapError(err, "failed to check configuration file")
	}

	// Convert config to state
	state := TunnelConfigState{
		Tunnel:     config.Tunnel,
		Ingress:    config.Ingress,
		ConfigPath: configPath,
	}

	return state, domain.StateUp, nil
}
