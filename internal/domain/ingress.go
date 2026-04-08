// Package domain provides core domain models for Moley.
package domain

import (
	"fmt"
	"time"
)

type TargetProtocol string

const (
	ProtocolTCP   TargetProtocol = "tcp"
	ProtocolHTTP  TargetProtocol = "http"
	ProtocolHTTPS TargetProtocol = "https"
)

type TargetConfig struct {
	Port     int            `yaml:"port" json:"port" validate:"required,min=1,max=65535"`
	Hostname string         `yaml:"hostname" json:"hostname" validate:"required"`
	Protocol TargetProtocol `yaml:"protocol" json:"protocol" validate:"required,oneof=http https tcp"`
}

func (t *TargetConfig) GetTargetURL() string {
	return fmt.Sprintf("%s://%s:%d", t.Protocol, t.Hostname, t.Port)
}

type AppConfig struct {
	Target TargetConfig  `yaml:"target" json:"target" validate:"required"`
	Expose ExposeConfig  `yaml:"expose" json:"expose" validate:"required"`
	Access *AccessConfig `yaml:"access,omitempty" json:"access,omitempty" validate:"omitempty"`
}

type AccessPolicyDecision string

const (
	AccessPolicyDecisionAllow  AccessPolicyDecision = "allow"
	AccessPolicyDecisionBypass AccessPolicyDecision = "bypass"
)

type AccessPolicyConfig struct {
	Decision AccessPolicyDecision `yaml:"decision" json:"decision" validate:"required,oneof=allow bypass"`
	Emails   []string             `yaml:"emails,omitempty" json:"emails,omitempty" validate:"omitempty,dive,email"`
	Domains  []string             `yaml:"domains,omitempty" json:"domains,omitempty" validate:"omitempty,dive,fqdn"`
}

type AccessConfig struct {
	Session   string              `yaml:"session,omitempty" json:"session,omitempty" validate:"required"`
	Providers []string            `yaml:"providers,omitempty" json:"providers,omitempty"`
	Policy    *AccessPolicyConfig `yaml:"policy" json:"policy" validate:"required"`
}

// GetSessionDuration parses Session.
// Go's time.Duration format is used (e.g., "24h", "30m").
func (c *AccessConfig) GetSessionDuration() (time.Duration, error) {
	d, err := time.ParseDuration(c.Session)
	if err != nil {
		return 0, fmt.Errorf("invalid session %q: %w", c.Session, err)
	}
	return d, nil
}

type ExposeConfig struct {
	Subdomain string `yaml:"subdomain" json:"subdomain" validate:"required"`
}

type IngressMode string

const (
	// IngressModeWildcard creates a single *.zone DNS record.
	// Cloudflared routes requests by hostname in ingress rules.
	// Best for: development, frequently changing apps, faster iteration.
	// DNS: *.example.com → tunnel (1 record)
	IngressModeWildcard IngressMode = "wildcard"

	// IngressModeSubdomain creates individual DNS records per app.
	// Each app gets its own subdomain.domain DNS entry.
	// Best for: production, explicit DNS control, stable apps.
	// DNS: api.example.com → tunnel, app.example.com → tunnel (N records)
	IngressModeSubdomain IngressMode = "subdomain"
)

type Ingress struct {
	Zone string      `yaml:"zone" json:"zone" validate:"required"`
	Apps []AppConfig `yaml:"apps" json:"apps" validate:"required,dive"`
	Mode IngressMode `yaml:"mode" json:"mode" validate:"required,oneof=wildcard subdomain"`
}

// HasAccessConfig returns true if any app has access protection configured.
func (i *Ingress) HasAccessConfig() bool {
	for _, app := range i.Apps {
		if app.Access != nil {
			return true
		}
	}
	return false
}

// FQDN returns the fully qualified domain name for a subdomain within a zone.
func FQDN(subdomain, zone string) string {
	return subdomain + "." + zone
}
