package config

import (
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"

	"github.com/stupside/moley/v2/internal/shared"
)

const (
	tag = "yaml"
)

// Manager manages configuration loading and persistence
type Manager[T any] struct {
	k         *koanf.Koanf
	path      string
	validator *validator.Validate
}

// Option configures a Manager
type Option[T any] func(*Manager[T])

// New creates a new config manager with functional options
func New[T any](path string, defaultConfig *T, opts ...Option[T]) *Manager[T] {
	m := &Manager[T]{
		k:         koanf.New("."),
		path:      path,
		validator: validator.New(),
	}

	// Initialize with default config
	if err := m.k.Load(structs.Provider(defaultConfig, "yaml"), nil); err != nil {
		panic(shared.WrapError(err, "load default config failed"))
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// WithSources adds configuration sources
func WithSources[T any](sources ...Source) Option[T] {
	return func(m *Manager[T]) {
		for _, source := range sources {
			source.Load(m.k) // Ignore errors - sources might not exist
		}
	}
}

// save persists configuration to file
func (m *Manager[T]) save() error {

	// Validate before saving
	if err := m.validator.Struct(m.Get()); err != nil {
		return shared.WrapError(err, "validation failed")
	}

	data, err := m.k.Marshal(yaml.Parser())
	if err != nil {
		return shared.WrapError(err, "marshal failed")
	}

	if err := os.MkdirAll(filepath.Dir(m.path), 0755); err != nil {
		return shared.WrapError(err, "create dir failed")
	}

	return shared.WrapError(
		os.WriteFile(m.path, data, 0600),
		"write file failed",
	)
}

// Update updates configuration and saves it
func (m *Manager[T]) Update(fn func(*T)) error {
	config := m.Get()
	fn(config)
	return m.save()
}

// Override replaces the entire configuration and saves it
func (m *Manager[T]) Override(config *T) error {
	if err := m.k.Load(structs.Provider(config, tag), nil); err != nil {
		return shared.WrapError(err, "load config failed")
	}
	return m.save()
}

// Get returns the configuration
func (m *Manager[T]) Get() *T {
	config := new(T)
	if err := m.k.UnmarshalWithConf("", config, koanf.UnmarshalConf{
		Tag: tag,
	}); err != nil {
		panic(shared.WrapError(err, "unmarshal failed"))
	}
	return config
}
