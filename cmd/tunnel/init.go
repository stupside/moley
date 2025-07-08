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
	Short: fmt.Sprintf("Init a new %s tunnel configuration file", tunnel.TunneConfigFile),
	Long:  fmt.Sprintf("This command initializes a new %s tunnel configuration file with default settings.", tunnel.TunneConfigFile),
	RunE:  execInit,
}

func execInit(cmd *cobra.Command, args []string) error {
	logger.Infof("Initializing tunnel configuration file", map[string]interface{}{
		"command": cmd.Name(),
	})
	if _, err := os.Stat(tunnel.TunneConfigFile); err == nil {
		logger.Warnf("Configuration file already exists", map[string]interface{}{
			"config_file": tunnel.TunneConfigFile,
		})
		return fmt.Errorf("configuration file already exists")
	}
	data, err := yaml.Marshal(tunnel.GetDefaultConfig())
	if err != nil {
		logger.Errorf("Failed to marshal default config", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(tunnel.TunneConfigFile, data, 0644); err != nil {
		logger.Errorf("Failed to write config file", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("failed to write config file: %w", err)
	}
	logger.Debugf("Tunnel configuration file created", map[string]interface{}{
		"config_file": tunnel.TunneConfigFile,
	})
	return nil
}
