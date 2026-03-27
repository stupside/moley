package tunnel

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/platform/framework"
)

// Handler name constants used for DAG dependency wiring.
const (
	HandlerTunnelCreate = "tunnel-create"
	HandlerTunnelConfig = "tunnel-config"
	HandlerTunnelRun    = "tunnel-run"
	HandlerDNSRecord    = "dns-record"
)

func (s *Service) createOrchestrator(_ context.Context) (*framework.Reconciler, error) {
	orchestrator, err := framework.NewReconciler()
	if err != nil {
		return nil, fmt.Errorf("failed to create resource orchestrator: %w", err)
	}

	// tunnel-create — no dependencies
	framework.Register(orchestrator, newCreateHandler(s.tunnelService),
		func(reg *framework.OutputRegistry) ([]CreateInput, error) {
			return []CreateInput{
				{
					Name:       s.tunnel.Name,
					Persistent: s.tunnel.Persistent,
				},
			}, nil
		},
	)

	// tunnel-config — depends on tunnel-create (needs TunnelUUID)
	framework.Register(orchestrator, newConfigHandler(s.tunnelService),
		func(reg *framework.OutputRegistry) ([]ConfigInput, error) {
			create, ok := framework.GetOutput[CreateOutput](reg, HandlerTunnelCreate, s.tunnel.Name)
			if !ok {
				return nil, fmt.Errorf("%s: missing upstream output from %s", HandlerTunnelConfig, HandlerTunnelCreate)
			}
			return []ConfigInput{
				{
					TunnelName: s.tunnel.Name,
					TunnelUUID: create.TunnelUUID,
					Persistent: s.tunnel.Persistent,
					Ingress: &domain.Ingress{
						Zone: s.ingress.Zone,
						Apps: s.ingress.Apps,
						Mode: s.ingress.Mode,
					},
				},
			}, nil
		},
		HandlerTunnelCreate,
	)

	// tunnel-run — depends on tunnel-config (needs ConfigPath + ContentHash)
	framework.Register(orchestrator, newRunHandler(s.tunnelService),
		func(reg *framework.OutputRegistry) ([]RunInput, error) {
			config, ok := framework.GetOutput[ConfigOutput](reg, HandlerTunnelConfig, s.tunnel.Name)
			if !ok {
				return nil, fmt.Errorf("%s: missing upstream output from %s", HandlerTunnelRun, HandlerTunnelConfig)
			}
			return []RunInput{
				{
					TunnelName:  s.tunnel.Name,
					ConfigPath:  config.ConfigPath,
					ContentHash: config.ContentHash,
				},
			}, nil
		},
		HandlerTunnelConfig,
	)

	// dns-record — depends on tunnel-create (needs TunnelUUID)
	framework.Register(orchestrator, newRecordHandler(s.dnsService),
		func(reg *framework.OutputRegistry) ([]RecordInput, error) {
			create, ok := framework.GetOutput[CreateOutput](reg, HandlerTunnelCreate, s.tunnel.Name)
			if !ok {
				return nil, fmt.Errorf("%s: missing upstream output from %s", HandlerDNSRecord, HandlerTunnelCreate)
			}

			var inputs []RecordInput

			switch s.ingress.Mode {
			case domain.IngressModeWildcard:
				inputs = []RecordInput{
					{
						Zone:       s.ingress.Zone,
						Subdomain:  "*",
						TunnelName: s.tunnel.Name,
						TunnelUUID: create.TunnelUUID,
						Persistent: s.tunnel.Persistent,
					},
				}
			case domain.IngressModeSubdomain:
				inputs = make([]RecordInput, 0, len(s.ingress.Apps))
				for _, app := range s.ingress.Apps {
					inputs = append(inputs, RecordInput{
						Zone:       s.ingress.Zone,
						Subdomain:  app.Expose.Subdomain,
						TunnelName: s.tunnel.Name,
						TunnelUUID: create.TunnelUUID,
						Persistent: s.tunnel.Persistent,
					})
				}
			default:
				return nil, fmt.Errorf("invalid ingress mode: %s", s.ingress.Mode)
			}

			return inputs, nil
		},
		HandlerTunnelCreate,
	)

	return orchestrator, nil
}
