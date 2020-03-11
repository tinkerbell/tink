package workflow

import (
	"context"
	"fmt"
	"log"

	"github.com/packethost/rover/client"
	"github.com/packethost/rover/protos/workflow"
	"github.com/spf13/cobra"
)

var (
	fTemplate = "template"
	fTarget   = "target"
	template  string
	target    string
)

// createCmd represents the create subcommand for worflow command
var createCmd = &cobra.Command{
	Use:     "create",
	Short:   "create a workflow",
	Example: "rover workflow create [flags]",
	PreRunE: func(c *cobra.Command, args []string) error {
		tmp, _ := c.Flags().GetString(fTemplate)
		err := validateID(tmp)
		if err != nil {
			return err
		}
		tar, _ := c.Flags().GetString(fTarget)
		err = validateID(tar)
		return err
	},
	Run: func(c *cobra.Command, args []string) {
		createWorkflow(c, args)
	},
}

func addFlags() {
	flags := createCmd.PersistentFlags()
	flags.StringVarP(&template, "template", "t", "", "workflow template")
	flags.StringVarP(&target, "target", "r", "", "workflow target")

	createCmd.MarkPersistentFlagRequired(fTarget)
	createCmd.MarkPersistentFlagRequired(fTemplate)
}

func createWorkflow(c *cobra.Command, args []string) {
	req := workflow.CreateRequest{Template: template, Target: target}
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
