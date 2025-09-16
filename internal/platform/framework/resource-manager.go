package framework

import (
	"context"
	"encoding/json"

	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/shared"
)

// ResourceManager manages resources of a specific type with full type safety
type ResourceManager[TConfig any, TState any] struct {
	handler     ResourceHandler[TConfig, TState]
	registry    *ResourceRegistry
	handlerName string
}

// NewResourceManager creates a type-safe resource manager for a specific handler type
func NewResourceManager[TConfig any, TState any](
	handler ResourceHandler[TConfig, TState],
	registry *ResourceRegistry,
) *ResourceManager[TConfig, TState] {
	return &ResourceManager[TConfig, TState]{
		handler:     handler,
		registry:    registry,
		handlerName: handler.Name(),
	}
}

// Reconcile ensures the desired resources match the actual state
func (rm *ResourceManager[TConfig, TState]) Reconcile(
	ctx context.Context,
	desiredConfigs []TConfig,
) error {
	// Get current persisted state for this handler
	currentRecords := rm.getCurrentRecords()

	// Determine what actions are needed
	toRemove, toAdd, toUpdate := rm.computeActions(desiredConfigs, currentRecords)

	// Execute in order: remove, add, update
	if err := rm.removeResources(ctx, toRemove); err != nil {
		return shared.WrapError(err, "failed to remove resources")
	}

	if err := rm.addResources(ctx, toAdd); err != nil {
		return shared.WrapError(err, "failed to add resources")
	}

	if err := rm.updateResources(ctx, toUpdate); err != nil {
		return shared.WrapError(err, "failed to update resources")
	}

	return nil
}

// getCurrentRecords gets all current resource records for this handler from persistent storage
func (rm *ResourceManager[TConfig, TState]) getCurrentRecords() []ResourceRecord[TConfig, TState] {
	var records []ResourceRecord[TConfig, TState]

	for _, persistentEntry := range rm.registry.Entries {
		if persistentEntry.HandlerName == rm.handlerName {
			// Convert JSON data back to typed ResourceRecord
			var record ResourceRecord[TConfig, TState]
			if err := rm.unmarshalData(persistentEntry.Data, &record); err != nil {
				logger.Debugf("Failed to unmarshal persistent entry", map[string]any{
					"handler": rm.handlerName,
					"error":   err.Error(),
				})
				continue
			}

			records = append(records, record)
		}
	}

	return records
}

// unmarshalData converts interface{} data back to typed struct using JSON marshaling
func (rm *ResourceManager[TConfig, TState]) unmarshalData(data any, target *ResourceRecord[TConfig, TState]) error {
	// Marshal the data to JSON bytes
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return shared.WrapError(err, "failed to marshal data to JSON")
	}

	// Unmarshal back to typed struct
	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return shared.WrapError(err, "failed to unmarshal JSON to typed struct")
	}

	return nil
}

// computeActions determines what resources need to be added, removed, or updated
func (rm *ResourceManager[TConfig, TState]) computeActions(
	desired []TConfig,
	current []ResourceRecord[TConfig, TState],
) (
	toRemove []ResourceRecord[TConfig, TState],
	toAdd []TConfig,
	toUpdate []struct {
		old ResourceRecord[TConfig, TState]
		new TConfig
	},
) {
	// Build map of current configs
	currentMap := make(map[string]ResourceRecord[TConfig, TState])
	for _, record := range current {
		key := rm.getConfigKey(record.Config)
		currentMap[key] = record
	}

	// Build map of desired configs
	desiredMap := make(map[string]TConfig)
	for _, config := range desired {
		key := rm.getConfigKey(config)
		desiredMap[key] = config
	}

	// Find what to remove (current but not desired)
	for key, record := range currentMap {
		if _, exists := desiredMap[key]; !exists {
			toRemove = append(toRemove, record)
		}
	}

	// Find what to add or update
	for key, desiredConfig := range desiredMap {
		if currentRecord, exists := currentMap[key]; exists {
			// Check if config changed
			if !rm.handler.Equals(currentRecord.Config, desiredConfig) {
				toUpdate = append(toUpdate, struct {
					old ResourceRecord[TConfig, TState]
					new TConfig
				}{old: currentRecord, new: desiredConfig})
			}
			// If config unchanged, no action needed
		} else {
			// New resource
			toAdd = append(toAdd, desiredConfig)
		}
	}

	return toRemove, toAdd, toUpdate
}

