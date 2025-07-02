package cf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/stupside/moley/internal/config"
	"github.com/stupside/moley/internal/domain"
	"github.com/stupside/moley/internal/errors"
	"github.com/stupside/moley/internal/logger"
	"github.com/stupside/moley/internal/services"

	"gopkg.in/yaml.v3"
)

// ingressService implements the IngressService interface for Cloudflare
type ingressService struct {
	services.IngressService

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
	logger.Debugf("Generating Cloudflare ingress configuration", map[string]interface{}{
		"tunnel": domainTunnel.GetName(),
	})
	credentialsFile, err := c.tunnelService.GetCredentialsPath(ctx, domainTunnel)
	if err != nil {
		logger.Errorf("Failed to get credentials path for ingress config", map[string]interface{}{
			"tunnel": domainTunnel.GetName(),
			"error":  err.Error(),
		})
		return nil, errors.NewConfigError(errors.ErrCodeInvalidConfig, "failed to determine credentials file path", err)
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
		logger.Errorf("Failed to marshal ingress configuration to YAML", map[string]interface{}{
			"tunnel": domainTunnel.GetName(),
			"error":  err.Error(),
		})
		return nil, errors.NewConfigError(errors.ErrCodeInvalidConfig, "failed to marshal Cloudflare configuration", err)
	}
	logger.Infof("Cloudflare ingress configuration generated successfully", map[string]interface{}{
		"tunnel": domainTunnel.GetName(),
	})
	return yamlData, nil
}

// GetConfigurationPath returns the path where the Cloudflare configuration file should be stored
func (c *ingressService) GetConfigurationPath(ctx context.Context, domainTunnel *domain.Tunnel) (string, error) {
	logger.Debugf("Getting ingress configuration file path", map[string]interface{}{
		"tunnel": domainTunnel.GetName(),
	})
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Errorf("Failed to get user home directory for ingress config path", map[string]interface{}{
			"tunnel": domainTunnel.GetName(),
			"error":  err.Error(),
		})
		return "", errors.NewConfigError(errors.ErrCodePermissionDenied, "failed to get user home directory", err)
	}
	cloudflaredFolder := filepath.Join(homeDir, config.ConfigFileFolder, "tunnels")
	if err := os.MkdirAll(cloudflaredFolder, 0755); err != nil {
		logger.Errorf("Failed to create directory for ingress config", map[string]interface{}{
			"tunnel": domainTunnel.GetName(),
			"dir":    cloudflaredFolder,
			"error":  err.Error(),
		})
		return "", errors.NewConfigError(errors.ErrCodePermissionDenied, "failed to create directory", err)
	}
	configPath := filepath.Join(cloudflaredFolder, fmt.Sprintf("%s.yml", domainTunnel.GetName()))
	logger.Debugf("Ingress configuration file path determined", map[string]interface{}{
		"tunnel": domainTunnel.GetName(),
		"file":   configPath,
	})
	return configPath, nil
}
