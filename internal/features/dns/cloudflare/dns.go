// Package cloudflare provides Cloudflare-specific implementations.
package cloudflare

import (
	"context"
	"fmt"

	logger "github.com/stupside/moley/v2/internal/platform/logging"

	cfgo "github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/dns"
	"github.com/cloudflare/cloudflare-go/v3/option"
	"github.com/cloudflare/cloudflare-go/v3/zones"
)

type DNSService struct {
	client *cfgo.Client
	dryRun bool
}

func NewDNSService(client *cfgo.Client, dryRun bool) *DNSService {
	return &DNSService{
		dryRun: dryRun,
		client: client,
	}
}

func cnameTarget(tunnelUUID string) string {
	return tunnelUUID + ".cfargotunnel.com"
}

func (c *DNSService) RouteRecord(ctx context.Context, tunnelUUID string, zoneName string, subdomain string) error {
	if c.dryRun {
		logger.Debug("Dry run: skipping DNS record creation")
		return nil
	}

	zoneID, err := c.getZoneID(ctx, zoneName)
	if err != nil {
		return fmt.Errorf("failed to get zone ID for zone %s: %w", zoneName, err)
	}

	dnsContent := cnameTarget(tunnelUUID)

	name := subdomain + "." + zoneName

	records, err := c.listRecords(ctx, zoneID, dnsContent)
	if err != nil {
		return fmt.Errorf("failed to check existing DNS records: %w", err)
	}
	for _, r := range records {
		if r.Name == name {
			logger.Debugf("DNS record already exists, skipping creation", map[string]any{
				"subdomain": subdomain,
				"zone":      zoneName,
			})
			return nil
		}
	}

	_, err = c.client.DNS.Records.New(ctx, dns.RecordNewParams{
		ZoneID: cfgo.F(zoneID),
		Record: dns.RecordParam{
			Name:    cfgo.F(name),
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

func (c *DNSService) getZoneID(ctx context.Context, zoneName string) (string, error) {
	pager := c.client.Zones.ListAutoPaging(ctx, zones.ZoneListParams{
		Name: cfgo.F(zoneName),
	})

	var found bool
	var zoneID string
	for pager.Next() {
		if found {
			return "", fmt.Errorf("multiple zones found for %s", zoneName)
		}
		zoneID = pager.Current().ID
		found = true
	}
	if err := pager.Err(); err != nil {
		return "", fmt.Errorf("failed to list zones: %w", err)
	}
	if !found {
		return "", fmt.Errorf("zone %s not found", zoneName)
	}

	return zoneID, nil
}

func (c *DNSService) listRecords(ctx context.Context, zoneID, content string) ([]dns.RecordListResponse, error) {
	pager := c.client.DNS.Records.ListAutoPaging(ctx, dns.RecordListParams{
		ZoneID:  cfgo.F(zoneID),
		Content: cfgo.F(content),
	})

	var records []dns.RecordListResponse
	for pager.Next() {
		records = append(records, pager.Current())
	}
	if err := pager.Err(); err != nil {
		return nil, fmt.Errorf("failed to list DNS records: %w", err)
	}

	return records, nil
}

func (c *DNSService) DeleteRecord(ctx context.Context, tunnelUUID string, zoneName string, subdomain string) error {
	if c.dryRun {
		logger.Debug("Dry run: skipping DNS record deletion")
		return nil
	}

	zoneID, err := c.getZoneID(ctx, zoneName)
	if err != nil {
		return fmt.Errorf("failed to get zone ID for zone %s: %w", zoneName, err)
	}

	dnsContent := cnameTarget(tunnelUUID)

	records, err := c.listRecords(ctx, zoneID, dnsContent)
	if err != nil {
		return fmt.Errorf("failed to get DNS records for tunnel %s in zone %s: %w", tunnelUUID, zoneName, err)
	}

	name := subdomain + "." + zoneName
	for _, record := range records {
		if record.Name == name {
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

func (c *DNSService) RecordExists(ctx context.Context, tunnelUUID string, zoneName string, subdomain string) (bool, error) {
	if c.dryRun {
		return true, nil
	}

	zoneID, err := c.getZoneID(ctx, zoneName)
	if err != nil {
		return false, fmt.Errorf("failed to get zone ID: %w", err)
	}

	dnsContent := cnameTarget(tunnelUUID)

	records, err := c.listRecords(ctx, zoneID, dnsContent)
	if err != nil {
		return false, err
	}

	name := subdomain + "." + zoneName
	for _, r := range records {
		if r.Name == name {
			return true, nil
		}
	}
	return false, nil
}
