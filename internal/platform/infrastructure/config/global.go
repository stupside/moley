package config

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/stupside/moley/internal/platform/infrastructure/paths"
	"github.com/stupside/moley/internal/shared"
)

// GlobalConfigManager manages global configuration
type GlobalConfigManager struct {
	*ConfigManager[GlobalConfig]
}

// NewGlobalConfigManager creates a new global configuration manager
func NewGlobalConfigManager(cmd *cobra.Command) (*GlobalConfigManager, error) {
	// Get global config path
	globalPath, err := getGlobalConfigPath()
	if err != nil {
		return nil, shared.WrapError(err, "failed to get global config path")
	}

	// Create config manager
	configManager := NewConfigManager(globalPath, getDefaultGlobalConfig())

	// Bind environment variables and CLI flags
	configManager.BindEnv("MOLEY")
	configManager.BindFlags(cmd)

	// Load config
	if err := configManager.Load(); err != nil {
		return nil, shared.WrapError(err, "failed to load global config")
	}

	return &GlobalConfigManager{
		ConfigManager: configManager,
	}, nil
}

// GetGlobalConfig returns the global configuration
func (gcm *GlobalConfigManager) GetGlobalConfig() *GlobalConfig {
	return gcm.GetConfig()
}

// UpdateGlobalConfig updates the global configuration and saves it
func (gcm *GlobalConfigManager) UpdateGlobalConfig(updater func(*GlobalConfig)) error {
	return gcm.UpdateConfig(updater)
}

// getGlobalConfigPath returns the path to the global configuration file
func getGlobalConfigPath() (string, error) {
	userFolderPath, err := paths.GetUserFolderPath()
	if err != nil {
		return "", shared.WrapError(err, "failed to get user folder path")
	}
	return filepath.Join(userFolderPath, "config.yml"), nil
}

// getDefaultGlobalConfig returns the default global configuration
func getDefaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		Cloudflare: struct {
			Token string `mapstructure:"token" yaml:"token" validate:"required"`
		}{
			Token: "<cloudflare_token>", // User must set this
		},
	}
}
