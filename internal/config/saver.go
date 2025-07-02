package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// SaveMoleyConfigToFile saves the Moley configuration to file
func SaveMoleyConfigToFile(cfg *MoleyConfig) error {
	// Get the path to the configuration file
	configPath, err := GetConfigFilePath()
	if err != nil {
		return fmt.Errorf("failed to get config file path: %w", err)
	}

	// Marshal config to YAML and write to file
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write the YAML data to the configuration file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
