package services

import (
	"context"

	"github.com/stupside/moley/internal/domain"
)

// TunnelService defines the interface for managing Cloudflare tunnels
type TunnelService interface {
	TunnelGetter
	TunnelDeleter
	TunnelCreater
}

// TunnelGetter defines methods for retrieving information about Cloudflare tunnels
type TunnelGetter interface {
	GetID(ctx context.Context, tunnel *domain.Tunnel) (string, error)
	GetToken(ctx context.Context, tunnel *domain.Tunnel) (*struct {
		TunnelId  string
		AccountId string
	}, error)
	GetAccountID(ctx context.Context, tunnel *domain.Tunnel) (string, error)
	GetCredentialsPath(ctx context.Context, tunnel *domain.Tunnel) (string, error)
}

// TunnelDeleter defines methods for deleting Cloudflare tunnels
type TunnelDeleter interface {
	DeleteTunnel(ctx context.Context, tunnel *domain.Tunnel) error
}

type TunnelCreater interface {
	CreateTunnel(ctx context.Context, tunnel *domain.Tunnel) (string, error)
}
