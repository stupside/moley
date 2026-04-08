package tunnel

import (
	"context"
	"fmt"

	logger "github.com/stupside/moley/v2/internal/platform/logging"

	"github.com/urfave/cli/v3"
)

var stopCmd = &cli.Command{
	Name:   "stop",
	Usage:  "Bring the tunnel down",
	Action: execStop,
}

func execStop(ctx context.Context, cmd *cli.Command) error {
	logger.Infof("Bringing tunnel down", map[string]any{
		"dry":    cmd.Bool(dryRunFlag),
		"config": cmd.String(configPathFlag),
	})

	tunnelService, err := buildTunnelService(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to build tunnel service: %w", err)
	}

	if err := tunnelService.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop tunnel service: %w", err)
	}

	return nil
}
