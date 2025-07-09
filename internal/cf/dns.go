package cf

import (
	"context"
	"fmt"

	"github.com/stupside/moley/internal/domain"
	"github.com/stupside/moley/internal/logger"
	"github.com/stupside/moley/internal/services"
	"github.com/stupside/moley/internal/shared"

	"github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/dns"
	"github.com/cloudflare/cloudflare-go/v3/option"
	"github.com/cloudflare/cloudflare-go/v3/zones"
)

// dnsService implements the DNSService interface for Cloudflare
type dnsService struct {
	cfApiClient   *cloudflare.Client
	tunnelService services.TunnelService
}

// NewDNSService creates a new Cloudflare DNS service
func NewDNSService(token string, tunnelService services.TunnelService) (*dnsService, error) {
	cfApiClient := cloudflare.NewClient(
		option.WithAPIToken(token),
	)

	return &dnsService{
		cfApiClient:   cfApiClient,
		tunnelService: tunnelService,
	}, nil
}

func (c *dnsService) GetContent(ctx context.Context, domainTunnel *domain.Tunnel) (string, error) {
	tunnelID, err := c.tunnelService.GetID(ctx, domainTunnel)
	var ErrGetTunnelID = shared.WrapError(shared.ErrConfigRead, "failed to get tunnel ID")
	if err != nil {
		return "", shared.WrapError(ErrGetTunnelID, err.Error())
	}
	return fmt.Sprintf("%s.%s", tunnelID, "cfargotunnel.com"), nil
}

// Route creates DNS routes for multiple hostnames in a tunnel
func (c *dnsService) Route(ctx context.Context, domainTunnel *domain.Tunnel, domainDNS *domain.DNS) error {
	logger.Debug(fmt.Sprintf("Routing DNS records for tunnel: %s", domainTunnel.GetName()))

	for _, subdomain := range domainDNS.GetSubdomains() {
		logger.Debug(fmt.Sprintf("Creating DNS record for subdomain: %s", subdomain))

		// TODO: If a record already exists for this subdomain, it will crash.
		// This should be handled by checking if the record exists before creating it.

		var ErrCloudflaredRoute = shared.WrapError(shared.ErrConfigWrite, "cloudflared tunnel dns create failed")
		_, err := execCloudflared(ctx, "tunnel", "route", "dns", domainTunnel.GetName(), subdomain)
		if err != nil {
			return shared.WrapError(ErrCloudflaredRoute, err.Error())
		}
		logger.Debug(fmt.Sprintf("DNS record created successfully for subdomain: %s", subdomain))
	}
	logger.Debug(fmt.Sprintf("All DNS records routed for tunnel: %s", domainTunnel.GetName()))
	return nil
}

// GetZoneID retrieves the zone ID for a given tunnel
func (c *dnsService) GetZoneID(ctx context.Context, domainDNS *domain.DNS) (string, error) {
	// Get the zone ID using the account ID
	zones, err := c.cfApiClient.Zones.List(ctx, zones.ZoneListParams{
		Name: cloudflare.F(domainDNS.GetZone()),
	})
	var ErrListZones = shared.WrapError(shared.ErrConfigRead, "failed to list zones")
	if err != nil {
		return "", shared.WrapError(ErrListZones, err.Error())
	}

	if len(zones.Result) == 0 {
		var ErrNoZones = shared.WrapError(shared.ErrConfigNotFound, "no zones found for the specified name")
		return "", ErrNoZones
	}

	return zones.Result[0].ID, nil
}

// GetRecords lists DNS records for a zone with optional filtering
func (c *dnsService) GetRecords(ctx context.Context, domainTunnel *domain.Tunnel, domainDNS *domain.DNS) ([]dns.RecordListResponse, error) {
	// Get the zone ID for the tunnel
	zoneId, err := c.GetZoneID(ctx, domainDNS)
	var ErrGetZoneID = shared.WrapError(shared.ErrConfigRead, "failed to get zone ID")
	if err != nil {
		return nil, shared.WrapError(ErrGetZoneID, err.Error())
	}

	// Get the DNS content for the tunnel
	dnsContent, err := c.GetContent(ctx, domainTunnel)
	var ErrGetContent = shared.WrapError(shared.ErrConfigRead, "failed to get tunnel DNS content")
	if err != nil {
		return nil, shared.WrapError(ErrGetContent, err.Error())
	}

	// List DNS records for the zone with the specified content
	records, err := c.cfApiClient.DNS.Records.List(ctx, dns.RecordListParams{
		ZoneID:  cloudflare.F(zoneId),
		Content: cloudflare.F(dnsContent),
	})
	var ErrListDNSRecords = shared.WrapError(shared.ErrConfigRead, "failed to list DNS records")
	if err != nil {
		return nil, shared.WrapError(ErrListDNSRecords, err.Error())
	}

	return records.Result, nil
}

// DeleteRecords deletes all DNS records associated with a tunnel
func (c *dnsService) DeleteRecords(ctx context.Context, domainTunnel *domain.Tunnel, domainDNS *domain.DNS) error {

	// Get the zone ID for the tunnel
	zoneId, err := c.GetZoneID(ctx, domainDNS)
	var ErrGetZoneID = fmt.Errorf("failed to get zone ID: %w", shared.ErrConfigRead)
	if err != nil {
		return shared.WrapError(ErrGetZoneID, err.Error())
	}

	// Get the DNS records for the tunnel
	dnsRecords, err := c.GetRecords(ctx, domainTunnel, domainDNS)
	var ErrGetRecords = fmt.Errorf("failed to list DNS records: %w", shared.ErrConfigRead)
	if err != nil {
		return shared.WrapError(ErrGetRecords, err.Error())
	}

	// For each DNS record, delete it
	for _, record := range dnsRecords {
		var ErrDeleteRecord = fmt.Errorf("failed to delete DNS record %s: %w", record.Name, shared.ErrConfigWrite)
		_, err := c.cfApiClient.DNS.Records.Delete(ctx, record.ID, dns.RecordDeleteParams{
			ZoneID: cloudflare.F(zoneId),
		})
		if err != nil {
			return shared.WrapError(ErrDeleteRecord, err.Error())
		}
	}

	return nil
}