// getConfigKey generates a unique key for a config (for deduplication)
func (rm *ResourceManager[TConfig, TState]) getConfigKey(config TConfig) string {
	// For now, use JSON serialization as key
	// In production, might want more efficient hashing
	data, _ := json.Marshal(config)
	return string(data)
}

// removeResources removes the specified resources
func (rm *ResourceManager[TConfig, TState]) removeResources(
	ctx context.Context,
	toRemove []ResourceRecord[TConfig, TState],
) error {
	for _, record := range toRemove {
		logger.Infof("Removing resource", map[string]any{
			"handler": rm.handlerName,
		})

		if err := rm.handler.Destroy(ctx, record.State); err != nil {
			return shared.WrapError(err, "failed to destroy resource")
		}

		// Remove from persistent storage
		if err := rm.removeFromRegistry(record); err != nil {
			return shared.WrapError(err, "failed to remove from registry")
		}
	}
	return nil
}

// addResources creates the specified resources
func (rm *ResourceManager[TConfig, TState]) addResources(
	ctx context.Context,
	toAdd []TConfig,
) error {
	for _, config := range toAdd {
		logger.Infof("Adding resource", map[string]any{
			"handler": rm.handlerName,
		})

		state, err := rm.handler.Create(ctx, config)
		if err != nil {
			return shared.WrapError(err, "failed to create resource")
		}

		// Verify it's actually up
		status, err := rm.handler.Status(ctx, state)
		if err != nil {
			return shared.WrapError(err, "failed to verify resource status")
		}

		if status != domain.StateUp {
			return shared.WrapError(nil, "resource created but not in up state")
		}

		// Add to persistent storage
		record := ResourceRecord[TConfig, TState]{
			Config: config,
			State:  state,
		}
		if err := rm.addToRegistry(record); err != nil {
			return shared.WrapError(err, "failed to add to registry")
		}
	}
	return nil
}

// updateResources updates the specified resources
func (rm *ResourceManager[TConfig, TState]) updateResources(
	ctx context.Context,
	toUpdate []struct {
		old ResourceRecord[TConfig, TState]
		new TConfig
	},
) error {
	for _, update := range toUpdate {
		logger.Infof("Updating resource", map[string]any{
			"handler": rm.handlerName,
		})

		// Remove old resource
		if err := rm.handler.Destroy(ctx, update.old.State); err != nil {
			return shared.WrapError(err, "failed to destroy old resource during update")
		}

		// Create new resource
		newState, err := rm.handler.Create(ctx, update.new)
		if err != nil {
			return shared.WrapError(err, "failed to create new resource during update")
		}

		// Verify it's up
		status, err := rm.handler.Status(ctx, newState)
		if err != nil {
			return shared.WrapError(err, "failed to verify updated resource status")
		}

		if status != domain.StateUp {
			return shared.WrapError(nil, "updated resource not in up state")
		}

		// Update persistent storage
		if err := rm.removeFromRegistry(update.old); err != nil {
			return shared.WrapError(err, "failed to remove old record from registry")
		}

		newRecord := ResourceRecord[TConfig, TState]{
			Config: update.new,
			State:  newState,
		}
		if err := rm.addToRegistry(newRecord); err != nil {
			return shared.WrapError(err, "failed to add updated record to registry")
		}
	}
	return nil
}

// ResourceRecord represents the business data for a managed resource (config + runtime state)
type ResourceRecord[TConfig any, TState any] struct {
	State  TState  `json:"state"`
	Config TConfig `json:"config"`
}

// addToRegistry adds a resource record to persistent storage
func (rm *ResourceManager[TConfig, TState]) addToRegistry(record ResourceRecord[TConfig, TState]) error {
	persistentEntry := PersistentResourceEntry{
		HandlerName: rm.handlerName,
		Data:        record,
	}

	return rm.registry.Add(persistentEntry)
}

// removeFromRegistry removes a resource record from persistent storage
func (rm *ResourceManager[TConfig, TState]) removeFromRegistry(record ResourceRecord[TConfig, TState]) error {
	persistentEntry := PersistentResourceEntry{
		HandlerName: rm.handlerName,
		Data:        record,
	}

	return rm.registry.Remove(persistentEntry)
}

// Stop removes all resources managed by this typed manager
func (rm *ResourceManager[TConfig, TState]) Stop(ctx context.Context) error {
	currentRecords := rm.getCurrentRecords()

	logger.Debugf("Stop found resources", map[string]any{
		"handler": rm.handlerName,
		"count":   len(currentRecords),
	})

	return rm.removeResources(ctx, currentRecords)
}
