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

const (
	detachFlag = "detach"
)

var runCmd = &cli.Command{
	Name:  "run",
	Usage: "Run a Cloudflare tunnel",
	Description: "Run a Cloudflare tunnel with the specified configuration. This command will start the tunnel service.",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  detachFlag,
			Value: false,
			Usage: "Run the tunnel in the background (detached mode)",
		},
	},
	Action: execRun,
}

func execRun(ctx context.Context, cmd *cli.Command) error {
	detach := cmd.Bool(detachFlag)
	dryRun := cmd.Bool(dryRunFlag)
	configPath := cmd.String(configPathFlag)

	logger.Infof("Starting tunnel", map[string]any{
		"dry":    dryRun,
		"detach": detach,
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

	// Create framework config for dry-run support
	frameworkConfig := &framework.Config{
		DryRun: dryRun,
	}

	cfTunnel := cloudflare.NewTunnelService(frameworkConfig)
	cfDNS, err := cloudflare.NewDNSService(globalMgr.Get().Cloudflare.Token, cfTunnel, frameworkConfig)
	if err != nil {
		return shared.WrapError(err, "create Cloudflare DNS service failed")
	}

	// Extract tunnel and ingress from config
	tunnelConfig := tunnelMgr.Get()
	tunnelService := tunnel.NewService(tunnelConfig.Tunnel, tunnelConfig.Ingress, cfDNS, cfTunnel)

	if detach {
		if err := tunnelService.Start(ctx); err != nil {
			return shared.WrapError(err, "failed to run tunnel service")
		}
	} else {
		if err := shared.StartManaged(ctx, tunnelService); err != nil {
			return shared.WrapError(err, "failed to start tunnel service")
		}
	}

	logger.Info("Run completed")
	return nil
}

