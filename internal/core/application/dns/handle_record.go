// Package dns provides the DNS record lifecycle handler for the reconciler.
package dns

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/core/ports"
	"github.com/stupside/moley/v2/internal/platform/framework"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
)

const HandlerName = "dns-record"

type RecordInput struct {
	Zone       string `json:"zone"`
	Subdomain  string `json:"subdomain"`
	TunnelName string `json:"tunnel_name"`
	TunnelUUID string `json:"tunnel_uuid"` // included for hash-based change detection
	Persistent bool   `json:"persistent"`
}

func (i RecordInput) tunnel() *domain.Tunnel {
	return &domain.Tunnel{Name: i.TunnelName, Persistent: i.Persistent}
}

type RecordOutput struct {
	Zone       string `json:"zone"`
	Subdomain  string `json:"subdomain"`
	Persistent bool   `json:"persistent"`
	TunnelName string `json:"tunnel_name"`
}

func (o RecordOutput) tunnel() *domain.Tunnel {
	return &domain.Tunnel{Name: o.TunnelName, Persistent: o.Persistent}
}

type recordHandler struct {
	dnsService ports.DNSService
}

var _ framework.Lifecycle[RecordInput, RecordOutput] = (*recordHandler)(nil)

func NewHandler(dnsService ports.DNSService) *recordHandler {
	return &recordHandler{dnsService: dnsService}
}

func (h *recordHandler) Name() string {
	return HandlerName
}

func (h *recordHandler) Key(input RecordInput) string {
	return fmt.Sprintf("%s:%s", input.Zone, input.Subdomain)
}

func (h *recordHandler) Create(ctx context.Context, input RecordInput) (RecordOutput, error) {
	logger.Debugf("Creating DNS record", map[string]any{
		"zone":      input.Zone,
		"subdomain": input.Subdomain,
	})

	if err := h.dnsService.RouteRecord(ctx, input.tunnel(), input.Zone, input.Subdomain); err != nil {
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

func (h *recordHandler) Destroy(ctx context.Context, output RecordOutput) error {
	logger.Debugf("Deleting DNS record", map[string]any{
		"zone":      output.Zone,
		"subdomain": output.Subdomain,
	})

	if err := h.dnsService.DeleteRecord(ctx, output.tunnel(), output.Zone, output.Subdomain); err != nil {
		return fmt.Errorf("failed to delete DNS record for subdomain %s: %w", output.Subdomain, err)
	}

	logger.Infof("DNS record deleted", map[string]any{"subdomain": output.Subdomain})
	return nil
}

func (h *recordHandler) Check(ctx context.Context, output RecordOutput) (framework.Status, error) {
	return h.checkExists(ctx, output.tunnel(), output.Zone, output.Subdomain)
}

func (h *recordHandler) Recover(ctx context.Context, input RecordInput) (RecordOutput, framework.Status, error) {
	status, err := h.checkExists(ctx, input.tunnel(), input.Zone, input.Subdomain)
	return RecordOutput{
		Zone:       input.Zone,
		Subdomain:  input.Subdomain,
		TunnelName: input.TunnelName,
		Persistent: input.Persistent,
	}, status, err
}

func (h *recordHandler) checkExists(ctx context.Context, tunnel *domain.Tunnel, zone, subdomain string) (framework.Status, error) {
	exists, err := h.dnsService.RecordExists(ctx, tunnel, zone, subdomain)
	if err != nil {
		return framework.StatusUnknown, fmt.Errorf("failed to check DNS record existence: %w", err)
	}
	if exists {
		return framework.StatusUp, nil
	}
	return framework.StatusDown, nil
}
