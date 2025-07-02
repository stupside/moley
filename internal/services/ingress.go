package services

import (
	"context"

	"github.com/stupside/moley/internal/domain"
)

// IngressService defines the interface for managing ingress configurations for Cloudflare tunnels
type IngressService interface {
	IngressGetter
}

// IngressGetter defines methods for retrieving and writing Cloudflare tunnel ingress configurations
type IngressGetter interface {
	GetConfiguration(ctx context.Context, tunnel *domain.Tunnel, dns *domain.DNS) ([]byte, error)
	GetConfigurationPath(ctx context.Context, tunnel *domain.Tunnel) (string, error)
}
