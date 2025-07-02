package config

import (
	"fmt"
	"os"
)

const (
	configFileName   = "config.yml"
	ConfigFileFolder = ".moley"
)

type CloudflareConfig struct {
	Token string `mapstructure:"token" yaml:"token"`
}

type MoleyConfig struct {
	Cloudflare CloudflareConfig `mapstructure:"cloudflare" yaml:"cloudflare"`
}

// NewConfig creates a new Config instance with default values
func NewConfig() *MoleyConfig {
	return &MoleyConfig{
		Cloudflare: CloudflareConfig{},
	}
}

// GetConfigFilePath returns the full path to the config file
func GetConfigFilePath() (string, error) {
	// Get the path to the config folder
	configFolderPath, err := GetConfigFolderPath()
	if err != nil {
		return "", fmt.Errorf("failed to get config folder path: %w", err)
	}

	// Return the full path to the config file <homedir>/.moley/config.yaml
	configFilePath := fmt.Sprintf("%s/%s", configFolderPath, configFileName)

	// Ensure the config file exists
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		// Create the config file if it does not exist
		file, err := os.Create(configFilePath)
		if err != nil {
			return "", fmt.Errorf("failed to create config file: %w", err)
		}
		defer file.Close()
	}

	return configFilePath, nil
}

// GetConfigFolderPath returns the path to the config folder
func GetConfigFolderPath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Construct the path to the config folder <homedir>/.moley
	configFolderPath := fmt.Sprintf("%s/%s", homedir, ConfigFileFolder)

	// Ensure the config folder exists
	if err := os.MkdirAll(configFolderPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create config folder: %w", err)
	}

	return configFolderPath, nil
}
