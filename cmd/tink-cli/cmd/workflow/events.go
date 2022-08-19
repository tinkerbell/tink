package workflow

import (
	"context"
	"encoding/json"
	"errors"
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

type Options struct {
	// Format specifies the format you want the list of resources printed
	// out. By default, it is table, but it can be JSON ar CSV.
	Format string
}

func NewEventsOptions() Options {
	return Options{}
}

func NewShowCommand(opt Options) *cobra.Command {
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
		RunE: func(c *cobra.Command, args []string) error {
			allEvents := fetchEvents(args)

			switch opt.Format {
			case "json":
				b, err := json.Marshal(struct {
					Data interface{} `json:"data"`
				}{Data: allEvents})
				if err != nil {
					return err
				}
				fmt.Fprint(c.OutOrStdout(), string(b))
			default:
				render(allEvents)
			}

			return nil
		},
	}
	cmd.PersistentFlags().StringVarP(&opt.Format, "format", "", "table", "The format you expect the list to be printed out. Currently supported format are table, JSON and CSV")
	return cmd
}

func fetchEvents(args []string) [][]interface{} {
	allEvents := make([][]interface{}, 0)
	for _, arg := range args {
		req := workflow.GetRequest{Id: arg}
		events, err := client.WorkflowClient.ShowWorkflowEvents(context.Background(), &req)
		if err != nil {
			log.Fatal(err)
		}
		// var wf *workflow.Workflow
		err = nil
		for event, err := events.Recv(); err == nil && event != nil; event, err = events.Recv() {
			allEvents = append(allEvents, []interface{}{event.WorkerId, event.TaskName, event.ActionName, event.Seconds, event.Message, event.ActionStatus})
		}
		if err != nil && !errors.Is(err, io.EOF) {
			log.Fatal(err)
		}
	}

	return allEvents
}

func render(allEvents [][]interface{}) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{hWorkerID, hTaskName, hActionName, hExecutionTime, hMessage, hStatus})

	listEvents(t, allEvents)
	t.Render()
}

func listEvents(t table.Writer, allEvents [][]interface{}) {
	for _, event := range allEvents {
		t.AppendRows([]table.Row{event})
	}
}
