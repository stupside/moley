package framework

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/shared"
)

// ResourceManager manages the lifecycle of multiple resources.
type ResourceManager struct {
	handlers  map[string]ResourceHandler
	resources []Resource
	lock      *ResourceLock
}

// NewResourceManager creates a new resource manager.
func NewResourceManager(handlers map[string]ResourceHandler, resources []Resource) (*ResourceManager, error) {
	lock, err := LoadResourceLock()
	if err != nil {
		return nil, shared.WrapError(err, "failed to load resource lock")
	}

	return &ResourceManager{
		handlers:  handlers,
		resources: resources,
		lock:      lock,
	}, nil
}

// Start brings resources up by comparing desired state vs previous actual state.
func (rm *ResourceManager) Start(ctx context.Context) error {
	logger.Debug("Starting resources")

	removed, added := rm.lock.DiffResources(rm.resources)

	// Remove obsolete resources first, then add new ones
	if err := rm.processResources(ctx, domain.StateDown, removed); err != nil {
		return shared.WrapError(err, "failed to process removed resources")
	}

	if err := rm.processResources(ctx, domain.StateUp, added); err != nil {
		return shared.WrapError(err, "failed to process added resources")
	}

	logger.Info("Resources started")
	return nil
}

// Stop brings all resources down.
func (rm *ResourceManager) Stop(ctx context.Context) error {
	logger.Debug("Stopping resources")

	if err := rm.processResources(ctx, domain.StateDown, rm.resources); err != nil {
		return shared.WrapError(err, "failed to stop resources")
	}

	logger.Info("Resources stopped")
	return nil
}

func (rm *ResourceManager) processResources(ctx context.Context, state domain.State, resources []Resource) error {
	// Process resources in reverse order when tearing down
	if state == domain.StateDown {
		for i := len(resources) - 1; i >= 0; i-- {
			if err := rm.processResource(ctx, state, resources[i]); err != nil {
				return err
			}
		}
		return nil
	}

	// Forward order for bringing up
	for _, resource := range resources {
		if err := rm.processResource(ctx, state, resource); err != nil {
			return err
		}
	}
	return nil
}

func (rm *ResourceManager) processResource(ctx context.Context, state domain.State, resource Resource) error {
	handler, ok := rm.handlers[resource.Handler]
	if !ok {
		return fmt.Errorf("unknown handler: %s", resource.Handler)
	}

	currentStatus, err := handler.Status(ctx, resource.Payload)
	if err != nil {
		return shared.WrapError(err, fmt.Sprintf("failed to get status for resource %s", handler.Name(ctx)))
	}

	// Perform action if needed
	if currentStatus == state {
		logger.Infof("Skipping resource, already in desired state", map[string]any{
			"action":   state,
			"resource": handler.Name(ctx),
		})
	} else {
		logger.Infof("Processing resource", map[string]any{
			"action":   state,
			"resource": handler.Name(ctx),
		})

		if err := rm.executeAction(ctx, state, handler, resource.Payload); err != nil {
			return shared.WrapError(err, fmt.Sprintf("failed to process resource %s", handler.Name(ctx)))
		}
	}

	// Sync lockfile with actual final state
	return rm.lock.SyncResource(ctx, handler, resource)
}

func (rm *ResourceManager) executeAction(ctx context.Context, state domain.State, handler ResourceHandler, payload any) error {
	switch state {
	case domain.StateUp:
		return handler.Up(ctx, payload)
	case domain.StateDown:
		return handler.Down(ctx, payload)
	default:
		return fmt.Errorf("unknown state: %s", state)
	}
}
