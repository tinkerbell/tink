package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/cli/tink/cmd/target"
)

// templateCmd represents the template sub-command
var targetCmd = &cobra.Command{
	Use:     "target",
	Short:   "tink target client",
	Example: "tink target [command]",
	Args: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("%v requires arguments", c.UseLine())
		}
		return nil
	},
}

func init() {
	targetCmd.AddCommand(target.SubCommands...)
	rootCmd.AddCommand(targetCmd)
}
