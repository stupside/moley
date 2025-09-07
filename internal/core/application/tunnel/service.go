// Package tunnel provides tunnel orchestration functionality.
package tunnel

import (
	"context"

	"github.com/stupside/moley/internal/core/domain"
	"github.com/stupside/moley/internal/core/ports"
	"github.com/stupside/moley/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/internal/shared"
)

type Service struct {
	tunnel        *domain.Tunnel
	ingress       *domain.Ingress
	dnsService    ports.DNSService
	tunnelService ports.TunnelService
}

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
		"tunnelID": s.tunnel.ID,
		"zone":     s.ingress.Zone,
	})

	rm, err := s.createResourceManager(ctx)
	if err != nil {
		return shared.WrapError(err, "failed to create resource manager")
	}

	if err := rm.Start(ctx); err != nil {
		return shared.WrapError(err, "failed to start resources")
	}

	if err := s.tunnelService.Run(ctx, s.tunnel); err != nil {
		return shared.WrapError(err, "failed to run tunnel")
	}

	logger.Info("Tunnel service started")
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	logger.Infof("Stopping tunnel service", map[string]any{
		"tunnelID": s.tunnel.ID,
		"zone":     s.ingress.Zone,
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
