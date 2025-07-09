package steps

import (
	"context"
	"fmt"
	"os"

	"github.com/stupside/moley/internal/domain"
	"github.com/stupside/moley/internal/logger"
	"github.com/stupside/moley/internal/services"
	"github.com/stupside/moley/internal/shared"
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
	apps := i.dns.GetApps()
	zone := i.dns.GetZone()

	logger.Info(fmt.Sprintf("Generating ingress configuration for tunnel: %s", i.tunnel.GetName()))
	for _, app := range apps {
		hostname := fmt.Sprintf("%s.%s", app.Expose.Subdomain, zone)
		logger.Info(fmt.Sprintf("  → Configuring ingress: %s → %s", hostname, app.Target.GetTargetURL()))
	}

	configYAML, err := i.ingressService.GetConfiguration(ctx, i.tunnel, i.dns)
	if err != nil {
		return shared.WrapError(err, "failed to generate tunnel configuration")
	}

	configYAMLPath, err := i.ingressService.GetConfigurationPath(ctx, i.tunnel)
	if err != nil {
		return shared.WrapError(err, "failed to determine configuration file path")
	}

	if err := os.WriteFile(configYAMLPath, configYAML, 0644); err != nil {
		return shared.WrapError(err, "failed to write tunnel configuration file")
	}

	logger.Info(fmt.Sprintf("Ingress configuration file created: %s", configYAMLPath))
	return nil
}

// Down removes the Cloudflare tunnel configuration file
func (i *IngressStep) Down(ctx context.Context) error {
	logger.Debug(fmt.Sprintf("Removing ingress configuration for tunnel: %s", i.tunnel.GetName()))

	configYAMLPath, err := i.ingressService.GetConfigurationPath(ctx, i.tunnel)
	if err != nil {
		return shared.WrapError(err, "failed to determine configuration file path")
	}

	if err := os.Remove(configYAMLPath); err != nil {
		return shared.WrapError(err, "failed to remove config file")
	}

	logger.Debug(fmt.Sprintf("Ingress configuration file removed: %s", configYAMLPath))
	return nil
}
