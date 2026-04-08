package cloudflare

import (
	"context"
	"fmt"
	"time"

	cfgo "github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/zero_trust"

	"github.com/stupside/moley/v2/internal/domain"
	accessusecase "github.com/stupside/moley/v2/internal/features/access/usecase"
	logger "github.com/stupside/moley/v2/internal/platform/logging"
)

type AccessService struct {
	client    *cfgo.Client
	dryRun    bool
	accountID string
}

func NewAccessService(client *cfgo.Client, accountID string, dryRun bool) *AccessService {
	return &AccessService{
		client:    client,
		accountID: accountID,
		dryRun:    dryRun,
	}
}

func (s *AccessService) CreateApplication(ctx context.Context, params accessusecase.AccessApplicationParams) (string, error) {
	if s.dryRun {
		logger.Debug("Dry run: skipping Access Application creation")
		return "dry-run-access-app", nil
	}

	sessionDuration := params.SessionDuration
	d, err := time.ParseDuration(sessionDuration)
	if err != nil {
		return "", fmt.Errorf("invalid session_duration %q: %w", sessionDuration, err)
	}
	sessionDuration = d.String()

	body := zero_trust.AccessApplicationNewParamsBodySelfHostedApplication{
		Name:            cfgo.F(params.Name),
		Domain:          cfgo.F(params.Domain),
		Type:            cfgo.F(string(zero_trust.ApplicationTypeSelfHosted)),
		SessionDuration: cfgo.F(sessionDuration),
		Policies: cfgo.F([]zero_trust.AccessApplicationNewParamsBodySelfHostedApplicationPolicyUnion{
			zero_trust.AccessApplicationNewParamsBodySelfHostedApplicationPoliciesObject{
				Name:     cfgo.F(fmt.Sprintf("%s-policy", params.Name)),
				Decision: cfgo.F(mapDecision(params.Decision)),
				Include:  cfgo.F(buildIncludeRules(params.Emails, params.EmailDomains)),
			},
		}),
	}

	if len(params.IdentityProviders) > 0 {
		idpIDs, err := s.resolveIdentityProviders(ctx, params.IdentityProviders)
		if err != nil {
			return "", fmt.Errorf("failed to resolve identity providers: %w", err)
		}
		body.AllowedIdPs = cfgo.F(idpIDs)
	}

	app, err := s.client.ZeroTrust.Access.Applications.New(ctx, zero_trust.AccessApplicationNewParams{
		AccountID: cfgo.F(s.accountID),
		Body:      body,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create Access Application: %w", err)
	}

	logger.Debugf("Access Application created", map[string]any{
		"app_id": app.ID,
		"domain": params.Domain,
	})

	return app.ID, nil
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
		app := pager.Current()
		if app.Domain == domain {
			return app.ID, true, nil
		}
	}
	if err := pager.Err(); err != nil {
		return "", false, fmt.Errorf("failed to list Access Applications: %w", err)
	}

	return "", false, nil
}

func (s *AccessService) resolveIdentityProviders(ctx context.Context, providerTypes []string) ([]zero_trust.AllowedIdPsParam, error) {
	pager := s.client.ZeroTrust.IdentityProviders.ListAutoPaging(ctx, zero_trust.IdentityProviderListParams{
		AccountID: cfgo.F(s.accountID),
	})

	typeSet := make(map[string]struct{}, len(providerTypes))
	for _, t := range providerTypes {
		typeSet[t] = struct{}{}
	}

	matched := make(map[string]bool, len(providerTypes))
	var ids []zero_trust.AllowedIdPsParam
	for pager.Next() {
		idp := pager.Current()
		if _, ok := typeSet[string(idp.Type)]; ok {
			ids = append(ids, idp.ID)
			matched[string(idp.Type)] = true
		}
	}
	if err := pager.Err(); err != nil {
		return nil, fmt.Errorf("failed to list identity providers: %w", err)
	}

	for _, t := range providerTypes {
		if !matched[t] {
			logger.Warnf("Identity provider not found on account", map[string]any{"type": t})
		}
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("none of the configured identity providers %v are available on this account", providerTypes)
	}

	return ids, nil
}

func mapDecision(d domain.AccessPolicyDecision) zero_trust.Decision {
	switch d {
	case domain.AccessPolicyDecisionBypass:
		return zero_trust.DecisionBypass
	default:
		return zero_trust.DecisionAllow
	}
}

func buildIncludeRules(emails []string, emailDomains []string) []zero_trust.AccessRuleUnionParam {
	rules := make([]zero_trust.AccessRuleUnionParam, 0, len(emails)+len(emailDomains))

	for _, email := range emails {
		rules = append(rules, zero_trust.EmailRuleParam{
			Email: cfgo.F(zero_trust.EmailRuleEmailParam{
				Email: cfgo.F(email),
			}),
		})
	}

	for _, d := range emailDomains {
		rules = append(rules, zero_trust.DomainRuleParam{
			EmailDomain: cfgo.F(zero_trust.DomainRuleEmailDomainParam{
				Domain: cfgo.F(d),
			}),
		})
	}

	return rules
}
