package shared

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Runnable interface {
	Stop(ctx context.Context) error
	Start(ctx context.Context) error
}

func Run(ctx context.Context, r Runnable) error {
	sigCtx, cancel := signal.NotifyContext(ctx, osInterruptSignals()...)
	defer cancel()

	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		if err := r.Start(sigCtx); err != nil {
			errCh <- err
		}
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

func osInterruptSignals() []os.Signal {
	return []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP}
}

func newStopContext(parent context.Context) (context.Context, context.CancelFunc) {
	base := context.WithoutCancel(parent)
	return context.WithTimeout(base, 30*time.Second)
}
