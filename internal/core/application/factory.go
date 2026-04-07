package application

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/core/application/access"
	"github.com/stupside/moley/v2/internal/core/application/dns"
	"github.com/stupside/moley/v2/internal/core/application/tunnel"
	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/platform/framework"
)

func (s *Service) createOrchestrator(_ context.Context) (*framework.Reconciler, error) {
	orchestrator, err := framework.NewReconciler()
	if err != nil {
		return nil, fmt.Errorf("failed to create resource orchestrator: %w", err)
	}

	// tunnel-create — no dependencies
	framework.Register(orchestrator, tunnel.NewCreateHandler(s.tunnelService),
		func(reg *framework.OutputRegistry) ([]tunnel.CreateInput, error) {
			return []tunnel.CreateInput{
				{
					Name:       s.tunnel.Name,
					Persistent: s.tunnel.Persistent,
				},
			}, nil
		},
	)

	// tunnel-config — depends on tunnel-create (needs TunnelUUID)
	framework.Register(orchestrator, tunnel.NewConfigHandler(s.tunnelService),
		func(reg *framework.OutputRegistry) ([]tunnel.ConfigInput, error) {
			create, ok := framework.GetOutput[tunnel.CreateOutput](reg, tunnel.CreateHandlerName, s.tunnel.Name)
			if !ok {
				return nil, fmt.Errorf("%s: missing upstream output from %s", tunnel.ConfigHandlerName, tunnel.CreateHandlerName)
			}
			return []tunnel.ConfigInput{
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
		tunnel.CreateHandlerName,
	)

	// tunnel-run — depends on tunnel-config (needs ConfigPath + ContentHash)
	framework.Register(orchestrator, tunnel.NewRunHandler(s.tunnelService),
		func(reg *framework.OutputRegistry) ([]tunnel.RunInput, error) {
			config, ok := framework.GetOutput[tunnel.ConfigOutput](reg, tunnel.ConfigHandlerName, s.tunnel.Name)
			if !ok {
				return nil, fmt.Errorf("%s: missing upstream output from %s", tunnel.RunHandlerName, tunnel.ConfigHandlerName)
			}
			return []tunnel.RunInput{
				{
					TunnelName:  s.tunnel.Name,
					ConfigPath:  config.ConfigPath,
					ContentHash: config.ContentHash,
				},
			}, nil
		},
		tunnel.ConfigHandlerName,
	)

	// dns-record — depends on tunnel-create (needs TunnelUUID)
	framework.Register(orchestrator, dns.NewHandler(s.dnsService),
		func(reg *framework.OutputRegistry) ([]dns.RecordInput, error) {
			create, ok := framework.GetOutput[tunnel.CreateOutput](reg, tunnel.CreateHandlerName, s.tunnel.Name)
			if !ok {
				return nil, fmt.Errorf("%s: missing upstream output from %s", dns.HandlerName, tunnel.CreateHandlerName)
			}

			var inputs []dns.RecordInput

			switch s.ingress.Mode {
			case domain.IngressModeWildcard:
				inputs = []dns.RecordInput{
					{
						Zone:       s.ingress.Zone,
						Subdomain:  "*",
						TunnelName: s.tunnel.Name,
						TunnelUUID: create.TunnelUUID,
						Persistent: s.tunnel.Persistent,
					},
				}
			case domain.IngressModeSubdomain:
				inputs = make([]dns.RecordInput, 0, len(s.ingress.Apps))
				for _, app := range s.ingress.Apps {
					inputs = append(inputs, dns.RecordInput{
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
		tunnel.CreateHandlerName,
	)

	// access-app — depends on dns-record (protects already-routed subdomains)
	if s.accessService != nil && s.ingress.HasAccessConfig() {
		framework.Register(orchestrator, access.NewHandler(s.accessService),
			func(reg *framework.OutputRegistry) ([]access.AppInput, error) {
				var inputs []access.AppInput
				for _, app := range s.ingress.Apps {
					if app.Access == nil {
						continue
					}
					inputs = append(inputs, access.AppInput{
						Zone:              s.ingress.Zone,
						Subdomain:         app.Expose.Subdomain,
						SessionDuration:   app.Access.SessionDuration,
						Decision:          app.Access.Policy.Decision,
						IdentityProviders: app.Access.IdentityProviders,
						Emails:            app.Access.Policy.Include.Emails,
						EmailDomains:      app.Access.Policy.Include.EmailDomains,
					})
				}
				return inputs, nil
			},
			dns.HandlerName,
		)
	}

	return orchestrator, nil
}
