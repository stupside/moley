package cmd

import (
	"github.com/stupside/moley/v2/cmd/config"
	"github.com/stupside/moley/v2/cmd/tunnel"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
	"github.com/stupside/moley/v2/internal/shared"
	"github.com/stupside/moley/v2/internal/version"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	logLevelFlag = "log-level"
)

var rootCmd = &cobra.Command{
	Use:           "moley",
	Short:         "Expose local services through Cloudflare Tunnel",
	Long:          "Expose local development services through Cloudflare Tunnel using your own domain names.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logLevel, err := cmd.Flags().GetString(logLevelFlag)
		if err != nil {
			return shared.WrapError(err, "failed to get log level")
		}
		logLevelValue, err := zerolog.ParseLevel(logLevel)
		if err != nil {
			return shared.WrapError(err, "failed to parse log level")
		}
		logger.InitLogger(logLevelValue)
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Version = version.Version

	addLoggingFlags(rootCmd.PersistentFlags())

	rootCmd.AddCommand(config.Cmd)
	rootCmd.AddCommand(tunnel.Cmd)

	rootCmd.AddCommand(&cobra.Command{
		Use:   "info",
		Short: "Show detailed build information",
		Run: func(cmd *cobra.Command, args []string) {
			level, _ := cmd.Flags().GetString(logLevelFlag)
			logger.Infof("Build information", map[string]any{
				"commit":    version.Commit,
				"version":   version.Version,
				"logLevel":  level,
				"buildTime": version.BuildTime,
			})
		},
	})
}

func addLoggingFlags(flags *pflag.FlagSet) {
	flags.String(logLevelFlag, zerolog.LevelInfoValue, "Log level (trace, debug, info, warn, error, fatal, panic, disabled)")
}
