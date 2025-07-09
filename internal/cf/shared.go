package cf

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/stupside/moley/internal/shared"
)

// execCloudflared is an internal helper function to execute cloudflared commands
func execCloudflared(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "cloudflared", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return "", shared.WrapError(err, fmt.Sprintf("cloudflared command failed: %s", cmd.Args))
	}

	return out.String(), nil
}
