// Package ports declares provider-agnostic interfaces.
package ports

import (
	"context"

	"github.com/stupside/moley/v2/internal/core/domain"
)

type TunnelService interface {
	Run(ctx context.Context, tunnel *domain.Tunnel) (int, error)
	GetID(ctx context.Context, tunnel *domain.Tunnel) (string, error)
	GetAccountID(ctx context.Context, tunnel *domain.Tunnel) (string, error)
	GetCredentialsPath(ctx context.Context, tunnel *domain.Tunnel) (string, error)
	DeleteTunnel(ctx context.Context, tunnel *domain.Tunnel) error
	CreateTunnel(ctx context.Context, tunnel *domain.Tunnel) (string, error)
	TunnelExists(ctx context.Context, tunnel *domain.Tunnel) (bool, error)
	SaveConfiguration(ctx context.Context, tunnel *domain.Tunnel, ingress *domain.Ingress) error
	DeleteConfiguration(ctx context.Context, tunnel *domain.Tunnel) error
	GetConfigurationPath(ctx context.Context, tunnel *domain.Tunnel) (string, error)
}
