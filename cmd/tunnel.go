package cmd

import (
	"github.com/stupside/moley/cmd/tunnel"
	"github.com/stupside/moley/internal/platform/infrastructure/config"
	"github.com/stupside/moley/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/internal/shared"

	"github.com/spf13/cobra"
)

var tunnelCmd = &cobra.Command{
	Use:   "tunnel",
	Short: "Manage Cloudflare tunnels",
	Long:  "Create, configure, and run Cloudflare tunnels.",
}

func init() {
	tunnelCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize a new tunnel configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load (or create) tunnel config; creation writes default if file doesn't exist
			if _, err := config.NewTunnelConfigManager("moley.yml"); err != nil {
				return shared.WrapError(err, "failed to initialize tunnel config")
			}
			logger.Info("Initialized tunnel configuration at ./moley.yml")
			return nil
		},
	})

	// Add commands to the tunnel command
	tunnelCmd.AddCommand(tunnel.RunCmd)

	// Register the tunnel command with the root command
	rootCmd.AddCommand(tunnelCmd)
}
