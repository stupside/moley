package config

import (
	"github.com/spf13/cobra"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/config"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/shared"
)

const (
	cloudflareTokenFlag = "cloudflare.token"
)

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set Moley configuration values",
	Long:  "Set Moley configuration values using command-line flags.",
	RunE:  execSet,
}

func execSet(cmd *cobra.Command, args []string) error {
	logger.Info("Editing configuration")

	// Load global config
	manager, err := config.NewGlobalConfigManager(cmd)
	if err != nil {
		return shared.WrapError(err, "failed to create global config manager")
	}

	if err := manager.UpdateGlobalConfig(func(cfg *config.GlobalConfig) {
		cfg.Cloudflare.Token = cmd.Flag(cloudflareTokenFlag).Value.String()
	}); err != nil {
		return shared.WrapError(err, "failed to update global config")
	}

	logger.Info("Configuration saved successfully")
	return nil
}

func init() {
	setCmd.Flags().String(cloudflareTokenFlag, "", "Cloudflare API token")
	if err := setCmd.MarkFlagRequired(cloudflareTokenFlag); err != nil {
		logger.LogError(err, "failed to mark flag as required")
	}
}
