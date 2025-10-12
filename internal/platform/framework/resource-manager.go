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
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return shared.WrapError(err, "failed to marshal data to JSON")
	}

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
		new TConfig
		old ResourceRecord[TConfig, TState]
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
					new TConfig
					old ResourceRecord[TConfig, TState]
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

// createAndVerify creates a resource and verifies it's in StateUp
func (rm *ResourceManager[TConfig, TState]) createAndVerify(ctx context.Context, config TConfig) (TState, error) {
	state, err := rm.handler.Create(ctx, config)
	if err != nil {
		return state, shared.WrapError(err, "failed to create resource")
	}

	if err := rm.errorIfNotUp(ctx, state); err != nil {
		return state, shared.WrapError(err, "failed to verify created resource")
	}

	return state, nil
}

// errorIfNotUp checks if a resource is in StateUp and returns an error if not
func (rm *ResourceManager[TConfig, TState]) errorIfNotUp(ctx context.Context, state TState) error {
	status, err := rm.handler.CheckFromState(ctx, state)
	if err != nil {
		return shared.WrapError(err, "failed to verify resource status")
	}
	if status != domain.StateUp {
		return shared.WrapError(nil, "resource not in up state")
	}
	return nil
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

		// Check if resource already exists (important for persistent resources)
		existingState, status, err := rm.handler.CheckFromConfig(ctx, config)

		var state TState

		if status == domain.StateUp {
			// Resource already exists, reuse it
			logger.Infof("Resource already exists, reusing", map[string]any{
				"handler": rm.handlerName,
			})
			state = existingState

			// Verify the existing state is still valid
			if err := rm.errorIfNotUp(ctx, state); err != nil {
				// Resource was detected but is no longer up, recreate it
				logger.Warnf("Existing resource is stale, recreating", map[string]any{
					"handler": rm.handlerName,
				})
				if state, err = rm.createAndVerify(ctx, config); err != nil {
					return err
				}
			}
		} else {
			// Resource doesn't exist or state is unknown - create it
			if status == domain.StateUnknown {
				logger.Warnf("Unable to check if resource exists, attempting creation", map[string]any{
					"handler": rm.handlerName,
					"error":   err,
				})
			}

			if state, err = rm.createAndVerify(ctx, config); err != nil {
				return err
			}
		}

		// Add to persistent storage
		if err := rm.addToRegistry(ResourceRecord[TConfig, TState]{
			State:  state,
			Config: config,
		}); err != nil {
			return shared.WrapError(err, "failed to add to registry")
		}
	}
	return nil
}

// updateResources updates the specified resources
func (rm *ResourceManager[TConfig, TState]) updateResources(
	ctx context.Context,
	toUpdate []struct {
		new TConfig
		old ResourceRecord[TConfig, TState]
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

		// Create and verify new resource
		newState, err := rm.createAndVerify(ctx, update.new)
		if err != nil {
			return shared.WrapError(err, "failed to create updated resource")
		}

		// Update persistent storage
		if err := rm.removeFromRegistry(update.old); err != nil {
			return shared.WrapError(err, "failed to remove old record from registry")
		}

		if err := rm.addToRegistry(ResourceRecord[TConfig, TState]{
			Config: update.new,
			State:  newState,
		}); err != nil {
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
	return rm.registry.Add(PersistentResourceEntry{
		HandlerName: rm.handlerName,
		Data:        record,
	})
}

// removeFromRegistry removes a resource record from persistent storage
func (rm *ResourceManager[TConfig, TState]) removeFromRegistry(record ResourceRecord[TConfig, TState]) error {
	return rm.registry.Remove(PersistentResourceEntry{
		HandlerName: rm.handlerName,
		Data:        record,
	})
}

// Stop removes all resources managed by this typed manager (tracked + detected)
func (rm *ResourceManager[TConfig, TState]) Stop(ctx context.Context, configs []TConfig) error {
	// Get currently tracked resources
	currentRecords := rm.getCurrentRecords()

	// Create a map of tracked resources for efficient lookup
	trackedMap := make(map[string]ResourceRecord[TConfig, TState])
	for _, record := range currentRecords {
		key := rm.getConfigKey(record.Config)
		trackedMap[key] = record
	}

	// Collect all resources to remove (tracked + detected)
	allResourcesToRemove := currentRecords

	// Check for untracked resources matching configs
	for _, config := range configs {
		key := rm.getConfigKey(config)

		// Skip if already tracked
		if _, exists := trackedMap[key]; exists {
			continue
		}

		// Try to check untracked resource directly from config
		state, status, err := rm.handler.CheckFromConfig(ctx, config)

		switch status {
		case domain.StateUp:
			// Resource is running and untracked, add it to removal list
			logger.Infof("Found untracked running resource", map[string]any{
				"handler": rm.handlerName,
			})
			allResourcesToRemove = append(allResourcesToRemove, ResourceRecord[TConfig, TState]{
				State:  state,
				Config: config,
			})

		case domain.StateUnknown:
			// Unable to determine state, skip for safety
			logger.Warnf("Unable to determine untracked resource state, skipping", map[string]any{
				"handler": rm.handlerName,
				"error":   err,
			})
		}
	}

	logger.Debugf("Total resources to remove", map[string]any{
		"handler":      rm.handlerName,
		"remove_count": len(allResourcesToRemove),
	})

	return rm.removeResources(ctx, allResourcesToRemove)
}
