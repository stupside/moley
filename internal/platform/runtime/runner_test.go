package runtime

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"
)

type runnable struct {
	start func(context.Context) error
	stop  func(context.Context) error
	wait  func(context.Context) error
}

func (r runnable) Start(ctx context.Context) error { return r.start(ctx) }
func (r runnable) Stop(ctx context.Context) error  { return r.stop(ctx) }
func (r runnable) Wait(ctx context.Context) error  { return r.wait(ctx) }

func TestStartupFailurePreservesCleanupError(t *testing.T) {
	startErr, stopErr := errors.New("provisioning failed"), errors.New("cleanup failed")
	c := newController()
	err := c.run(context.Background(), runnable{
		start: func(context.Context) error { return startErr },
		stop: func(ctx context.Context) error {
			if ctx.Err() != nil {
				t.Error("cleanup context canceled")
			}
			if _, ok := ctx.Deadline(); !ok {
				t.Error("cleanup must have a deadline")
			}
			return stopErr
		},
	})
	if !errors.Is(err, startErr) || !errors.Is(err, stopErr) {
		t.Fatalf("lost error: %v", err)
	}
	if !c.machine.Is(string(cleanupFailed)) {
		t.Fatal(c.machine.Current())
	}
}

func TestCancellationJoinsStartupBeforeCleanup(t *testing.T) {
	for _, detachAfterStart := range []bool{false, true} {
		t.Run(map[bool]string{false: "foreground", true: "detached"}[detachAfterStart], func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				release := make(chan struct{})
				startFinished, stopCalls := false, 0
				c := newController()
				result := make(chan error, 1)
				go func() {
					result <- runSession(c, ctx, runnable{
						start: func(ctx context.Context) error {
							<-ctx.Done()
							<-release
							startFinished = true
							return nil // A successful result must not override requested cancellation.
						},
						stop: func(ctx context.Context) error {
							if !startFinished {
								t.Error("cleanup raced startup")
							}
							if ctx.Err() != nil {
								t.Error("cleanup inherited cancellation")
							}
							stopCalls++
							return nil
						},
					}, detachAfterStart)
				}()
				synctest.Wait()
				cancel()
				cancel()
				// All workers are now blocked; assert the intermediate state without sleeps.
				synctest.Wait()
				if stopCalls != 0 || !c.machine.Is(string(canceling)) {
					t.Fatalf("premature cleanup: calls %d, state %s", stopCalls, c.machine.Current())
				}
				close(release)
				synctest.Wait()
				if err := <-result; err != nil {
					t.Fatal(err)
				}
				if stopCalls != 1 || !c.machine.Is(string(stopped)) {
					t.Fatalf("calls %d, state %s", stopCalls, c.machine.Current())
				}
			})
		})
	}
}

func TestImmediateChildExitCleansUp(t *testing.T) {
	for _, exitErr := range []error{nil, errors.New("child failed")} {
		c := newController()
		calls := 0
		err := c.run(context.Background(), runnable{
			start: func(context.Context) error { return nil },
			wait:  func(context.Context) error { return exitErr },
			stop:  func(context.Context) error { calls++; return nil },
		})
		if err == nil || exitErr != nil && !errors.Is(err, exitErr) {
			t.Fatalf("unexpected result: %v", err)
		}
		if calls != 1 || !c.machine.Is(string(stopped)) {
			t.Fatalf("calls %d state %s", calls, c.machine.Current())
		}
	}
}

func TestRunningCancellationWaitsForObserver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	observerFinished := false
	c := newController()
	err := c.run(ctx, runnable{
		start: func(context.Context) error { return nil },
		wait:  func(ctx context.Context) error { cancel(); <-ctx.Done(); observerFinished = true; return ctx.Err() },
		stop: func(context.Context) error {
			if !observerFinished {
				t.Error("observer still running")
			}
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDetachedStartupDoesNotWaitOrStop(t *testing.T) {
	c := newController()
	var startupContext context.Context
	err := c.runDetached(context.Background(), lifecycleOnly{
		start:    func(ctx context.Context) error { startupContext = ctx; return nil },
		stopOnly: func(context.Context) error { t.Error("detached session stopped"); return nil },
	})
	if err != nil || !c.machine.Is(string(detached)) {
		t.Fatalf("%v %s", err, c.machine.Current())
	}
	if startupContext.Err() == nil {
		t.Fatal("startup context was not released")
	}
}

func TestStandaloneStopDoesNotStart(t *testing.T) {
	calls := 0
	err := StopManaged(context.Background(), stopOnly(func(context.Context) error { calls++; return nil }))
	if err != nil || calls != 1 {
		t.Fatalf("calls %d error %v", calls, err)
	}
}

func TestCancellationPreservesJoinedFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	diskErr := errors.New("save failed")
	err := newController().run(ctx, runnable{
		start: func(context.Context) error { cancel(); return errors.Join(context.Canceled, diskErr) },
		stop:  func(context.Context) error { return nil },
	})
	if !errors.Is(err, diskErr) {
		t.Fatalf("lost persistence failure: %v", err)
	}
}

func TestDeadlineDuringStartupIsAnError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		err := newController().run(ctx, runnable{
			start: func(ctx context.Context) error { <-ctx.Done(); return nil },
			stop: func(ctx context.Context) error {
				if ctx.Err() != nil {
					t.Error("cleanup context expired")
				}
				return nil
			},
		})
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("lost deadline: %v", err)
		}
	})
}

func TestCleanupDeadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := newController()
		before := time.Now()
		err := c.cleanup(context.Background(), runnable{
			stop: func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() },
		}, nil)
		if !errors.Is(err, context.DeadlineExceeded) || !c.machine.Is(string(cleanupFailed)) {
			t.Fatalf("error %v, state %s", err, c.machine.Current())
		}
		if elapsed := time.Since(before); elapsed != 30*time.Second {
			t.Fatalf("cleanup budget: %v", elapsed)
		}
	})
}

func runSession(c *controller, ctx context.Context, r Runnable, detached bool) error {
	if detached {
		return c.runDetached(ctx, r)
	}
	return c.run(ctx, r)
}

// These types intentionally lack capabilities their callers do not need.
type stopOnly func(context.Context) error

func (f stopOnly) Stop(ctx context.Context) error { return f(ctx) }

type lifecycleOnly struct {
	stopOnly
	start func(context.Context) error
}

func (l lifecycleOnly) Start(ctx context.Context) error { return l.start(ctx) }
