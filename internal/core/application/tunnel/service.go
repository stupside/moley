// Package tunnel provides tunnel orchestration functionality.
package tunnel

import (
	"context"

	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/core/ports"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/shared"
)

type Service struct {
	tunnel        *domain.Tunnel
	ingress       *domain.Ingress
	dnsService    ports.DNSService
	tunnelService ports.TunnelService
}

var _ shared.Runnable = (*Service)(nil)

func NewService(tunnel *domain.Tunnel, ingress *domain.Ingress, dnsService ports.DNSService, tunnelService ports.TunnelService) *Service {
	return &Service{
		tunnel:        tunnel,
		ingress:       ingress,
		dnsService:    dnsService,
		tunnelService: tunnelService,
	}
}

func (s *Service) Start(ctx context.Context) error {
	logger.Infof("Starting tunnel service", map[string]any{
		"zone":     s.ingress.Zone,
		"tunnelID": s.tunnel.ID,
	})

	rm, err := s.createResourceManager(ctx)
	if err != nil {
		return shared.WrapError(err, "failed to create resource manager")
	}

	if err := rm.Start(ctx); err != nil {
		return shared.WrapError(err, "failed to start resources")
	}

	logger.Info("Tunnel service started")

	return ctx.Err()
}

func (s *Service) Stop(ctx context.Context) error {
	logger.Infof("Stopping tunnel service", map[string]any{
		"zone":     s.ingress.Zone,
		"tunnelID": s.tunnel.ID,
	})

	rm, err := s.createResourceManager(ctx)
	if err != nil {
		return shared.WrapError(err, "failed to create resource manager")
	}

	if err := rm.Stop(ctx); err != nil {
		return shared.WrapError(err, "failed to stop resources")
	}

	logger.Info("Tunnel service stopped")
	return nil
}
