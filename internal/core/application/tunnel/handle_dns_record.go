package tunnel

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/core/ports"
	"github.com/stupside/moley/v2/internal/platform/framework"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
)

type RecordInput struct {
	Zone       string `json:"zone"`
	Subdomain  string `json:"subdomain"`
	TunnelName string `json:"tunnel_name"`
	TunnelUUID string `json:"tunnel_uuid"` // included for hash-based change detection
	Persistent bool   `json:"persistent"`
}

func (i RecordInput) Tunnel() *domain.Tunnel {
	return &domain.Tunnel{Name: i.TunnelName, Persistent: i.Persistent}
}

type RecordOutput struct {
	Zone       string `json:"zone"`
	Subdomain  string `json:"subdomain"`
	Persistent bool   `json:"persistent"`
	TunnelName string `json:"tunnel_name"`
}

func (o RecordOutput) Tunnel() *domain.Tunnel {
	return &domain.Tunnel{Name: o.TunnelName, Persistent: o.Persistent}
}

type RecordHandler struct {
	dnsService ports.DNSService
}

var _ framework.Lifecycle[RecordInput, RecordOutput] = (*RecordHandler)(nil)

func newRecordHandler(dnsService ports.DNSService) *RecordHandler {
	return &RecordHandler{
		dnsService: dnsService,
	}
}

func (h *RecordHandler) Name() string {
	return HandlerDNSRecord
}

func (h *RecordHandler) Key(input RecordInput) string {
	return fmt.Sprintf("%s:%s", input.Zone, input.Subdomain)
}

func (h *RecordHandler) Create(ctx context.Context, input RecordInput) (RecordOutput, error) {
	logger.Debugf("Creating DNS record", map[string]any{
		"zone":      input.Zone,
		"subdomain": input.Subdomain,
	})

	if err := h.dnsService.RouteRecord(ctx, input.Tunnel(), input.Zone, input.Subdomain); err != nil {
		return RecordOutput{}, fmt.Errorf("failed to create DNS record for subdomain %s: %w", input.Subdomain, err)
	}

	logger.Infof("DNS record created", map[string]any{"subdomain": input.Subdomain})
	return RecordOutput{
		Zone:       input.Zone,
		Subdomain:  input.Subdomain,
		TunnelName: input.TunnelName,
		Persistent: input.Persistent,
	}, nil
}

func (h *RecordHandler) Destroy(ctx context.Context, output RecordOutput) error {
	logger.Debugf("Deleting DNS record", map[string]any{
		"zone":      output.Zone,
		"subdomain": output.Subdomain,
	})

	if err := h.dnsService.DeleteRecord(ctx, output.Tunnel(), output.Zone, output.Subdomain); err != nil {
		return fmt.Errorf("failed to delete DNS record for subdomain %s: %w", output.Subdomain, err)
	}

	logger.Infof("DNS record deleted", map[string]any{"subdomain": output.Subdomain})
	return nil
}

func (h *RecordHandler) Check(ctx context.Context, output RecordOutput) (framework.Status, error) {
	return h.checkExists(ctx, output.Tunnel(), output.Zone, output.Subdomain)
}

func (h *RecordHandler) Recover(ctx context.Context, input RecordInput) (RecordOutput, framework.Status, error) {
	status, err := h.checkExists(ctx, input.Tunnel(), input.Zone, input.Subdomain)

	return RecordOutput{
		Zone:       input.Zone,
		Subdomain:  input.Subdomain,
		TunnelName: input.TunnelName,
		Persistent: input.Persistent,
	}, status, err
}

func (h *RecordHandler) checkExists(ctx context.Context, tunnel *domain.Tunnel, zone, subdomain string) (framework.Status, error) {
	exists, err := h.dnsService.RecordExists(ctx, tunnel, zone, subdomain)
	if err != nil {
		return framework.StatusUnknown, fmt.Errorf("failed to check DNS record existence: %w", err)
	}
	if exists {
		return framework.StatusUp, nil
	}
	return framework.StatusDown, nil
}
