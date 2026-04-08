package config

import (
	"fmt"
	"path/filepath"

	platformconfig "github.com/stupside/moley/v2/internal/platform/config"
	"github.com/stupside/moley/v2/internal/platform/paths"
)

// GlobalConfig represents the global application configuration
type GlobalConfig struct {
	Cloudflare struct {
		Token string `yaml:"token" validate:"required"`
	} `yaml:"cloudflare"`
}

// NewGlobalManager creates a new global configuration manager
func NewGlobalManager() (*platformconfig.Manager[GlobalConfig], error) {
	path, err := globalConfigPath()
	if err != nil {
		return nil, fmt.Errorf("get global config path failed: %w", err)
	}

	mgr, err := platformconfig.New(path, defaultGlobalConfig(),
		platformconfig.WithSources[GlobalConfig](platformconfig.FileSource(path)),
		platformconfig.WithSources[GlobalConfig](platformconfig.EnvSource("MOLEY")),
	)
	if err != nil {
		return nil, fmt.Errorf("create global config manager failed: %w", err)
	}

	return mgr, nil
}

func globalConfigPath() (string, error) {
	userFolderPath, err := paths.GetUserFolderPath()
	if err != nil {
		return "", fmt.Errorf("get user folder path failed: %w", err)
	}
	return filepath.Join(userFolderPath, "config.yml"), nil
}

func defaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		Cloudflare: struct {
			Token string `yaml:"token" validate:"required"`
		}{},
	}
}
