// Package ports declares provider-agnostic interfaces.
package ports

import (
	"context"

	"github.com/stupside/moley/v2/internal/core/domain"
)

type DNSService interface {
	DeleteRecord(ctx context.Context, tunnel *domain.Tunnel, zoneName string, subdomain string) error
	RouteRecord(ctx context.Context, tunnel *domain.Tunnel, zoneName string, subdomain string) error
	RecordExists(ctx context.Context, tunnel *domain.Tunnel, zoneName string, subdomain string) (bool, error)
}
