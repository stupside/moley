package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"

	cfgo "github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/zero_trust"

	accessusecase "github.com/stupside/moley/v2/internal/features/access/usecase"
	logger "github.com/stupside/moley/v2/internal/platform/logging"
)

type AccessService struct {
	client    *cfgo.Client
	dryRun    bool
	accountID string
}

func NewAccessService(client *cfgo.Client, accountID string, dryRun bool) *AccessService {
	return &AccessService{client: client, accountID: accountID, dryRun: dryRun}
}

func (s *AccessService) appsPath() string {
	return fmt.Sprintf("accounts/%s/access/apps", s.accountID)
}

type policyRef struct {
	ID         string `json:"id"`
	Precedence int    `json:"precedence"`
}

type accessAppBody struct {
	Name        string                     `json:"name"`
	Domain      string                     `json:"domain"`
	Type        zero_trust.ApplicationType `json:"type"`
	AllowedIdPs []string                   `json:"allowed_idps"`
	Policies    []policyRef                `json:"policies"`
	Extra       map[string]any             `json:"-"`
}

func (b accessAppBody) MarshalJSON() ([]byte, error) {
	m := make(map[string]any, len(b.Extra)+5)
	maps.Copy(m, b.Extra)
	m["name"] = b.Name
	m["domain"] = b.Domain
	m["type"] = b.Type
	if len(b.AllowedIdPs) > 0 {
		m["allowed_idps"] = b.AllowedIdPs
	}
	if len(b.Policies) > 0 {
		m["policies"] = b.Policies
	}
	return json.Marshal(m)
}

func (s *AccessService) CreateApplication(ctx context.Context, params accessusecase.AccessApplicationParams) (string, error) {
	if s.dryRun {
		logger.Debug("Dry run: skipping Access Application creation")
		return "dry-run-access-app", nil
	}

	body := accessAppBody{
		Name:   params.Name,
		Domain: params.Domain,
		Type:   zero_trust.ApplicationTypeSelfHosted,
		Extra:  params.Access.Raw,
	}

	if len(params.Access.Providers) > 0 {
		ids, err := s.resolveIdentityProviders(ctx, params.Access.Providers)
		if err != nil {
			return "", fmt.Errorf("failed to resolve identity providers: %w", err)
		}
		body.AllowedIdPs = ids
	}

	if len(params.PolicyIDs) > 0 {
		body.Policies = make([]policyRef, len(params.PolicyIDs))
		for i, id := range params.PolicyIDs {
			body.Policies[i] = policyRef{ID: id, Precedence: i + 1}
		}
	}

	var env struct {
		Result struct {
			ID string `json:"id"`
		} `json:"result"`
	}
	if err := s.client.Post(ctx, s.appsPath(), body, &env); err != nil {
		return "", fmt.Errorf("failed to create Access Application: %w", err)
	}

	logger.Debugf("Access Application created", map[string]any{"app_id": env.Result.ID, "domain": params.Domain})
	return env.Result.ID, nil
}

func (s *AccessService) DeleteApplication(ctx context.Context, appID string) error {
	if s.dryRun {
		logger.Debug("Dry run: skipping Access Application deletion")
		return nil
	}
	_, err := s.client.ZeroTrust.Access.Applications.Delete(ctx, appID, zero_trust.AccessApplicationDeleteParams{
		AccountID: cfgo.F(s.accountID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete Access Application %s: %w", appID, err)
	}
	return nil
}

func (s *AccessService) FindApplication(ctx context.Context, domain string) (string, bool, error) {
	if s.dryRun {
		return "dry-run-access-app", true, nil
	}
	pager := s.client.ZeroTrust.Access.Applications.ListAutoPaging(ctx, zero_trust.AccessApplicationListParams{
		AccountID: cfgo.F(s.accountID),
	})
	for pager.Next() {
		if app := pager.Current(); app.Domain == domain {
			return app.ID, true, nil
		}
	}
	if err := pager.Err(); err != nil {
		return "", false, fmt.Errorf("failed to list Access Applications: %w", err)
	}
	return "", false, nil
}

func (s *AccessService) resolveIdentityProviders(ctx context.Context, types []string) ([]string, error) {
	want := make(map[string]struct{}, len(types))
	for _, t := range types {
		want[t] = struct{}{}
	}

	pager := s.client.ZeroTrust.IdentityProviders.ListAutoPaging(ctx, zero_trust.IdentityProviderListParams{
		AccountID: cfgo.F(s.accountID),
	})

	found := make(map[string]struct{}, len(types))
	var ids []string
	for pager.Next() {
		idp := pager.Current()
		t := string(idp.Type)
		if _, inWant := want[t]; inWant {
			if _, alreadyFound := found[t]; !alreadyFound {
				found[t] = struct{}{}
				ids = append(ids, idp.ID)
			}
		}
	}
	if err := pager.Err(); err != nil {
		return nil, fmt.Errorf("failed to list identity providers: %w", err)
	}

	for _, t := range types {
		if _, ok := found[t]; !ok {
			logger.Warnf("Identity provider not found on account", map[string]any{"type": t})
		}
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("none of the configured identity providers %v are available on this account", types)
	}
	return ids, nil
}
