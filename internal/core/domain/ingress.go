// Package domain provides core domain models for Moley.
package domain

import "fmt"

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

type ExposeConfig struct {
	Subdomain string `yaml:"subdomain" json:"subdomain" validate:"required"`
}

type AppConfig struct {
	Target TargetConfig `yaml:"target" json:"target" validate:"required"`
	Expose ExposeConfig `yaml:"expose" json:"expose" validate:"required"`
}

type Ingress struct {
	Zone string      `yaml:"zone" json:"zone" validate:"required"`
	Apps []AppConfig `yaml:"apps" json:"apps" validate:"required,dive"`
}
