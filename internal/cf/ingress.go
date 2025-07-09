package cf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/stupside/moley/internal/config"
	"github.com/stupside/moley/internal/domain"
	"github.com/stupside/moley/internal/logger"
	"github.com/stupside/moley/internal/services"
	"github.com/stupside/moley/internal/shared"

	"gopkg.in/yaml.v3"
)

// ingressService implements the IngressService interface for Cloudflare
type ingressService struct {
	tunnelService services.TunnelService
}

// NewIngressService creates a new Cloudflare Ingress service
func NewIngressService(tunnelService services.TunnelService) *ingressService {
	return &ingressService{
		tunnelService: tunnelService,
	}
}

// GetConfiguration generates the Cloudflare configuration for a tunnel and its associated DNS
func (c *ingressService) GetConfiguration(ctx context.Context, domainTunnel *domain.Tunnel, domainDNS *domain.DNS) ([]byte, error) {
	logger.Debug(fmt.Sprintf("Generating Cloudflare ingress configuration for tunnel: %s", domainTunnel.GetName()))

	credentialsFile, err := c.tunnelService.GetCredentialsPath(ctx, domainTunnel)
	if err != nil {
		return nil, shared.WrapError(err, "failed to determine credentials file path")
	}

	// Create configuration
	cloudflareConfig := struct {
		Tunnel          string `yaml:"tunnel"`
		CredentialsFile string `yaml:"credentials-file"`
		Ingress         []struct {
			Service  string `yaml:"service"`
			Hostname string `yaml:"hostname"`
		} `yaml:"ingress"`
		Logfile  string `yaml:"logfile,omitempty"`
		Loglevel string `yaml:"loglevel,omitempty"`
	}{
		Tunnel:          domainTunnel.GetName(),
		Logfile:         "cloudflared.log",
		Loglevel:        "info",
		CredentialsFile: credentialsFile,
		Ingress: make([]struct {
			Service  string `yaml:"service"`
			Hostname string `yaml:"hostname"`
		}, len(domainDNS.GetApps())),
	}

	// Build ingress rules
	for idx, app := range domainDNS.GetApps() {
		cloudflareConfig.Ingress[idx] = struct {
			Service  string `yaml:"service"`
			Hostname string `yaml:"hostname"`
		}{
			Service:  app.Target.GetTargetURL(),
			Hostname: fmt.Sprintf("%s.%s", app.Expose.Subdomain, domainDNS.GetZone()),
		}
	}

	// Add catch-all rule for unmatched requests
	cloudflareConfig.Ingress = append(cloudflareConfig.Ingress, struct {
		Service  string `yaml:"service"`
		Hostname string `yaml:"hostname"`
	}{
		Service:  "http_status:404",
		Hostname: "*",
	})

	// Marshal to YAML
	yamlData, err := yaml.Marshal(&cloudflareConfig)
	if err != nil {
		return nil, shared.WrapError(err, "failed to marshal Cloudflare configuration")
	}

	logger.Debug("Cloudflare ingress configuration generated successfully")
	return yamlData, nil
}

// GetConfigurationPath returns the path where the Cloudflare configuration file should be stored
func (c *ingressService) GetConfigurationPath(ctx context.Context, domainTunnel *domain.Tunnel) (string, error) {
	logger.Debug(fmt.Sprintf("Getting ingress configuration file path for tunnel: %s", domainTunnel.GetName()))

	globalConfigDir, err := config.GetGlobalFolderPath()
	if err != nil {
		return "", shared.WrapError(err, "failed to get global config directory")
	}

	tunnelsFolder := filepath.Join(globalConfigDir, "tunnels")
	if err := os.MkdirAll(tunnelsFolder, 0755); err != nil {
		return "", shared.WrapError(err, "failed to create tunnels directory")
	}

	tunnelFile := filepath.Join(tunnelsFolder, fmt.Sprintf("%s.yml", domainTunnel.GetName()))
	logger.Debug(fmt.Sprintf("Ingress configuration file path: %s", tunnelFile))
	return tunnelFile, nil
}
