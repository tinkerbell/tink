package cmd

import (
	"fmt"

	"github.com/tinkerbell/tink/cmd/tinkerbell/cmd/workflow"
	"github.com/spf13/cobra"
)

// workflowCmd represents the workflow sub-command
var workflowCmd = &cobra.Command{
	Use:     "workflow",
	Short:   "tinkerbell workflow client",
	Example: "tinkerbell workflow [command]",
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
