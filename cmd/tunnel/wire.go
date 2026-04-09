package tunnel

import (
	"context"
	"fmt"

	appconfig "github.com/stupside/moley/v2/internal/app/config"
	application "github.com/stupside/moley/v2/internal/app/session"
	accesscf "github.com/stupside/moley/v2/internal/features/access/cloudflare"
	accessusecase "github.com/stupside/moley/v2/internal/features/access/usecase"
	dnscf "github.com/stupside/moley/v2/internal/features/dns/cloudflare"
	tunnelcf "github.com/stupside/moley/v2/internal/features/tunnel/cloudflare"

	cfgo "github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/option"
	"github.com/urfave/cli/v3"
)

// buildTunnelService creates the Cloudflare adapters and returns a ready-to-use tunnel service.
func buildTunnelService(ctx context.Context, cmd *cli.Command) (*application.Service, error) {
	dryRun := cmd.Bool(dryRunFlag)
	configPath := cmd.String(configPathFlag)

	tunnelMgr, err := appconfig.NewTunnelManager(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create tunnel config manager: %w", err)
	}

	globalMgr, err := appconfig.NewGlobalManager()
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

	cfClient := cfgo.NewClient(
		option.WithAPIToken(globalConfig.Cloudflare.Token),
	)

	cfTunnel, err := tunnelcf.NewTunnelService(ctx, cfClient, tunnelConfig.Ingress.Zone, dryRun)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloudflare tunnel service: %w", err)
	}

	cfDNS := dnscf.NewDNSService(cfClient, dryRun)

	var cfAccess accessusecase.AccessManager
	var cfPolicy accessusecase.PolicyManager
	if tunnelConfig.Access.HasPolicies() || tunnelConfig.Ingress.HasAccessConfig() {
		svc := accesscf.NewAccessService(cfClient, cfTunnel.AccountID(), dryRun)
		cfAccess = svc
		cfPolicy = svc
	}

	return application.NewService(tunnelConfig.Tunnel, tunnelConfig.Ingress, tunnelConfig.Access, cfDNS, cfTunnel, cfTunnel, cfTunnel, cfAccess, cfPolicy), nil
}
