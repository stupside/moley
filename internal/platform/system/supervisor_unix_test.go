//go:build !windows

package system

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestStubbornProcessHelper(t *testing.T) {
	if os.Getenv("MOLEY_PROCESS_TEST") != "ignore-term" {
		return
	}
	signal.Ignore(syscall.SIGTERM)
	fmt.Println("ready")
	for {
		time.Sleep(time.Hour)
	}
}

func TestStopEscalatesAndReapsUnresponsiveChild(t *testing.T) {
	var p Processes
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	cmd := exec.Command(os.Args[0], "-test.run=^TestStubbornProcessHelper$")
	cmd.Env = append(os.Environ(), "MOLEY_PROCESS_TEST=ignore-term")
	cmd.SysProcAttr = GetProcessAttributes()
	output, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	pid, err := p.Start(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Process.Kill(); <-p.lookup(pid).done }()
	ready := make(chan bool, 1)
	go func() { scanner := bufio.NewScanner(output); ready <- scanner.Scan() && scanner.Text() == "ready" }()
	select {
	case ok := <-ready:
		if !ok {
			t.Fatal("child did not become ready")
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if err := p.Stop(ctx, pid, ""); err != nil {
		t.Fatal(err)
	}
	select {
	case <-p.lookup(pid).done:
	default:
		t.Fatal("child not reaped")
	}
}

func TestRecoveredZombieIsAlreadyExited(t *testing.T) {
	cmd := helperCommand("exit")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	// Deliberately leave the child unreaped until observation has completed.
	defer cmd.Wait()
	var recovered Processes
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := recovered.Wait(ctx, cmd.Process.Pid, GetProcessCommand(cmd.Process.Pid)); err != nil {
		t.Fatalf("exited child was treated as running: %v", err)
	}
}
