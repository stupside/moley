// Package session orchestrates the full lifecycle of a tunnel session:
// tunnel infrastructure, DNS routing, and Cloudflare Access protection.
package session

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/domain"
	accessusecase "github.com/stupside/moley/v2/internal/features/access/usecase"
	dnsusecase "github.com/stupside/moley/v2/internal/features/dns/usecase"
	tunnelusecase "github.com/stupside/moley/v2/internal/features/tunnel/usecase"
	logger "github.com/stupside/moley/v2/internal/platform/logging"
	shared "github.com/stupside/moley/v2/internal/platform/runtime"
)

type Service struct {
	tunnel             *domain.Tunnel
	ingress            *domain.Ingress
	access             *domain.Access
	dnsService         dnsusecase.DNSRouter
	tunnelCreator      tunnelusecase.TunnelCreator
	tunnelConfigurator tunnelusecase.TunnelConfigurator
	tunnelRunner       tunnelusecase.TunnelRunner
	accessService      accessusecase.AccessManager
	policyService      accessusecase.PolicyManager
}

var _ shared.Runnable = (*Service)(nil)

func NewService(
	tunnel *domain.Tunnel,
	ingress *domain.Ingress,
	access *domain.Access,
	dnsService dnsusecase.DNSRouter,
	tunnelCreator tunnelusecase.TunnelCreator,
	tunnelConfigurator tunnelusecase.TunnelConfigurator,
	tunnelRunner tunnelusecase.TunnelRunner,
	accessService accessusecase.AccessManager,
	policyService accessusecase.PolicyManager,
) *Service {
	return &Service{
		tunnel:             tunnel,
		ingress:            ingress,
		access:             access,
		dnsService:         dnsService,
		tunnelCreator:      tunnelCreator,
		tunnelConfigurator: tunnelConfigurator,
		tunnelRunner:       tunnelRunner,
		accessService:      accessService,
		policyService:      policyService,
	}
}

func (s *Service) Start(ctx context.Context) error {
	logger.Infof("Starting tunnel service", map[string]any{
		"zone":   s.ingress.Zone,
		"tunnel": s.tunnel.Ref(),
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
		"tunnel": s.tunnel.Ref(),
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
