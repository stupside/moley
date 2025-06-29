package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate shell completion scripts",
	Long: `To load completions:

Bash:
  $ source <(moly completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ moly completion bash > /etc/bash_completion.d/moly
  # macOS:
  $ moly completion bash > /usr/local/etc/bash_completion.d/moly

Zsh:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  $ moly completion zsh > "${fpath[1]}/_moly"

Fish:
  $ moly completion fish | source
  $ moly completion fish > ~/.config/fish/completions/moly.fish
`,
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: []string{"bash", "zsh", "fish"},
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			RootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			RootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			RootCmd.GenFishCompletion(os.Stdout, true)
		}
	},
}

func init() {
	RootCmd.AddCommand(completionCmd)
}
