// Package config provides configuration management for Moley.
package config

import (
	"github.com/stupside/moley/internal/core/domain"
)

type GlobalConfig struct {
	Cloudflare struct {
		Token string `mapstructure:"token" yaml:"token" validate:"required"`
	} `mapstructure:"cloudflare" yaml:"cloudflare"`
}

type TunnelConfig struct {
	Tunnel  *domain.Tunnel  `mapstructure:"tunnel" yaml:"tunnel" validate:"required"`
	Ingress *domain.Ingress `mapstructure:"ingress" yaml:"ingress" validate:"required"`
}
