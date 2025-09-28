package config

import (
	"github.com/urfave/cli/v3"
)

var Cmd = &cli.Command{
	Name:        "config",
	Usage:       "Edit Moley configuration",
	Description: "Edit Moley configuration. You can set any value in the Moley config file using command-line flags.",
	Commands: []*cli.Command{
		setCmd,
	},
}
