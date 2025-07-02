package cmd

import (
	"moley/internal/config"
	"moley/internal/logger"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Edit Moley configuration",
	Long:  "Edit Moley configuration. You can set any value in the Moley config file using command-line flags.",
	RunE:  execConfig,
}

func execConfig(cmd *cobra.Command, args []string) error {
	logger.Infof("Editing configuration", map[string]interface{}{
		"command": cmd.Name(),
	})

	flagConfig, err := config.LoadFromFlags(cmd)
	if err != nil {
		logger.Errorf("Failed to load configuration from flags", map[string]interface{}{"error": err.Error()})
		return err
	}

	manager := config.GetManager()
	currentConfig := manager.Get()

	currentConfig.Cloudflare.Token = flagConfig.Cloudflare.Token

	manager.Set(currentConfig)
	if err := manager.Save(); err != nil {
		logger.Errorf("Failed to save configuration", map[string]interface{}{"error": err.Error()})
		return err
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
