package cf

import (
	"bytes"
	"context"
	"fmt"
	"moley/internal/errors"
	"os/exec"
)

// execCloudflared is an internal helper function to execute cloudflared commands
func execCloudflared(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "cloudflared", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return "", errors.NewExecutionError(errors.ErrCodeCommandFailed, fmt.Sprintf("cloudflared command failed: %s", cmd.Args), err)
	}

	return out.String(), nil
}
