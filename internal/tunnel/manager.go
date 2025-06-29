package tunnel

import (
	"moley/internal/cloudflare"
	"moley/internal/config"
)

// Manager handles tunnel deployment and cleanup operations
type Manager struct {
	config         *config.MoleyConfig
	tunnelName     string
	tunnelID       string
	zoneID         string
	cf             *cloudflare.CloudflareClient
	configFilename string
}

// NewManager creates a new tunnel manager
func NewManager(config *config.MoleyConfig, tunnelName string) *Manager {
	if config == nil {
		panic("config cannot be nil")
	}
	if tunnelName == "" {
		panic("tunnelName cannot be empty")
	}

	return &Manager{
		config:     config,
		tunnelName: tunnelName,
	}
}

// GetConfigFilename returns the generated Cloudflare tunnel configuration filename
func (m *Manager) GetConfigFilename() string {
	return m.configFilename
}
