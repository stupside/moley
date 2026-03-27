package tunnel

import (
	"context"
	"fmt"

	"github.com/stupside/moley/v2/internal/platform/infrastructure/config"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"

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

				logger.Info("Initializing new tunnel configuration")

				configPath := cmd.String(configPathFlag)

				mgr, err := config.NewTunnelManager(configPath)
				if err != nil {
					return fmt.Errorf("initialize tunnel config failed: %w", err)
				}

				example, err := config.ExampleTunnelConfig()
				if err != nil {
					return fmt.Errorf("create example tunnel config failed: %w", err)
				}

				if err := mgr.Override(example); err != nil {
					return fmt.Errorf("override tunnel config failed: %w", err)
				}

				logger.Infof("Tunnel configuration initialized", map[string]any{
					"config": configPath,
				})
				return nil
			},
		},
	},
}
