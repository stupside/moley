package tunnel

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stupside/moley/v2/internal/core/application/tunnel"
	"github.com/stupside/moley/v2/internal/platform/adapters/cloudflare"
	"github.com/stupside/moley/v2/internal/platform/framework"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/config"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/shared"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Bring the tunnel down",
	RunE:  execStop,
}

func execStop(cmd *cobra.Command, args []string) error {
	dryRun := viper.GetBool(dryRunFlag)
	configPath := viper.GetString("config")

	logger.Infof("Bringing tunnel down", map[string]any{
		"dry":    dryRun,
		"config": configPath,
	})

	tunnelConfigManager, err := config.NewTunnelConfigManager(configPath)
	if err != nil {
		return shared.WrapError(err, "failed to create tunnel config manager")
	}

	// Build adapters (Cloudflare) implementing ports
	globalConfigManager, err := config.NewGlobalConfigManager(cmd)
	if err != nil {
		return shared.WrapError(err, "failed to create global config manager")
	}

	// Create framework config for dry-run support
	frameworkConfig := &framework.Config{
		DryRun: dryRun,
	}

	cfTunnel := cloudflare.NewTunnelService(frameworkConfig)
	cfDNS, err := cloudflare.NewDNSService(globalConfigManager.GetGlobalConfig().Cloudflare.Token, cfTunnel, frameworkConfig)
	if err != nil {
		return shared.WrapError(err, "failed to create Cloudflare DNS service")
	}

	// Extract tunnel and ingress from config
	tunnelConfig := tunnelConfigManager.GetTunnelConfig()

	tunnelService := tunnel.NewService(tunnelConfig.Tunnel, tunnelConfig.Ingress, cfDNS, cfTunnel)

	if err := tunnelService.Stop(cmd.Context()); err != nil {
		return shared.WrapError(err, "failed to stop tunnel service")
	}

	return nil
}
