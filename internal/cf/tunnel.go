package cf

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/stupside/moley/internal/domain"
	"github.com/stupside/moley/internal/logger"
	"github.com/stupside/moley/internal/shared"
)

// tunnelService implements the TunnelService interface for Cloudflare
type tunnelService struct {
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
		return "", shared.WrapError(err, "failed to get tunnel token")
	}

	return token.TunnelId, nil
}

// GetZoneID retrieves the zone ID for a given tunnel
func (c *tunnelService) GetAccountID(ctx context.Context, domainTunnel *domain.Tunnel) (string, error) {
	token, err := c.GetToken(ctx, domainTunnel)
	if err != nil {
		return "", shared.WrapError(err, "failed to get tunnel token")
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
		return nil, shared.WrapError(err, "failed to get tunnel token")
	}

	// Decode the base64-encoded JSON output
	rawJson, err := base64.StdEncoding.DecodeString(output)
	if err != nil {
		return nil, shared.WrapError(err, "failed to decode tunnel token")
	}

	// Unmarshal the JSON into a struct
	var info struct {
		TunnelId  string `json:"t"`
		AccountId string `json:"a"`
	}
	if err := json.Unmarshal(rawJson, &info); err != nil {
		return nil, shared.WrapError(err, "failed to parse tunnel token JSON")
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
	logger.Debug(fmt.Sprintf("Creating Cloudflare tunnel: %s", domainTunnel.GetName()))

	_, err := execCloudflared(ctx, "tunnel", "create", domainTunnel.GetName())
	if err != nil {
		return "", shared.WrapError(err, "failed to create tunnel")
	}

	tokenInfo, err := c.GetToken(ctx, domainTunnel)
	if err != nil {
		return "", shared.WrapError(err, "failed to get tunnel token after creation")
	}

	logger.Debug(fmt.Sprintf("Cloudflare tunnel created with ID: %s", tokenInfo.TunnelId))
	return tokenInfo.TunnelId, nil
}

// DeleteTunnel deletes a tunnel and cleans up its resources
func (c *tunnelService) DeleteTunnel(ctx context.Context, domainTunnel *domain.Tunnel) error {
	logger.Debug(fmt.Sprintf("Deleting Cloudflare tunnel: %s", domainTunnel.GetName()))

	if _, err := execCloudflared(ctx, "tunnel", "cleanup", domainTunnel.GetName()); err != nil {
		logger.Warn(fmt.Sprintf("Tunnel cleanup before deletion failed: %s", err.Error()))
		return shared.WrapError(err, "failed to cleanup tunnel")
	}

	if _, err := execCloudflared(ctx, "tunnel", "delete", domainTunnel.GetName()); err != nil {
		return shared.WrapError(err, "failed to delete tunnel")
	}

	logger.Debug("Cloudflare tunnel deleted successfully")
	return nil
}

// GetCredentialsPath returns the path to the tunnel credentials file
func (c *tunnelService) GetCredentialsPath(ctx context.Context, domainTunnel *domain.Tunnel) (string, error) {
	logger.Debug(fmt.Sprintf("Getting credentials path for tunnel: %s", domainTunnel.GetName()))

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", shared.WrapError(err, "failed to get user home directory")
	}

	tunnelID, err := c.GetID(ctx, domainTunnel)
	if err != nil {
		return "", shared.WrapError(err, "failed to get tunnel ID")
	}

	credentialsFile := filepath.Join(homeDir, fmt.Sprintf(".cloudflared/%s.json", tunnelID))
	if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
		return "", shared.WrapError(err, "credentials file does not exist")
	} else if err != nil {
		return "", shared.WrapError(err, "failed to access credentials file")
	}

	logger.Debug(fmt.Sprintf("Tunnel credentials file found: %s", credentialsFile))
	return credentialsFile, nil
}
