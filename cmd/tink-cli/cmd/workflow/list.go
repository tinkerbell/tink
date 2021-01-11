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
	quiet bool
	t     table.Writer

	hCreatedAt = "Created At"
	hUpdatedAt = "Updated At"
)

// listCmd represents the list subcommand for workflow command
func NewListCommand(cl *client.MetaClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "list all workflows",
		Example: "tink workflow list",
		Deprecated: `This command is deprecated and it will change at some
	point. Please try what follows:

	# If you want to retrieve a single workflow you know by ID
	tink workflow get [id]
	# You can print it in JSON and CSV as well
	tink workflow get -o json [id]

	# Get a list of available workflows
	tink workflow get [id]
`,
		Args: func(c *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("%v takes no arguments", c.UseLine())
			}
			return nil
		},
		Run: func(c *cobra.Command, args []string) {
			if quiet {
				listWorkflows(cl)
				return
			}
			t = table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{hID, hTemplate, hDevice, hCreatedAt, hUpdatedAt})
			listWorkflows(cl)
			t.Render()
		},
	}
	flags := cmd.Flags()
	flags.BoolVarP(&quiet, "quiet", "q", false, "only display workflow IDs")
	return cmd
}

func listWorkflows(cl *client.MetaClient) {
	list, err := cl.WorkflowClient.ListWorkflows(context.Background(), &workflow.Empty{})
	if err != nil {
		log.Fatal(err)
	}

	var wf *workflow.Workflow
	for wf, err = list.Recv(); err == nil && wf.Id != ""; wf, err = list.Recv() {
		printOutput(wf)
	}

	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
}

func printOutput(wf *workflow.Workflow) {
	if quiet {
		fmt.Println(wf.Id)
	} else {
		cr := wf.CreatedAt
		up := wf.UpdatedAt
		t.AppendRows([]table.Row{
			{wf.Id, wf.Template, wf.Hardware, time.Unix(cr.Seconds, 0), time.Unix(up.Seconds, 0)},
		})
	}
}
