package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/delete"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/get"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/workflow"
)

func NewWorkflowCommand() *cobra.Command {
	cmd := &cobra.Command{
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

	cmd.AddCommand(workflow.NewCreateCommand())
	cmd.AddCommand(workflow.NewDataCommand())
	cmd.AddCommand(delete.NewDeleteCommand(workflow.NewDeleteOptions()))
	cmd.AddCommand(workflow.NewShowCommand())
	cmd.AddCommand(workflow.NewListCommand())
	cmd.AddCommand(workflow.NewStateCommand())

	// If the variable TINK_CLI_VERSION is not set to 0.0.0 use the old get
	// command
	getCmd := workflow.GetCmd
	if v := os.Getenv("TINK_CLI_VERSION"); v != "0.0.0" {
		getCmd = get.NewGetCommand(workflow.NewGetOptions())
	}
	cmd.AddCommand(getCmd)
	return cmd
}
