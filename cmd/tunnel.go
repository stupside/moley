package cmd

import (
	"github.com/stupside/moley/cmd/tunnel"

	"github.com/spf13/cobra"
)

var tunnelCmd = &cobra.Command{
	Use:   "tunnel",
	Short: "Manage Cloudflare tunnels",
	Long:  "Commands for creating, configuring, and running Cloudflare tunnels.",
}

func init() {
	tunnelCmd.AddCommand(tunnel.RunCmd)
	tunnelCmd.AddCommand(tunnel.InitCmd)
	rootCmd.AddCommand(tunnelCmd)
}
