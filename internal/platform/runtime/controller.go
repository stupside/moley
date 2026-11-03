package runtime

import (
	"context"
	"errors"
	"time"
)

func (c *controller) cleanup(ctx context.Context, r Stopper, cause error) error {
	if err := c.advance(stop); err != nil {
		return errors.Join(cause, err)
	}
	stopCtx, cancel := newStopContext(ctx)
	defer cancel()
	err := r.Stop(stopCtx)
	end := cleaned
	if err != nil {
		end = cleanupError
	}
	return errors.Join(cause, err, c.advance(end))
}

func (c *controller) startup(ctx context.Context, r Lifecycle) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if err := c.advance(start); err != nil {
		return false, err
	}
	startCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- r.Start(startCtx) }()

	var startErr error
	select {
	case startErr = <-result:
	case <-ctx.Done():
		cancel()
		if err := c.advance(cancelStart); err != nil {
			return false, err
		}
		// Cleanup must never race a worker that can still create resources.
		startErr = <-result
	}
	if ctx.Err() != nil {
		if c.machine.Is(string(starting)) {
			if err := c.advance(cancelStart); err != nil {
				return false, err
			}
		}
		// Cancellation is a normal foreground shutdown. Preserve other failures.
		return false, c.cleanup(ctx, r, shutdownCause(ctx, startErr))
	}
	if startErr != nil {
		return false, c.cleanup(ctx, r, startErr)
	}
	return true, nil
}

func (c *controller) runDetached(ctx context.Context, r Lifecycle) error {
	ready, err := c.startup(ctx, r)
	if !ready {
		return err
	}
	return c.advance(detach)
}

func (c *controller) run(ctx context.Context, r Runnable) error {
	ready, err := c.startup(ctx, r)
	if !ready {
		return err
	}

	if err := c.advance(started); err != nil {
		return err
	}

	// Wait already accepts cancellation. Calling it directly avoids an extra
	// goroutine, result channel, and join protocol for the same operation.
	exitErr := r.Wait(ctx)

	if ctx.Err() != nil {
		return c.cleanup(ctx, r, shutdownCause(ctx, exitErr))
	}
	if exitErr == nil {
		exitErr = errors.New("tunnel process exited unexpectedly")
	}
	return c.cleanup(ctx, r, exitErr)
}

func newStopContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(parent), 30*time.Second)
}

// A joined cancellation must not hide unrelated provisioning or persistence errors.
func onlyCanceled(err error) bool {
	if many, ok := err.(interface{ Unwrap() []error }); ok {
		children := many.Unwrap()
		if len(children) == 0 {
			return false
		}
		for _, child := range children {
			if !onlyCanceled(child) {
				return false
			}
		}
		return true
	}
	if wrapped := errors.Unwrap(err); wrapped != nil {
		return onlyCanceled(wrapped)
	}
	return err == context.Canceled
}

func shutdownCause(ctx context.Context, err error) error {
	if onlyCanceled(err) {
		err = nil
	}
	if ctx.Err() == context.DeadlineExceeded {
		return errors.Join(err, ctx.Err())
	}
	return err
}
