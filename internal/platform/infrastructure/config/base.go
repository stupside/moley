package config

import (
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/stupside/moley/internal/shared"
)

// ConfigManager is a generic configuration manager
type ConfigManager[T any] struct {
	path   string
	viper  *viper.Viper
	config *T
}

// NewConfigManager creates a new generic configuration manager
func NewConfigManager[T any](path string, defaultConfig *T) *ConfigManager[T] {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	return &ConfigManager[T]{
		viper:  v,
		path:   path,
		config: defaultConfig,
	}
}

// Load loads the configuration from file or creates default if not exists
func (cm *ConfigManager[T]) Load() error {
	// Try to read existing config
	if err := cm.viper.ReadInConfig(); err == nil {
		// Config file exists, unmarshal it into the existing config (which has defaults)
		if err := cm.viper.Unmarshal(cm.config); err != nil {
			return shared.WrapError(err, "failed to unmarshal config")
		}
	} else {
		// Config file doesn't exist, create default
		// Create config directory if it doesn't exist
		configDir := filepath.Dir(cm.path)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return shared.WrapError(err, "failed to create config directory")
		}

		// Save default config
		if err := cm.Save(); err != nil {
			return shared.WrapError(err, "failed to save default config")
		}
	}

	// Validate config
	if err := cm.Validate(); err != nil {
		return shared.WrapError(err, "config validation failed")
	}

	return nil
}

// Save writes the configuration to file
func (cm *ConfigManager[T]) Save() error {
	data, err := yaml.Marshal(cm.config)
	if err != nil {
		return shared.WrapError(err, "failed to marshal config")
	}

	// Restrictive permissions by default (configs may contain secrets)
	if err := os.WriteFile(cm.path, data, 0600); err != nil {
		return shared.WrapError(err, "failed to write config file")
	}

	return nil
}

// Validate validates the configuration using struct tags
func (cm *ConfigManager[T]) Validate() error {
	return validator.New().Struct(cm.config)
}

// GetConfig returns the current configuration
func (cm *ConfigManager[T]) GetConfig() *T {
	return cm.config
}

// UpdateConfig updates the configuration and saves it
func (cm *ConfigManager[T]) UpdateConfig(updater func(*T)) error {
	updater(cm.config)
	return cm.Save()
}

// BindEnv binds environment variables to the config using viper
func (cm *ConfigManager[T]) BindEnv(prefix string) {
	cm.viper.SetEnvPrefix(prefix)
	cm.viper.AutomaticEnv()
}

// BindFlags binds CLI flags to the config using viper/cobra
func (cm *ConfigManager[T]) BindFlags(cmd *cobra.Command) {
	_ = cm.viper.BindPFlags(cmd.Flags())
}
