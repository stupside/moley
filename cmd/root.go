package cmd

import (
	"fmt"
	"moley/config"

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
	RootCmd.SetVersionTemplate(`{{with .Name}}{{printf "%s " .}}{{end}}{{printf "version %s" .Version}}
`)
}
