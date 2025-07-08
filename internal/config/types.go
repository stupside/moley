package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stupside/moley/internal/shared"
)

// GlobalConfig represents the main configuration structure
type GlobalConfig struct {
	Cloudflare CloudflareConfig `mapstructure:"cloudflare" yaml:"cloudflare"`
}

// CloudflareConfig holds Cloudflare-specific configuration
type CloudflareConfig struct {
	Token string `mapstructure:"token" yaml:"token"`
}

// GetDefaultConfig returns the default global configuration
func GetDefaultConfig() *GlobalConfig {
	return &GlobalConfig{
		Cloudflare: CloudflareConfig{},
	}
}

// NewGlobalConfigManager creates a new global configuration manager using BaseConfigManager
func NewGlobalConfigManager(cmd *cobra.Command) (*shared.BaseConfigManager[GlobalConfig], error) {
	configPath, err := GetGlobalConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	options := []shared.WithOption{
		shared.WithBindFlags(cmd),
		// shared.WithBindEnv("MOLEY"),
	}

	return shared.NewConfigManager(configPath, GetDefaultConfig(), options...), nil
}
