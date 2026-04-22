// Package cloudflare provides Cloudflare-specific implementations.
package cloudflare

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	cfgo "github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/zero_trust"
	"github.com/cloudflare/cloudflare-go/v3/zones"

	"github.com/stupside/moley/v2/internal/domain"
	logger "github.com/stupside/moley/v2/internal/platform/logging"
	"github.com/stupside/moley/v2/internal/platform/paths"
	"go.yaml.in/yaml/v3"
)

type TunnelService struct {
	client    *cfgo.Client
	accountID string
	dryRun    bool
}

func NewTunnelService(ctx context.Context, client *cfgo.Client, zoneName string, dryRun bool) (*TunnelService, error) {
	svc := &TunnelService{
		client: client,
		dryRun: dryRun,
	}

	if svc.dryRun {
		svc.accountID = "dry-run-account"
		return svc, nil
	}

	accountID, err := svc.resolveAccountID(ctx, zoneName)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve account ID from zone %s: %w", zoneName, err)
	}
	svc.accountID = accountID

	return svc, nil
}

func (c *TunnelService) resolveAccountID(ctx context.Context, zoneName string) (string, error) {
	pager := c.client.Zones.ListAutoPaging(ctx, zones.ZoneListParams{
		Name: cfgo.F(zoneName),
	})
	for pager.Next() {
		return pager.Current().Account.ID, nil
	}
	if err := pager.Err(); err != nil {
		return "", fmt.Errorf("failed to look up zone: %w", err)
	}
	return "", fmt.Errorf("zone %q not found — check your token has Zone > Zone > Read permission", zoneName)
}

func (c *TunnelService) AccountID() string {
	return c.accountID
}

func (c *TunnelService) Run(ctx context.Context, tunnel *domain.Tunnel) (int, error) {
	if c.dryRun {
		logger.Debug("Dry run: skipping tunnel process start")
		return 0, nil
	}

	configPath, err := c.GetConfigurationPath(ctx, tunnel)
	if err != nil {
		return 0, fmt.Errorf("failed to get tunnel configuration path: %w", err)
	}

	// Pass the UUID (not the name) so cloudflared doesn't do a name→UUID
	// lookup via the Cloudflare control plane, which would require cert.pem
	// from `cloudflared tunnel login`.
	tunnelUUID, err := c.GetID(ctx, tunnel)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve tunnel UUID: %w", err)
	}

	cfCommand := newCommand(ctx, "tunnel", "--config", configPath, "run", tunnelUUID)
	pid, err := cfCommand.execAsync()
	if err != nil {
		return 0, fmt.Errorf("failed to run tunnel: %w", err)
	}
	return pid, nil
}

// findTunnel looks up a tunnel by name and returns its UUID, or empty string if not found.
func (c *TunnelService) findTunnel(ctx context.Context, tunnel *domain.Tunnel) (string, error) {
	name := tunnel.GetName()

	pager := c.client.ZeroTrust.Tunnels.ListAutoPaging(ctx, zero_trust.TunnelListParams{
		AccountID: cfgo.F(c.accountID),
		Name:      cfgo.F(name),
		IsDeleted: cfgo.F(false),
	})

	for pager.Next() {
		t := pager.Current()
		if t.Name == name {
			return t.ID, nil
		}
	}
	if err := pager.Err(); err != nil {
		return "", fmt.Errorf("failed to list tunnels: %w", err)
	}

	return "", nil
}

func (c *TunnelService) GetID(ctx context.Context, tunnel *domain.Tunnel) (string, error) {
	if c.dryRun {
		return tunnel.GetName(), nil
	}

	tunnelID, err := c.findTunnel(ctx, tunnel)
	if err != nil {
		return "", err
	}
	if tunnelID == "" {
		return "", fmt.Errorf("tunnel %s not found", tunnel.GetName())
	}
	return tunnelID, nil
}

