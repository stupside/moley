package config

import (
	"moley/internal/errors"
	"moley/internal/logger"
	"sync"
)

// Manager provides a simple way to manage global configuration
type Manager struct {
	config *MoleyConfig
}

var (
	// once is used to ensure that the Manager instance is created only once
	once sync.Once
	// instance holds the singleton instance of the Manager
	instance *Manager
)

// GetManager returns the singleton config manager instance
func GetManager() *Manager {
	// Use sync.Once to ensure that the instance is created only once
	once.Do(func() {
		instance = &Manager{}
	})
	return instance
}

// Load loads the configuration from file and caches it
func (m *Manager) Load() error {
	config, err := LoadMoleyConfig()
	if err != nil {
		logger.Errorf("Failed to load Moley config", map[string]interface{}{"error": err.Error()})
		return err
	}
	m.Set(config)
	return nil
}

// Get returns the cached configuration
func (m *Manager) Get() *MoleyConfig {
	return m.config
}

// Set updates the cached configuration
func (m *Manager) Set(config *MoleyConfig) {
	m.config = config
}

// Save saves the current configuration to file
func (m *Manager) Save() error {
	config := m.Get()
	if config == nil {
		return errors.NewConfigError(errors.ErrCodeInvalidConfig, "no configuration loaded", nil)
	}
	return SaveMoleyConfigToFile(config)
}
