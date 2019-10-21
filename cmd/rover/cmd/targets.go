package cmd

import (
	"fmt"

	"github.com/packethost/rover/cmd/rover/cmd/targets"
	"github.com/spf13/cobra"
)

// templateCmd represents the template sub-command
var targetCmd = &cobra.Command{
	Use:     "targets",
	Short:   "rover targets client",
	Example: "rover targets [command]",
	Args: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("%v requires arguments", c.UseLine())
		}
		return nil
	},
}

func init() {
	targetCmd.AddCommand(targets.SubCommands...)
	rootCmd.AddCommand(targetCmd)
}
