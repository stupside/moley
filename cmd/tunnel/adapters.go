package tunnel

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/core/application"
	"github.com/stupside/moley/v2/internal/core/ports"
	"github.com/stupside/moley/v2/internal/platform/adapters/cloudflare"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/config"

	cfgo "github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/option"
	"github.com/urfave/cli/v3"
)

// buildTunnelService creates the Cloudflare adapters and returns a ready-to-use tunnel service.
func buildTunnelService(ctx context.Context, cmd *cli.Command) (*application.Service, error) {
	dryRun := cmd.Bool(dryRunFlag)
	configPath := cmd.String(configPathFlag)

	tunnelMgr, err := config.NewTunnelManager(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create tunnel config manager: %w", err)
	}

	globalMgr, err := config.NewGlobalManager(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create global config manager: %w", err)
	}

	globalConfig, err := globalMgr.Get(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get global config: %w", err)
	}

	tunnelConfig, err := tunnelMgr.Get(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get tunnel config: %w", err)
	}

	adapterConfig := &cloudflare.Config{
		DryRun: dryRun,
	}

	cfClient := cfgo.NewClient(
		option.WithAPIToken(globalConfig.Cloudflare.Token),
	)

	cfTunnel, err := cloudflare.NewTunnelService(ctx, cfClient, tunnelConfig.Ingress.Zone, adapterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloudflare tunnel service: %w", err)
	}

	cfDNS := cloudflare.NewDNSService(cfClient, cfTunnel, adapterConfig)

	var cfAccess ports.AccessService
	if tunnelConfig.Ingress.HasAccessConfig() {
		cfAccess = cloudflare.NewAccessService(cfClient, cfTunnel.AccountID(), adapterConfig)
	}

	return application.NewService(tunnelConfig.Tunnel, tunnelConfig.Ingress, cfDNS, cfTunnel, cfAccess), nil
}
