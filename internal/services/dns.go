package services

import (
	"context"
	"moley/internal/domain"

	"github.com/cloudflare/cloudflare-go/v3/dns"
)

// DNSService defines the interface for managing DNS records for Cloudflare tunnels
type DNSService interface {
	DNSRouter
	DNSGetter
	DNSDeleter
}

// DNSRouter defines methods for routing DNS records for a Cloudflare tunnel
type DNSRouter interface {
	Route(ctx context.Context, tunnel *domain.Tunnel, dns *domain.DNS) error
}

// DNSGetter defines methods for retrieving DNS content and records for a Cloudflare tunnel
type DNSGetter interface {
	GetZoneID(ctx context.Context, dns *domain.DNS) (string, error)
	GetContent(ctx context.Context, tunnel *domain.Tunnel) (string, error)
	GetRecords(ctx context.Context, tunnel *domain.Tunnel, dns *domain.DNS) ([]dns.RecordListResponse, error)
}

// DNSDeleter defines methods for deleting DNS records for a Cloudflare tunnel
type DNSDeleter interface {
	DeleteRecords(ctx context.Context, tunnel *domain.Tunnel, dns *domain.DNS) error
}
