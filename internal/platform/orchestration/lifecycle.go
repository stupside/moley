// Package orchestration provides DAG-based resource lifecycle management.
package orchestration

import "context"

// Status represents the operational status of a resource (up, down, or unknown).
type Status string

const (
	StatusUp      Status = "up"
	StatusDown    Status = "down"
	StatusUnknown Status = "unknown"
)

// Lifecycle defines how to create, destroy, and check a specific resource type.
// TInput represents the full input (including upstream outputs).
// TOutput represents the runtime output after resource creation.
// Change detection is handled by the framework via input hashing — no Equals() needed.
type Lifecycle[TInput any, TOutput any] interface {
	// Name returns the handler name for identification
	Name() string
	// Key generates a unique key for the given input
	Key(input TInput) string
	// Create provisions the resource using the given input
	Create(ctx context.Context, input TInput) (TOutput, error)
	// Destroy removes the resource using its output
	Destroy(ctx context.Context, output TOutput) error
	// Check verifies the resource status from its output
	Check(ctx context.Context, output TOutput) (Status, error)
	// Recover discovers a resource from its input when no lock entry exists
	Recover(ctx context.Context, input TInput) (TOutput, Status, error)
}
