package config

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "config",
	Short: "Edit Moley configuration",
	Long:  "Edit Moley configuration. You can set any value in the Moley config file using command-line flags.",
}

func init() {
	Cmd.AddCommand(setCmd)
}
