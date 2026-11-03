package tunnel

import (
	"context"
	"errors"
	"testing"

	"github.com/stupside/moley/v2/internal/domain"
	framework "github.com/stupside/moley/v2/internal/platform/orchestration"
)

// A lifecycle adapter does not need to implement session observation or cleanup.
type runnerStub struct {
	identity ProcessIdentity
	alive    bool
	checkErr error
	stopped  ProcessIdentity
}

func (r *runnerStub) Run(context.Context, *domain.Tunnel) (ProcessIdentity, error) {
	return r.identity, nil
}
func (r *runnerStub) Check(context.Context, ProcessIdentity) (bool, error) {
	return r.alive, r.checkErr
}
func (r *runnerStub) Stop(_ context.Context, identity ProcessIdentity) error {
	r.stopped = identity
	return nil
}

func TestRunCheckPreservesObservationError(t *testing.T) {
	denied := errors.New("cannot inspect process")
	for _, tc := range []struct {
		name  string
		alive bool
		err   error
		want  framework.Status
	}{
		{"alive", true, nil, framework.StatusUp},
		{"exited", false, nil, framework.StatusDown},
		{"unobservable", false, denied, framework.StatusUnknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			adapter := &runnerStub{alive: tc.alive, checkErr: tc.err}
			status, err := NewRunHandler(adapter).Check(context.Background(), RunOutput{})
			if status != tc.want || !errors.Is(err, tc.err) {
				t.Fatalf("got %s, %v", status, err)
			}
		})
	}
}

func TestRunLifecyclePreservesAdapterIdentity(t *testing.T) {
	adapter := &runnerStub{identity: ProcessIdentity{PID: 42, Command: "custom-runner"}}
	handler := NewRunHandler(adapter)
	output, err := handler.Create(context.Background(), RunInput{TunnelName: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if output.Process != adapter.identity || output.TunnelName != "test" {
		t.Fatalf("lost identity: %+v", output)
	}
	if err := handler.Destroy(context.Background(), output); err != nil {
		t.Fatal(err)
	}
	if adapter.stopped != adapter.identity {
		t.Fatalf("stopped wrong process: %+v", adapter.stopped)
	}
}
