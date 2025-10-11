package tunnel

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/core/domain"
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

	// Add all apps for this ingress
	tunnelConfigConfigs := []TunnelConfigConfig{
		{
			Tunnel: s.tunnel,
			Ingress: &domain.Ingress{
				Zone: s.ingress.Zone,
				Apps: s.ingress.Apps,
			},
		},
	}
	framework.AddManager(orchestrator, tunnelConfigHandler, tunnelConfigConfigs)

	// Add tunnel run management
	tunnelRunHandler := newTunnelRunHandler(s.tunnelService)
	tunnelRunConfigs := []TunnelRunConfig{
		{Tunnel: s.tunnel},
	}
	framework.AddManager(orchestrator, tunnelRunHandler, tunnelRunConfigs)

	// Add DNS record management
	dnsHandler := newDNSRecordHandler(s.dnsService)

	var dnsConfigs []DNSRecordConfig

	switch s.ingress.Mode {
	case domain.IngressModeWildcard:
		// Wildcard DNS: single *.zone record
		dnsConfigs = []DNSRecordConfig{
			{
				Zone:      s.ingress.Zone,
				Tunnel:    s.tunnel,
				Subdomain: "*",
			},
		}
	case domain.IngressModeSubdomain:
		// Individual DNS: one record per app
		dnsConfigs = make([]DNSRecordConfig, 0, len(s.ingress.Apps))
		for _, app := range s.ingress.Apps {
			dnsConfigs = append(dnsConfigs, DNSRecordConfig{
				Zone:      s.ingress.Zone,
				Tunnel:    s.tunnel,
				Subdomain: app.Expose.Subdomain,
			})
		}
	default:
		// Should never happen due to validation, but handle gracefully
		return nil, shared.WrapError(
			fmt.Errorf("unknown ingress mode: %s", s.ingress.Mode),
			"invalid ingress mode",
		)
	}

	// Add all desired DNS records - reconciliation will add/remove/update as needed
	framework.AddManager(orchestrator, dnsHandler, dnsConfigs)

	return orchestrator, nil
}