func (c *TunnelService) Create(ctx context.Context, tunnel *domain.Tunnel) (string, error) {
	logger.Info("Creating Cloudflare tunnel")

	if c.dryRun {
		logger.Debug("Dry run: skipping tunnel creation")
		return tunnel.GetName(), nil
	}

	// Generate a random 32-byte tunnel secret
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("failed to generate tunnel secret: %w", err)
	}
	tunnelSecret := base64.StdEncoding.EncodeToString(secret)

	result, err := c.client.ZeroTrust.Tunnels.New(ctx, zero_trust.TunnelNewParams{
		AccountID:    cfgo.F(c.accountID),
		Name:         cfgo.F(tunnel.GetName()),
		ConfigSrc:    cfgo.F(zero_trust.TunnelNewParamsConfigSrcLocal),
		TunnelSecret: cfgo.F(tunnelSecret),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create tunnel: %w", err)
	}

	// Save credentials file so cloudflared can use it for `tunnel run`
	if err := c.saveCredentials(result.ID, result.AccountTag, tunnelSecret, tunnel.GetName()); err != nil {
		return "", fmt.Errorf("failed to save tunnel credentials: %w", err)
	}

	logger.Debugf("Tunnel created successfully", map[string]any{
		"name":     tunnel.GetName(),
		"tunnelID": result.ID,
		"account":  result.AccountTag,
	})
	return result.ID, nil
}

// tunnelCredentials is the format cloudflared expects in ~/.cloudflared/{tunnelID}.json.
type tunnelCredentials struct {
	AccountTag   string `json:"AccountTag"`
	TunnelSecret string `json:"TunnelSecret"`
	TunnelID     string `json:"TunnelID"`
	TunnelName   string `json:"TunnelName"`
}

func cloudflaredCredPath(tunnelID string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".cloudflared", tunnelID+".json"), nil
}

// saveCredentials writes the credentials file that cloudflared needs to run the tunnel.
func (c *TunnelService) saveCredentials(tunnelID, accountTag, tunnelSecret, tunnelName string) error {
	credPath, err := cloudflaredCredPath(tunnelID)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(credPath), 0700); err != nil {
		return fmt.Errorf("failed to create credentials directory: %w", err)
	}

	creds := tunnelCredentials{
		AccountTag:   accountTag,
		TunnelSecret: tunnelSecret,
		TunnelID:     tunnelID,
		TunnelName:   tunnelName,
	}

	data, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	if err := os.WriteFile(credPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	logger.Debugf("Credentials saved", map[string]any{
		"path": credPath,
	})
	return nil
}

