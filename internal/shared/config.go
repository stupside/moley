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

	logger.Debugf("Checking if initial configuration is provided", map[string]interface{}{"path": m.path})
	if _, exists := configs[m.path]; exists {
		if _, ok := configs[m.path].(*T); ok {
			return ErrConfigAlreadyLoaded
		}
		return ErrConfigAlreadyLoadedInvalidType
	}
	logger.Info("Configuration file not found")

	if err := m.Save(m.initial, false); err != nil {
		return fmt.Errorf("%w: %w", ErrConfigSave, err)
	}

	return nil
}

// Save writes the configuration to the file after validating it if required
func (m *BaseConfigManager[T]) Save(config *T, validate bool) error {
	if config == nil {
		return ErrConfigNil
	}

	if validate {
		logger.Debug("Validating configuration")
		if err := validation.ValidateStruct(config); err != nil {
			return fmt.Errorf("%w: %w", ErrConfigValidation, err)
		}
		logger.Info("Configuration validation successful")
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrConfigMarshal, err)
	}

	logger.Debugf("Writing configuration to file", map[string]interface{}{"path": m.path})
	if err := os.WriteFile(m.path, data, 0755); err != nil {
		return fmt.Errorf("%w: %w", ErrConfigWrite, err)
	}
	logger.Infof("Written configuration to file successfully", map[string]interface{}{"path": m.path})

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

	logger.Debug("Loading configuration from")

	// Try to read config file if it exists
	if m.IsFound() {
		if err := m.vp.ReadInConfig(); err != nil {
			if os.IsNotExist(err) {
				logger.Debug("Configuration file not found, using default values and bound flags")
			} else {
				return nil, fmt.Errorf("%w: %w", ErrConfigRead, err)
			}
		} else {
			logger.Infof("Configuration loaded successfully", map[string]interface{}{"path": m.path})
		}
	} else {
		logger.Debug("Configuration file not found, using default values and bound flags")
	}

	// Unmarshal the configuration into the specified type
	// This will include values from config file (if exists) and bound flags
	config := new(T)
	if err := m.vp.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConfigUnmarshal, err)
	}

	if validate {
		logger.Debug("Validating loaded configuration")
		if err := validation.ValidateStruct(config); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrConfigValidation, err)
		}
		logger.Info("Configuration validation successful")
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
