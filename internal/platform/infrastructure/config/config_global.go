package config

import (
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/stupside/moley/v2/internal/platform/infrastructure/paths"
	"github.com/stupside/moley/v2/internal/shared"
)

// GlobalConfig represents the global application configuration
type GlobalConfig struct {
	Cloudflare struct {
		Token string `yaml:"token" validate:"required"`
	} `yaml:"cloudflare"`
}

// GlobalManager manages global configuration
type GlobalManager = Manager[GlobalConfig]

// NewGlobalManager creates a new global configuration manager
func NewGlobalManager(cmd *cli.Command) (*GlobalManager, error) {
	path, err := globalConfigPath()
	if err != nil {
		return nil, shared.WrapError(err, "get global config path failed")
	}

	mgr := New(path, defaultGlobalConfig(),
		WithSources[GlobalConfig](FileSource(path)),
		WithSources[GlobalConfig](EnvSource("MOLEY")),
	)

	return mgr, nil
}

func globalConfigPath() (string, error) {
	userFolderPath, err := paths.GetUserFolderPath()
	if err != nil {
		return "", shared.WrapError(err, "get user folder path failed")
	}
	return filepath.Join(userFolderPath, "config.yml"), nil
}

func defaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		Cloudflare: struct {
			Token string `yaml:"token" validate:"required"`
		}{
			Token: "<cloudflare_token>",
		},
	}
}
