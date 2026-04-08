package orchestration

import (
	"context"
	"errors"
	"fmt"

	logger "github.com/stupside/moley/v2/internal/platform/logging"
)

// Reconciler manages the lifecycle of multiple typed resources in dependency order.
type Reconciler struct {
	lockFile *LockFile
	outputs  *OutputRegistry
	nodes    []node
	nodeMap  map[string]node
}

// NewReconciler creates a new reconciler backed by the lock file registry.
func NewReconciler() (*Reconciler, error) {
	lf, err := LoadLockFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load registry: %w", err)
	}

	return &Reconciler{
		lockFile: lf,
		outputs:  newOutputRegistry(),
		nodeMap:  make(map[string]node),
	}, nil
}

// Register adds a typed resource with its input resolver and dependency list.
func Register[TInput any, TOutput any](
	r *Reconciler,
	handler Lifecycle[TInput, TOutput],
	resolver InputResolver[TInput],
	deps ...string,
) {
	n := &typedNode[TInput, TOutput]{
		handler:  handler,
		resolver: resolver,
		deps:     deps,
	}
	r.nodes = append(r.nodes, n)
	r.nodeMap[handler.Name()] = n
}

// Start reconciles all registered resources in topological (dependency) order.
func (r *Reconciler) Start(ctx context.Context) error {
	defer func() { _ = r.lockFile.Close() }()

	logger.Debug("Starting reconciliation")

	// Purge orphaned lock entries
	registered := make(map[string]bool, len(r.nodes))
	for _, n := range r.nodes {
		registered[n.name()] = true
	}
	if err := r.lockFile.PurgeOrphans(registered); err != nil {
		logger.Warnf("Failed to purge orphans", map[string]any{"error": err.Error()})
	}

	// Topological sort
	sorted, err := r.topoSort()
	if err != nil {
		return fmt.Errorf("dependency resolution failed: %w", err)
	}

	// Load existing outputs from lock file into registry for recovery
	r.loadOutputs("")

	for _, n := range sorted {
		logger.Debugf("Reconciling", map[string]any{
			"resource": n.name(),
		})

		// Resolve inputs from upstream outputs
		if err := n.resolve(r.outputs); err != nil {
			return fmt.Errorf("input resolution failed: %s: %w", n.name(), err)
		}

		if err := n.reconcile(ctx, r.lockFile); err != nil {
			return fmt.Errorf("reconciliation failed: %s: %w", n.name(), err)
		}

		// Publish this node's outputs for downstream consumers
		r.loadOutputs(n.name())
	}

	logger.Info("Reconciliation completed")
	return nil
}

// Stop tears down all registered resources in reverse topological order.
func (r *Reconciler) Stop(ctx context.Context) error {
	defer func() { _ = r.lockFile.Close() }()

	logger.Debug("Stopping resources")

	sorted, err := r.topoSort()
	if err != nil {
		return fmt.Errorf("dependency resolution failed: %w", err)
	}

	// Load existing outputs for resolver use
	r.loadOutputs("")

	// Resolve all inputs first (needed for Stop to find untracked resources)
	for _, n := range sorted {
		if err := n.resolve(r.outputs); err != nil {
			logger.Warnf("Failed to resolve inputs during stop", map[string]any{
				"resource": n.name(),
				"error":    err.Error(),
			})
		}
	}

	var errs []error

	// Stop in reverse order
	for i := len(sorted) - 1; i >= 0; i-- {
		n := sorted[i]
		logger.Debugf("Stopping", map[string]any{
			"resource": n.name(),
		})

		if err := n.stop(ctx, r.lockFile); err != nil {
			logger.Warnf("Stop failed, continuing cleanup", map[string]any{
				"resource": n.name(),
				"error":    err.Error(),
			})
			errs = append(errs, fmt.Errorf("stop failed: %s: %w", n.name(), err))
		}
	}

	logger.Info("Resources stopped")
	return errors.Join(errs...)
}

// topoSort returns nodes in dependency order using Kahn's algorithm.
func (r *Reconciler) topoSort() ([]node, error) {
	inDegree := make(map[string]int)
	for _, n := range r.nodes {
		inDegree[n.name()] = 0
	}
	for _, n := range r.nodes {
		for _, dep := range n.dependencies() {
			if _, exists := r.nodeMap[dep]; !exists {
				return nil, fmt.Errorf("handler %q depends on unknown handler %q", n.name(), dep)
			}
			inDegree[n.name()]++
		}
	}

	var queue []node
	for _, n := range r.nodes {
		if inDegree[n.name()] == 0 {
			queue = append(queue, n)
		}
	}

	var sorted []node
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		sorted = append(sorted, current)

		for _, n := range r.nodes {
			for _, dep := range n.dependencies() {
				if dep == current.name() {
					inDegree[n.name()]--
					if inDegree[n.name()] == 0 {
						queue = append(queue, n)
					}
				}
			}
		}
	}

	if len(sorted) != len(r.nodes) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return sorted, nil
}

// extractOutput extracts the "output" field from entry data, handling both
// map[string]any (loaded from disk) and concrete Snapshot structs (just created in memory).
func extractOutput(data any) any {
	var snap struct {
		Output any `json:"output"`
	}
	if err := unmarshalData(data, &snap); err != nil {
		return nil
	}
	return snap.Output
}

// loadOutputs populates the output registry from lock file entries.
// If handlerName is empty, all entries are loaded; otherwise only matching entries.
func (r *Reconciler) loadOutputs(handlerName string) {
	for _, entry := range r.lockFile.Entries {
		if handlerName != "" && entry.HandlerName != handlerName {
			continue
		}
		if output := extractOutput(entry.Data); output != nil {
			r.outputs.set(entry.HandlerName, entry.Key, output)
		}
	}
}
