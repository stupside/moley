package framework

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	RegistryFile = "moley.lock"
)

// PersistentResourceEntry represents a persisted resource in the registry file
type PersistentResourceEntry struct {
	Data        any    `json:"data"`
	HandlerName string `json:"handler_name"`
}

// ResourceRegistry manages persistent storage of resource state
type ResourceRegistry struct {
	Entries []PersistentResourceEntry `json:"entries"`
}

// LoadResourceRegistry loads the resource registry from persistent storage
func LoadResourceRegistry() (*ResourceRegistry, error) {
	registry := &ResourceRegistry{Entries: make([]PersistentResourceEntry, 0)}

	if _, err := os.Stat(RegistryFile); os.IsNotExist(err) {
		return registry, nil
	}

	data, err := os.ReadFile(RegistryFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read resource registry file: %w", err)
	}

	if err := json.Unmarshal(data, registry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resource registry: %w", err)
	}

	if registry.Entries == nil {
		registry.Entries = make([]PersistentResourceEntry, 0)
	}

	return registry, nil
}

// Save persists the resource registry to storage
func (rr *ResourceRegistry) Save() error {
	data, err := json.Marshal(rr)
	if err != nil {
		return fmt.Errorf("failed to marshal resource registry: %w", err)
	}

	if err := os.WriteFile(RegistryFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write resource registry file: %w", err)
	}

	return nil
}

// Add adds a persistent resource entry to the registry
func (rr *ResourceRegistry) Add(entry PersistentResourceEntry) error {
	rr.Entries = append(rr.Entries, entry)
	return rr.Save()
}

// Remove removes a persistent resource entry from the registry
func (rr *ResourceRegistry) Remove(entry PersistentResourceEntry) error {
	for i, existing := range rr.Entries {
		if existing.HandlerName == entry.HandlerName {
			rr.Entries = append(rr.Entries[:i], rr.Entries[i+1:]...)
			break
		}
	}
	return rr.Save()
}
