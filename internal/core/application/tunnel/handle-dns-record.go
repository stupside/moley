package tunnel

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/core/ports"
	"github.com/stupside/moley/v2/internal/platform/framework"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/shared"
)

// DNSRecordConfig represents the desired DNS record configuration
type DNSRecordConfig struct {
	Zone      string         `json:"zone"`
	Tunnel    *domain.Tunnel `json:"tunnel"`
	Subdomain string         `json:"subdomain"`
}

// DNSRecordState represents the runtime state of a created DNS record
type DNSRecordState struct {
	Zone      string         `json:"zone"`
	Tunnel    *domain.Tunnel `json:"tunnel"`
	Subdomain string         `json:"subdomain"`
}

// DNSRecordHandler manages DNS record lifecycle with type safety
type DNSRecordHandler struct {
	dnsService ports.DNSService
}

// Ensure DNSRecordHandler implements the typed interface
var _ framework.ResourceHandler[DNSRecordConfig, DNSRecordState] = (*DNSRecordHandler)(nil)

func newDNSRecordHandler(dnsService ports.DNSService) *DNSRecordHandler {
	return &DNSRecordHandler{
		dnsService: dnsService,
	}
}

func (h *DNSRecordHandler) Name() string {
	return "dns-record"
}

func (h *DNSRecordHandler) Create(ctx context.Context, config DNSRecordConfig) (DNSRecordState, error) {
	logger.Debugf("Creating DNS record", map[string]any{
		"zone":      config.Zone,
		"subdomain": config.Subdomain,
	})

	if err := h.dnsService.RouteRecord(ctx, config.Tunnel, config.Zone, config.Subdomain); err != nil {
		return DNSRecordState{}, shared.WrapError(err, fmt.Sprintf("failed to create DNS record for subdomain %s", config.Subdomain))
	}

	state := DNSRecordState{
		Zone:      config.Zone,
		Tunnel:    config.Tunnel,
		Subdomain: config.Subdomain,
	}

	logger.Infof("DNS record created", map[string]any{
		"subdomain": config.Subdomain,
	})

	return state, nil
}

func (h *DNSRecordHandler) Destroy(ctx context.Context, state DNSRecordState) error {
	logger.Debugf("Deleting DNS record", map[string]any{
		"zone":      state.Zone,
		"subdomain": state.Subdomain,
	})

	if err := h.dnsService.DeleteRecord(ctx, state.Tunnel, state.Zone, state.Subdomain); err != nil {
		return shared.WrapError(err, fmt.Sprintf("failed to delete DNS record for subdomain %s", state.Subdomain))
	}

	logger.Infof("DNS record deleted", map[string]any{
		"subdomain": state.Subdomain,
	})

	return nil
}

func (h *DNSRecordHandler) Status(ctx context.Context, state DNSRecordState) (domain.State, error) {
	exists, err := h.dnsService.RecordExists(ctx, state.Tunnel, state.Zone, state.Subdomain)
	if err != nil {
		return domain.StateDown, shared.WrapError(err, "failed to check DNS record existence")
	}

	if exists {
		return domain.StateUp, nil
	}
	return domain.StateDown, nil
}

func (h *DNSRecordHandler) Equals(a, b DNSRecordConfig) bool {
	return a.Zone == b.Zone &&
		a.Subdomain == b.Subdomain &&
		a.Tunnel.ID == b.Tunnel.ID
}
