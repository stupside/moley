package tunnel

import (
	"fmt"

	"github.com/stupside/moley/internal/feats/tunnel"
	"github.com/stupside/moley/internal/logger"

	"github.com/spf13/cobra"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: fmt.Sprintf("Init a new %s tunnel configuration file", tunnel.TunnelConfigFile),
	Long:  fmt.Sprintf("This command initializes a new %s tunnel configuration file with default settings.", tunnel.TunnelConfigFile),
	RunE:  execInit,
}

func execInit(cmd *cobra.Command, args []string) error {
	logger.Info("Initializing tunnel configuration file")

	tunnelConfigManager := tunnel.NewTunnelConfigManager()

	if err := tunnelConfigManager.Init(); err != nil {
		return fmt.Errorf("failed to initialize tunnel configuration file: %w", err)
	}

	logger.Info("Tunnel configuration file created")
	return nil
}
