package cf

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"moley/internal/domain"
	"moley/internal/errors"
	"moley/internal/logger"
	"moley/internal/services"
	"os"
	"path/filepath"
)

// tunnelService implements the TunnelService interface for Cloudflare
type tunnelService struct {
	services.TunnelService
}

// NewTunnelService creates a new Cloudflare Tunnel service
func NewTunnelService() *tunnelService {
	return &tunnelService{}
}

// GetID retrieves the tunnel ID for a given tunnel
func (c *tunnelService) GetID(ctx context.Context, domainTunnel *domain.Tunnel) (string, error) {
	// Get the tunnel token using cloudflared
	token, err := c.GetToken(ctx, domainTunnel)
	if err != nil {
		return "", errors.NewExecutionError(errors.ErrCodeCommandFailed, "failed to get tunnel token", err)
	}

	return token.TunnelId, nil
}

// GetZoneID retrieves the zone ID for a given tunnel
func (c *tunnelService) GetAccountID(ctx context.Context, domainTunnel *domain.Tunnel) (string, error) {
	token, err := c.GetToken(ctx, domainTunnel)
	if err != nil {
		return "", errors.NewExecutionError(errors.ErrCodeCommandFailed, "failed to get tunnel token", err)
	}
	return token.AccountId, nil
}

// GetToken retrieves the tunnel token, which contains the zone ID and tunnel ID
func (c *tunnelService) GetToken(ctx context.Context, domainTunnel *domain.Tunnel) (*struct {
	TunnelId  string
	AccountId string
}, error) {
	// Execute the cloudflared command to get the tunnel token
	output, err := execCloudflared(ctx, "tunnel", "token", domainTunnel.GetName())
	if err != nil {
		return nil, errors.NewExecutionError(errors.ErrCodeCommandFailed, "failed to get tunnel token", err)
	}

	// Decode the base64-encoded JSON output
	rawJson, err := base64.StdEncoding.DecodeString(output)
	if err != nil {
		return nil, errors.NewExecutionError(errors.ErrCodeCommandFailed, "failed to decode tunnel token", err)
	}

	// Unmarshal the JSON into a struct
	var info struct {
		TunnelId  string `json:"t"`
		AccountId string `json:"a"`
	}
	if err := json.Unmarshal(rawJson, &info); err != nil {
		return nil, errors.NewExecutionError(errors.ErrCodeCommandFailed, "failed to parse tunnel token JSON", err)
	}

	return &struct {
		TunnelId  string
		AccountId string
	}{
		TunnelId:  info.TunnelId,
		AccountId: info.AccountId,
	}, nil
}

// CreateTunnel creates a tunnel with the specified name
func (c *tunnelService) CreateTunnel(ctx context.Context, domainTunnel *domain.Tunnel) (string, error) {
	logger.Debugf("Creating Cloudflare tunnel", map[string]interface{}{
		"tunnel": domainTunnel.GetName(),
	})
	_, err := execCloudflared(ctx, "tunnel", "create", domainTunnel.GetName())
	if err != nil {
		logger.Errorf("Cloudflare tunnel creation failed", map[string]interface{}{
			"tunnel": domainTunnel.GetName(),
			"error":  err.Error(),
		})
		return "", errors.NewExecutionError(errors.ErrCodeCommandFailed, "failed to create tunnel", err)
	}
	tokenInfo, err := c.GetToken(ctx, domainTunnel)
	if err != nil {
		logger.Errorf("Failed to get tunnel token after creation", map[string]interface{}{
			"tunnel": domainTunnel.GetName(),
			"error":  err.Error(),
		})
		return "", errors.NewExecutionError(errors.ErrCodeCommandFailed, "failed to get tunnel token", err)
	}
	logger.Infof("Cloudflare tunnel created successfully", map[string]interface{}{
		"tunnel":    domainTunnel.GetName(),
		"tunnel_id": tokenInfo.TunnelId,
	})
	return tokenInfo.TunnelId, nil
}

// DeleteTunnel deletes a tunnel and cleans up its resources
func (c *tunnelService) DeleteTunnel(ctx context.Context, domainTunnel *domain.Tunnel) error {
	logger.Debugf("Deleting Cloudflare tunnel", map[string]interface{}{
		"tunnel": domainTunnel.GetName(),
	})
	if _, err := execCloudflared(ctx, "tunnel", "cleanup", domainTunnel.GetName()); err != nil {
		logger.Warnf("Tunnel cleanup before deletion failed", map[string]interface{}{
			"tunnel": domainTunnel.GetName(),
			"error":  err.Error(),
		})
		return errors.NewExecutionError(errors.ErrCodeCommandFailed, "failed to cleanup tunnel", err)
	}
	if _, err := execCloudflared(ctx, "tunnel", "delete", domainTunnel.GetName()); err != nil {
		logger.Errorf("Cloudflare tunnel deletion failed", map[string]interface{}{
			"tunnel": domainTunnel.GetName(),
			"error":  err.Error(),
		})
		return errors.NewExecutionError(errors.ErrCodeCommandFailed, "failed to delete tunnel", err)
	}
	logger.Infof("Cloudflare tunnel deleted successfully", map[string]interface{}{
		"tunnel": domainTunnel.GetName(),
	})
	return nil
}

// GetCredentialsPath returns the path to the tunnel credentials file
func (c *tunnelService) GetCredentialsPath(ctx context.Context, domainTunnel *domain.Tunnel) (string, error) {
	logger.Debugf("Getting credentials path for tunnel", map[string]interface{}{
		"tunnel": domainTunnel.GetName(),
	})
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Errorf("Failed to get user home directory for credentials path", map[string]interface{}{
			"tunnel": domainTunnel.GetName(),
			"error":  err.Error(),
		})
		return "", errors.NewConfigError(errors.ErrCodePermissionDenied, "failed to get user home directory", err)
	}
	tunnelID, err := c.GetID(ctx, domainTunnel)
	if err != nil {
		logger.Errorf("Failed to get tunnel ID for credentials path", map[string]interface{}{
			"tunnel": domainTunnel.GetName(),
			"error":  err.Error(),
		})
		return "", errors.NewConfigError(errors.ErrCodeInvalidConfig, "failed to get tunnel ID", err)
	}
	credentialsFile := filepath.Join(homeDir, fmt.Sprintf(".cloudflared/%s.json", tunnelID))
	if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
		logger.Warnf("Tunnel credentials file does not exist", map[string]interface{}{
			"tunnel": domainTunnel.GetName(),
			"file":   credentialsFile,
		})
		return "", errors.NewConfigError(errors.ErrCodeInvalidConfig, "credentials file does not exist", err)
	} else if err != nil {
		logger.Errorf("Failed to access tunnel credentials file", map[string]interface{}{
			"tunnel": domainTunnel.GetName(),
			"file":   credentialsFile,
			"error":  err.Error(),
		})
		return "", errors.NewConfigError(errors.ErrCodePermissionDenied, "failed to access credentials file", err)
	}
	logger.Debugf("Tunnel credentials file found", map[string]interface{}{
		"tunnel": domainTunnel.GetName(),
		"file":   credentialsFile,
	})
	return credentialsFile, nil
}
