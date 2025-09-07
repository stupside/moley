package tunnel

import (
	"context"

	"github.com/stupside/moley/internal/platform/framework"
	"github.com/stupside/moley/internal/shared"
)

func (s *Service) createResourceManager(ctx context.Context) (*framework.ResourceManager, error) {
	dnsHandler := &DNSRecordHandler{dnsService: s.dnsService}
	tunnelCreateHandler := &TunnelCreateHandler{tunnelService: s.tunnelService}
	tunnelConfigHandler := &TunnelConfigHandler{tunnelService: s.tunnelService}

	type spec struct {
		payload any
		handler framework.ResourceHandler
	}

	specs := []spec{
		{handler: tunnelCreateHandler, payload: &TunnelCreatePayload{Tunnel: s.tunnel}},
		{handler: tunnelConfigHandler, payload: &TunnelConfigPayload{
			Tunnel:  s.tunnel,
			Ingress: s.ingress,
		}},
	}

	for _, app := range s.ingress.Apps {
		specs = append(specs, spec{
			handler: dnsHandler,
			payload: &DNSRecordPayload{
				Tunnel:    s.tunnel,
				Zone:      s.ingress.Zone,
				Subdomain: app.Expose.Subdomain,
			},
		})
	}

	handlers := make(map[string]framework.ResourceHandler, len(specs))
	resources := make([]framework.Resource, 0, len(specs))

	for _, spt := range specs {
		r, err := framework.NewResource(ctx, spt.handler, spt.payload)
		if err != nil {
			return nil, shared.WrapError(err, "failed to create resource")
		}
		handlers[spt.handler.Name(ctx)] = spt.handler
		resources = append(resources, *r)
	}

	return framework.NewResourceManager(handlers, resources)
}
