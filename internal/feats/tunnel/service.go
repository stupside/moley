package tunnel

import (
	"context"
	"fmt"
	"moley/internal/cf"
	"moley/internal/config"
	"moley/internal/domain"
	"moley/internal/errors"
	"moley/internal/feats/tunnel/steps"
	"moley/internal/logger"
	"moley/internal/services"
	"os"
	"os/exec"
	"strings"
)

// Service orchestrates tunnel deployment and cleanup
type Service struct {
	// steps holds the ordered steps to deploy and clean up the tunnel
	steps []steps.Step
	// tunnel holds the tunnel information
	tunnel *domain.Tunnel
	// ingress holds the DNS configuration for the tunnel
	ingress services.IngressGetter
}

// NewService creates a new tunnel manager
func NewService(moleyConfig *config.MoleyConfig, config *TunnelConfig, tunnelName string) (*Service, error) {

	// Create DNS configuration
	dns := domain.NewDNS(config.Zone, config.Apps)
	// Create tunnel using the provided name
	tunnel, err := domain.NewTunnel(tunnelName)
	if err != nil {
		return nil, errors.NewConfigError(errors.ErrCodeInvalidConfig, "failed to create tunnel", err)
	}

	// Create tunnel service
	tunnelService := cf.NewTunnelService()
	// Create ingress service
	ingressService := cf.NewIngressService(tunnelService)

	// Create dns service
	dnsService, err := cf.NewDNSService(moleyConfig.Cloudflare.Token, tunnelService)
	if err != nil {
		return nil, errors.NewConfigError(errors.ErrCodeInvalidConfig, "failed to create Cloudflare client", err)
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
		// Initialize steps in the order they should be executed
		// 1. Create the tunnel
		// 2. Create DNS records for the tunnel
		// 3. Generate the ingress configuration file
		steps: []steps.Step{tunnelStep, dnsStep, ingressStep},
	}, nil
}

// Deploy deploys the tunnel without running cloudflared
func (m *Service) Deploy(ctx context.Context) error {
	logger.Debugf("Deploying tunnel", map[string]interface{}{
		"tunnel": m.tunnel.GetName(),
	})

	steps := make([]steps.Step, 0)
	for _, step := range m.steps {
		logger.Debugf("Running step", map[string]interface{}{
			"step":   step.Name(),
			"tunnel": m.tunnel.GetName(),
		})
		// Execute the step
		if err := step.Up(ctx); err != nil {
			// If the step fails, revert all previous steps
			m.Revert(ctx, steps)
			return errors.NewExecutionError(errors.ErrCodeCommandFailed, "step failed during deployment", err)
		}
		logger.Debugf("Step completed successfully", map[string]interface{}{
			"step":   step.Name(),
			"tunnel": m.tunnel.GetName(),
		})
		// Add the step to the list of completed steps
		steps = append(steps, step)
	}
	logger.Debugf("Tunnel deployed successfully", map[string]interface{}{
		"tunnel": m.tunnel.GetName(),
	})
	return nil
}

// Run runs the cloudflared tunnel
func (m *Service) Run(ctx context.Context) error {
	logger.Infof("Starting tunnel service", map[string]interface{}{
		"tunnel": m.tunnel.GetName(),
	})
	configPath, err := m.ingress.GetConfigurationPath(ctx, m.tunnel)
	if err != nil {
		logger.Errorf("Failed to get configuration path", map[string]interface{}{
			"tunnel": m.tunnel.GetName(),
			"error":  err.Error(),
		})
		return errors.NewConfigError(errors.ErrCodeInvalidConfig, "failed to determine configuration file path", err)
	}
	cmd := exec.CommandContext(ctx, "cloudflared", "tunnel", "--config", configPath, "run", m.tunnel.GetName())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logger.Errorf("cloudflared tunnel process failed", map[string]interface{}{
			"tunnel": m.tunnel.GetName(),
			"error":  err.Error(),
		})
		return errors.NewExecutionError(errors.ErrCodeCommandFailed, "cloudflared tunnel failed", err)
	}
	logger.Infof("Tunnel service exited normally", map[string]interface{}{
		"tunnel": m.tunnel.GetName(),
	})
	return nil
}

// Revert tunnel resources
func (m *Service) Revert(ctx context.Context, steps []steps.Step) error {
	logger.Infof("Reverting tunnel resources", map[string]interface{}{
		"tunnel": m.tunnel.GetName(),
	})
	var cleanupErrors []string
	for i := len(steps) - 1; i >= 0; i-- {
		step := steps[i]
		logger.Debugf("Cleaning up step", map[string]interface{}{
			"step":   step.Name(),
			"tunnel": m.tunnel.GetName(),
		})
		if err := step.Down(ctx); err != nil {
			logger.Warnf("Step cleanup failed", map[string]interface{}{
				"step":   step.Name(),
				"tunnel": m.tunnel.GetName(),
				"error":  err.Error(),
			})
			cleanupErrors = append(cleanupErrors, step.Name()+": "+err.Error())
		}
	}
	if len(cleanupErrors) > 0 {
		logger.Warnf("Cleanup completed with errors", map[string]interface{}{
			"tunnel": m.tunnel.GetName(),
			"errors": cleanupErrors,
		})
		return errors.NewExecutionError(errors.ErrCodeCommandFailed, "cleanup completed with errors", fmt.Errorf("%s", strings.Join(cleanupErrors, "; ")))
	}
	logger.Infof("Tunnel resources cleaned up successfully", map[string]interface{}{
		"tunnel": m.tunnel.GetName(),
	})
	return nil
}

// Cleanup cleans up the tunnel resources
func (m *Service) Cleanup(ctx context.Context) error {
	logger.Infof("Cleaning up tunnel resources", map[string]interface{}{
		"tunnel": m.tunnel.GetName(),
	})
	return m.Revert(ctx, m.steps)
}
