package cf

import (
	"context"
	"fmt"
	"moley/internal/domain"
	"moley/internal/errors"
	"moley/internal/logger"
	"moley/internal/services"

	"github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/dns"
	"github.com/cloudflare/cloudflare-go/v3/option"
	"github.com/cloudflare/cloudflare-go/v3/zones"
)

// dnsService implements the DNSService interface for Cloudflare
type dnsService struct {
	services.DNSService

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
	if err != nil {
		return "", errors.NewConfigError(errors.ErrCodeInvalidConfig, "failed to get tunnel ID", err)
	}

	return fmt.Sprintf("%s.%s", tunnelID, "cfargotunnel.com"), nil
}

// Route creates DNS routes for multiple hostnames in a tunnel
func (c *dnsService) Route(ctx context.Context, domainTunnel *domain.Tunnel, domainDNS *domain.DNS) error {
	logger.Debugf("Routing DNS records for tunnel", map[string]interface{}{
		"tunnel":     domainTunnel.GetName(),
		"subdomains": domainDNS.GetSubdomains(),
	})

	for _, subdomain := range domainDNS.GetSubdomains() {
		logger.Debugf("Creating DNS record", map[string]interface{}{
			"tunnel":    domainTunnel.GetName(),
			"subdomain": subdomain,
		})

		// TODO: If a record already exists for this subdomain, it will crash.
		// This should be handled by checking if the record exists before creating it.

		_, err := execCloudflared(ctx, "tunnel", "route", "dns", domainTunnel.GetName(), subdomain)
		if err != nil {
			return errors.NewExecutionError(errors.ErrCodeCommandFailed, "cloudflared tunnel dns create failed", err)
		}
		logger.Infof("DNS record created successfully", map[string]interface{}{
			"tunnel":    domainTunnel.GetName(),
			"subdomain": subdomain,
		})
	}
	logger.Debugf("All DNS records routed for tunnel", map[string]interface{}{
		"tunnel":     domainTunnel.GetName(),
		"subdomains": domainDNS.GetSubdomains(),
	})
	return nil
}

// GetZoneID retrieves the zone ID for a given tunnel
func (c *dnsService) GetZoneID(ctx context.Context, domainDNS *domain.DNS) (string, error) {
	// Get the zone ID using the account ID
	zones, err := c.cfApiClient.Zones.List(ctx, zones.ZoneListParams{
		Name: cloudflare.F(domainDNS.GetZone()),
	})
	if err != nil {
		return "", errors.NewExecutionError(errors.ErrCodeCommandFailed, "failed to list zones", err)
	}

	if len(zones.Result) == 0 {
		return "", errors.NewExecutionError(errors.ErrCodeCommandFailed, "no zones found for the specified name", nil)
	}

	return zones.Result[0].ID, nil
}

// GetRecords lists DNS records for a zone with optional filtering
func (c *dnsService) GetRecords(ctx context.Context, domainTunnel *domain.Tunnel, domainDNS *domain.DNS) ([]dns.RecordListResponse, error) {
	// Get the zone ID for the tunnel
	zoneId, err := c.GetZoneID(ctx, domainDNS)
	if err != nil {
		return nil, errors.NewExecutionError(errors.ErrCodeCommandFailed, "failed to get zone ID", err)
	}

	// Get the DNS content for the tunnel
	dnsContent, err := c.GetContent(ctx, domainTunnel)
	if err != nil {
		return nil, errors.NewExecutionError(errors.ErrCodeCommandFailed, "failed to get tunnel DNS content", err)
	}

	// List DNS records for the zone with the specified content
	records, err := c.cfApiClient.DNS.Records.List(ctx, dns.RecordListParams{
		ZoneID:  cloudflare.F(zoneId),
		Content: cloudflare.F(dnsContent),
	})
	if err != nil {
		return nil, errors.NewExecutionError(errors.ErrCodeCommandFailed, "failed to list DNS records", err)
	}

	return records.Result, nil
}

// DeleteRecords deletes all DNS records associated with a tunnel
func (c *dnsService) DeleteRecords(ctx context.Context, domainTunnel *domain.Tunnel, domainDNS *domain.DNS) error {

	// Get the zone ID for the tunnel
	zoneId, err := c.GetZoneID(ctx, domainDNS)
	if err != nil {
		return errors.NewExecutionError(errors.ErrCodeCommandFailed, "failed to get zone ID", err)
	}

	// Get the DNS records for the tunnel
	dnsRecords, err := c.GetRecords(ctx, domainTunnel, domainDNS)
	if err != nil {
		return errors.NewExecutionError(errors.ErrCodeCommandFailed, "failed to list DNS records", err)
	}

	// For each DNS record, delete it
	for _, record := range dnsRecords {
		_, err := c.cfApiClient.DNS.Records.Delete(ctx, record.ID, dns.RecordDeleteParams{
			ZoneID: cloudflare.F(zoneId),
		})
		if err != nil {
			return errors.NewExecutionError(errors.ErrCodeCommandFailed, fmt.Sprintf("failed to delete DNS record %s", record.Name), err)
		}
	}

	return nil
}
