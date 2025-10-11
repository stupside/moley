package config

import (
	"github.com/google/uuid"
	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/shared"
)

// TunnelConfig represents tunnel-specific configuration
type TunnelConfig struct {
	Tunnel  *domain.Tunnel  `yaml:"tunnel" validate:"required"`
	Ingress *domain.Ingress `yaml:"ingress" validate:"required"`
}

// TunnelManager manages tunnel configuration
type TunnelManager = Manager[TunnelConfig]

// NewTunnelManager creates a new tunnel configuration manager
func NewTunnelManager(path string) (*TunnelManager, error) {
	defaultConfig, err := defaultTunnelConfig()
	if err != nil {
		return nil, shared.WrapError(err, "create default tunnel config failed")
	}

	mgr, err := New(path, defaultConfig,
		WithSources[TunnelConfig](FileSource(path)),
		WithSources[TunnelConfig](EnvSource("MOLEY_TUNNEL")),
	)
	if err != nil {
		return nil, shared.WrapError(err, "create tunnel config manager failed")
	}

	return mgr, nil
}

func ExampleTunnelConfig() (*TunnelConfig, error) {
	id := uuid.New().String()

	tunnel, err := domain.NewTunnel(id)
	if err != nil {
		return nil, shared.WrapError(err, "create tunnel failed")
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
		return nil, shared.WrapError(err, "create tunnel failed")
	}

	return &TunnelConfig{
		Tunnel: tunnel,
		Ingress: &domain.Ingress{
			Mode: domain.IngressModeSubdomain,
			Apps: []domain.AppConfig{
				{},
			},
		},
	}, nil
}
