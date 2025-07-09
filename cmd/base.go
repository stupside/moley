package cmd

import (
	"fmt"

	"github.com/stupside/moley/internal/logger"
	"github.com/stupside/moley/internal/version"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "moley",
	Short: "A simple CLI tool for exposing local services through Cloudflare Tunnel",
	Long:  "Moley makes it easy to expose your local development services through Cloudflare Tunnel with your own domain names.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Set the version for --version flag
	rootCmd.Version = version.Version

	// Add info command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "info",
		Short: "Show detailed build information",
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("Moley Build Information:")
			logger.Info(fmt.Sprintf("  Version:    %s", version.Version))
			logger.Info(fmt.Sprintf("  Commit:     %s", version.Commit))
			logger.Info(fmt.Sprintf("  Build Time: %s", version.BuildTime))
		},
	})
}
