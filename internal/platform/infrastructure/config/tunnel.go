// Package config provides configuration management for Moley.
// It handles both global configuration (API tokens, etc.) and tunnel-specific configuration
// (tunnel settings, ingress rules, DNS configuration) with proper validation and defaults.
package config

import (
	"github.com/google/uuid"
	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/shared"
)

// TunnelConfigManager manages tunnel configuration
type TunnelConfigManager struct {
	*ConfigManager[TunnelConfig]
}

// NewTunnelConfigManager creates a new tunnel configuration manager
func NewTunnelConfigManager(configPath string) (*TunnelConfigManager, error) {
	// Get default tunnel config
	defaultTunnelConfig, err := getDefaultTunnelConfig()
	if err != nil {
		return nil, shared.WrapError(err, "failed to get default tunnel configuration")
	}

	// Create config manager
	configManager := NewConfigManager(configPath, defaultTunnelConfig)

	// Load config
	if err := configManager.Load(); err != nil {
		return nil, shared.WrapError(err, "failed to load tunnel configuration")
	}

	return &TunnelConfigManager{
		ConfigManager: configManager,
	}, nil
}

// GetTunnelConfig returns the tunnel configuration
func (tcm *TunnelConfigManager) GetTunnelConfig() *TunnelConfig {
	return tcm.GetConfig()
}

// UpdateTunnelConfig updates the tunnel configuration and saves it
func (tcm *TunnelConfigManager) UpdateTunnelConfig(updater func(*TunnelConfig)) error {
	return tcm.UpdateConfig(updater)
}

// getDefaultTunnelConfig returns the default tunnel configuration
func getDefaultTunnelConfig() (*TunnelConfig, error) {
	// Generate a random ID for the tunnel
	id := uuid.New().String()

	// Create tunnel with validation
	tunnel, err := domain.NewTunnel(id)
	if err != nil {
		return nil, shared.WrapError(err, "failed to create default tunnel configuration")
	}

	return &TunnelConfig{
		Tunnel:  tunnel,
		Ingress: domain.NewDefaultIngress(),
	}, nil
}
