package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
)

const (
	ResourceLockFile = "moley.lock"
)

type ResourceLock struct {
	Resources map[string]Resource `json:"resources"`
}

func LoadResourceLock() (*ResourceLock, error) {
	lock := &ResourceLock{Resources: make(map[string]Resource)}

	if _, err := os.Stat(ResourceLockFile); os.IsNotExist(err) {
		return lock, nil
	}

	data, err := os.ReadFile(ResourceLockFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read resource lock file: %w", err)
	}

	if err := json.Unmarshal(data, lock); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resource lock: %w", err)
	}

	if lock.Resources == nil {
		lock.Resources = make(map[string]Resource)
	}

	return lock, nil
}

func (rl *ResourceLock) Save() error {
	data, err := json.Marshal(rl)
	if err != nil {
		return fmt.Errorf("failed to marshal resource lock: %w", err)
	}

	if err := os.WriteFile(ResourceLockFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write resource lock file: %w", err)
	}

	return nil
}

func (rl *ResourceLock) Add(resource Resource) error {
	hash, err := resource.Hash()
	if err != nil {
		return fmt.Errorf("failed to hash resource: %w", err)
	}

	rl.Resources[hash] = resource
	return rl.Save()
}

func (rl *ResourceLock) Remove(resource Resource) error {
	hash, err := resource.Hash()
	if err != nil {
		return fmt.Errorf("failed to hash resource: %w", err)
	}

	delete(rl.Resources, hash)
	return rl.Save()
}

func (rl *ResourceLock) Has(resource Resource) bool {
	hash, err := resource.Hash()
	if err != nil {
		return false
	}

	_, exists := rl.Resources[hash]
	return exists
}

// DiffResources compares desired vs actual resources and returns what to remove/add.
func (rl *ResourceLock) DiffResources(desiredResources []Resource) (removed, added []Resource) {
	// Find resources to add (desired but not in lock)
	for _, resource := range desiredResources {
		if !rl.Has(resource) {
			added = append(added, resource)
		}
	}

	// Find resources to remove (in lock but not desired)
	desired := make(map[string]bool)
	for _, resource := range desiredResources {
		if hash, err := resource.Hash(); err == nil {
			desired[hash] = true
		}
	}

	for hash, resource := range rl.Resources {
		if !desired[hash] {
			removed = append(removed, resource)
		}
	}

	return removed, added
}

// SyncResource syncs a single resource's state with the lockfile.
func (rl *ResourceLock) SyncResource(ctx context.Context, handler ResourceHandler, resource Resource) error {
	finalStatus, err := handler.Status(ctx, resource.Payload)
	if err != nil {
		logger.Debugf("Failed to get final status for lockfile sync", map[string]any{
			"resource": handler.Name(ctx),
			"error":    err.Error(),
		})
		return nil // Don't fail the entire operation for lockfile sync issues
	}

	if finalStatus == domain.StateUp {
		if err := rl.Add(resource); err != nil {
			logger.Debugf("Failed to add resource to lockfile", map[string]any{
				"resource": handler.Name(ctx),
				"error":    err.Error(),
			})
		}
	} else {
		if err := rl.Remove(resource); err != nil {
			logger.Debugf("Failed to remove resource from lockfile", map[string]any{
				"resource": handler.Name(ctx),
				"error":    err.Error(),
			})
		}
	}
	return nil
}
