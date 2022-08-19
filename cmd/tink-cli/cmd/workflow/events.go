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

const shortDescr = `show all events for a workflow`

const longDescr = `Prints a table containing all events for a workflow.
You can specify the kind of output you want to receive.
It can be table or json.
`

const exampleDescr = `# Lists all events for a workflow in table output format.
tink workflow events [id]

# List a single template in json output format.
tink workflow events [id] --format json
`

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
	// out. By default, it is table, but it can also be JSON.
	Format string
}

type Event struct {
	WorkerId string          `json:"worker_id"`
	Tasks    map[string]Task `json:"tasks"`
}

type Task struct {
	Actions map[string]Action `json:"actions"`
}

type Action struct {
	Stages []ActionStage `json:"stages"`
}

type ActionStage struct {
	ExecutionTime int64          `json:"execution_time"`
	Message       string         `json:"message"`
	Status        workflow.State `json:"status"`
}

func NewEventsOptions() Options {
	return Options{}
}

func NewShowCommand(opt Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "events [id]",
		Short:                 shortDescr,
		Long:                  longDescr,
		Example:               exampleDescr,
		DisableFlagsInUseLine: true,
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
					Events interface{} `json:"events"`
				}{Events: getFormattedEvents(allEvents)})

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

func fetchEvents(args []string) []*workflow.WorkflowActionStatus {
	allEvents := make([]*workflow.WorkflowActionStatus, 0)

	for _, arg := range args {
		req := workflow.GetRequest{Id: arg}
		events, err := client.WorkflowClient.ShowWorkflowEvents(context.Background(), &req)
		if err != nil {
			log.Fatal(err)
		}

		err = nil
		for event, err := events.Recv(); err == nil && event != nil; event, err = events.Recv() {
			allEvents = append(allEvents, event)

		}
		if err != nil && !errors.Is(err, io.EOF) {
			log.Fatal(err)
		}
	}

	return allEvents
}

func getFormattedEvents(events []*workflow.WorkflowActionStatus) []Event {
	formattedEvents := make([]Event, 0)
	allEvents := make(map[string]Event)

	for _, ev := range events {
		constructEvent(ev, allEvents)
	}

	for _, ev := range allEvents {
		formattedEvents = append(formattedEvents, ev)
	}

	return formattedEvents
}

func constructEvent(event *workflow.WorkflowActionStatus, allEvents map[string]Event) {

	stage := ActionStage{
		ExecutionTime: event.GetSeconds(),
		Message:       event.GetMessage(),
		Status:        event.GetActionStatus(),
	}

	var ev Event

	if e, ok := allEvents[event.GetWorkerId()]; ok {
		ev = e
	}

	if len(ev.Tasks) == 0 {
		ev = Event{
			WorkerId: event.GetWorkerId(),
			Tasks:    make(map[string]Task, 0),
		}
	}

	var task Task

	if t, ok := ev.Tasks[event.GetTaskName()]; ok {
		task = t
	}

	if len(task.Actions) == 0 {
		task = Task{
			Actions: make(map[string]Action, 0),
		}
	}

	var action Action

	if a, ok := task.Actions[event.GetActionName()]; ok {
		action = a
	}

	if len(action.Stages) == 0 {
		action = Action{
			Stages: make([]ActionStage, 0),
		}
	}

	action.Stages = append(action.Stages, stage)

	task.Actions[event.GetActionName()] = action

	ev.Tasks[event.GetTaskName()] = task

	allEvents[event.GetWorkerId()] = ev
}

func render(allEvents []*workflow.WorkflowActionStatus) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{hWorkerID, hTaskName, hActionName, hExecutionTime, hMessage, hStatus})

	listEvents(t, allEvents)
	t.Render()
}

func listEvents(t table.Writer, allEvents []*workflow.WorkflowActionStatus) {
	for _, event := range allEvents {
		t.AppendRows([]table.Row{{event.WorkerId, event.TaskName, event.ActionName, event.Seconds, event.Message, event.ActionStatus}})
	}
}
