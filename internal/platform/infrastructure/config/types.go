package config

import (
	"github.com/knadh/koanf/v2"
)

// Source represents a configuration source
type Source interface {
	Load(*koanf.Koanf) error
	Name() string
}
