package workflow

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/jedib0t/go-pretty/table"
	"github.com/packethost/rover/client"
	"github.com/packethost/rover/protos/workflow"
	"github.com/spf13/cobra"
)

// getCmd represents the get subcommand for workflow command
var stateCmd = &cobra.Command{
	Use:     "state [id]",
	Short:   "get the current workflow context",
	Example: "rover workflow state [id]",
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
			wfProgress := calWorkflowProgress(wf.CurrentActionIndex, wf.TotalNumberOfActions, int(wf.CurrentActionState))
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

func calWorkflowProgress(cur int64, total int64, state int) string {
	if total == 0 || (cur == 0 && state != 2) {
		return "0%"
	}
	var taskCompleted int64
	if state == 2 {
		taskCompleted = cur + 1
	} else {
		taskCompleted = cur
	}
	progress := (taskCompleted * 100) / total
	fmt.Println("Value of progress  is ", progress)
	perc := strconv.Itoa(int(progress)) + "%"

	return fmt.Sprintf("%s", perc)
}

func init() {
	SubCommands = append(SubCommands, stateCmd)
}
