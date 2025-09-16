package shared

import (
	"context"
	"os/signal"
	"time"

	"github.com/stupside/moley/v2/internal/shared/sys"
)

type Runnable interface {
	Stop(ctx context.Context) error
	Start(ctx context.Context) error
}

func StartManaged(ctx context.Context, r Runnable) error {
	sigCtx, cancel := signal.NotifyContext(ctx, sys.GetShutdownSignals()...)
	defer cancel()

	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		if err := r.Start(sigCtx); err != nil {
			errCh <- err
		}
		<-ctx.Done()
	}()

	select {
	case err := <-errCh:
		stopCtx, cancel := newStopContext(sigCtx)
		defer cancel()
		if stopErr := r.Stop(stopCtx); stopErr != nil {
			return err
		}
		return err
	case <-sigCtx.Done():
		stopCtx, cancel := newStopContext(sigCtx)
		defer cancel()
		if stopErr := r.Stop(stopCtx); stopErr != nil {
			return stopErr
		}
		return nil
	}
}

func newStopContext(parent context.Context) (context.Context, context.CancelFunc) {
	base := context.WithoutCancel(parent)
	return context.WithTimeout(base, 30*time.Second)
}
