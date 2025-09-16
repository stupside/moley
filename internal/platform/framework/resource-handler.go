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
	ResourceNamer
	ResourceChecker[TConfig, TState]
	ResourceCreator[TConfig, TState]
	ResourceDestroyer[TState]
	// Equals compares two configurations for equality
	// Used to detect when resources need to be updated
	Equals(a, b TConfig) bool
}

type ResourceNamer interface {
	// Name returns the resource name for identification
	Name() string
}

type ResourceCreator[TConfig any, TState any] interface {
	// Create provisions the resource using the given configuration
	// Returns the runtime state needed to manage the resource
	Create(ctx context.Context, config TConfig) (TState, error)
}

type ResourceDestroyer[TState any] interface {
	// Destroy removes the resource using its runtime state
	Destroy(ctx context.Context, state TState) error
}

type ResourceChecker[TConfig any, TState any] interface {
	// CheckFromState checks the resource status from its runtime state
	CheckFromState(ctx context.Context, state TState) (domain.State, error)
	// CheckFromConfig checks the resource status from its configuration
	CheckFromConfig(ctx context.Context, config TConfig) (TState, domain.State, error)
}
