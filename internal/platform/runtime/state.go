package runtime

import (
	"context"
	"fmt"

	"github.com/looplab/fsm"
	logger "github.com/stupside/moley/v2/internal/platform/logging"
)

type state string
type event string

const (
	idle          state = "idle"
	starting      state = "starting"
	canceling     state = "canceling"
	running       state = "running"
	stopping      state = "stopping"
	stopped       state = "stopped"
	cleanupFailed state = "cleanup_failed"
	detached      state = "detached"
	start         event = "start"
	started       event = "started"
	cancelStart   event = "cancel_start"
	stop          event = "stop"
	cleaned       event = "cleaned"
	cleanupError  event = "cleanup_error"
	detach        event = "detach"
)

// controller has one owner. Workers return results; they never mutate the FSM.
type controller struct{ machine *fsm.FSM }

func newController() *controller {
	transition := func(e event, dst state, src ...state) fsm.EventDesc {
		sources := make([]string, len(src))
		for i, s := range src {
			sources[i] = string(s)
		}
		return fsm.EventDesc{Name: string(e), Src: sources, Dst: string(dst)}
	}
	return &controller{machine: fsm.NewFSM(string(idle), fsm.Events{
		transition(start, starting, idle),
		transition(started, running, starting),
		transition(cancelStart, canceling, starting),
		transition(stop, stopping, idle, starting, canceling, running),
		transition(cleaned, stopped, stopping),
		transition(cleanupError, cleanupFailed, stopping),
		transition(detach, detached, starting),
	}, fsm.Callbacks{})}
}

func (c *controller) advance(e event) error {
	from := c.machine.Current()
	// Events only change local state. Operation cancellation is handled separately.
	if err := c.machine.Event(context.Background(), string(e)); err != nil {
		return fmt.Errorf("session transition: %w", err)
	}
	logger.Debugf("Session transition", map[string]any{"from": from, "to": c.machine.Current(), "event": e})
	return nil
}
