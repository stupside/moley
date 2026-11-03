package system

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

type child struct {
	process *os.Process
	done    chan struct{}
	err     error // published by closing done
}

// Processes owns Wait for children started by this invocation. Recovered PIDs
// are observed separately because they are not children of this process.
type Processes struct {
	mu       sync.Mutex
	children map[int]*child
}

func (p *Processes) lookup(pid int) *child {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.children[pid]
}

func (p *Processes) Start(ctx context.Context, cmd *exec.Cmd) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	c := &child{process: cmd.Process, done: make(chan struct{})}
	p.mu.Lock()
	if p.children == nil {
		p.children = make(map[int]*child)
	}
	p.children[cmd.Process.Pid] = c
	p.mu.Unlock()
	go func() {
		c.err = cmd.Wait()
		close(c.done)
	}()
	return cmd.Process.Pid, nil
}

func (p *Processes) Wait(ctx context.Context, pid int, command string) error {
	if c := p.lookup(pid); c != nil {
		select {
		case <-c.done:
			return c.err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return waitForProcess(ctx, pid, command)
}

func waitForProcess(ctx context.Context, pid int, command string) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		running, err := CheckProcessIdentity(ctx, pid, command)
		if err != nil || !running {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// waitStopped reports observation failures, but not a child's exit status.
// An unsuccessful exit is still a successfully stopped process.
func (p *Processes) waitStopped(ctx context.Context, pid int, command string) error {
	if c := p.lookup(pid); c != nil {
		select {
		case <-c.done:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return waitForProcess(ctx, pid, command)
}

func (p *Processes) Stop(ctx context.Context, pid int, command string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c := p.lookup(pid)
	var proc *os.Process
	if c != nil {
		select {
		case <-c.done:
			return nil
		default:
		}
		proc = c.process
	} else {
		running, err := CheckProcessIdentity(ctx, pid, command)
		if err != nil || !running {
			return err
		}
		proc, err = os.FindProcess(pid)
		if err != nil {
			return err
		}
		defer proc.Release()
	}
	if err := TerminateProcess(proc); err != nil && !errors.Is(err, os.ErrProcessDone) {
		// Only a confirmed exit makes a failed signal harmless.
		if c != nil {
			select {
			case <-c.done:
				return nil
			default:
				return err
			}
		}
		running, checkErr := CheckProcessIdentity(ctx, pid, command)
		if checkErr != nil || running {
			return errors.Join(err, checkErr)
		}
		return nil
	}
	graceful, cancel := context.WithTimeout(ctx, 5*time.Second)
	err := p.waitStopped(graceful, pid, command)
	cancel()
	if err == nil {
		return nil
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	// Recheck a recovered identity immediately before escalating.
	if c == nil {
		running, err := CheckProcessIdentity(ctx, pid, command)
		if err != nil || !running {
			return err
		}
	}
	if err := proc.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}
	return p.waitStopped(ctx, pid, command)
}

// StopOwned also cleans up children created before reconciliation could save
// their outputs, for example when post-creation verification fails.
func (p *Processes) StopOwned(ctx context.Context) error {
	p.mu.Lock()
	pids := make([]int, 0, len(p.children))
	for pid := range p.children {
		pids = append(pids, pid)
	}
	p.mu.Unlock()
	var errs []error
	for _, pid := range pids {
		if err := p.Stop(ctx, pid, ""); err != nil {
			errs = append(errs, fmt.Errorf("stop process %d: %w", pid, err))
		}
	}
	return errors.Join(errs...)
}
