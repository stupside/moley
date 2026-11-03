package cloudflare

import (
	"context"
	"errors"
	"testing"

	"github.com/stupside/moley/v2/internal/domain"
	tunnel "github.com/stupside/moley/v2/internal/features/tunnel/usecase"
)

func TestRunnerDryRunDoesNotRequireConfigurationOrSignal(t *testing.T) {
	r := NewRunner(nil, true)
	identity, err := r.Run(context.Background(), &domain.Tunnel{Name: "test"})
	if err != nil || identity.PID != 0 {
		t.Fatalf("dry run: %+v, %v", identity, err)
	}
	// An invalid PID would fail validation in the OS adapter. Dry-run must not reach it.
	if err := r.Stop(context.Background(), tunnel.ProcessIdentity{PID: -1}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := r.Wait(ctx, identity); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestRunnerRejectsUnverifiableProcess(t *testing.T) {
	r := NewRunner(nil, false)
	alive, err := r.Check(context.Background(), tunnel.ProcessIdentity{PID: -1, Command: "cloudflared"})
	if alive || err == nil {
		t.Fatalf("unverifiable identity: %v, %v", alive, err)
	}
}

type unavailableConfiguration struct{ err error }

func (c unavailableConfiguration) GetConfigurationPath(context.Context, *domain.Tunnel) (string, error) {
	return "", c.err
}
func (c unavailableConfiguration) GetID(context.Context, *domain.Tunnel) (string, error) {
	panic("must not resolve UUID after configuration failure")
}

func TestRunnerConfigurationFailureDoesNotLaunch(t *testing.T) {
	unavailable := errors.New("configuration unavailable")
	r := NewRunner(unavailableConfiguration{err: unavailable}, false)
	if _, err := r.Run(context.Background(), &domain.Tunnel{}); !errors.Is(err, unavailable) {
		t.Fatal(err)
	}
	if err := r.StopOwned(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestRealRunnerDoesNotAdoptDryRunSnapshot(t *testing.T) {
	r := NewRunner(nil, false)
	identity := tunnel.ProcessIdentity{}
	alive, err := r.Check(context.Background(), identity)
	if err != nil || alive {
		t.Fatalf("synthetic identity was adopted: %v, %v", alive, err)
	}
	if err := r.Wait(context.Background(), identity); err == nil {
		t.Fatal("wait accepted a synthetic identity")
	}
}
