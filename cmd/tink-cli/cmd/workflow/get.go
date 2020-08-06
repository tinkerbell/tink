package workflow

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/workflow"
)

var (
	hID       = "Workflow ID"
	hTemplate = "Template ID"
	hDevice   = "Hardware device"
)

// getCmd represents the get subcommand for workflow command
var getCmd = &cobra.Command{
	Use:     "get [id]",
	Short:   "get a workflow",
	Example: "tink workflow get [id]",
	Args: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("%v requires an argument", c.UseLine())
		}
		return validateID(args[0])
	},
	Run: func(c *cobra.Command, args []string) {
		for _, arg := range args {
			req := workflow.GetRequest{Id: arg}
			w, err := client.WorkflowClient.GetWorkflow(context.Background(), &req)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(w.Data)
		}
	},
}

func init() {
	getCmd.DisableFlagsInUseLine = true
	SubCommands = append(SubCommands, getCmd)
}
