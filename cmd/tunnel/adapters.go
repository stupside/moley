package tunnel

import (
	"context"
	"fmt"

	apptunnel "github.com/stupside/moley/v2/internal/core/application/tunnel"
	"github.com/stupside/moley/v2/internal/platform/adapters/cloudflare"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/config"

	cfgo "github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/option"
	"github.com/urfave/cli/v3"
)

// buildTunnelService creates the Cloudflare adapters and returns a ready-to-use tunnel service.
func buildTunnelService(ctx context.Context, cmd *cli.Command) (*apptunnel.Service, error) {
	dryRun := cmd.Bool(dryRunFlag)
	configPath := cmd.String(configPathFlag)

	tunnelMgr, err := config.NewTunnelManager(configPath)
	if err != nil {
		return nil, fmt.Errorf("create tunnel config manager failed: %w", err)
	}

	globalMgr, err := config.NewGlobalManager(cmd)
	if err != nil {
		return nil, fmt.Errorf("create global config manager failed: %w", err)
	}

	globalConfig, err := globalMgr.Get(true)
	if err != nil {
		return nil, fmt.Errorf("get global config failed: %w", err)
	}

	tunnelConfig, err := tunnelMgr.Get(true)
	if err != nil {
		return nil, fmt.Errorf("get tunnel config failed: %w", err)
	}

	adapterConfig := &cloudflare.Config{
		DryRun: dryRun,
	}

	cfClient := cfgo.NewClient(
		option.WithAPIToken(globalConfig.Cloudflare.Token),
	)

	cfTunnel, err := cloudflare.NewTunnelService(ctx, cfClient, tunnelConfig.Ingress.Zone, adapterConfig)
	if err != nil {
		return nil, fmt.Errorf("create Cloudflare tunnel service failed: %w", err)
	}

	cfDNS := cloudflare.NewDNSService(cfClient, cfTunnel, adapterConfig)

	return apptunnel.NewService(tunnelConfig.Tunnel, tunnelConfig.Ingress, cfDNS, cfTunnel), nil
}
