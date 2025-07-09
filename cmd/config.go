package cmd

import (
	"github.com/stupside/moley/internal/config"
	"github.com/stupside/moley/internal/logger"
	"github.com/stupside/moley/internal/shared"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Edit Moley configuration",
	Long:  "Edit Moley configuration. You can set any value in the Moley config file using command-line flags.",
	RunE:  execConfig,
}

func execConfig(cmd *cobra.Command, args []string) error {
	logger.Info("Editing configuration")

	globalConfigManager, err := config.NewGlobalConfigManager(cmd)
	if err != nil {
		return shared.WrapError(err, "failed to get global config manager")
	}

	// Load configuration (this will now work even if file doesn't exist)
	globalConfig, err := globalConfigManager.Load(false)
	if err != nil {
		return shared.WrapError(err, "failed to load global configuration")
	}

	// Save the configuration with validation
	if err := globalConfigManager.Save(globalConfig, true); err != nil {
		return shared.WrapError(err, "failed to save configuration")
	}

	logger.Info("Configuration saved successfully")
	return nil
}

func init() {
	configCmd.Flags().String("cloudflare.token", "", "Cloudflare API token")
	if err := configCmd.MarkFlagRequired("cloudflare.token"); err != nil {
		logger.LogFatal(err, "Failed to mark flag as required")
	}

	rootCmd.AddCommand(configCmd)
}
