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
  $ source <(moley completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ moley completion bash > /etc/bash_completion.d/moley
  # macOS:
  $ moley completion bash > /usr/local/etc/bash_completion.d/moley

Zsh:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  $ moley completion zsh > "${fpath[1]}/_moley"

Fish:
  $ moley completion fish | source
  $ moley completion fish > ~/.config/fish/completions/moley.fish
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
