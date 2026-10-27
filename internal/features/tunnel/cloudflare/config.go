// Package cloudflare provides Cloudflare-specific implementations.
package cloudflare

import (
	"cmp"
	"regexp"
	"slices"

	"github.com/stupside/moley/v2/internal/domain"
	tunnelusecase "github.com/stupside/moley/v2/internal/features/tunnel/usecase"
)

// runConfig is the YAML config format for `cloudflared tunnel run`.
type runConfig struct {
	Tunnel          string        `yaml:"tunnel" validate:"required"`
	Logfile         string        `yaml:"logfile,omitempty"`
	Loglevel        string        `yaml:"loglevel,omitempty"`
	CredentialsFile string        `yaml:"credentials-file" validate:"required"`
	Ingress         []ingressRule `yaml:"ingress" validate:"required"`
}

type ingressRule struct {
	Hostname      string         `yaml:"hostname,omitempty"`
	Path          string         `yaml:"path,omitempty"`
	Service       string         `yaml:"service" validate:"required"`
	OriginRequest *originRequest `yaml:"originRequest,omitempty"`
}

type originRequest struct {
	domain.OriginRequest `yaml:",inline"`
	Access               *originAccess `yaml:"access,omitempty"`
}

type originAccess struct {
	Required bool     `yaml:"required"`
	TeamName string   `yaml:"teamName"`
	AudTag   []string `yaml:"audTag"`
}

// buildIngress returns the cloudflared rules for the apps, ending with a 404 catch-all.
// cloudflared picks the first match, so longer paths go first.
func buildIngress(ingress *domain.Ingress, guards map[string]tunnelusecase.AccessGuard) []ingressRule {
	rules := make([]ingressRule, 0, len(ingress.Apps)+1)
	for _, app := range ingress.Apps {
		var origin *originRequest
		if app.Target.OriginRequest != nil {
			origin = &originRequest{OriginRequest: *app.Target.OriginRequest}
		}
		if g, ok := guards[app.Expose.Host(ingress.Zone)]; ok {
			if origin == nil {
				origin = &originRequest{}
			}
			origin.Access = &originAccess{Required: true, TeamName: g.TeamName, AudTag: []string{g.AUD}}
		}
		rules = append(rules, ingressRule{
			Hostname:      domain.FQDN(app.Expose.Subdomain, ingress.Zone),
			Path:          pathPattern(app.Expose.Path),
			Service:       app.Target.GetTargetURL(),
			OriginRequest: origin,
		})
	}
	slices.SortStableFunc(rules, func(a, b ingressRule) int {
		return cmp.Compare(len(b.Path), len(a.Path))
	})
	return append(rules, ingressRule{Service: "http_status:404"})
}

// pathPattern matches a path prefix and everything below it: /api matches /api/v1, not /apiary.
func pathPattern(prefix string) string {
	if prefix == "" {
		return ""
	}
	return "^" + regexp.QuoteMeta(prefix) + "(/|$)"
}
