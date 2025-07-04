package tunnel

import (
	"fmt"
	"os"

	"github.com/stupside/moley/internal/feats/tunnel"
	"github.com/stupside/moley/internal/logger"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a new moley.yml tunnel configuration file",
	Long:  "Creates a new moley.yml tunnel configuration file with default settings. This is the first step to get started with moley.",
	RunE:  execInit,
}

func execInit(cmd *cobra.Command, args []string) error {
	logger.Infof("Initializing tunnel configuration file", map[string]interface{}{
		"command": cmd.Name(),
	})
	if _, err := os.Stat("moley.yml"); err == nil {
		logger.Warnf("Configuration file already exists", map[string]interface{}{
			"config_file": "moley.yml",
		})
		return fmt.Errorf("configuration file already exists")
	}
	data, err := yaml.Marshal(tunnel.GetDefaultConfig())
	if err != nil {
		logger.Errorf("Failed to marshal default config", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile("moley.yml", data, 0644); err != nil {
		logger.Errorf("Failed to write config file", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("failed to write config file: %w", err)
	}
	logger.Debug("Tunnel configuration file created: moley.yml")
	return nil
}
