package cmd

import (
	"fmt"

	"github.com/stupside/moley/internal/config"
	"github.com/stupside/moley/internal/logger"

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
		return fmt.Errorf("failed to get global config manager: %w", err)
	}

	// Load configuration (this will now work even if file doesn't exist)
	globalConfig, err := globalConfigManager.Load(false)
	if err != nil {
		return fmt.Errorf("failed to load global configuration: %w", err)
	}

	// Save the configuration with validation
	if err := globalConfigManager.Save(globalConfig, true); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	logger.Debug("Configuration saved")
	return nil
}

func init() {
	configCmd.Flags().String("cloudflare.token", "", "Cloudflare API token")
	if err := configCmd.MarkFlagRequired("cloudflare.token"); err != nil {
		logger.Fatalf("Failed to mark flag as required", map[string]interface{}{
			"error": err.Error(),
		})
	}

	rootCmd.AddCommand(configCmd)
}
