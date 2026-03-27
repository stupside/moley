// Package cloudflare provides Cloudflare-specific implementations.
package cloudflare

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/core/ports"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"

	cfgo "github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/dns"
	"github.com/cloudflare/cloudflare-go/v3/option"
	"github.com/cloudflare/cloudflare-go/v3/zones"
)

type dnsService struct {
	client   *cfgo.Client
	tunnelService ports.TunnelService
	config        *Config
	// zoneIDCache caches zone name → zone ID lookups within this service instance
	zoneIDCache map[string]string
}


func NewDNSService(client *cfgo.Client, tunnelService ports.TunnelService, config *Config) *dnsService {
	return &dnsService{
		tunnelService: tunnelService,
		config:        config,
		client:   client,
		zoneIDCache:   make(map[string]string),
	}
}

func (c *dnsService) getContent(ctx context.Context, tunnel *domain.Tunnel) (string, error) {
	tunnelID, err := c.tunnelService.GetID(ctx, tunnel)
	if err != nil {
		return "", fmt.Errorf("failed to get tunnel ID: %w", err)
	}
	return tunnelID + ".cfargotunnel.com", nil
}

func (c *dnsService) RouteRecord(ctx context.Context, tunnel *domain.Tunnel, zoneName string, subdomain string) error {
	if c.config.IsDryRun() {
		logger.Debug("Dry run: skipping DNS record creation")
		return nil
	}

	zoneID, err := c.getZoneID(ctx, zoneName)
	if err != nil {
		return fmt.Errorf("failed to get zone ID for zone %s: %w", zoneName, err)
	}

	dnsContent, err := c.getContent(ctx, tunnel)
	if err != nil {
		return fmt.Errorf("failed to get DNS content: %w", err)
	}

	recordName := subdomain + "." + zoneName

	_, err = c.client.DNS.Records.New(ctx, dns.RecordNewParams{
		ZoneID: cfgo.F(zoneID),
		Record: dns.RecordParam{
			Name:    cfgo.F(recordName),
			Proxied: cfgo.F(true),
			TTL:     cfgo.F(dns.TTL1), // automatic
		},
	},
		option.WithJSONSet("type", "CNAME"),
		option.WithJSONSet("content", dnsContent),
	)
	if err != nil {
		return fmt.Errorf("failed to create DNS record for subdomain %s: %w", subdomain, err)
	}
	return nil
}

func (c *dnsService) getZoneID(ctx context.Context, zoneName string) (string, error) {
	if id, ok := c.zoneIDCache[zoneName]; ok {
		return id, nil
	}

	z, err := c.client.Zones.List(ctx, zones.ZoneListParams{
		Name: cfgo.F(zoneName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list zones: %w", err)
	}
	if len(z.Result) == 0 {
		return "", fmt.Errorf("zone %s not found", zoneName)
	}
	if len(z.Result) > 1 {
		return "", fmt.Errorf("multiple zones found for %s", zoneName)
	}

	c.zoneIDCache[zoneName] = z.Result[0].ID
	return z.Result[0].ID, nil
}

func (c *dnsService) getRecords(ctx context.Context, tunnel *domain.Tunnel, zoneName string) ([]dns.RecordListResponse, error) {
	zoneID, err := c.getZoneID(ctx, zoneName)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone ID: %w", err)
	}
	dnsContent, err := c.getContent(ctx, tunnel)
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS content: %w", err)
	}
	records, err := c.client.DNS.Records.List(ctx, dns.RecordListParams{
		ZoneID:  cfgo.F(zoneID),
		Content: cfgo.F(dnsContent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list DNS records: %w", err)
	}
	return records.Result, nil
}

func (c *dnsService) DeleteRecord(ctx context.Context, tunnel *domain.Tunnel, zoneName string, subdomain string) error {
	if c.config.IsDryRun() {
		logger.Debug("Dry run: skipping DNS record deletion")
		return nil
	}

	records, err := c.getRecords(ctx, tunnel, zoneName)
	if err != nil {
		return fmt.Errorf("failed to get DNS records for tunnel %s in zone %s: %w", tunnel.GetName(), zoneName, err)
	}

	zoneID, err := c.getZoneID(ctx, zoneName)
	if err != nil {
		return fmt.Errorf("failed to get zone ID for deletion: %w", err)
	}

	expectedName := subdomain + "." + zoneName
	for _, record := range records {
		if record.Name == expectedName {
			_, err := c.client.DNS.Records.Delete(ctx, record.ID, dns.RecordDeleteParams{
				ZoneID: cfgo.F(zoneID),
			})
			if err != nil {
				return fmt.Errorf("failed to delete DNS record %s: %w", record.Name, err)
			}
			return nil
		}
	}
	logger.Debugf("DNS record not found, skipping deletion", map[string]any{
		"subdomain": subdomain,
		"zone":      zoneName,
	})
	return nil
}

func (c *dnsService) RecordExists(ctx context.Context, tunnel *domain.Tunnel, zoneName string, subdomain string) (bool, error) {
	if c.config.IsDryRun() {
		return true, nil
	}

	records, err := c.getRecords(ctx, tunnel, zoneName)
	if err != nil {
		return false, err
	}
	expected := subdomain + "." + zoneName
	for _, r := range records {
		if r.Name == expected {
			return true, nil
		}
	}
	return false, nil
}
