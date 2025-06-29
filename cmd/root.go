package cmd

import (
	"fmt"
	"moley/config"
	"moley/internal/version"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "moley",
	Short: "A simple CLI tool for exposing local services through Cloudflare Tunnel",
	Long:  "Moley makes it easy to expose your local development services through Cloudflare Tunnel with your own domain names.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		if err := config.Load(); err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
		return nil
	},
}

// Execute runs the root command
func Execute() error {
	return RootCmd.Execute()
}

func init() {
	// Only set the version for --version flag
	RootCmd.Version = version.Version

	RootCmd.AddCommand(&cobra.Command{
		Use:   "info",
		Short: "Show detailed build information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version %s\nCommit: %s\nBuildTime: %s\n", version.Version, version.Commit, version.BuildTime)
		},
	})
}
