package workflow

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/workflow"
)

var (
	hCreatedAt = "Created At"
	hUpdatedAt = "Updated At"
)

// listCmd represents the list subcommand for workflow command
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "list all workflows",
	Example: "tink workflow list",
	Args: func(c *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("%v takes no arguments", c.UseLine())
		}
		return nil
	},
	Run: func(c *cobra.Command, args []string) {
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{hID, hTemplate, hTarget, hCreatedAt, hUpdatedAt})
		listWorkflows(c, t)
		t.Render()

	},
}

func listWorkflows(c *cobra.Command, t table.Writer) {
	list, err := client.WorkflowClient.ListWorkflows(context.Background(), &workflow.Empty{})
	if err != nil {
		log.Fatal(err)
	}

	var wf *workflow.Workflow
	err = nil
	for wf, err = list.Recv(); err == nil && wf.Id != ""; wf, err = list.Recv() {
		cr := *wf.CreatedAt
		up := *wf.UpdatedAt
		t.AppendRows([]table.Row{
			{wf.Id, wf.Template, wf.Target, time.Unix(cr.Seconds, 0), time.Unix(up.Seconds, 0)},
		})
	}

	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
}

func init() {
	listCmd.DisableFlagsInUseLine = true
	SubCommands = append(SubCommands, listCmd)
}
