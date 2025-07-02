package steps

import (
	"context"
	"moley/internal/domain"
	"moley/internal/logger"
	"moley/internal/services"

	"moley/internal/errors"
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
	logger.Debugf("Creating DNS records for tunnel", map[string]interface{}{
		"tunnel":     d.tunnel.GetName(),
		"subdomains": d.dns.GetSubdomains(),
	})
	if err := d.dnsService.Route(ctx, d.tunnel, d.dns); err != nil {
		logger.Warnf("DNS record creation failed", map[string]interface{}{
			"tunnel":     d.tunnel.GetName(),
			"subdomains": d.dns.GetSubdomains(),
			"error":      err.Error(),
		})
		return errors.NewExecutionError(errors.ErrCodeCommandFailed, "cloudflared tunnel dns create failed", err)
	}
	logger.Infof("DNS records created successfully", map[string]interface{}{
		"tunnel":     d.tunnel.GetName(),
		"subdomains": d.dns.GetSubdomains(),
	})
	return nil
}

// Down removes DNS records for all hostnames
func (d *DNSStep) Down(ctx context.Context) error {
	logger.Debugf("Deleting DNS records for tunnel", map[string]interface{}{
		"tunnel":     d.tunnel.GetName(),
		"subdomains": d.dns.GetSubdomains(),
	})
	if err := d.dnsService.DeleteRecords(ctx, d.tunnel, d.dns); err != nil {
		logger.Warnf("DNS record deletion failed", map[string]interface{}{
			"tunnel": d.tunnel.GetName(),
			"error":  err.Error(),
		})
		return errors.NewExecutionError(errors.ErrCodeCommandFailed, "failed to delete DNS records", err)
	}
	logger.Infof("DNS records deleted successfully", map[string]interface{}{
		"tunnel": d.tunnel.GetName(),
	})
	return nil
}
