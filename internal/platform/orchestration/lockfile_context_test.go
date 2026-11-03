package orchestration_test

import (
	"context"
	"errors"
	"testing"
	"time"

	framework "github.com/stupside/moley/v2/internal/platform/orchestration"
)

func TestLockAcquisitionHonorsCancellation(t *testing.T) {
	chdir(t)
	held, err := framework.LoadLockFile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer held.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	acquired, err := framework.LoadLockFile(ctx)
	if acquired != nil {
		acquired.Close()
		t.Fatal("acquired an already held lock")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline: %v", err)
	}
}

type failingStopHandler struct{ *testHandler }

func (h *failingStopHandler) Destroy(context.Context, testOutput) error {
	return errors.New("process still running")
}

func TestStopPreservesDependenciesOfFailedResource(t *testing.T) {
	chdir(t)
	dependency := newTestHandler("config")
	process := &failingStopHandler{newTestHandler("process")}
	independent := newTestHandler("independent")
	register := func(r *framework.Reconciler) {
		framework.Register(r, dependency, staticResolver("item"))
		framework.Register(r, process, staticResolver("item"), "config")
		framework.Register(r, independent, staticResolver("item"))
	}
	r, err := framework.NewReconciler(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	register(r)
	if err := r.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	r, err = framework.NewReconciler(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	register(r)
	if err := r.Stop(context.Background()); err == nil {
		t.Fatal("expected cleanup failure")
	}
	if dependency.destroyed["item"] {
		t.Fatal("destroyed dependency of live process")
	}
	if !independent.destroyed["item"] {
		t.Fatal("independent resource was not cleaned up")
	}
	lf, err := framework.LoadLockFile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer lf.Close()
	if len(lf.Entries) != 2 {
		t.Fatalf("expected retained process and config: %+v", lf.Entries)
	}
}
