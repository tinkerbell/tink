package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
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

	cl, err := client.NewFullClientFromGlobal()
	if err != nil {
		panic(err)
	}

	cmd.AddCommand(workflow.NewCreateCommand(cl))
	cmd.AddCommand(workflow.NewDataCommand(cl))
	cmd.AddCommand(workflow.NewDeleteCommand(cl))
	cmd.AddCommand(workflow.NewShowCommand(cl))
	cmd.AddCommand(workflow.NewListCommand(cl))
	cmd.AddCommand(workflow.NewStateCommand(cl))

	// If the variable TINK_CLI_VERSION is not set to 0.0.0 use the old get
	// command
	getCmd := workflow.GetCmd
	if v := os.Getenv("TINK_CLI_VERSION"); v != "0.0.0" {
		getCmd = workflow.NewGetCommand(cl)
	}
	cmd.AddCommand(getCmd)
	return cmd
}
