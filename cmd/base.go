package cmd

import (
	"context"
	"os"

	"github.com/stupside/moley/v2/cmd/config"
	"github.com/stupside/moley/v2/cmd/tunnel"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/version"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
)

const (
	logLevelFlag = "log-level"
)

var app = &cli.Command{
	Name:        "moley",
	Usage:       "Expose local services through Cloudflare Tunnel",
	Description: "Expose local development services through Cloudflare Tunnel using your own domain names.",
	Version:     version.Version,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  logLevelFlag,
			Value: zerolog.LevelInfoValue,
			Usage: "Log level (trace, debug, info, warn, error, fatal, panic, disabled)",
		},
	},
}

func Execute() error {
	return app.Run(context.Background(), os.Args)
}

func init() {
	app.Commands = []*cli.Command{
		config.Cmd,
		tunnel.Cmd,
		{
			Name:  "info",
			Usage: "Show detailed build information",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				level := cmd.String(logLevelFlag)
				logger.Infof("Build information", map[string]any{
					"commit":    version.Commit,
					"version":   version.Version,
					"logLevel":  level,
					"buildTime": version.BuildTime,
				})
				return nil
			},
		},
	}
}
