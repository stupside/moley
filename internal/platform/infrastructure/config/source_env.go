package config

import (
	"strings"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
)

// EnvSource loads config from environment variables
type EnvSource string

func (e EnvSource) Load(k *koanf.Koanf) error {
	prefix := string(e) + "_"
	return k.Load(env.Provider(prefix, ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(
			strings.TrimPrefix(s, prefix)), "_", ".")
	}), nil)
}

func (e EnvSource) Name() string {
	return "env:" + string(e)
}
