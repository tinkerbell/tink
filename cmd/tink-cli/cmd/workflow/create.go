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
	fTemplate = "template"
	fHardware = "hardware"
	template  string
	hardware  string
)

func NewCreateCommand(cl *client.MetaClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "create a workflow",
		Example: "tink workflow create [flags]",
		PreRunE: func(c *cobra.Command, args []string) error {
			tmp, _ := c.Flags().GetString(fTemplate)
			err := validateID(tmp)
			return err
		},
		Run: func(c *cobra.Command, args []string) {
			createWorkflow(cl, args)
		},
	}
	flags := cmd.PersistentFlags()
	flags.StringVarP(&template, "template", "t", "", "workflow template")
	flags.StringVarP(&hardware, "hardware", "r", "", "workflow targeted hardwares")

	_ = cmd.MarkPersistentFlagRequired(fHardware)
	_ = cmd.MarkPersistentFlagRequired(fTemplate)
	return cmd
}

func createWorkflow(cl *client.MetaClient, args []string) {
	req := workflow.CreateRequest{Template: template, Hardware: hardware}
	res, err := cl.WorkflowClient.CreateWorkflow(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created Workflow: ", res.Id)
}
