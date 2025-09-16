package tunnel

import (
	"github.com/stupside/moley/v2/internal/platform/infrastructure/config"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/shared"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	dryRunFlag     = "dry-run"
	configPathFlag = "config"
)

var Cmd = &cobra.Command{
	Use:   "tunnel",
	Short: "Manage Cloudflare tunnels",
}

func init() {
	Cmd.PersistentFlags().Bool(dryRunFlag, false, "Simulate actions without making any changes")
	if err := viper.BindPFlag(dryRunFlag, Cmd.PersistentFlags().Lookup(dryRunFlag)); err != nil {
		logger.Fatal("Failed to bind dry-run flag to Viper")
	}

	Cmd.PersistentFlags().String(configPathFlag, "moley.yml", "Path to the tunnel configuration file")
	if err := viper.BindPFlag(configPathFlag, Cmd.PersistentFlags().Lookup(configPathFlag)); err != nil {
		logger.Fatal("Failed to bind config flag to Viper")
	}

	Cmd.AddCommand(runCmd)

	Cmd.AddCommand(&cobra.Command{
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
}
