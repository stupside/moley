package session

import (
	"context"
	"errors"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/stupside/moley/v2/internal/domain"
	tunnel "github.com/stupside/moley/v2/internal/features/tunnel/usecase"
	framework "github.com/stupside/moley/v2/internal/platform/orchestration"
)

type sessionFixture struct {
	tunnelExists bool
	dnsExists    bool
	runs         int
	waitedPID    int
	events       []string
	exit         error
}

func (f *sessionFixture) Create(context.Context, *domain.Tunnel) (string, error) {
	f.tunnelExists = true
	return "uuid", nil
}
func (f *sessionFixture) Delete(context.Context, *domain.Tunnel) error {
	f.events = append(f.events, "tunnel")
	f.tunnelExists = false
	return nil
}
func (f *sessionFixture) Exists(context.Context, *domain.Tunnel) (bool, error) {
	return f.tunnelExists, nil
}
func (f *sessionFixture) GetID(context.Context, *domain.Tunnel) (string, error) { return "uuid", nil }
func (f *sessionFixture) SaveConfiguration(context.Context, *domain.Tunnel, *domain.Ingress, map[string]tunnel.AccessGuard) error {
	return os.WriteFile("cloudflared.yml", []byte("config"), 0600)
}
func (f *sessionFixture) GetConfigurationPath(context.Context, *domain.Tunnel) (string, error) {
	return "cloudflared.yml", nil
}
func (f *sessionFixture) DeleteConfiguration(context.Context, *domain.Tunnel) error {
	f.events = append(f.events, "config")
	return os.Remove("cloudflared.yml")
}
func (f *sessionFixture) RouteRecord(context.Context, string, string, string) error {
	f.dnsExists = true
	return nil
}
func (f *sessionFixture) DeleteRecord(context.Context, string, string, string) error {
	f.events = append(f.events, "dns")
	f.dnsExists = false
	return nil
}
func (f *sessionFixture) RecordExists(context.Context, string, string, string) (bool, error) {
	return f.dnsExists, nil
}

// This fixture supplies process observations without inspecting or signaling the OS.
func (f *sessionFixture) Run(context.Context, *domain.Tunnel) (tunnel.ProcessIdentity, error) {
	f.runs++
	return tunnel.ProcessIdentity{PID: 42, Command: "cloudflared"}, nil
}
func (f *sessionFixture) Wait(_ context.Context, identity tunnel.ProcessIdentity) error {
	f.waitedPID = identity.PID
	return f.exit
}
func (f *sessionFixture) Stop(context.Context, tunnel.ProcessIdentity) error {
	f.events = append(f.events, "process")
	return nil
}
func (f *sessionFixture) StopOwned(context.Context) error {
	f.events = append(f.events, "owned")
	return nil
}

func TestSessionObservesRecoveredProcessAndOrdersCleanup(t *testing.T) {
	t.Chdir(t.TempDir())
	fixture := &sessionFixture{exit: errors.New("process exited")}
	makeService := func() *Service {
		return NewService(&domain.Tunnel{Name: "test"}, &domain.Ingress{Zone: "example.com", Mode: domain.IngressModeWildcard}, &domain.Access{}, fixture, fixture, fixture, fixture, nil, nil)
	}
	ctx := context.Background()
	first := makeService()
	if err := first.Start(ctx); err != nil {
		t.Fatal(err)
	}
	recovered := makeService()
	if err := recovered.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if fixture.runs != 1 {
		t.Fatalf("started duplicate process: %d", fixture.runs)
	}
	if err := recovered.Wait(ctx); !errors.Is(err, fixture.exit) {
		t.Fatal(err)
	}
	if fixture.waitedPID != 42 {
		t.Fatalf("lost persisted process identity: %d", fixture.waitedPID)
	}
	if err := recovered.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	processIndex, configIndex := slices.Index(fixture.events, "process"), slices.Index(fixture.events, "config")
	if fixture.events[0] != "owned" || processIndex < 0 || configIndex <= processIndex {
		t.Fatalf("unsafe cleanup order: %v", fixture.events)
	}
}

func TestCanceledLockWaitDoesNotCleanUpAnotherSession(t *testing.T) {
	t.Chdir(t.TempDir())
	held, err := framework.LoadLockFile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer held.Close()
	fixture := &sessionFixture{}
	s := NewService(&domain.Tunnel{Name: "test"}, &domain.Ingress{}, &domain.Access{}, fixture, fixture, fixture, fixture, nil, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := s.Start(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected lock timeout: %v", err)
	}
	cleanup, cancelCleanup := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancelCleanup()
	if err := s.Stop(cleanup); err != nil {
		t.Fatalf("cleanup must be a no-op before acquiring ownership: %v", err)
	}
	if len(fixture.events) != 0 {
		t.Fatalf("cleaned up without acquiring the lock: %v", fixture.events)
	}
}

func (f *sessionFixture) Check(context.Context, tunnel.ProcessIdentity) (bool, error) {
	return true, nil
}
