// Package domain provides core domain models for Moley.
package domain

import "fmt"

type TargetConfig struct {
	Port     int    `mapstructure:"port" yaml:"port" json:"port" validate:"required,min=1,max=65535"`
	Hostname string `mapstructure:"hostname" yaml:"hostname" json:"hostname" validate:"required"`
}

func (t *TargetConfig) GetTargetURL() string {
	return fmt.Sprintf("http://%s:%d", t.Hostname, t.Port)
}

type ExposeConfig struct {
	Subdomain string `mapstructure:"subdomain" yaml:"subdomain" json:"subdomain" validate:"required"`
}

type AppConfig struct {
	Target TargetConfig `mapstructure:"target" yaml:"target" json:"target" validate:"required"`
	Expose ExposeConfig `mapstructure:"expose" yaml:"expose" json:"expose" validate:"required"`
}

type Ingress struct {
	Zone string      `mapstructure:"zone" yaml:"zone" json:"zone" validate:"required"`
	Apps []AppConfig `mapstructure:"apps" yaml:"apps" json:"apps" validate:"required,dive"`
}

// NewDefaultIngress creates a default ingress configuration for examples/templates
func NewDefaultIngress() *Ingress {
	return &Ingress{
		Zone: "example.com",
		Apps: []AppConfig{
			{
				Target: TargetConfig{
					Port:     8080,
					Hostname: "localhost",
				},
				Expose: ExposeConfig{
					Subdomain: "api",
				},
			},
		},
	}
}

func (i *Ingress) GetSubdomains() []string {
	subdomains := make([]string, 0, len(i.Apps))
	for _, app := range i.Apps {
		subdomains = append(subdomains, app.Expose.Subdomain)
	}
	return subdomains
}
