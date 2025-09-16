// Package framework provides resource lifecycle management.
package framework

import (
	"context"

	"github.com/stupside/moley/v2/internal/core/domain"
)

// ResourceHandler provides type-safe resource lifecycle management.
// TConfig represents the desired configuration (immutable input)
// TState represents the runtime state after resource creation (mutable state)
type ResourceHandler[TConfig any, TState any] interface {
	// Name returns the handler name for identification
	Name() string

	// Equals compares two configurations for equality
	// Used to detect when resources need to be updated
	Equals(a, b TConfig) bool

	// Create provisions the resource using the given configuration
	// Returns the runtime state needed to manage the resource
	Create(ctx context.Context, config TConfig) (TState, error)

	// Destroy removes the resource using its runtime state
	Destroy(ctx context.Context, state TState) error

	// Status checks the current state of the resource (idempotent)
	Status(ctx context.Context, state TState) (domain.State, error)
}
