package tunnel

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/domain"
	logger "github.com/stupside/moley/v2/internal/platform/logging"
	framework "github.com/stupside/moley/v2/internal/platform/orchestration"
)

type TunnelCreator interface {
	Create(ctx context.Context, tunnel *domain.Tunnel) (string, error)
	Delete(ctx context.Context, tunnel *domain.Tunnel) error
	Exists(ctx context.Context, tunnel *domain.Tunnel) (bool, error)
	GetID(ctx context.Context, tunnel *domain.Tunnel) (string, error)
}

const CreateHandlerName = "tunnel-create"

type CreateInput struct {
	Name       string `json:"name"`
	Persistent bool   `json:"persistent"`
}

func (i CreateInput) tunnel() *domain.Tunnel {
	return &domain.Tunnel{Name: i.Name, Persistent: i.Persistent}
}

type CreateOutput struct {
	Name       string `json:"name"`
	Persistent bool   `json:"persistent"`
	TunnelUUID string `json:"tunnel_uuid"`
}

func (o CreateOutput) tunnel() *domain.Tunnel {
	return &domain.Tunnel{Name: o.Name, Persistent: o.Persistent}
}

type createHandler struct {
	tunnelService TunnelCreator
}

var _ framework.Lifecycle[CreateInput, CreateOutput] = (*createHandler)(nil)

func NewCreateHandler(tunnelService TunnelCreator) *createHandler {
	return &createHandler{
		tunnelService: tunnelService,
	}
}

func (h *createHandler) Name() string {
	return CreateHandlerName
}

func (h *createHandler) Key(input CreateInput) string {
	return input.Name
}

func (h *createHandler) Create(ctx context.Context, input CreateInput) (CreateOutput, error) {
	logger.Debug("Creating tunnel")

	tunnelUUID, err := h.tunnelService.Create(ctx, input.tunnel())
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

func (h *createHandler) Destroy(ctx context.Context, output CreateOutput) error {
	if output.Persistent {
		logger.Info("Tunnel is persistent, skipping deletion")
		return nil
	}

	logger.Debug("Deleting tunnel")

	if err := h.tunnelService.Delete(ctx, output.tunnel()); err != nil {
		return fmt.Errorf("tunnel delete failed: %w", err)
	}

	logger.Info("Tunnel deleted")
	return nil
}

func (h *createHandler) Check(ctx context.Context, output CreateOutput) (framework.Status, error) {
	return h.checkExists(ctx, output.tunnel())
}

func (h *createHandler) Recover(ctx context.Context, input CreateInput) (CreateOutput, framework.Status, error) {
	tunnel := input.tunnel()

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

func (h *createHandler) checkExists(ctx context.Context, tunnel *domain.Tunnel) (framework.Status, error) {
	exists, err := h.tunnelService.Exists(ctx, tunnel)
	if err != nil {
		return framework.StatusUnknown, fmt.Errorf("failed to check tunnel existence: %w", err)
	}
	if exists {
		return framework.StatusUp, nil
	}
	return framework.StatusDown, nil
}
