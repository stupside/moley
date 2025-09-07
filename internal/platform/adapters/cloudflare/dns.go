// Package cloudflare provides Cloudflare-specific implementations.
package cloudflare

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/core/ports"
	"github.com/stupside/moley/v2/internal/platform/framework"
	"github.com/stupside/moley/v2/internal/shared"

	"github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/dns"
	"github.com/cloudflare/cloudflare-go/v3/option"
	"github.com/cloudflare/cloudflare-go/v3/zones"
)

type dnsService struct {
	cfApiClient   *cloudflare.Client
	tunnelService ports.TunnelService
	config        *framework.Config
}

func NewDNSService(token string, tunnelService ports.TunnelService, config *framework.Config) (*dnsService, error) {
	return &dnsService{
		tunnelService: tunnelService,
		config:        config,
		cfApiClient: cloudflare.NewClient(
			option.WithAPIToken(token),
		),
	}, nil
}

func (c *dnsService) GetContent(ctx context.Context, tunnel *domain.Tunnel) (string, error) {
	tunnelID, err := c.tunnelService.GetID(ctx, tunnel)
	if err != nil {
		return "", shared.WrapError(err, "failed to get tunnel ID")
	}
	return fmt.Sprintf("%s.%s", tunnelID, "cfargotunnel.com"), nil
}

func (c *dnsService) RouteRecord(ctx context.Context, tunnel *domain.Tunnel, zoneName string, subdomain string) error {
	_, err := framework.RunWithDryRunGuard(c.config, func() (string, error) {
		cfCommand := NewCommand(ctx, "tunnel", "route", "dns", tunnel.GetName(), subdomain)
		return cfCommand.Exec()
	}, "")
	if err != nil {
		return shared.WrapError(err, fmt.Sprintf("failed to create DNS record for subdomain %s", subdomain))
	}
	return nil
}

func (c *dnsService) GetZoneID(ctx context.Context, zoneName string) (string, error) {
	zones, err := c.cfApiClient.Zones.List(ctx, zones.ZoneListParams{
		Name: cloudflare.F(zoneName),
	})
	if err != nil {
		return "", shared.WrapError(err, "failed to list zones")
	}
	if len(zones.Result) == 0 {
		return "", fmt.Errorf("zone %s not found", zoneName)
	}
	if len(zones.Result) > 1 {
		return "", fmt.Errorf("multiple zones found for %s", zoneName)
	}
	return zones.Result[0].ID, nil
}

func (c *dnsService) GetRecords(ctx context.Context, tunnel *domain.Tunnel, zoneName string) ([]dns.RecordListResponse, error) {
	zoneId, err := c.GetZoneID(ctx, zoneName)
	if err != nil {
		return nil, shared.WrapError(err, "failed to get zone ID")
	}
	dnsContent, err := c.GetContent(ctx, tunnel)
	if err != nil {
		return nil, shared.WrapError(err, "failed to get DNS content")
	}
	records, err := c.cfApiClient.DNS.Records.List(ctx, dns.RecordListParams{
		ZoneID:  cloudflare.F(zoneId),
		Content: cloudflare.F(dnsContent),
	})
	if err != nil {
		return nil, shared.WrapError(err, "failed to list DNS records")
	}
	return records.Result, nil
}

func (c *dnsService) DeleteRecord(ctx context.Context, tunnel *domain.Tunnel, zoneName string, subdomain string) error {
	zoneId, err := c.GetZoneID(ctx, zoneName)
	if err != nil {
		return shared.WrapError(err, fmt.Sprintf("failed to get zone ID for zone %s", zoneName))
	}
	dnsRecords, err := c.GetRecords(ctx, tunnel, zoneName)
	if err != nil {
		return shared.WrapError(err, fmt.Sprintf("failed to get DNS records for tunnel %s in zone %s", tunnel.GetName(), zoneName))
	}
	expectedRecordName := fmt.Sprintf("%s.%s", subdomain, zoneName)
	for _, record := range dnsRecords {
		if record.Name == expectedRecordName {
			_, err := framework.RunWithDryRunGuard(c.config, func() (*dns.RecordDeleteResponse, error) {
				return c.cfApiClient.DNS.Records.Delete(ctx, record.ID, dns.RecordDeleteParams{
					ZoneID: cloudflare.F(zoneId),
				})
			}, nil)
			if err != nil {
				return shared.WrapError(err, fmt.Sprintf("failed to delete DNS record %s", record.Name))
			}
			return nil
		}
	}
	return fmt.Errorf("DNS record for subdomain %s not found in zone %s", subdomain, zoneName)
}

func (c *dnsService) RecordExists(ctx context.Context, tunnel *domain.Tunnel, zoneName string, subdomain string) (bool, error) {
	records, err := c.GetRecords(ctx, tunnel, zoneName)
	if err != nil {
		return false, err
	}
	expected := fmt.Sprintf("%s.%s", subdomain, zoneName)
	for _, r := range records {
		if r.Name == expected {
			return true, nil
		}
	}
	return false, nil
}
