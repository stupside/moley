package cloudflare

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/shared"
)

type Cloudflared struct {
	ctx context.Context
	cmd *exec.Cmd
}

func NewCommand(ctx context.Context, args ...string) *Cloudflared {
	return &Cloudflared{
		ctx: ctx,
		cmd: exec.CommandContext(ctx, "cloudflared", args...),
	}
}

// ExecSync runs the command synchronously and waits for completion, returning output.
func (c *Cloudflared) ExecSync() (string, error) {
	args := c.formatArgs()

	logger.Debugf("Executing cloudflared command synchronously", map[string]any{
		"args": args,
	})

	out, err := c.cmd.CombinedOutput()
	if err != nil {
		return "", shared.WrapError(err, fmt.Sprintf("cloudflared failed: %s", args))
	}

	output := strings.TrimSpace(string(out))
	logger.Debugf("Cloudflared command completed", map[string]any{
		"args":   args,
		"output": output,
	})

	return output, nil
}

// ExecAsync runs the command in the background and returns the PID immediately.
func (c *Cloudflared) ExecAsync() (int, error) {
	args := c.formatArgs()

	logger.Debugf("Starting cloudflared command in background", map[string]any{
		"args": args,
	})

	c.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err := c.cmd.Start(); err != nil {
		return 0, shared.WrapError(err, fmt.Sprintf("failed to start cloudflared: %s", args))
	}

	pid := c.cmd.Process.Pid
	logger.Infof("Cloudflared command started in background", map[string]any{
		"pid":  pid,
		"args": args,
	})

	return pid, nil
}

// formatArgs formats command arguments for logging.
func (c *Cloudflared) formatArgs() string {
	return strings.Join(c.cmd.Args[1:], " ")
}
