package config

import (
	"fmt"
	"strings"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
)

type EnvSource string

func (e EnvSource) Load(k *koanf.Koanf) error {
	prefix := string(e) + "_"
	return k.Load(env.Provider(prefix, ".", func(s string) string {
		s = strings.TrimPrefix(s, prefix)
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, "__", ".")
		return s
	}), nil)
}

func (e EnvSource) Name() string {
	return fmt.Sprintf("env(%s)", string(e))
}
