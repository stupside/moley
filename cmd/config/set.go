package config

import (
	"context"
	"fmt"

	appconfig "github.com/stupside/moley/v2/internal/app/config"
	logger "github.com/stupside/moley/v2/internal/platform/logging"

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

	mgr, err := appconfig.NewGlobalManager()
	if err != nil {
		return fmt.Errorf("create global config manager failed: %w", err)
	}

	if err := mgr.Update(func(cfg *appconfig.GlobalConfig) {
		cfg.Cloudflare.Token = cmd.String(cloudflareTokenFlag)
	}); err != nil {
		return fmt.Errorf("update global config failed: %w", err)
	}

	logger.Info("Configuration saved successfully")
	return nil
}
