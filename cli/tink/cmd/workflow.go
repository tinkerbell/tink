package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/cli/tink/cmd/workflow"
)

// workflowCmd represents the workflow sub-command
var workflowCmd = &cobra.Command{
	Use:     "workflow",
	Short:   "tink workflow client",
	Example: "tink workflow [command]",
	Args: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("%v requires arguments", c.UseLine())
		}
		return nil
	},
}

func init() {
	workflowCmd.AddCommand(workflow.SubCommands...)
	rootCmd.AddCommand(workflowCmd)
}
