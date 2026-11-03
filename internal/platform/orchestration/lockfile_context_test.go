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
