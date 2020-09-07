package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// completionCmd returns the completion command that, when run, generates a
// bash or zsh completion script for the CLI
func completionCmd(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:
	Bash:
	$ source <(tink-cli completion bash)
	Bash (3.2.x):
	$ eval "$(tink-cli completion bash)"
	# To load completions for each session, execute once:
	Linux:
	  $ tink-cli completion bash > /etc/bash_completion.d/tink-cli
	MacOS:
	  $ tink-cli completion bash > /usr/local/etc/bash_completion.d/tink-cli
	Zsh:
	$ source <(tink-cli completion zsh)
	# To load completions for each session, execute once:
	$ tink-cli completion zsh > "${fpath[1]}/_tink-cli"
	Fish:
	$ tink-cli completion fish | source
	# To load completions for each session, execute once:
	$ tink-cli completion fish > ~/.config/fish/completions/tink-cli.fish
	`,
		Hidden:    true,
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletion(os.Stdout)
			}
			// ValidArgs make this error response dead-code
			return fmt.Errorf("unknown shell: %q", args[0])
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func init() {
	rootCmd.AddCommand(completionCmd(rootCmd.CalledAs()))
}
