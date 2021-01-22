package workflow

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/workflow"
)

func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "delete [id]",
		Short:                 "delete a workflow",
		Example:               "tink workflow delete [id]",
		DisableFlagsInUseLine: true,
		Args: func(c *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("%v requires an argument", c.UseLine())
			}
			for _, arg := range args {
				if _, err := uuid.Parse(arg); err != nil {
					return fmt.Errorf("invalid uuid: %s", arg)
				}
			}
			return nil
		},
		Run: func(c *cobra.Command, args []string) {
			for _, arg := range args {
				req := workflow.GetRequest{Id: arg}
				if _, err := client.WorkflowClient.DeleteWorkflow(context.Background(), &req); err != nil {
					log.Fatal(err)
				}
			}
		},
	}
	return cmd
}
