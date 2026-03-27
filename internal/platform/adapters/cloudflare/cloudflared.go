package cloudflare

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/shared/sys"
)

type cloudflaredCmd struct {
	cmd *exec.Cmd
}

func newCommand(ctx context.Context, args ...string) *cloudflaredCmd {
	return &cloudflaredCmd{
		cmd: exec.CommandContext(ctx, "cloudflared", args...),
	}
}

// execAsync runs the command in the background and returns the PID immediately.
func (c *cloudflaredCmd) execAsync() (int, error) {
	args := c.formatArgs()

	logger.Debugf("Starting cloudflared command in background", map[string]any{
		"args": args,
	})

	c.cmd.SysProcAttr = sys.GetProcessAttributes()

	if err := c.cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start cloudflared: %s: %w", args, err)
	}

	pid := c.cmd.Process.Pid
	logger.Infof("Cloudflared command started in background", map[string]any{
		"pid":  pid,
		"args": args,
	})

	return pid, nil
}

// formatArgs formats command arguments for logging.
func (c *cloudflaredCmd) formatArgs() string {
	return strings.Join(c.cmd.Args[1:], " ")
}
