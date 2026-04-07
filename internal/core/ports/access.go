package ports

import (
	"context"

	"github.com/stupside/moley/v2/internal/core/domain"
)

type AccessApplicationParams struct {
	Name              string
	Domain            string
	SessionDuration   string
	Decision          domain.AccessPolicyDecision
	IdentityProviders []string
	Emails            []string
	EmailDomains      []string
}

type AccessService interface {
	CreateApplication(ctx context.Context, params AccessApplicationParams) (string, error)
	DeleteApplication(ctx context.Context, appID string) error
	FindApplication(ctx context.Context, domain string) (string, bool, error)
}
