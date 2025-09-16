package tunnel

import (
	"context"

	"github.com/stupside/moley/v2/internal/platform/framework"
	"github.com/stupside/moley/v2/internal/shared"
)

func (s *Service) createResourceManager(_ context.Context) (*framework.ResourceOrchestrator, error) {
	// Create the resource orchestrator
	orchestrator, err := framework.NewResourceOrchestrator()
	if err != nil {
		return nil, shared.WrapError(err, "failed to create resource orchestrator")
	}

	// Add tunnel creation management
	tunnelCreateHandler := newTunnelCreateHandler(s.tunnelService)
	tunnelCreateConfigs := []TunnelCreateConfig{
		{Tunnel: s.tunnel},
	}
	framework.AddManager(orchestrator, tunnelCreateHandler, tunnelCreateConfigs)

	// Add tunnel configuration management
	tunnelConfigHandler := newTunnelConfigHandler(s.tunnelService)
	tunnelConfigConfigs := []TunnelConfigConfig{
		{
			Tunnel:  s.tunnel,
			Ingress: s.ingress,
		},
	}
	framework.AddManager(orchestrator, tunnelConfigHandler, tunnelConfigConfigs)

	// Add tunnel run management
	tunnelRunHandler := newTunnelRunHandler(s.tunnelService)
	tunnelRunConfigs := []TunnelRunConfig{
		{Tunnel: s.tunnel},
	}
	framework.AddManager(orchestrator, tunnelRunHandler, tunnelRunConfigs)

	// Add DNS record management for each app
	dnsHandler := newDNSRecordHandler(s.dnsService)
	var dnsConfigs []DNSRecordConfig
	for _, app := range s.ingress.Apps {
		dnsConfigs = append(dnsConfigs, DNSRecordConfig{
			Tunnel:    s.tunnel,
			Zone:      s.ingress.Zone,
			Subdomain: app.Expose.Subdomain,
		})
	}
	framework.AddManager(orchestrator, dnsHandler, dnsConfigs)

	return orchestrator, nil
}
