package tunnel

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/stupside/moley/internal/cf"
	"github.com/stupside/moley/internal/config"
	"github.com/stupside/moley/internal/domain"
	"github.com/stupside/moley/internal/feats/tunnel/steps"
	"github.com/stupside/moley/internal/logger"
	"github.com/stupside/moley/internal/services"
	"github.com/stupside/moley/internal/shared"
)

// Service orchestrates tunnel deployment and cleanup
type Service struct {
	// steps holds the ordered steps to deploy and clean up the tunnel
	steps []steps.Step
	// tunnel holds the tunnel information
	tunnel *domain.Tunnel
	// ingress holds the DNS configuration for the tunnel
	ingress services.IngressGetter
	// config holds the tunnel configuration
	config *TunnelConfig
}

// NewService creates a new tunnel manager
func NewService(globalConfig *config.GlobalConfig, tunnelConfig *TunnelConfig, tunnelName string) (*Service, error) {

	// Create DNS configuration
	dns := domain.NewDNS(tunnelConfig.Zone, tunnelConfig.Apps)
	// Create tunnel using the provided name
	tunnel, err := domain.NewTunnel(tunnelName)
	if err != nil {
		return nil, shared.WrapError(err, "failed to create tunnel")
	}

	// Create tunnel service
	tunnelService := cf.NewTunnelService()
	// Create ingress service
	ingressService := cf.NewIngressService(tunnelService)

	// Create dns service
	dnsService, err := cf.NewDNSService(globalConfig.Cloudflare.Token, tunnelService)
	if err != nil {
		return nil, shared.WrapError(err, "failed to create Cloudflare client")
	}

	// Create steps
	tunnelStep := steps.NewTunnelStep(tunnelService, tunnel)
	// Create DNS step
	dnsStep := steps.NewDNSStep(dnsService, tunnel, dns)
	// Create ingress step
	ingressStep := steps.NewIngressStep(ingressService, tunnel, dns)

	return &Service{
		tunnel:  tunnel,
		ingress: ingressService,
		config:  tunnelConfig,
		// Initialize steps in the order they should be executed
		// 1. Create the tunnel
		// 2. Create DNS records for the tunnel
		// 3. Generate the ingress configuration file
		steps: []steps.Step{tunnelStep, dnsStep, ingressStep},
	}, nil
}

// Deploy deploys the tunnel without running cloudflared
func (m *Service) Deploy(ctx context.Context) error {
	logger.Info(fmt.Sprintf("Deploying tunnel: %s with %d app(s)", m.tunnel.GetName(), len(m.config.Apps)))

	steps := make([]steps.Step, 0)
	for _, step := range m.steps {

		// Execute the step
		if err := step.Up(ctx); err != nil {
			// If the step fails, revert all previous steps
			m.Revert(ctx, steps)
			return shared.WrapError(err, fmt.Sprintf("step %s failed during deployment", step.Name()))
		}

		logger.Debug(fmt.Sprintf("Step completed: %s", step.Name()))
		// Add the step to the list of completed steps
		steps = append(steps, step)
	}

	logger.Info(fmt.Sprintf("Tunnel %s deployed successfully", m.tunnel.GetName()))
	return nil
}

// Run runs the cloudflared tunnel
func (m *Service) Run(ctx context.Context) error {
	logger.Info(fmt.Sprintf("Starting tunnel service: %s", m.tunnel.GetName()))

	configPath, err := m.ingress.GetConfigurationPath(ctx, m.tunnel)
	if err != nil {
		return shared.WrapError(err, "failed to determine configuration file path")
	}

	// Use warn level for cleaner output and redirect to log file
	cmd := exec.CommandContext(ctx, "cloudflared", "tunnel",
		"--config", configPath,
		"--logfile", "cloudflared.log",
		"run", m.tunnel.GetName())

	logger.Debug("Executing cloudflared tunnel run command")
	if err := cmd.Run(); err != nil {
		return shared.WrapError(err, "cloudflared tunnel process failed")
	}

	logger.Info(fmt.Sprintf("Tunnel service exited normally: %s", m.tunnel.GetName()))
	return nil
}

// Revert tunnel resources
func (m *Service) Revert(ctx context.Context, steps []steps.Step) error {
	logger.Info(fmt.Sprintf("Reverting tunnel resources: %s", m.tunnel.GetName()))

	var cleanupErrors []string
	for i := len(steps) - 1; i >= 0; i-- {
		step := steps[i]
		logger.Debug(fmt.Sprintf("Cleaning up step: %s", step.Name()))

		if err := step.Down(ctx); err != nil {
			logger.Warn(fmt.Sprintf("Step cleanup failed: %s - %s", step.Name(), err.Error()))
			cleanupErrors = append(cleanupErrors, step.Name()+": "+err.Error())
		}
	}

	if len(cleanupErrors) > 0 {
		errMsg := strings.Join(cleanupErrors, "; ")
		logger.Warn(fmt.Sprintf("Cleanup completed with errors: %s", errMsg))
		return shared.WrapError(fmt.Errorf("%s", errMsg), "cleanup completed with errors")
	}

	logger.Info(fmt.Sprintf("Tunnel resources cleaned up successfully: %s", m.tunnel.GetName()))
	return nil
}

// Cleanup cleans up the tunnel resources
func (m *Service) Cleanup(ctx context.Context) error {
	logger.Info(fmt.Sprintf("Cleaning up tunnel resources: %s", m.tunnel.GetName()))
	return m.Revert(ctx, m.steps)
}
