package config

import (
	"context"

	"github.com/stupside/moley/v2/internal/platform/infrastructure/config"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/shared"

	"github.com/urfave/cli/v3"
)

const (
	cloudflareTokenFlag = "cloudflare.token"
)

var setCmd = &cli.Command{
	Name:        "set",
	Usage:       "Set Moley configuration values",
	Description: "Set Moley configuration values using command-line flags.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     cloudflareTokenFlag,
			Usage:    "Cloudflare API token",
			Required: true,
		},
	},
	Action: execSet,
}

func execSet(ctx context.Context, cmd *cli.Command) error {
	logger.Info("Editing configuration")

	// Load global config
	mgr, err := config.NewGlobalManager(cmd)
	if err != nil {
		return shared.WrapError(err, "create global config manager failed")
	}

	if err := mgr.Update(func(cfg *config.GlobalConfig) {
		cfg.Cloudflare.Token = cmd.String(cloudflareTokenFlag)
	}); err != nil {
		return shared.WrapError(err, "update global config failed")
	}

	logger.Info("Configuration saved successfully")
	return nil
}

