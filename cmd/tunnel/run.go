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

const (
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a Cloudflare tunnel",
	Long:  "Run a Cloudflare tunnel with the specified configuration. This command will start the tunnel service.",
	RunE:  execRun,
}

func execRun(cmd *cobra.Command, args []string) error {
	dryRun := viper.GetBool(dryRunFlag)
	configPath := viper.GetString(configPathFlag)

	logger.Infof("Starting tunnel", map[string]any{
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

		if err := shared.StartManaged(cmd.Context(), tunnelService); err != nil {
			return shared.WrapError(err, "failed to start tunnel service")
		}
	}

	logger.Info("Run completed")
	return nil
}

func init() {
	}
}
