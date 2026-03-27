package tunnel

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/core/ports"
	"github.com/stupside/moley/v2/internal/platform/framework"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
)

type CreateInput struct {
	Name       string `json:"name"`
	Persistent bool   `json:"persistent"`
}

func (i CreateInput) Tunnel() *domain.Tunnel {
	return &domain.Tunnel{Name: i.Name, Persistent: i.Persistent}
}

type CreateOutput struct {
	Name       string `json:"name"`
	Persistent bool   `json:"persistent"`
	TunnelUUID string `json:"tunnel_uuid"`
}

func (o CreateOutput) Tunnel() *domain.Tunnel {
	return &domain.Tunnel{Name: o.Name, Persistent: o.Persistent}
}

type CreateHandler struct {
	tunnelService ports.TunnelService
}

var _ framework.Lifecycle[CreateInput, CreateOutput] = (*CreateHandler)(nil)

func newCreateHandler(tunnelService ports.TunnelService) *CreateHandler {
	return &CreateHandler{
		tunnelService: tunnelService,
	}
}

func (h *CreateHandler) Name() string {
	return HandlerTunnelCreate
}

func (h *CreateHandler) Key(input CreateInput) string {
	return input.Name
}

func (h *CreateHandler) Create(ctx context.Context, input CreateInput) (CreateOutput, error) {
	logger.Debug("Creating tunnel")

	tunnelUUID, err := h.tunnelService.Create(ctx, input.Tunnel())
	if err != nil {
		return CreateOutput{}, fmt.Errorf("tunnel create failed: %w", err)
	}

	logger.Info("Tunnel created")
	return CreateOutput{
		Name:       input.Name,
		Persistent: input.Persistent,
		TunnelUUID: tunnelUUID,
	}, nil
}

func (h *CreateHandler) Destroy(ctx context.Context, output CreateOutput) error {
	if output.Persistent {
		logger.Info("Tunnel is persistent, skipping deletion")
		return nil
	}

	logger.Debug("Deleting tunnel")

	if err := h.tunnelService.Delete(ctx, output.Tunnel()); err != nil {
		return fmt.Errorf("tunnel delete failed: %w", err)
	}

	logger.Info("Tunnel deleted")
	return nil
}

func (h *CreateHandler) Check(ctx context.Context, output CreateOutput) (framework.Status, error) {
	return h.checkExists(ctx, output.Tunnel())
}

func (h *CreateHandler) Recover(ctx context.Context, input CreateInput) (CreateOutput, framework.Status, error) {
	tunnel := input.Tunnel()

	status, err := h.checkExists(ctx, tunnel)
	if status != framework.StatusUp {
		return CreateOutput{}, status, err
	}

	tunnelUUID, err := h.tunnelService.GetID(ctx, tunnel)
	if err != nil {
		return CreateOutput{}, framework.StatusUnknown, fmt.Errorf("failed to get tunnel ID during recovery: %w", err)
	}

	return CreateOutput{
		Name:       input.Name,
		Persistent: input.Persistent,
		TunnelUUID: tunnelUUID,
	}, framework.StatusUp, nil
}

func (h *CreateHandler) checkExists(ctx context.Context, tunnel *domain.Tunnel) (framework.Status, error) {
	exists, err := h.tunnelService.Exists(ctx, tunnel)
	if err != nil {
		return framework.StatusUnknown, fmt.Errorf("failed to check tunnel existence: %w", err)
	}
	if exists {
		return framework.StatusUp, nil
	}
	return framework.StatusDown, nil
}
