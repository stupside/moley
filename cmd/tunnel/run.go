package tunnel

import (
	"context"
	"fmt"

	logger "github.com/stupside/moley/v2/internal/platform/logging"
	shared "github.com/stupside/moley/v2/internal/platform/runtime"

	"github.com/urfave/cli/v3"
)

const (
	detachFlag = "detach"
)

var runCmd = &cli.Command{
	Name:        "run",
	Usage:       "Run a Cloudflare tunnel",
	Description: "Run a Cloudflare tunnel with the specified configuration. This command will start the tunnel service.",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  detachFlag,
			Value: false,
			Usage: "Run the tunnel in the background (detached mode)",
		},
	},
	Action: execRun,
}

func execRun(ctx context.Context, cmd *cli.Command) error {
	detach := cmd.Bool(detachFlag)

	logger.Infof("Starting tunnel", map[string]any{
		"dry":    cmd.Bool(dryRunFlag),
		"detach": detach,
		"config": cmd.String(configPathFlag),
	})

	tunnelService, err := buildTunnelService(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to build tunnel service: %w", err)
	}

	if detach {
		if err := tunnelService.Start(ctx); err != nil {
			return fmt.Errorf("failed to run tunnel service: %w", err)
		}
	} else {
		if err := shared.StartManaged(ctx, tunnelService); err != nil {
			return fmt.Errorf("failed to start tunnel service: %w", err)
		}
	}

	logger.Info("Run completed")
	return nil
}
