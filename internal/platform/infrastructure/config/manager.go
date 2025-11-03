package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"

	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/shared"
)

const (
	configTag = "yaml"
)

// Manager manages configuration loading and persistence
type Manager[T any] struct {
	k         *koanf.Koanf
	path      string
	validator *validator.Validate
}

// ConfigOption configures a Manager
type ConfigOption[T any] func(*Manager[T]) error

// New creates a new config manager with functional options
func New[T any](path string, defaultConfig *T, opts ...ConfigOption[T]) (*Manager[T], error) {
	m := &Manager[T]{
		k:         koanf.New("."),
		path:      path,
		validator: validator.New(),
	}

	// Initialize with default config
	if err := m.k.Load(structs.Provider(defaultConfig, "yaml"), nil); err != nil {
		return nil, shared.WrapError(err, "load default config failed")
	}

	// Apply options (sources like files, env vars)
	for _, opt := range opts {
		if err := opt(m); err != nil {
			logger.Debugf("Config option failed", map[string]any{
				"error": err.Error(),
			})
		}
	}

	return m, nil
}

// WithSources adds configuration sources
func WithSources[T any](sources ...Source) ConfigOption[T] {
	return func(m *Manager[T]) error {
		for _, source := range sources {
			if err := source.Load(m.k); err != nil {
				return shared.WrapError(err, fmt.Sprintf("load source %s failed", source.Name()))
			}
		}
		return nil
	}
}

// save persists the current koanf state to file
// This assumes the config has already been validated
func (m *Manager[T]) save() error {
	// Marshal to YAML and write to file
	data, err := m.k.Marshal(yaml.Parser())
	if err != nil {
		return shared.WrapError(err, "marshal failed")
	}

	if err := os.MkdirAll(filepath.Dir(m.path), 0755); err != nil {
		return shared.WrapError(err, "create dir failed")
	}

	if err := os.WriteFile(m.path, data, 0600); err != nil {
		return shared.WrapError(err, "write file failed")
	}

	return nil
}

// validate validates the configuration
func (m *Manager[T]) validate(config *T) error {
	if err := m.validator.Struct(config); err != nil {
		return shared.WrapError(err, "validate config failed")
	}
	return nil
}

// Get returns the configuration
func (m *Manager[T]) Get(validate bool) (*T, error) {
	config := new(T)
	if err := m.k.UnmarshalWithConf("", config, koanf.UnmarshalConf{
		Tag: configTag,
	}); err != nil {
		return nil, shared.WrapError(err, "unmarshal failed")
	}

	if validate {
		if err := m.validate(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// Update updates configuration and saves it
func (m *Manager[T]) Update(fn func(*T)) error {
	// Get current config to modify
	config, err := m.Get(false)
	if err != nil {
		return shared.WrapError(err, "get config failed")
	}

	// Apply modifications
	fn(config)

	// Validate before loading into koanf
	if err := m.validate(config); err != nil {
		return err
	}

	// Load the modified config into koanf
	if err := m.k.Load(structs.Provider(config, configTag), nil); err != nil {
		return shared.WrapError(err, "load updated config failed")
	}

	// Persist to disk
	return m.save()
}

// Override replaces the entire configuration and saves it
func (m *Manager[T]) Override(config *T) error {
	// Validate before loading into koanf
	if err := m.validate(config); err != nil {
		return err
	}

	// Load the new config into koanf
	if err := m.k.Load(structs.Provider(config, configTag), nil); err != nil {
		return shared.WrapError(err, "load config failed")
	}

	// Persist to disk
	return m.save()
}
