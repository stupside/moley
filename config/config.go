package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Load initializes Viper configuration from file and environment variables
func Load() error {
	// Set default config file name
	viper.SetConfigName("moley")
	viper.SetConfigType("yml")

	// Look for config in current directory
	viper.AddConfigPath(".")

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read configuration file: %w", err)
		}
		// Config file not found is not an error, we can use defaults
	}

	// Bind environment variables
	viper.SetEnvPrefix("MOLEY")
	viper.AutomaticEnv()

	return nil
}

// GetConfigFilePath returns the path to the currently loaded config file
func GetConfigFilePath() string {
	return viper.ConfigFileUsed()
}

// CreateDefaultConfig creates a default moley.yml configuration file
func CreateDefaultConfig() error {
	defaultConfig := `# Moley Configuration
# This file contains the configuration for Moley tunnel management

# Cloudflare configuration
cloudflare:
  api_token: ""  # Your Cloudflare API token

# Tunnel configuration
tunnel:
  name: "moley-tunnel"  # Name for the tunnel
  port: 8080           # Local port to expose
  hostname: ""         # Custom hostname (optional, will be auto-generated)

# Zone configuration
zone: local.example.com

# Applications to expose
apps:
  - port: 3000
    subdomain: api
  - port: 8080
    subdomain: web
`

	configPath := filepath.Join(".", "moley.yml")
	if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to create default configuration file: %w", err)
	}

	return nil
}
