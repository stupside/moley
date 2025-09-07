package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stupside/moley/v2/cmd/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Edit Moley configuration",
	Long:  "Edit Moley configuration. You can set any value in the Moley config file using command-line flags.",
}

func init() {
	configCmd.AddCommand(config.SetCmd)
	rootCmd.AddCommand(configCmd)
}
