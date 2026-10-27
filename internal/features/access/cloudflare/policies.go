package cloudflare

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/domain"

	cfgo "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/zero_trust"
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
	if _, err := s.client.ZeroTrust.Access.Policies.Delete(ctx, policyID, zero_trust.AccessPolicyDeleteParams{
		AccountID: cfgo.F(s.accountID),
	}); err != nil {
		return fmt.Errorf("failed to delete policy %s: %w", policyID, err)
	}
	return nil
}

func (s *AccessService) FindPolicy(ctx context.Context, name string) (string, bool, error) {
	if s.dryRun {
		return "dry-run-policy", true, nil
	}
	pager := s.client.ZeroTrust.Access.Policies.ListAutoPaging(ctx, zero_trust.AccessPolicyListParams{
		AccountID: cfgo.F(s.accountID),
	})
	for pager.Next() {
		if p := pager.Current(); p.Name == name {
			return p.ID, true, nil
		}
	}
	if err := pager.Err(); err != nil {
		return "", false, fmt.Errorf("failed to list policies: %w", err)
	}
	return "", false, nil
}
