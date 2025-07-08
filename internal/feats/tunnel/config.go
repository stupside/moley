package tunnel

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/stupside/moley/internal/domain"
	"github.com/stupside/moley/internal/logger"
	"github.com/stupside/moley/internal/validation"
)

const (
	TunneConfigFile = "moley.yml"
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

func LoadConfigFromFile(filePath string) (*TunnelConfig, error) {
	v := viper.New()

	v.SetConfigName("moley")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Errorf("Configuration file not found", map[string]interface{}{"config_file": TunneConfigFile})
			return nil, fmt.Errorf("configuration file not found")
		}
		logger.Errorf("Failed to read configuration file", map[string]interface{}{"config_file": TunneConfigFile, "error": err.Error()})
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	var tunnelConfig TunnelConfig
	if err := v.Unmarshal(&tunnelConfig); err != nil {
		logger.Errorf("Failed to unmarshal tunnel configuration", map[string]interface{}{"error": err.Error()})
		return nil, fmt.Errorf("failed to unmarshal tunnel configuration: %w", err)
	}

	if err := validation.ValidateStruct(&tunnelConfig); err != nil {
		logger.Errorf("Tunnel configuration validation failed", map[string]interface{}{"error": err.Error()})
		return nil, fmt.Errorf("tunnel configuration error: %w", err)
	}

	return &tunnelConfig, nil
}
