// Package runtime coordinates session startup, observation, and cleanup.
package runtime

import (
	"context"
	"os/signal"

	sys "github.com/stupside/moley/v2/internal/platform/system"
)

// Stopper is sufficient for standalone cleanup.
type Stopper interface{ Stop(context.Context) error }

// Lifecycle supports startup and cleanup without requiring ongoing observation.
// Start must finish its provisioning work before returning. Both methods must
// honor cancellation; cleanup receives its own bounded context.
type Lifecycle interface {
	Stopper
	Start(context.Context) error
}

// Runnable adds the observation required by a foreground session.
type Runnable interface {
	Lifecycle
	// Wait returns when the process exits or ctx is canceled.
	Wait(context.Context) error
}

func StartManaged(ctx context.Context, r Runnable) error {
	sigCtx, cancel := signal.NotifyContext(ctx, sys.GetShutdownSignals()...)
	defer cancel()
	return newController().run(sigCtx, r)
}

func StartDetached(ctx context.Context, r Lifecycle) error {
	sigCtx, cancel := signal.NotifyContext(ctx, sys.GetShutdownSignals()...)
	defer cancel()
	return newController().runDetached(sigCtx, r)
}

func StopManaged(ctx context.Context, r Stopper) error {
	sigCtx, cancel := signal.NotifyContext(ctx, sys.GetShutdownSignals()...)
	defer cancel()
	return newController().cleanup(sigCtx, r, nil)
}
