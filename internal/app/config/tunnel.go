package config

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/stupside/moley/v2/internal/domain"
	platformconfig "github.com/stupside/moley/v2/internal/platform/config"
)

// TunnelConfig represents tunnel-specific configuration
type TunnelConfig struct {
	Tunnel  *domain.Tunnel  `yaml:"tunnel" validate:"required"`
	Ingress *domain.Ingress `yaml:"ingress" validate:"required"`
}

// NewTunnelManager creates a new tunnel configuration manager
func NewTunnelManager(path string) (*platformconfig.Manager[TunnelConfig], error) {
	defaultConfig, err := defaultTunnelConfig()
	if err != nil {
		return nil, fmt.Errorf("create default tunnel config failed: %w", err)
	}

	mgr, err := platformconfig.New(path, defaultConfig,
		platformconfig.WithSources[TunnelConfig](platformconfig.FileSource(path)),
		platformconfig.WithSources[TunnelConfig](platformconfig.EnvSource("MOLEY_TUNNEL")),
	)
	if err != nil {
		return nil, fmt.Errorf("create tunnel config manager failed: %w", err)
	}

	return mgr, nil
}

func ExampleTunnelConfig() (*TunnelConfig, error) {
	id := uuid.New().String()

	tunnel, err := domain.NewTunnel(id)
	if err != nil {
		return nil, fmt.Errorf("create tunnel failed: %w", err)
	}

	return &TunnelConfig{
		Tunnel: tunnel,
		Ingress: &domain.Ingress{
			Zone: "moley.dev",
			Mode: domain.IngressModeSubdomain,
			Apps: []domain.AppConfig{
				{
					Target: domain.TargetConfig{
						Port:     3000,
						Hostname: "localhost",
						Protocol: domain.ProtocolHTTP,
					},
					Expose: domain.ExposeConfig{
						Subdomain: "api",
					},
				},
			},
		},
	}, nil
}

func defaultTunnelConfig() (*TunnelConfig, error) {
	id := uuid.New().String()

	tunnel, err := domain.NewTunnel(id)
	if err != nil {
		return nil, fmt.Errorf("create tunnel failed: %w", err)
	}

	return &TunnelConfig{
		Tunnel: tunnel,
		Ingress: &domain.Ingress{
			Mode: domain.IngressModeSubdomain,
			Apps: []domain.AppConfig{},
		},
	}, nil
}
