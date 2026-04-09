package cloudflare

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/domain"
)

func (s *AccessService) policiesPath() string {
	return fmt.Sprintf("accounts/%s/access/policies", s.accountID)
}

func (s *AccessService) CreatePolicy(ctx context.Context, policy domain.Policy) (string, error) {
	if s.dryRun {
		return "dry-run-policy", nil
	}
	var env struct {
		Result struct {
			ID string `json:"id"`
		} `json:"result"`
	}
	if err := s.client.Post(ctx, s.policiesPath(), policy, &env); err != nil {
		return "", fmt.Errorf("failed to create policy: %w", err)
	}
	return env.Result.ID, nil
}

func (s *AccessService) DeletePolicy(ctx context.Context, policyID string) error {
	if s.dryRun {
		return nil
	}
	if err := s.client.Delete(ctx, s.policiesPath()+"/"+policyID, nil, nil); err != nil {
		return fmt.Errorf("failed to delete policy %s: %w", policyID, err)
	}
	return nil
}

func (s *AccessService) FindPolicy(ctx context.Context, name string) (string, bool, error) {
	if s.dryRun {
		return "dry-run-policy", true, nil
	}
	var env struct {
		Result []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"result"`
	}
	if err := s.client.Get(ctx, s.policiesPath(), nil, &env); err != nil {
		return "", false, fmt.Errorf("failed to list policies: %w", err)
	}
	for _, p := range env.Result {
		if p.Name == name {
			return p.ID, true, nil
		}
	}
	return "", false, nil
}
