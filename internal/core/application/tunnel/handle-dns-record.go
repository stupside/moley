package tunnel

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/core/ports"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/shared"
)

type DNSRecordHandler struct {
	dnsService ports.DNSService
}

type DNSRecordPayload struct {
	Zone      string         `json:"zone"`
	Tunnel    *domain.Tunnel `json:"tunnel"`
	Subdomain string         `json:"subdomain"`
}

func NewDNSRecordHandler(dnsService ports.DNSService) *DNSRecordHandler {
	return &DNSRecordHandler{
		dnsService: dnsService,
	}
}

func (h *DNSRecordHandler) Name(ctx context.Context) string {
	return "dns-record"
}

func (h *DNSRecordHandler) Up(ctx context.Context, payload any) error {
	dnsRecordPayload, err := castPayload[DNSRecordPayload](payload)
	if err != nil {
		return err
	}

	logger.Debugf("Creating DNS record", map[string]any{
		"zone":      dnsRecordPayload.Zone,
		"subdomain": dnsRecordPayload.Subdomain,
	})

	if err := h.dnsService.RouteRecord(ctx, dnsRecordPayload.Tunnel, dnsRecordPayload.Zone, dnsRecordPayload.Subdomain); err != nil {
		return shared.WrapError(err, fmt.Sprintf("failed to create DNS record for subdomain %s", dnsRecordPayload.Subdomain))
	}

	logger.Infof("DNS record created", map[string]any{
		"subdomain": dnsRecordPayload.Subdomain,
	})
	return nil
}

func (h *DNSRecordHandler) Down(ctx context.Context, payload any) error {
	dnsRecordPayload, err := castPayload[DNSRecordPayload](payload)
	if err != nil {
		return err
	}

	logger.Debugf("Deleting DNS record", map[string]any{
		"zone":      dnsRecordPayload.Zone,
		"subdomain": dnsRecordPayload.Subdomain,
	})

	if err := h.dnsService.DeleteRecord(ctx, dnsRecordPayload.Tunnel, dnsRecordPayload.Zone, dnsRecordPayload.Subdomain); err != nil {
		return shared.WrapError(err, fmt.Sprintf("failed to delete DNS record for subdomain %s", dnsRecordPayload.Subdomain))
	}

	logger.Infof("DNS record deleted", map[string]any{
		"subdomain": dnsRecordPayload.Subdomain,
	})
	return nil
}

func (h *DNSRecordHandler) Status(ctx context.Context, payload any) (domain.State, error) {
	dnsRecordPayload, err := castPayload[DNSRecordPayload](payload)
	if err != nil {
		return "", err
	}
	exists, err := h.dnsService.RecordExists(ctx, dnsRecordPayload.Tunnel, dnsRecordPayload.Zone, dnsRecordPayload.Subdomain)
	if err != nil {
		return domain.StateDown, nil
	}
	if exists {
		return domain.StateUp, nil
	}
	return domain.StateDown, nil
}
