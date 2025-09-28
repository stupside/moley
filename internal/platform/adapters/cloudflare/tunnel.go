// Package cloudflare provides Cloudflare-specific implementations.
package cloudflare

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/platform/framework"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/paths"
	"github.com/stupside/moley/v2/internal/shared"
	"go.yaml.in/yaml/v3"
)

type tunnelService struct {
	config *framework.Config
}

func NewTunnelService(config *framework.Config) *tunnelService {
	return &tunnelService{
		config: config,
	}
}

func (c *tunnelService) Run(ctx context.Context, tunnel *domain.Tunnel) (int, error) {
	configPath, err := c.GetConfigurationPath(ctx, tunnel)
	if err != nil {
		return 0, shared.WrapError(err, "failed to get tunnel configuration path")
	}

	cfCommand := NewCommand(ctx, "tunnel", "--config", configPath, "run", tunnel.GetName())
	if pid, err := cfCommand.ExecAsync(); err != nil {
		return 0, shared.WrapError(err, "failed to run tunnel")
	} else {
		return pid, nil
	}
}

func (c *tunnelService) GetID(ctx context.Context, tunnel *domain.Tunnel) (string, error) {
	token, err := c.GetToken(ctx, tunnel)
	if err != nil {
		return "", shared.WrapError(err, "failed to get tunnel token")
	}

	return token.TunnelID, nil
}

func (c *tunnelService) GetAccountID(ctx context.Context, tunnel *domain.Tunnel) (string, error) {
	token, err := c.GetToken(ctx, tunnel)
	if err != nil {
		return "", shared.WrapError(err, "failed to get tunnel token")
	}
	return token.AccountID, nil
}

func (c *tunnelService) GetToken(ctx context.Context, tunnel *domain.Tunnel) (*struct {
	TunnelID  string
	AccountID string
}, error) {
	output, err := framework.RunWithDryRunGuard(c.config, func() (string, error) {
		cfCommand := NewCommand(ctx, "tunnel", "token", tunnel.GetName())
		out, err := cfCommand.ExecSync()
		if err != nil {
			return "", shared.WrapError(err, "failed to get tunnel token")
		}
		return out, nil
	}, base64.StdEncoding.EncodeToString(fmt.Appendf([]byte{}, `{"t":"%s","a":"%s"}`, tunnel.GetName(), tunnel.GetName())))
	if err != nil {
		return nil, shared.WrapError(err, "failed to get tunnel token")
	}

	rawJson, err := base64.StdEncoding.DecodeString(output)
	if err != nil {
		return nil, shared.WrapError(err, "failed to decode tunnel token")
	}

	var info struct {
		TunnelID  string `json:"t"`
		AccountID string `json:"a"`
	}
	if err := json.Unmarshal(rawJson, &info); err != nil {
		return nil, shared.WrapError(err, "failed to parse tunnel token JSON")
	}

	return &struct {
		TunnelID  string
		AccountID string
	}{
		TunnelID:  info.TunnelID,
		AccountID: info.AccountID,
	}, nil
}

func (c *tunnelService) CreateTunnel(ctx context.Context, tunnel *domain.Tunnel) (string, error) {
	logger.Info("Creating Cloudflare tunnel")

	if _, err := framework.RunWithDryRunGuard(c.config, func() (string, error) {
		cfCommand := NewCommand(ctx, "tunnel", "create", tunnel.GetName())
		out, err := cfCommand.ExecSync()
		if err != nil {
			return "", shared.WrapError(err, "failed to create tunnel")
		}
		return out, nil
	}, ""); err != nil {
		return "", shared.WrapError(err, "failed to create tunnel")
	}

	tokenInfo, err := c.GetToken(ctx, tunnel)
	if err != nil {
		return "", shared.WrapError(err, "failed to get tunnel token after creation")
	}

	logger.Debugf("Tunnel created successfully", map[string]any{
		"name":     tunnel.GetName(),
		"tunnelID": tokenInfo.TunnelID,
		"account":  tokenInfo.AccountID,
	})
	return tokenInfo.TunnelID, nil
}

func (c *tunnelService) DeleteTunnel(ctx context.Context, tunnel *domain.Tunnel) error {
	logger.Info("Deleting Cloudflare tunnel")

	if _, err := framework.RunWithDryRunGuard(c.config, func() (string, error) {
		cfCommand := NewCommand(ctx, "tunnel", "cleanup", tunnel.GetName())
		out, err := cfCommand.ExecSync()
		if err != nil {
			return "", shared.WrapError(err, "failed to cleanup tunnel")
		}
		return out, nil
	}, ""); err != nil {
		logger.Warn(fmt.Sprintf("Tunnel cleanup before deletion failed: %s", err.Error()))
		return shared.WrapError(err, "failed to cleanup tunnel")
	}

	if _, err := framework.RunWithDryRunGuard(c.config, func() (string, error) {
		cfCommand := NewCommand(ctx, "tunnel", "delete", tunnel.GetName())
		out, err := cfCommand.ExecSync()
		if err != nil {
			return "", shared.WrapError(err, "failed to delete tunnel")
		}
		return out, nil
	}, ""); err != nil {
		return shared.WrapError(err, "failed to delete tunnel")
	}

	logger.Debug("Cloudflare tunnel deleted successfully")
	return nil
}

