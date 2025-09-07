// Package ports declares provider-agnostic interfaces.
package ports

import (
	"context"

	"github.com/stupside/moley/v2/internal/core/domain"
)

type DNSService interface {
	GetZoneID(ctx context.Context, zoneName string) (string, error)
	GetContent(ctx context.Context, tunnel *domain.Tunnel) (string, error)
	DeleteRecord(ctx context.Context, tunnel *domain.Tunnel, zoneName string, subdomain string) error
	RouteRecord(ctx context.Context, tunnel *domain.Tunnel, zoneName string, subdomain string) error
	RecordExists(ctx context.Context, tunnel *domain.Tunnel, zoneName string, subdomain string) (bool, error)
}
