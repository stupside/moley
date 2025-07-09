package steps

import (
	"context"
	"fmt"

	"github.com/stupside/moley/internal/domain"
	"github.com/stupside/moley/internal/logger"
	"github.com/stupside/moley/internal/services"
	"github.com/stupside/moley/internal/shared"
)

// DNSStep handles DNS record operations
type DNSStep struct {
	// dns holds the DNS configuration for the tunnel
	dns *domain.DNS
	// tunnel holds the tunnel information for which DNS records are managed
	tunnel *domain.Tunnel
	// dnsService is the service responsible for DNS operations
	dnsService services.DNSService
}

// NewDNSStep creates a new DNS step
func NewDNSStep(dnsService services.DNSService, tunnel *domain.Tunnel, dns *domain.DNS) *DNSStep {
	return &DNSStep{
		dns:        dns,
		tunnel:     tunnel,
		dnsService: dnsService,
	}
}

// Name returns the step name
func (d *DNSStep) Name() string {
	return "DNS"
}

// Up creates DNS records for all hostnames
func (d *DNSStep) Up(ctx context.Context) error {
	apps := d.dns.GetApps()
	zone := d.dns.GetZone()

	for _, app := range apps {
		hostname := fmt.Sprintf("%s.%s", app.Expose.Subdomain, zone)
		logger.Info(fmt.Sprintf("Setting up DNS record: %s → %s", hostname, app.Target.GetTargetURL()))
	}

	if err := d.dnsService.Route(ctx, d.tunnel, d.dns); err != nil {
		return shared.WrapError(err, "cloudflared tunnel dns create failed")
	}

	logger.Info(fmt.Sprintf("DNS records created successfully for %d app(s)", len(apps)))
	return nil
}

// Down removes DNS records for all hostnames
func (d *DNSStep) Down(ctx context.Context) error {
	logger.Debug(fmt.Sprintf("Deleting DNS records for tunnel: %s", d.tunnel.GetName()))

	if err := d.dnsService.DeleteRecords(ctx, d.tunnel, d.dns); err != nil {
		return shared.WrapError(err, "failed to delete DNS records")
	}

	logger.Debug("DNS records deleted successfully")
	return nil
}
