// Package application orchestrates the full lifecycle of a tunnel session:
// tunnel infrastructure, DNS routing, and Cloudflare Access protection.
package application

import (
	"context"
	"fmt"

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
	accessService ports.AccessService
}

var _ shared.Runnable = (*Service)(nil)

func NewService(tunnel *domain.Tunnel, ingress *domain.Ingress, dnsService ports.DNSService, tunnelService ports.TunnelService, accessService ports.AccessService) *Service {
	return &Service{
		tunnel:        tunnel,
		ingress:       ingress,
		dnsService:    dnsService,
		tunnelService: tunnelService,
		accessService: accessService,
	}
}

func (s *Service) Start(ctx context.Context) error {
	logger.Infof("Starting tunnel service", map[string]any{
		"zone":   s.ingress.Zone,
		"tunnel": s.tunnel.Name,
	})

	orch, err := s.createOrchestrator(ctx)
	if err != nil {
		return fmt.Errorf("failed to create orchestrator: %w", err)
	}

	if err := orch.Start(ctx); err != nil {
		return fmt.Errorf("failed to start resources: %w", err)
	}

	logger.Info("Tunnel service started")
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	logger.Infof("Stopping tunnel service", map[string]any{
		"zone":   s.ingress.Zone,
		"tunnel": s.tunnel.Name,
	})

	orch, err := s.createOrchestrator(ctx)
	if err != nil {
		return fmt.Errorf("failed to create orchestrator: %w", err)
	}

	if err := orch.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop resources: %w", err)
	}

	logger.Info("Tunnel service stopped")
	return nil
}
