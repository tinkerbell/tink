package workflow

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jedib0t/go-pretty/table"
	"github.com/packethost/rover/client"
	"github.com/packethost/rover/protos/workflow"
	"github.com/spf13/cobra"
)

var (
	hTaskName      = "Task Name"
	hActionName    = "Action Name"
	hExecutionTime = "Execution Time"
	hMessage       = "Message"
	hStatus        = "Action Status"
)

// showCmd represents the events subcommand for workflow command
var showCmd = &cobra.Command{
	Use:     "events [id]",
	Short:   "showe all events for a workflow",
	Example: "rover workflow events [id]",
	Args: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("%v takes an arguments", c.UseLine())
		}
		return nil
	},
	Run: func(c *cobra.Command, args []string) {
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{hID, hTaskName, hActionName, hExecutionTime, hMessage, hStatus})
		listEvents(c, t, args)
		t.Render()

	},
}

func listEvents(c *cobra.Command, t table.Writer, args []string) {
	for _, arg := range args {
		req := workflow.GetRequest{Id: arg}
		events, err := client.WorkflowClient.ShowWorkflowEvents(context.Background(), &req)
		if err != nil {
			log.Fatal(err)
		}
		//var wf *workflow.Workflow
		err = nil
		for event, err := events.Recv(); err == nil && event != nil; event, err = events.Recv() {
			t.AppendRows([]table.Row{
				{event.WorkflowId, event.TaskName, event.ActionName, event.Seconds, event.Message, event.ActionStatus},
			})
		}
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
	}
}

func init() {
	showCmd.DisableFlagsInUseLine = true
	SubCommands = append(SubCommands, showCmd)
}
