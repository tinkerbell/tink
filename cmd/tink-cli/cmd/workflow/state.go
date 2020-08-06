package workflow

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/workflow"
)

// getCmd represents the get subcommand for workflow command
var stateCmd = &cobra.Command{
	Use:     "state [id]",
	Short:   "get the current workflow state",
	Example: "tink workflow state [id]",
	Args: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("%v requires an argument", c.UseLine())
		}
		return validateID(args[0])
	},
	Run: func(c *cobra.Command, args []string) {
		for _, arg := range args {
			req := workflow.GetRequest{Id: arg}
			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"Field Name", "Values"})
			wf, err := client.WorkflowClient.GetWorkflowContext(context.Background(), &req)
			if err != nil {
				log.Fatal(err)
			}
			wfProgress := calWorkflowProgress(wf.CurrentActionIndex, wf.TotalNumberOfActions, wf.CurrentActionState)
			t.AppendRow(table.Row{"Workflow ID", wf.WorkflowId})
			t.AppendRow(table.Row{"Workflow Progress", wfProgress})
			t.AppendRow(table.Row{"Current Task", wf.CurrentTask})
			t.AppendRow(table.Row{"Current Action", wf.CurrentAction})
			t.AppendRow(table.Row{"Current Worker", wf.CurrentWorker})
			t.AppendRow(table.Row{"Current Action State", wf.CurrentActionState})

			t.Render()

		}
	},
}

func calWorkflowProgress(cur int64, total int64, state workflow.ActionState) string {
	if total == 0 || (cur == 0 && state != workflow.ActionState_ACTION_SUCCESS) {
		return "0%"
	}
	var taskCompleted int64
	if state == workflow.ActionState_ACTION_SUCCESS {
		taskCompleted = cur + 1
	} else {
		taskCompleted = cur
	}
	progress := (taskCompleted * 100) / total
	perc := strconv.Itoa(int(progress)) + "%"

	return perc
}

func init() {
	SubCommands = append(SubCommands, stateCmd)
}
