package workflow

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/workflow"
)

var (
	hWorkerID      = "Worker ID"
	hTaskName      = "Task Name"
	hActionName    = "Action Name"
	hExecutionTime = "Execution Time"
	hMessage       = "Message"
	hStatus        = "Action Status"
)

func NewShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "events [id]",
		Short:                 "show all events for a workflow",
		DisableFlagsInUseLine: true,
		Example:               "tink workflow events [id]",
		Args: func(c *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("%v takes an arguments", c.UseLine())
			}
			return nil
		},
		Run: func(c *cobra.Command, args []string) {
			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{hWorkerID, hTaskName, hActionName, hExecutionTime, hMessage, hStatus})
			listEvents(t, args)
			t.Render()
		},
	}
	return cmd
}

func listEvents(t table.Writer, args []string) {
	for _, arg := range args {
		req := workflow.GetRequest{Id: arg}
		events, err := client.WorkflowClient.ShowWorkflowEvents(context.Background(), &req)
		if err != nil {
			log.Fatal(err)
		}
		// var wf *workflow.Workflow
		err = nil
		for event, err := events.Recv(); err == nil && event != nil; event, err = events.Recv() {
			t.AppendRows([]table.Row{
				{event.WorkerId, event.TaskName, event.ActionName, event.Seconds, event.Message, event.ActionStatus},
			})
		}
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
	}
}
