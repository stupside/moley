package tunnel

import (
	"context"

	"github.com/stupside/moley/v2/internal/platform/infrastructure/config"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/shared"

	"github.com/urfave/cli/v3"
)

const (
	dryRunFlag     = "dry-run"
	configPathFlag = "config"
)

var Cmd = &cli.Command{
	Name:  "tunnel",
	Usage: "Manage Cloudflare tunnels",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  dryRunFlag,
			Value: false,
			Usage: "Simulate actions without making any changes",
		},
		&cli.StringFlag{
			Name:  configPathFlag,
			Value: "moley.yml",
			Usage: "Path to the tunnel configuration file",
		},
	},
	Commands: []*cli.Command{
		runCmd,
		stopCmd,
		{
			Name:  "init",
			Usage: "Initialize a new tunnel configuration file",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				// Load (or create) tunnel config; creation writes default if file doesn't exist
				mgr, err := config.NewTunnelManager("moley.yml")
				if err != nil {
					return shared.WrapError(err, "initialize tunnel config failed")
				}

				example, err := config.ExampleTunnelConfig()
				if err != nil {
					return shared.WrapError(err, "create example tunnel config failed")
				}

				if err := mgr.Override(example); err != nil {
					return shared.WrapError(err, "override tunnel config failed")
				}

				logger.Info("Initialized tunnel configuration at ./moley.yml")
				return nil
			},
		},
	},
}
