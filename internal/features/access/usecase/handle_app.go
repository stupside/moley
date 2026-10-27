// Package access provides the Cloudflare Access application lifecycle handler for the reconciler.
package access

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/domain"
	logger "github.com/stupside/moley/v2/internal/platform/logging"
	framework "github.com/stupside/moley/v2/internal/platform/orchestration"
)

type AccessApplicationParams struct {
	Name      string
	Domain    string
	Access    domain.AccessConfig
	PolicyIDs []string
}

type AccessManager interface {
	CreateApplication(ctx context.Context, params AccessApplicationParams) (string, error)
	DeleteApplication(ctx context.Context, appID string) error
	FindApplication(ctx context.Context, domain string) (string, bool, error)
}

const HandlerName = "access-app"

type AppInput struct {
	Zone      string              `json:"zone"`
	Subdomain string              `json:"subdomain"`
	Path      string              `json:"path,omitempty"`
	Access    domain.AccessConfig `json:"access"`
	PolicyIDs []string            `json:"policy_ids,omitempty"`
}

func (i AppInput) host() string {
	return domain.ExposeConfig{Subdomain: i.Subdomain, Path: i.Path}.Host(i.Zone)
}

type AppOutput struct {
	Zone      string `json:"zone"`
	AppID     string `json:"app_id"`
	Subdomain string `json:"subdomain"`
	Path      string `json:"path,omitempty"`
}

func (o AppOutput) Host() string {
	return domain.ExposeConfig{Subdomain: o.Subdomain, Path: o.Path}.Host(o.Zone)
}

type appHandler struct {
	accessService AccessManager
}

var _ framework.Lifecycle[AppInput, AppOutput] = (*appHandler)(nil)

func NewHandler(accessService AccessManager) *appHandler {
	return &appHandler{accessService: accessService}
}

func (h *appHandler) Name() string {
	return HandlerName
}

func (h *appHandler) Key(input AppInput) string {
	return AppKey(input.Zone, input.Subdomain, input.Path)
}

func AppKey(zone, subdomain, path string) string {
	return fmt.Sprintf("%s:%s%s", zone, subdomain, path)
}

func (h *appHandler) Create(ctx context.Context, input AppInput) (AppOutput, error) {
	host := input.host()
	logger.Debugf("Creating Access Application", map[string]any{"domain": host})

	appID, err := h.accessService.CreateApplication(ctx, AccessApplicationParams{
		Name:      fmt.Sprintf("moley-%s", host),
		Domain:    host,
		Access:    input.Access,
		PolicyIDs: input.PolicyIDs,
	})
	if err != nil {
		return AppOutput{}, fmt.Errorf("failed to create Access Application for %s: %w", host, err)
	}

	logger.Infof("Access Application created", map[string]any{"domain": host, "app_id": appID})
	return AppOutput{
		Zone:      input.Zone,
		Subdomain: input.Subdomain,
		Path:      input.Path,
		AppID:     appID,
	}, nil
}
func (h *appHandler) Destroy(ctx context.Context, output AppOutput) error {
	host := output.Host()
	logger.Debugf("Deleting Access Application", map[string]any{"domain": host, "app_id": output.AppID})

	if err := h.accessService.DeleteApplication(ctx, output.AppID); err != nil {
		return fmt.Errorf("failed to delete Access Application for %s: %w", host, err)
	}

	logger.Infof("Access Application deleted", map[string]any{"domain": host})
	return nil
}

func (h *appHandler) Check(ctx context.Context, output AppOutput) (framework.Status, error) {
	_, exists, err := h.accessService.FindApplication(ctx, output.Host())
	if err != nil {
		return framework.StatusUnknown, fmt.Errorf("failed to check Access Application: %w", err)
	}
	if exists {
		return framework.StatusUp, nil
	}
	return framework.StatusDown, nil
}

func (h *appHandler) Recover(ctx context.Context, input AppInput) (AppOutput, framework.Status, error) {
	appID, exists, err := h.accessService.FindApplication(ctx, input.host())
	if err != nil {
		return AppOutput{}, framework.StatusUnknown, err
	}
	if !exists {
		return AppOutput{}, framework.StatusDown, nil
	}
	return AppOutput{
		Zone:      input.Zone,
		Subdomain: input.Subdomain,
		Path:      input.Path,
		AppID:     appID,
	}, framework.StatusUp, nil
}
