package access

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/domain"
	logger "github.com/stupside/moley/v2/internal/platform/logging"
	framework "github.com/stupside/moley/v2/internal/platform/orchestration"
)

const PolicyHandlerName = "access-policies"

type PolicyManager interface {
	CreatePolicy(ctx context.Context, policy domain.Policy) (string, error)
	DeletePolicy(ctx context.Context, policyID string) error
	FindPolicy(ctx context.Context, name string) (string, bool, error)
}

type PolicyInput struct {
	Policy domain.Policy `json:"policy"`
}

type PolicyOutput struct {
	Name     string `json:"name"`
	PolicyID string `json:"policy_id"`
}

type policyHandler struct {
	policyService PolicyManager
}

var _ framework.Lifecycle[PolicyInput, PolicyOutput] = (*policyHandler)(nil)

func NewPolicyHandler(policyService PolicyManager) *policyHandler {
	return &policyHandler{policyService: policyService}
}

func (h *policyHandler) Name() string { return PolicyHandlerName }

func (h *policyHandler) Key(input PolicyInput) string {
	return input.Policy.Name
}

func (h *policyHandler) Create(ctx context.Context, input PolicyInput) (PolicyOutput, error) {
	id, err := h.policyService.CreatePolicy(ctx, input.Policy)
	if err != nil {
		return PolicyOutput{}, fmt.Errorf("failed to create policy %q: %w", input.Policy.Name, err)
	}
	logger.Infof("Access policy created", map[string]any{"name": input.Policy.Name, "id": id})
	return PolicyOutput{Name: input.Policy.Name, PolicyID: id}, nil
}

func (h *policyHandler) Destroy(ctx context.Context, output PolicyOutput) error {
	if err := h.policyService.DeletePolicy(ctx, output.PolicyID); err != nil {
		return fmt.Errorf("failed to delete policy %q: %w", output.Name, err)
	}
	logger.Infof("Access policy deleted", map[string]any{"name": output.Name})
	return nil
}

func (h *policyHandler) Check(ctx context.Context, output PolicyOutput) (framework.Status, error) {
	id, found, err := h.policyService.FindPolicy(ctx, output.Name)
	if err != nil {
		return framework.StatusUnknown, err
	}
	if !found || id != output.PolicyID {
		return framework.StatusDown, nil
	}
	return framework.StatusUp, nil
}

func (h *policyHandler) Recover(ctx context.Context, input PolicyInput) (PolicyOutput, framework.Status, error) {
	id, found, err := h.policyService.FindPolicy(ctx, input.Policy.Name)
	if err != nil {
		return PolicyOutput{}, framework.StatusUnknown, err
	}
	if !found {
		return PolicyOutput{}, framework.StatusDown, nil
	}
	return PolicyOutput{Name: input.Policy.Name, PolicyID: id}, framework.StatusUp, nil
}
