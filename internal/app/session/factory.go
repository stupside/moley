package session

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/domain"
	accessusecase "github.com/stupside/moley/v2/internal/features/access/usecase"
	dnsusecase "github.com/stupside/moley/v2/internal/features/dns/usecase"
	tunnelusecase "github.com/stupside/moley/v2/internal/features/tunnel/usecase"
	framework "github.com/stupside/moley/v2/internal/platform/orchestration"
)

func (s *Service) createOrchestrator(_ context.Context) (*framework.Reconciler, error) {
	orchestrator, err := framework.NewReconciler()
	if err != nil {
		return nil, fmt.Errorf("failed to create resource orchestrator: %w", err)
	}

	// tunnel-create — no dependencies
	framework.Register(orchestrator, tunnelusecase.NewCreateHandler(s.tunnelCreator),
		func(reg *framework.OutputRegistry) ([]tunnelusecase.CreateInput, error) {
			return []tunnelusecase.CreateInput{
				{
					Name:       s.tunnel.Name,
					Persistent: s.tunnel.Persistent,
				},
			}, nil
		},
	)

	// tunnel-config — depends on tunnel-create (needs TunnelUUID)
	framework.Register(orchestrator, tunnelusecase.NewConfigHandler(s.tunnelConfigurator),
		func(reg *framework.OutputRegistry) ([]tunnelusecase.ConfigInput, error) {
			create, ok := framework.GetOutput[tunnelusecase.CreateOutput](reg, tunnelusecase.CreateHandlerName, s.tunnel.Name)
			if !ok {
				return nil, fmt.Errorf("%s: missing upstream output from %s", tunnelusecase.ConfigHandlerName, tunnelusecase.CreateHandlerName)
			}
			return []tunnelusecase.ConfigInput{
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
		tunnelusecase.CreateHandlerName,
	)

	// tunnel-run — depends on tunnel-config (needs ConfigPath + ContentHash)
	framework.Register(orchestrator, tunnelusecase.NewRunHandler(s.tunnelRunner),
		func(reg *framework.OutputRegistry) ([]tunnelusecase.RunInput, error) {
			config, ok := framework.GetOutput[tunnelusecase.ConfigOutput](reg, tunnelusecase.ConfigHandlerName, s.tunnel.Name)
			if !ok {
				return nil, fmt.Errorf("%s: missing upstream output from %s", tunnelusecase.RunHandlerName, tunnelusecase.ConfigHandlerName)
			}
			return []tunnelusecase.RunInput{
				{
					TunnelName:  s.tunnel.Name,
					ConfigPath:  config.ConfigPath,
					ContentHash: config.ContentHash,
				},
			}, nil
		},
		tunnelusecase.ConfigHandlerName,
	)

	// dns-record — depends on tunnel-create (needs TunnelUUID)
	framework.Register(orchestrator, dnsusecase.NewHandler(s.dnsService),
		func(reg *framework.OutputRegistry) ([]dnsusecase.RecordInput, error) {
			create, ok := framework.GetOutput[tunnelusecase.CreateOutput](reg, tunnelusecase.CreateHandlerName, s.tunnel.Name)
			if !ok {
				return nil, fmt.Errorf("%s: missing upstream output from %s", dnsusecase.HandlerName, tunnelusecase.CreateHandlerName)
			}

			var inputs []dnsusecase.RecordInput

			switch s.ingress.Mode {
			case domain.IngressModeWildcard:
				inputs = []dnsusecase.RecordInput{
					{
						Zone:       s.ingress.Zone,
						Subdomain:  "*",
						TunnelName: s.tunnel.Name,
						TunnelUUID: create.TunnelUUID,
						Persistent: s.tunnel.Persistent,
					},
				}
			case domain.IngressModeSubdomain:
				inputs = make([]dnsusecase.RecordInput, 0, len(s.ingress.Apps))
				for _, app := range s.ingress.Apps {
					inputs = append(inputs, dnsusecase.RecordInput{
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
		tunnelusecase.CreateHandlerName,
	)

	// access-policies — no dependencies, account-level endpoint
	if s.policyService != nil && s.access.HasPolicies() {
		framework.Register(orchestrator, accessusecase.NewPolicyHandler(s.policyService),
			func(reg *framework.OutputRegistry) ([]accessusecase.PolicyInput, error) {
				inputs := make([]accessusecase.PolicyInput, len(s.access.Policies))
				for i, p := range s.access.Policies {
					inputs[i] = accessusecase.PolicyInput{Policy: p}
				}
				return inputs, nil
			},
		)
	}

	// access-app — depends on dns-record (ordering) and access-policies (policy IDs)
	if s.accessService != nil && s.ingress.HasAccessConfig() {
		deps := []string{dnsusecase.HandlerName}
		if s.policyService != nil && s.access.HasPolicies() {
			deps = append(deps, accessusecase.PolicyHandlerName)
		}
		framework.Register(orchestrator, accessusecase.NewHandler(s.accessService),
			func(reg *framework.OutputRegistry) ([]accessusecase.AppInput, error) {
				policyIDByName := make(map[string]string, len(s.access.Policies))
				for _, p := range s.access.Policies {
					out, ok := framework.GetOutput[accessusecase.PolicyOutput](reg, accessusecase.PolicyHandlerName, p.Name)
					if !ok {
						return nil, fmt.Errorf("%s: missing output for policy %q", accessusecase.HandlerName, p.Name)
					}
					policyIDByName[p.Name] = out.PolicyID
				}

				var inputs []accessusecase.AppInput
				for _, app := range s.ingress.Apps {
					if app.Access == nil {
						continue
					}
					policyIDs := make([]string, 0, len(app.Policies))
					for _, name := range app.Policies {
						id, ok := policyIDByName[name]
						if !ok {
							return nil, fmt.Errorf("policy %q referenced in app %q is not defined in access.policies", name, app.Expose.Subdomain)
						}
						policyIDs = append(policyIDs, id)
					}
					inputs = append(inputs, accessusecase.AppInput{
						Zone:      s.ingress.Zone,
						Subdomain: app.Expose.Subdomain,
						Access:    *app.Access,
						PolicyIDs: policyIDs,
					})
				}
				return inputs, nil
			},
			deps...,
		)
	}

	return orchestrator, nil
}
