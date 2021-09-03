package workflow

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/get"
	"github.com/tinkerbell/tink/protos/workflow"
)

var (
	hID       = "Workflow ID"
	hTemplate = "Template ID"
	hDevice   = "Hardware device"
)

// GetCmd represents the get subcommand for workflow command.
var GetCmd = &cobra.Command{
	Use:     "get [id]",
	Short:   "get a workflow",
	Example: "tink workflow get [id]",
	Deprecated: `This command is deprecated and it will change at some
	point. Please unset the environment variable TINK_CLI_VERSION and if
	you are doing some complex automation try using the following command:

	$ tink workflow get -o json [id]
`,

	DisableFlagsInUseLine: true,
	Args: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("%v requires an argument", c.UseLine())
		}
		return validateID(args[0])
	},
	Run: func(c *cobra.Command, args []string) {
		for _, arg := range args {
			req := workflow.GetRequest{Id: arg}
			w, err := client.WorkflowClient.GetWorkflow(context.Background(), &req)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(w.Data)
		}
	},
}

func init() {
}

type getWorkflow struct {
	get.Options
}

func (h *getWorkflow) RetrieveByID(ctx context.Context, cl *client.FullClient, requestedID string) (interface{}, error) {
	return cl.WorkflowClient.GetWorkflow(ctx, &workflow.GetRequest{
		Id: requestedID,
	})
}

func (h *getWorkflow) RetrieveData(ctx context.Context, cl *client.FullClient) ([]interface{}, error) {
	list, err := cl.WorkflowClient.ListWorkflows(ctx, &workflow.Empty{})
	if err != nil {
		return nil, err
	}

	data := []interface{}{}

	var w *workflow.Workflow
	for w, err = list.Recv(); err == nil && w.Id != ""; w, err = list.Recv() {
		data = append(data, w)
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	return data, nil
}

func (h *getWorkflow) PopulateTable(data []interface{}, t table.Writer) error {
	for _, v := range data {
		if w, ok := v.(*workflow.Workflow); ok {
			t.AppendRow(table.Row{
				w.Id, w.Template,
				w.State.String(),
				w.CreatedAt.AsTime().UTC().Format(time.RFC3339),
				w.UpdatedAt.AsTime().UTC().Format(time.RFC3339),
			})
		}
	}
	return nil
}

func NewGetOptions() get.Options {
	h := getWorkflow{}
	opt := get.Options{
		Headers:       []string{"ID", "Template ID", "State", "Created At", "Updated At"},
		RetrieveByID:  h.RetrieveByID,
		RetrieveData:  h.RetrieveData,
		PopulateTable: h.PopulateTable,
	}
	return opt
}
