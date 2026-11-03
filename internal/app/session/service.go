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
	framework "github.com/stupside/moley/v2/internal/platform/orchestration"
	shared "github.com/stupside/moley/v2/internal/platform/runtime"
)

// TunnelRuntime combines the lifecycle needed by the resource handler with
// session-level observation and cleanup of children not yet recorded on disk.
type TunnelRuntime interface {
	tunnelusecase.TunnelRunner
	Wait(context.Context, tunnelusecase.ProcessIdentity) error
	StopOwned(context.Context) error
}

type Service struct {
	// A failed lock acquisition gives this invocation no resources to clean up.
	// A fresh service used by standalone stop is allowed to acquire and clean up.
	startupBlocked     bool
	process            tunnelusecase.RunOutput
	tunnel             *domain.Tunnel
	ingress            *domain.Ingress
	access             *domain.Access
	dnsService         dnsusecase.DNSRouter
	tunnelCreator      tunnelusecase.TunnelCreator
	tunnelConfigurator tunnelusecase.TunnelConfigurator
	tunnelRunner       TunnelRuntime
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
	tunnelRunner TunnelRuntime,
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
	s.startupBlocked = true
	logger.Infof("Starting tunnel service", map[string]any{
		"zone":   s.ingress.Zone,
		"tunnel": s.tunnel.Ref(),
	})

	orch, err := s.createOrchestrator(ctx)
	if err != nil {
		return fmt.Errorf("failed to create orchestrator: %w", err)
	}

	s.startupBlocked = false

	if err := orch.Start(ctx); err != nil {
		return fmt.Errorf("failed to start resources: %w", err)
	}

	output, ok := framework.GetOutput[tunnelusecase.RunOutput](orch.Outputs(), tunnelusecase.RunHandlerName, s.tunnel.Ref())
	if !ok {
		return fmt.Errorf("missing tunnel process output after startup")
	}
	s.process = output

	logger.Info("Tunnel service started")
	return nil
}

func (s *Service) Wait(ctx context.Context) error {
	return s.tunnelRunner.Wait(ctx, s.process.Process)
}

func (s *Service) Stop(ctx context.Context) error {
	if s.startupBlocked {
		return nil
	}
	if err := s.tunnelRunner.StopOwned(ctx); err != nil {
		return err
	}
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
