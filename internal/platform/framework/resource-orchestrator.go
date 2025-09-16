package framework

import (
	"context"

	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/shared"
)

// ResourceOrchestrator manages the lifecycle of multiple typed resource managers
type ResourceOrchestrator struct {
	registry   *ResourceRegistry
	operations []ResourceOperation
}

// ResourceOperation represents a resource management operation
type ResourceOperation interface {
	Execute(ctx context.Context) error
	Name() string
	Stop(ctx context.Context) error
}

// NewResourceOrchestrator creates a new resource orchestrator
func NewResourceOrchestrator() (*ResourceOrchestrator, error) {
	registry, err := LoadResourceRegistry()
	if err != nil {
		return nil, shared.WrapError(err, "failed to load resource registry")
	}

	return &ResourceOrchestrator{
		registry:   registry,
		operations: make([]ResourceOperation, 0),
	}, nil
}

// AddManager adds a typed resource manager to the orchestrator
func AddManager[TConfig any, TState any](
	ro *ResourceOrchestrator,
	handler ResourceHandler[TConfig, TState],
	configs []TConfig,
) {
	manager := NewResourceManager(handler, ro.registry)
	operation := &ReconcileOperation[TConfig, TState]{
		manager: manager,
		configs: configs,
		name:    handler.Name(),
	}
	ro.operations = append(ro.operations, operation)
}

// Start executes all resource operations to bring resources up
func (ro *ResourceOrchestrator) Start(ctx context.Context) error {
	logger.Debug("Starting resource orchestration")

	for _, operation := range ro.operations {
		logger.Debugf("Executing operation", map[string]any{
			"operation": operation.Name(),
		})

		if err := operation.Execute(ctx); err != nil {
			return shared.WrapError(err, "operation failed: "+operation.Name())
		}
	}

	logger.Info("Resource orchestration completed")
	return nil
}

// Stop removes all resources managed by this orchestrator
func (ro *ResourceOrchestrator) Stop(ctx context.Context) error {
	logger.Debug("Stopping resource orchestration")

	// Execute operations in reverse order for proper teardown
	for i := len(ro.operations) - 1; i >= 0; i-- {
		operation := ro.operations[i]

		logger.Debugf("Stopping operation", map[string]any{
			"operation": operation.Name(),
		})

		// Stop all resources for this operation
		if err := operation.Stop(ctx); err != nil {
			return shared.WrapError(err, "stop operation failed: "+operation.Name())
		}
	}

	logger.Info("Resource orchestration stopped")
	return nil
}

// ReconcileOperation implements ResourceOperation for typed managers
type ReconcileOperation[TConfig any, TState any] struct {
	manager *ResourceManager[TConfig, TState]
	configs []TConfig
	name    string
}

func (tro *ReconcileOperation[TConfig, TState]) Execute(ctx context.Context) error {
	return tro.manager.Reconcile(ctx, tro.configs)
}

func (tro *ReconcileOperation[TConfig, TState]) Name() string {
	return tro.name
}

func (tro *ReconcileOperation[TConfig, TState]) Stop(ctx context.Context) error {
	return tro.manager.Stop(ctx)
}
