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

// createCmd represents the create subcommand for worflow command
var createCmd = &cobra.Command{
	Use:     "create",
	Short:   "create a workflow",
	Example: "tink workflow create [flags]",
	PreRunE: func(c *cobra.Command, args []string) error {
		tmp, _ := c.Flags().GetString(fTemplate)
		err := validateID(tmp)
		return err
	},
	Run: func(c *cobra.Command, args []string) {
		createWorkflow(c, args)
	},
}

func addFlags() {
	flags := createCmd.PersistentFlags()
	flags.StringVarP(&template, "template", "t", "", "workflow template")
	flags.StringVarP(&hardware, "hardware", "r", "", "workflow target hardwares")

	createCmd.MarkPersistentFlagRequired(fHardware)
	createCmd.MarkPersistentFlagRequired(fTemplate)
}

func createWorkflow(c *cobra.Command, args []string) {
	req := workflow.CreateRequest{Template: template, Target: hardware}
	res, err := client.WorkflowClient.CreateWorkflow(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created Workflow: ", res.Id)
}

func init() {
	addFlags()
	SubCommands = append(SubCommands, createCmd)
}
