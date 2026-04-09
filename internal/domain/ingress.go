// Package domain provides core domain models for Moley.
package domain

import (
	"encoding/json"
	"fmt"
	"maps"
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
	Target   TargetConfig   `yaml:"target" json:"target" validate:"required"`
	Expose   ExposeConfig   `yaml:"expose" json:"expose" validate:"required"`
	Access   *AccessConfig  `yaml:"access,omitempty" json:"access,omitempty" validate:"omitempty"`
	Policies []string `yaml:"policies,omitempty" json:"policies,omitempty"`
}

// AccessConfig holds the raw Cloudflare Access application configuration.
// Fields match the CF API shape and are passed through with minimal transformation.
// Only `providers` is processed by Moley (resolved to IdP UUIDs); everything else
// is forwarded to the Cloudflare API as-is.
type AccessConfig struct {
	Providers []string       `yaml:"providers,omitempty" json:"providers,omitempty"`
	Raw       map[string]any `yaml:",remain" json:"-"`
}

// MarshalJSON inlines Raw so policy changes are reflected in the input hash.
func (a AccessConfig) MarshalJSON() ([]byte, error) {
	m := make(map[string]any, len(a.Raw)+1)
	maps.Copy(m, a.Raw)
	if len(a.Providers) > 0 {
		m["providers"] = a.Providers
	}
	return json.Marshal(m)
}

type ExposeConfig struct {
	Subdomain string `yaml:"subdomain" json:"subdomain" validate:"required"`
}

type IngressMode string

const (
	// IngressModeWildcard — best for development and frequently changing apps; one DNS record covers all subdomains.
	IngressModeWildcard IngressMode = "wildcard"
	// IngressModeSubdomain — best for production; each app gets its own explicit DNS record.
	IngressModeSubdomain IngressMode = "subdomain"
)

type Ingress struct {
	Zone string      `yaml:"zone" json:"zone" validate:"required"`
	Apps []AppConfig `yaml:"apps" json:"apps" validate:"required,dive"`
	Mode IngressMode `yaml:"mode" json:"mode" validate:"required,oneof=wildcard subdomain"`
}

func (i *Ingress) HasAccessConfig() bool {
	for _, app := range i.Apps {
		if app.Access != nil {
			return true
		}
	}
	return false
}

func FQDN(subdomain, zone string) string {
	return subdomain + "." + zone
}
