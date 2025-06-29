package cloudflare

import (
	"context"
	"fmt"
	"moley/internal/config"
	"moley/internal/logger"
	"os/exec"

	cfapi "github.com/cloudflare/cloudflare-go"
)

// CloudflareClient wraps the Cloudflare API client and config
// and provides idiomatic methods for all operations.
type CloudflareClient struct {
	api    *cfapi.API
	config *config.MoleyConfig
}

// NewClient creates a new CloudflareClient
func NewClient(cfg *config.MoleyConfig) (*CloudflareClient, error) {
	token := cfg.GetAPIToken()
	if token == "" {
		return nil, fmt.Errorf("cloudflare.api_token must be set in moly.yml. Please create an API token in your Cloudflare dashboard with Zone:Read and DNS:Edit permissions")
	}
	api, err := cfapi.NewWithAPIToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloudflare API client: %w", err)
	}
	return &CloudflareClient{api: api, config: cfg}, nil
}

// ZoneID returns the zone ID for the configured zone
func (c *CloudflareClient) ZoneID(ctx context.Context) (string, error) {
	zone := c.config.Zone
	if zone == "" {
		return "", fmt.Errorf("zone must be set in config")
	}
	zoneID, err := c.api.ZoneIDByName(zone)
	if err != nil {
		return "", fmt.Errorf("failed to get zone ID for %s: %w", zone, err)
	}
	return zoneID, nil
}

// CreateDNSRecord creates a DNS record using cloudflared tunnel dns
func (c *CloudflareClient) CreateDNSRecord(ctx context.Context, tunnelName, hostname string) error {
	cmd := exec.CommandContext(ctx, "cloudflared", "tunnel", "route", "dns", tunnelName, hostname)

	if output, err := cmd.CombinedOutput(); err != nil {
		logger.Warnf("Failed to create DNS record via cloudflared", map[string]interface{}{
			"hostname": hostname,
			"tunnel":   tunnelName,
			"error":    err.Error(),
			"output":   string(output),
		})
		return fmt.Errorf("cloudflared tunnel dns create failed: %w", err)
	}

	logger.Infof("DNS record created via cloudflared", map[string]interface{}{
		"hostname": hostname,
		"tunnel":   tunnelName,
	})
	return nil
}

// ListDNSRecords lists DNS records for a zone with optional filtering
func (c *CloudflareClient) ListDNSRecords(ctx context.Context, zoneID string, params cfapi.ListDNSRecordsParams) ([]cfapi.DNSRecord, error) {
	zoneResource := cfapi.ZoneIdentifier(zoneID)
	records, _, err := c.api.ListDNSRecords(ctx, zoneResource, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list DNS records: %w", err)
	}
	return records, nil
}

// DeleteDNSRecord deletes a DNS record using the Cloudflare API
func (c *CloudflareClient) DeleteDNSRecord(ctx context.Context, tunnelName, hostname string) error {
	// Get tunnel ID from tunnel name
	tunnelID, err := GetTunnelIDByName(ctx, tunnelName)
	if err != nil {
		return fmt.Errorf("failed to get tunnel ID for %s: %w", tunnelName, err)
	}

	// Get zone ID
	zoneID, err := c.ZoneID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get zone ID: %w", err)
	}

	// List DNS records to find the one to delete
	params := cfapi.ListDNSRecordsParams{
		Name: hostname,
		Type: "CNAME",
	}

	records, err := c.ListDNSRecords(ctx, zoneID, params)
	if err != nil {
		return fmt.Errorf("failed to list DNS records: %w", err)
	}

	// Find the record pointing to our tunnel
	tunnelTarget := fmt.Sprintf("%s.cfargotunnel.com", tunnelID)
	for _, record := range records {
		if record.Content == tunnelTarget {
			// Delete the record using the API
			if err := c.api.DeleteDNSRecord(ctx, cfapi.ZoneIdentifier(zoneID), record.ID); err != nil {
				logger.Warnf("Failed to delete DNS record via API", map[string]interface{}{
					"hostname":  hostname,
					"record_id": record.ID,
					"error":     err.Error(),
				})
				return fmt.Errorf("failed to delete DNS record: %w", err)
			}

			logger.Infof("DNS record deleted via API", map[string]interface{}{
				"hostname":  hostname,
				"record_id": record.ID,
			})
			return nil
		}
	}

	logger.Warnf("DNS record not found for deletion", map[string]interface{}{
		"hostname": hostname,
		"tunnel":   tunnelName,
	})
	return nil
}
