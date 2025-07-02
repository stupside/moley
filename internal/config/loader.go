package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// LoadMoleyConfig loads the global Moley configuration from the user's home directory
func LoadMoleyConfig() (*MoleyConfig, error) {
	// Get the path to the configuration file
	configPath, err := GetConfigFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	// Create a new Viper instance
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// Try to read config file if it exists
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file doesn't exist, return default config
		return NewConfig(), nil
	}

	// Unmarshal the configuration into a MoleyConfig struct
	var moleyConfig MoleyConfig
	if err := v.Unmarshal(&moleyConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &moleyConfig, nil
}

// LoadFromFlags loads configuration from command line flags
func LoadFromFlags(cmd *cobra.Command) (*MoleyConfig, error) {
	// Create a new Viper instance
	v := viper.New()
	v.SetConfigType("yaml")

	// Bind flags
	if err := v.BindPFlags(cmd.Flags()); err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	// Unmarshal the configuration into a MoleyConfig struct
	var moleyConfig MoleyConfig
	if err := v.Unmarshal(&moleyConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal flag config: %w", err)
	}

	return &moleyConfig, nil
}