func (c *TunnelService) Delete(ctx context.Context, tunnel *domain.Tunnel) error {
	logger.Info("Deleting Cloudflare tunnel")

	if c.dryRun {
		logger.Debug("Dry run: skipping tunnel deletion")
		return nil
	}

	tunnelID, err := c.findTunnel(ctx, tunnel)
	if err != nil {
		return fmt.Errorf("failed to find tunnel for deletion: %w", err)
	}
	if tunnelID == "" {
		logger.Debug("Tunnel does not exist, skipping deletion")
		return nil
	}

	// Clean up connections first (best effort)
	_, err = c.client.ZeroTrust.Tunnels.Connections.Delete(ctx, tunnelID, zero_trust.TunnelConnectionDeleteParams{
		AccountID: cfgo.F(c.accountID),
	})
	if err != nil {
		logger.Warnf("Tunnel cleanup before deletion failed", map[string]any{
			"error": err.Error(),
		})
	}

	// Delete tunnel
	_, err = c.client.ZeroTrust.Tunnels.Delete(ctx, tunnelID, zero_trust.TunnelDeleteParams{
		AccountID: cfgo.F(c.accountID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete tunnel: %w", err)
	}

	// Clean up credentials file
	if credPath, err := cloudflaredCredPath(tunnelID); err == nil {
		if removeErr := os.Remove(credPath); removeErr != nil && !os.IsNotExist(removeErr) {
			logger.Warnf("Failed to remove credentials file", map[string]any{
				"path":  credPath,
				"error": removeErr.Error(),
			})
		}
	}

	logger.Debug("Cloudflare tunnel deleted successfully")
	return nil
}

func (c *TunnelService) SaveConfiguration(ctx context.Context, tunnel *domain.Tunnel, ingress *domain.Ingress) error {
	logger.Info("Saving Cloudflare configuration")

	tunnelUUID, err := c.GetID(ctx, tunnel)
	if err != nil {
		return fmt.Errorf("failed to resolve tunnel UUID: %w", err)
	}

	credentialsFile, err := cloudflaredCredPath(tunnelUUID)
	if err != nil {
		return fmt.Errorf("failed to build credentials file path: %w", err)
	}

	logger.Debugf("Using credentials file", map[string]any{
		"path": credentialsFile,
	})

	// Store the UUID (not the name) in the config so cloudflared never needs
	// to resolve a name via the control plane (which would require cert.pem).
	config := &runConfig{
		Tunnel:          tunnelUUID,
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
		config.Ingress = append(config.Ingress, ingressRule{
			Service:  app.Target.GetTargetURL(),
			Hostname: domain.FQDN(app.Expose.Subdomain, ingress.Zone),
		})
	}

	logger.Info("Adding catch-all ingress rule")
	config.Ingress = append(config.Ingress, ingressRule{
		Service: "http_status:404",
	})

	bytes, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	path, err := c.GetConfigurationPath(ctx, tunnel)
	if err != nil {
		return fmt.Errorf("failed to get tunnel configuration path: %w", err)
	}

	if err := os.WriteFile(path, bytes, 0600); err != nil {
		return fmt.Errorf("failed to save tunnel configuration: %w", err)
	}

	logger.Debug("Configuration saved")
	return nil
}

func (c *TunnelService) DeleteConfiguration(ctx context.Context, tunnel *domain.Tunnel) error {
	logger.Info("Deleting configuration")

	path, err := c.GetConfigurationPath(ctx, tunnel)
	if err != nil {
		return fmt.Errorf("failed to get tunnel configuration path: %w", err)
	}

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			logger.Debug("Configuration file does not exist, skipping deletion")
			return nil
		}
		return fmt.Errorf("failed to delete tunnel configuration: %w", err)
	}

	logger.Debug("Configuration deleted")
	return nil
}

func (c *TunnelService) Exists(ctx context.Context, tunnel *domain.Tunnel) (bool, error) {
	logger.Debugf("Checking if tunnel exists", map[string]any{
		"tunnelName": tunnel.GetName(),
	})

	if c.dryRun {
		return true, nil
	}

	tunnelID, err := c.findTunnel(ctx, tunnel)
	if err != nil {
		return false, fmt.Errorf("failed to check if tunnel exists: %w", err)
	}

	exists := tunnelID != ""
	logger.Debugf("Tunnel exists check result", map[string]any{
		"tunnelName": tunnel.GetName(),
		"exists":     exists,
	})
	return exists, nil
}

func (c *TunnelService) GetConfigurationPath(ctx context.Context, tunnel *domain.Tunnel) (string, error) {
	logger.Debug("Getting tunnel configuration path")

	base, err := paths.GetUserFolderPath()
	if err != nil {
		return "", fmt.Errorf("failed to get user folder path: %w", err)
	}

	tunnelsFolder := filepath.Join(base, "tunnels")
	if err := os.MkdirAll(tunnelsFolder, 0755); err != nil {
		return "", fmt.Errorf("failed to create tunnels directory: %w", err)
	}

	tunnelFile := filepath.Join(tunnelsFolder, tunnel.GetName()+".yml")

	logger.Debugf("Tunnel configuration path", map[string]any{
		"path": tunnelFile,
	})

	return tunnelFile, nil
}
