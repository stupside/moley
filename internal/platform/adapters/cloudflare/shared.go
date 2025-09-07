package cloudflare

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/stupside/moley/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/internal/shared"
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

func (c *Cloudflared) Exec() (string, error) {
	args := strings.Join(c.cmd.Args[1:], " ")

	logger.Debugf("Executing cloudflared command", map[string]any{
		"args": args,
	})

	out, err := c.cmd.CombinedOutput()
	if err != nil {
		return "", shared.WrapError(err, fmt.Sprintf("cloudflared failed: %s", args))
	}

	logger.Debugf("Cloudflared command output", map[string]any{
		"output": strings.TrimSpace(string(out)),
	})

	output := strings.TrimSpace(string(out))
	return output, nil
}
