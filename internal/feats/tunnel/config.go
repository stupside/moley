package tunnel

import (
	"github.com/stupside/moley/internal/domain"
	"github.com/stupside/moley/internal/shared"
)

const (
	TunnelConfigFile = "moley.yml"
)

// TunnelConfig holds the configuration for the tunnel feature
type TunnelConfig struct {
	// Zone is the DNS zone to use for the tunnel
	Zone string `mapstructure:"zone" yaml:"zone" json:"zone" validate:"required"`
	// Apps is a list of applications to expose via the tunnel
	Apps []domain.AppConfig `mapstructure:"apps" yaml:"apps" json:"apps" validate:"required,min=1"`
}

// GetDefaultConfig returns the default tunnel configuration
func GetDefaultConfig() *TunnelConfig {
	// Create a default DNS configuration with a single app
	dns := domain.NewDefaultDNS()

	return &TunnelConfig{
		Zone: dns.GetZone(),
		Apps: dns.GetApps(),
	}
}

func NewTunnelConfigManager() *shared.BaseConfigManager[TunnelConfig] {
	return shared.NewConfigManager(TunnelConfigFile, GetDefaultConfig())
}
