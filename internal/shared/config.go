package shared

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"github.com/stupside/moley/internal/logger"
	"github.com/stupside/moley/internal/validation"
	"gopkg.in/yaml.v3"
)

// Error definitions
var (
	configs = make(map[string]any)
)

// WithOption is a function type that modifies the viper.Viper instance
type BaseConfigManager[T any] struct {
	vp      *viper.Viper
	path    string
	initial *T
}

// NewConfigManager creates a new BaseConfigManager instance
func NewConfigManager[T any](path string, initial *T, options ...WithOption) *BaseConfigManager[T] {
	v := viper.New()
	v.SetConfigFile(path)

	for _, opt := range options {
		opt(v)
	}

	return &BaseConfigManager[T]{
		vp:      v,
		path:    path,
		initial: initial,
	}
}

// Init initializes the configuration manager by checking if the config file exists
func (m *BaseConfigManager[T]) Init() error {
	if m.IsFound() {
		return nil
	}

	logger.Debug("Config file not found, checking for initial configuration")
	if _, exists := configs[m.path]; exists {
		if _, ok := configs[m.path].(*T); ok {
			return WrapError(ErrConfigAlreadyLoaded, "config already loaded at this path")
		}
		return WrapError(ErrConfigAlreadyLoadedInvalidType, "config already loaded at this path has an invalid type")
	}

	if err := m.Save(m.initial, false); err != nil {
		return WrapError(err, "failed to save initial config")
	}

	return nil
}

// Save writes the configuration to the file after validating it if required
func (m *BaseConfigManager[T]) Save(config *T, validate bool) error {
	if config == nil {
		return WrapError(ErrConfigNil, "config is nil")
	}

	if validate {
		logger.Debug("Validating configuration")
		if err := validation.ValidateStruct(config); err != nil {
			return WrapError(ErrConfigValidation, err.Error())
		}
		logger.Debug("Configuration validation successful")
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return WrapError(ErrConfigMarshal, err.Error())
	}

	logger.Debug(fmt.Sprintf("Writing configuration file: %s", m.path))
	if err := os.WriteFile(m.path, data, 0755); err != nil {
		return WrapError(ErrConfigWrite, err.Error())
	}
	logger.Info("Configuration file written successfully")

	configs[m.path] = config

	return nil
}

// Load reads the configuration from the file and unmarshals it into the specified type
func (m *BaseConfigManager[T]) Load(validate bool) (*T, error) {
	if _, exists := configs[m.path]; exists {
		if cfg, ok := configs[m.path].(*T); ok {
			return cfg, nil
		}
	}

	logger.Debug(fmt.Sprintf("Loading configuration: %s", m.path))

	// Try to read config file if it exists
	if m.IsFound() {
		if err := m.vp.ReadInConfig(); err != nil {
			if os.IsNotExist(err) {
				logger.Debug("Configuration file not found, using defaults")
			} else {
				return nil, WrapError(ErrConfigRead, err.Error())
			}
		} else {
			logger.Debug("Configuration loaded from file")
		}
	} else {
		logger.Debug("Configuration file not found, using defaults")
	}

	// Unmarshal the configuration into the specified type
	config := new(T)
	if err := m.vp.Unmarshal(&config); err != nil {
		return nil, WrapError(ErrConfigUnmarshal, err.Error())
	}

	if validate {
		logger.Debug("Validating loaded configuration")
		if err := validation.ValidateStruct(config); err != nil {
			return nil, WrapError(ErrConfigValidation, err.Error())
		}
		logger.Debug("Configuration validation successful")
	}

	configs[m.path] = config

	return config, nil
}

// IsFound checks if the configuration file exists at the specified path
func (m *BaseConfigManager[T]) IsFound() bool {
	if _, err := os.Stat(m.path); os.IsNotExist(err) {
		return false
	}
	return true
}
