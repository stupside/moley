// Package cloudflare provides Cloudflare-specific implementations.
package cloudflare

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/domain"
	logger "github.com/stupside/moley/v2/internal/platform/logging"

	cfgo "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/zones"
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

	name := domain.FQDN(subdomain, zoneName)

	recordID, err := c.findRecord(ctx, zoneID, name, dnsContent)
	if err != nil {
		return fmt.Errorf("failed to check existing DNS records: %w", err)
	}
	if recordID != "" {
		logger.Debugf("DNS record already exists, skipping creation", map[string]any{
			"subdomain": subdomain,
			"zone":      zoneName,
		})
		return nil
	}

	_, err = c.client.DNS.Records.New(ctx, dns.RecordNewParams{
		ZoneID: cfgo.F(zoneID),
		Body: dns.CNAMERecordParam{
			Name:    cfgo.F(name),
			Type:    cfgo.F(dns.CNAMERecordTypeCNAME),
			Content: cfgo.F(dnsContent),
			Proxied: cfgo.F(true),
			TTL:     cfgo.F(dns.TTL1), // automatic
		},
	})
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

// findRecord returns the ID of the record matching name and content exactly, or "" if none exists.
func (c *DNSService) findRecord(ctx context.Context, zoneID, name, content string) (string, error) {
	pager := c.client.DNS.Records.ListAutoPaging(ctx, dns.RecordListParams{
		ZoneID:  cfgo.F(zoneID),
		Name:    cfgo.F(dns.RecordListParamsName{Exact: cfgo.F(name)}),
		Content: cfgo.F(dns.RecordListParamsContent{Exact: cfgo.F(content)}),
	})

	for pager.Next() {
		return pager.Current().ID, nil
	}
	if err := pager.Err(); err != nil {
		return "", fmt.Errorf("failed to list DNS records: %w", err)
	}

	return "", nil
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

	name := domain.FQDN(subdomain, zoneName)

	recordID, err := c.findRecord(ctx, zoneID, name, cnameTarget(tunnelUUID))
	if err != nil {
		return fmt.Errorf("failed to get DNS records for tunnel %s in zone %s: %w", tunnelUUID, zoneName, err)
	}
	if recordID == "" {
		logger.Debugf("DNS record not found, skipping deletion", map[string]any{
			"subdomain": subdomain,
			"zone":      zoneName,
		})
		return nil
	}

	_, err = c.client.DNS.Records.Delete(ctx, recordID, dns.RecordDeleteParams{
		ZoneID: cfgo.F(zoneID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete DNS record %s: %w", name, err)
	}
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

	recordID, err := c.findRecord(ctx, zoneID, domain.FQDN(subdomain, zoneName), cnameTarget(tunnelUUID))
	if err != nil {
		return false, err
	}
	return recordID != "", nil
}