func (c *tunnelService) GetCredentialsPath(ctx context.Context, tunnel *domain.Tunnel) (string, error) {
	logger.Debug("Getting tunnel credentials file path")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", shared.WrapError(err, "failed to get user home directory")
	}

	tunnelID, err := c.GetID(ctx, tunnel)
	if err != nil {
		return "", shared.WrapError(err, "failed to get tunnel ID")
	}

	credentialsFile, _ := framework.RunWithDryRunGuard(c.config, func() (string, error) {
		return filepath.Join(homeDir, fmt.Sprintf(".cloudflared/%s.json", tunnelID)), nil
	}, "path-placeholder")
	if !(c.config != nil && c.config.DryRun) {
		if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
			return "", shared.WrapError(err, "credentials file does not exist")
		} else if err != nil {
			return "", shared.WrapError(err, "failed to access credentials file")
		}
	}

	logger.Debugf("Tunnel credentials file path", map[string]any{
		"path": credentialsFile,
	})
	return credentialsFile, nil
}

func (c *tunnelService) SaveConfiguration(ctx context.Context, tunnel *domain.Tunnel, ingress *domain.Ingress) error {
	logger.Info("Saving Cloudflare configuration")

	credentialsFile, err := c.GetCredentialsPath(ctx, tunnel)
	if err != nil {
		return shared.WrapError(err, "failed to get credentials file path")
	}

	logger.Debugf("Using credentials file", map[string]any{
		"path": credentialsFile,
	})

	config := &CloudflaredConfig{
		Tunnel:          tunnel.GetName(),
		Logfile:         "cloudflared.log",
		Ingress:         nil,
		Loglevel:        "info",
		CredentialsFile: credentialsFile,
	}

	logger.Infof("Building ingress rules", map[string]any{
		"apps": len(ingress.Apps),
		"zone": ingress.Zone,
	})
	for _, app := range ingress.Apps {
		config.Ingress = append(config.Ingress, CloudflaredIngressConfig{
			Service:  app.Target.GetTargetURL(),
			Hostname: fmt.Sprintf("%s.%s", app.Expose.Subdomain, ingress.Zone),
		})
	}

	logger.Info("Adding catch-all ingress rule")
	config.Ingress = append(config.Ingress, CloudflaredIngressConfig{
		Service:  "http_status:404",
		Hostname: "*",
	})

	bytes, err := yaml.Marshal(config)
	if err != nil {
		return shared.WrapError(err, "failed to marshal configuration")
	}

	path, err := c.GetConfigurationPath(ctx, tunnel)
	if err != nil {
		return shared.WrapError(err, "failed to get tunnel configuration path")
	}

	if !(c.config != nil && c.config.DryRun) {
		if err := os.WriteFile(path, bytes, 0600); err != nil {
			return shared.WrapError(err, "failed to save tunnel configuration")
		}
	}

	logger.Debug("Configuration saved")
	return nil
}

func (c *tunnelService) DeleteConfiguration(ctx context.Context, tunnel *domain.Tunnel) error {
	logger.Info("Deleting configuration")

	path, err := c.GetConfigurationPath(ctx, tunnel)
	if err != nil {
		return shared.WrapError(err, "failed to get tunnel configuration path")
	}

	if !(c.config != nil && c.config.DryRun) {
		if err := os.Remove(path); err != nil {
			return shared.WrapError(err, "failed to delete tunnel configuration")
		}
	}

	logger.Debug("Configuration deleted")
	return nil
}

func (c *tunnelService) TunnelExists(ctx context.Context, tunnel *domain.Tunnel) (bool, error) {
	logger.Debugf("Checking if tunnel exists", map[string]any{
		"tunnelName": tunnel.GetName(),
	})

	exists, err := framework.RunWithDryRunGuard(c.config, func() (bool, error) {
		cfCommand := NewCommand(ctx, "tunnel", "list", "--output", "json")
		output, err := cfCommand.ExecSync()
		if err != nil {
			return false, shared.WrapError(err, "failed to list tunnels")
		}

		// Handle case where cloudflared returns "null" for empty lists
		if output == "null" {
			logger.Debug("No tunnels found (null output)")
			return false, nil
		}

		var tunnels []struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal([]byte(output), &tunnels); err != nil {
			return false, shared.WrapError(err, "failed to unmarshal tunnel list")
		}

		for _, t := range tunnels {
			if t.Name == tunnel.GetName() {
				logger.Debugf("Tunnel found", map[string]any{
					"tunnelName": tunnel.GetName(),
				})
				return true, nil
			}
		}
		logger.Debugf("Tunnel not found in list", map[string]any{
			"tunnelName":  tunnel.GetName(),
			"tunnelCount": len(tunnels),
		})
		return false, nil
	}, false)
	if err != nil {
		return false, shared.WrapError(err, "failed to check if tunnel exists")
	}
	logger.Debugf("Tunnel exists check result", map[string]any{
		"tunnelName": tunnel.GetName(),
		"exists":     exists,
	})
	return exists, nil
}

func (c *tunnelService) GetConfigurationPath(ctx context.Context, tunnel *domain.Tunnel) (string, error) {
	logger.Debug("Getting tunnel configuration path")

	base, err := paths.GetUserFolderPath()
	if err != nil {
		return "", shared.WrapError(err, "failed to get user folder path")
	}

	tunnelsFolder := filepath.Join(base, "tunnels")
	if err := os.MkdirAll(tunnelsFolder, 0755); err != nil {
		return "", shared.WrapError(err, "failed to create tunnels directory")
	}

	tunnelFile := filepath.Join(tunnelsFolder, fmt.Sprintf("%s.yml", tunnel.GetName()))

	logger.Debugf("Tunnel configuration path", map[string]any{
		"path": tunnelFile,
	})

	return tunnelFile, nil
}
