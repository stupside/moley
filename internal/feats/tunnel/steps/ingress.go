package steps

import (
	"context"
	"os"

	"github.com/stupside/moley/internal/domain"
	"github.com/stupside/moley/internal/errors"
	"github.com/stupside/moley/internal/logger"
	"github.com/stupside/moley/internal/services"
)

// IngressStep handles Cloudflare tunnel ingress configuration file operations
type IngressStep struct {
	// dns holds the DNS configuration for the tunnel
	dns *domain.DNS
	// tunnel holds the tunnel information for which ingress is configured
	tunnel *domain.Tunnel
	// ingressService is the service responsible for ingress operations
	ingressService services.IngressService
}

// NewIngressStep creates a new ingress config step
func NewIngressStep(ingressService services.IngressService, tunnel *domain.Tunnel, dns *domain.DNS) *IngressStep {
	return &IngressStep{
		dns:            dns,
		tunnel:         tunnel,
		ingressService: ingressService,
	}
}

// Name returns the step name
func (i *IngressStep) Name() string {
	return "Ingress"
}

// Up generates and writes the Cloudflare tunnel ingress configuration file
func (i *IngressStep) Up(ctx context.Context) error {
	logger.Debugf("Generating ingress configuration file", map[string]interface{}{
		"tunnel": i.tunnel.GetName(),
	})
	configYAML, err := i.ingressService.GetConfiguration(ctx, i.tunnel, i.dns)
	if err != nil {
		logger.Errorf("Failed to generate ingress configuration", map[string]interface{}{
			"tunnel": i.tunnel.GetName(),
			"error":  err.Error(),
		})
		return errors.NewConfigError(errors.ErrCodeInvalidConfig, "failed to generate tunnel configuration", err)
	}
	configYAMLPath, err := i.ingressService.GetConfigurationPath(ctx, i.tunnel)
	if err != nil {
		logger.Errorf("Failed to determine ingress config file path", map[string]interface{}{
			"tunnel": i.tunnel.GetName(),
			"error":  err.Error(),
		})
		return errors.NewConfigError(errors.ErrCodeInvalidConfig, "failed to determine configuration file path", err)
	}
	if err := os.WriteFile(configYAMLPath, configYAML, 0644); err != nil {
		logger.Errorf("Failed to write ingress config file", map[string]interface{}{
			"tunnel": i.tunnel.GetName(),
			"file":   configYAMLPath,
			"error":  err.Error(),
		})
		return errors.NewConfigError(errors.ErrCodePermissionDenied, "failed to write tunnel configuration file", err)
	}
	logger.Infof("Ingress configuration file created successfully", map[string]interface{}{
		"tunnel": i.tunnel.GetName(),
		"file":   configYAMLPath,
	})
	return nil
}

// Down removes the Cloudflare tunnel configuration file
func (i *IngressStep) Down(ctx context.Context) error {
	logger.Debugf("Removing ingress configuration file", map[string]interface{}{
		"tunnel": i.tunnel.GetName(),
	})
	configYAMLPath, err := i.ingressService.GetConfigurationPath(ctx, i.tunnel)
	if err != nil {
		logger.Errorf("Failed to determine ingress config file path for removal", map[string]interface{}{
			"tunnel": i.tunnel.GetName(),
			"error":  err.Error(),
		})
		return errors.NewConfigError(errors.ErrCodeInvalidConfig, "failed to determine configuration file path", err)
	}
	if err := os.Remove(configYAMLPath); err != nil {
		logger.Warnf("Failed to remove ingress config file", map[string]interface{}{
			"tunnel": i.tunnel.GetName(),
			"file":   configYAMLPath,
			"error":  err.Error(),
		})
		return errors.NewConfigError(errors.ErrCodePermissionDenied, "failed to remove config file", err)
	}
	logger.Infof("Ingress configuration file removed successfully", map[string]interface{}{
		"tunnel": i.tunnel.GetName(),
		"file":   configYAMLPath,
	})
	return nil
}
