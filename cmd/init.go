package cmd

import (
	"fmt"
	"os"

	"moley/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a new moly.yml config file",
	Long:  "Creates a new moly.yml configuration file with default settings. This is the first step to get started with Moly.",
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat("moly.yml"); err == nil {
			color.Yellow("moly.yml already exists. Overwrite? [y/N]: ")
			var resp string
			fmt.Scanln(&resp)
			if resp != "y" && resp != "Y" {
				color.Red("Aborted.")
				return
			}
		}
		err := config.CreateDefaultConfig()
		if err != nil {
			color.Red("Failed to create moly.yml: %v", err)
			return
		}
		color.Green("Created moly.yml!")
		color.Cyan("Next steps:")
		color.Cyan("  1. Edit moly.yml with your Cloudflare API token and domain")
		color.Cyan("  2. Configure your applications in the apps section")
		color.Cyan("  3. Run 'moly run' to deploy and start the tunnel")
		color.Cyan("  4. Use Ctrl+C to stop the tunnel and clean up resources")
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}
