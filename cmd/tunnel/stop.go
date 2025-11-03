package tunnel

import (
	"context"

	"github.com/stupside/moley/v2/internal/core/application/tunnel"
	"github.com/stupside/moley/v2/internal/platform/adapters/cloudflare"
	"github.com/stupside/moley/v2/internal/platform/framework"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/config"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/shared"

	"github.com/urfave/cli/v3"
)

var stopCmd = &cli.Command{
	Name:   "stop",
	Usage:  "Bring the tunnel down",
	Action: execStop,
}

func execStop(ctx context.Context, cmd *cli.Command) error {
	dryRun := cmd.Bool(dryRunFlag)
	configPath := cmd.String(configPathFlag)

	logger.Infof("Bringing tunnel down", map[string]any{
		"dry":    dryRun,
		"config": configPath,
	})

	tunnelMgr, err := config.NewTunnelManager(configPath)
	if err != nil {
		return shared.WrapError(err, "create tunnel config manager failed")
	}

	// Build adapters (Cloudflare) implementing ports
	globalMgr, err := config.NewGlobalManager(cmd)
	if err != nil {
		return shared.WrapError(err, "create global config manager failed")
	}

	// Extract global config for adapter setup
	globalConfig, err := globalMgr.Get(true)
	if err != nil {
		return shared.WrapError(err, "get global config failed")
	}

	// Create framework config for dry-run support
	frameworkConfig := &framework.Config{
		DryRun: dryRun,
	}

	cfTunnel := cloudflare.NewTunnelService(frameworkConfig)
	cfDNS, err := cloudflare.NewDNSService(globalConfig.Cloudflare.Token, cfTunnel, frameworkConfig)
	if err != nil {
		return shared.WrapError(err, "create Cloudflare DNS service failed")
	}

	// Extract tunnel and ingress from config
	tunnelConfig, err := tunnelMgr.Get(true)
	if err != nil {
		return shared.WrapError(err, "get tunnel config failed")
	}

	tunnelService := tunnel.NewService(tunnelConfig.Tunnel, tunnelConfig.Ingress, cfDNS, cfTunnel)

	if err := tunnelService.Stop(ctx); err != nil {
		return shared.WrapError(err, "failed to stop tunnel service")
	}

	return nil
}
