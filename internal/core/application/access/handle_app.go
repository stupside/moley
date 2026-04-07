// Package access provides the Cloudflare Access application lifecycle handler for the reconciler.
package access

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/core/ports"
	"github.com/stupside/moley/v2/internal/platform/framework"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
)

const HandlerName = "access-app"

type AppInput struct {
	Zone              string                     `json:"zone"`
	Subdomain         string                     `json:"subdomain"`
	SessionDuration   string                     `json:"session_duration"`
	Decision          domain.AccessPolicyDecision `json:"decision"`
	IdentityProviders []string                   `json:"identity_providers"`
	Emails            []string                   `json:"emails"`
	EmailDomains      []string                   `json:"email_domains"`
}

func (i AppInput) fqdn() string {
	return domain.FQDN(i.Subdomain, i.Zone)
}

type AppOutput struct {
	Zone      string `json:"zone"`
	Subdomain string `json:"subdomain"`
	AppID     string `json:"app_id"`
}

func (o AppOutput) fqdn() string {
	return domain.FQDN(o.Subdomain, o.Zone)
}

type appHandler struct {
	accessService ports.AccessService
}

var _ framework.Lifecycle[AppInput, AppOutput] = (*appHandler)(nil)

func NewHandler(accessService ports.AccessService) *appHandler {
	return &appHandler{accessService: accessService}
}

func (h *appHandler) Name() string {
	return HandlerName
}

func (h *appHandler) Key(input AppInput) string {
	return fmt.Sprintf("%s:%s", input.Zone, input.Subdomain)
}

func (h *appHandler) Create(ctx context.Context, input AppInput) (AppOutput, error) {
	fqdn := input.fqdn()

	logger.Debugf("Creating Access Application", map[string]any{"domain": fqdn})

	appID, err := h.accessService.CreateApplication(ctx, ports.AccessApplicationParams{
		Name:              fmt.Sprintf("moley-%s", fqdn),
		Domain:            fqdn,
		SessionDuration:   input.SessionDuration,
		Decision:          input.Decision,
		IdentityProviders: input.IdentityProviders,
		Emails:            input.Emails,
		EmailDomains:      input.EmailDomains,
	})
	if err != nil {
		return AppOutput{}, fmt.Errorf("failed to create Access Application for %s: %w", fqdn, err)
	}

	logger.Infof("Access Application created", map[string]any{"domain": fqdn, "app_id": appID})
	return AppOutput{
		Zone:      input.Zone,
		Subdomain: input.Subdomain,
		AppID:     appID,
	}, nil
}

func (h *appHandler) Destroy(ctx context.Context, output AppOutput) error {
	fqdn := output.fqdn()

	logger.Debugf("Deleting Access Application", map[string]any{"domain": fqdn, "app_id": output.AppID})

	if err := h.accessService.DeleteApplication(ctx, output.AppID); err != nil {
		return fmt.Errorf("failed to delete Access Application for %s: %w", fqdn, err)
	}

	logger.Infof("Access Application deleted", map[string]any{"domain": fqdn})
	return nil
}

func (h *appHandler) Check(ctx context.Context, output AppOutput) (framework.Status, error) {
	_, exists, err := h.accessService.FindApplication(ctx, output.fqdn())
	if err != nil {
		return framework.StatusUnknown, fmt.Errorf("failed to check Access Application: %w", err)
	}
	if exists {
		return framework.StatusUp, nil
	}
	return framework.StatusDown, nil
}

func (h *appHandler) Recover(ctx context.Context, input AppInput) (AppOutput, framework.Status, error) {
	appID, exists, err := h.accessService.FindApplication(ctx, input.fqdn())
	if err != nil {
		return AppOutput{}, framework.StatusUnknown, err
	}
	if !exists {
		return AppOutput{}, framework.StatusDown, nil
	}
	return AppOutput{
		Zone:      input.Zone,
		Subdomain: input.Subdomain,
		AppID:     appID,
	}, framework.StatusUp, nil
}
