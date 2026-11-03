package system

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestProcessHelper(t *testing.T) {
	switch os.Getenv("MOLEY_PROCESS_TEST") {
	case "exit":
		os.Exit(7)
	case "wait":
		for {
			time.Sleep(time.Hour)
		}
	}
}

func helperCommand(mode string) *exec.Cmd {
	cmd := exec.Command(os.Args[0], "-test.run=^TestProcessHelper$")
	cmd.Env = append(os.Environ(), "MOLEY_PROCESS_TEST="+mode)
	cmd.SysProcAttr = GetProcessAttributes()
	return cmd
}

func TestProcessExitCanBeObservedMoreThanOnce(t *testing.T) {
	var p Processes
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pid, err := p.Start(ctx, helperCommand("exit"))
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		var exit *exec.ExitError
		if err := p.Wait(ctx, pid, ""); !errors.As(err, &exit) || exit.ExitCode() != 7 {
			t.Fatalf("unexpected exit: %v", err)
		}
	}
	if err := p.Stop(ctx, pid, ""); err != nil {
		t.Fatal(err)
	}
}

func TestStartupCancellationLeavesChildForExplicitCleanup(t *testing.T) {
	var p Processes
	ctx, cancel := context.WithCancel(context.Background())
	pid, err := p.Start(ctx, helperCommand("wait"))
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup, c := context.WithTimeout(context.Background(), 10*time.Second)
		defer c()
		_ = p.StopOwned(cleanup)
	})
	cancel()
	observation, c := context.WithTimeout(context.Background(), 50*time.Millisecond)
	err = p.Wait(observation, pid, "")
	c()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("child did not survive startup cancellation: %v", err)
	}
	cleanup, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()
	if err := p.StopOwned(cleanup); err != nil {
		t.Fatal(err)
	}
	select {
	case <-p.lookup(pid).done:
	default:
		t.Fatal("Stop returned before Wait completed")
	}
	if err := p.StopOwned(cleanup); err != nil {
		t.Fatal("repeated cleanup:", err)
	}
}

func TestCanceledStartupDoesNotSpawn(t *testing.T) {
	var p Processes
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd := helperCommand("wait")
	if _, err := p.Start(ctx, cmd); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if cmd.Process != nil {
		t.Fatal("spawned after cancellation")
	}
}

func TestRecoveredProcessObservationAndStop(t *testing.T) {
	var owner, recovered Processes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pid, err := owner.Start(ctx, helperCommand("wait"))
	if err != nil {
		t.Fatal(err)
	}
	defer owner.StopOwned(ctx)
	if err := recovered.Stop(ctx, pid, GetProcessCommand(pid)); err != nil {
		t.Fatal(err)
	}
	if err := recovered.Wait(ctx, pid, ""); err != nil {
		t.Fatal(err)
	}
}

func TestRecoveredProcessRequiresVerifiableIdentity(t *testing.T) {
	var p Processes
	for _, pid := range []int{0, -1, os.Getpid()} {
		// Negative PID must never reach an OS signal. An empty command name must
		// never authorize signaling a live process owned by another invocation.
		if err := p.Stop(context.Background(), pid, ""); err == nil {
			t.Fatalf("accepted unverifiable identity for PID %d", pid)
		}
		if err := p.Wait(context.Background(), pid, ""); err == nil {
			t.Fatalf("reported unverifiable PID %d as exited", pid)
		}
	}
}
