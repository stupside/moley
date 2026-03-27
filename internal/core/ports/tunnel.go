// Package ports declares provider-agnostic interfaces.
package ports

import (
	"context"

	"github.com/stupside/moley/v2/internal/core/domain"
)

type TunnelService interface {
	Run(ctx context.Context, tunnel *domain.Tunnel) (int, error)
	GetID(ctx context.Context, tunnel *domain.Tunnel) (string, error)
	Create(ctx context.Context, tunnel *domain.Tunnel) (string, error)
	Delete(ctx context.Context, tunnel *domain.Tunnel) error
	Exists(ctx context.Context, tunnel *domain.Tunnel) (bool, error)
	SaveConfiguration(ctx context.Context, tunnel *domain.Tunnel, ingress *domain.Ingress) error
	DeleteConfiguration(ctx context.Context, tunnel *domain.Tunnel) error
	GetConfigurationPath(ctx context.Context, tunnel *domain.Tunnel) (string, error)
}
