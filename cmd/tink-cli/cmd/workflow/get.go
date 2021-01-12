package workflow

import (
	"context"
	"fmt"
	"io"
	"log"

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

// getCmd represents the get subcommand for workflow command
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

// NewGetCommand create the generic get command with everything required by the
// workflow resource to work
func NewGetCommand(cl *client.FullClient) *cobra.Command {
	cmd := get.NewGetCommand(get.CmdOpt{
		Headers: []string{"ID", "Template ID", "State", "Created At", "Updated At"},
		RetrieveData: func(ctx context.Context) ([]interface{}, error) {
			list, err := cl.WorkflowClient.ListWorkflows(ctx, &workflow.Empty{})
			if err != nil {
				return nil, err
			}

			data := []interface{}{}

			var w *workflow.Workflow
			for w, err = list.Recv(); err == nil && w.Id != ""; w, err = list.Recv() {
				data = append(data, w)
			}
			if err != nil && err != io.EOF {
				return nil, err
			}
			return data, nil
		},
		PopulateTable: func(data []interface{}, t table.Writer) error {
			for _, v := range data {
				if w, ok := v.(*workflow.Workflow); ok {
					t.AppendRow(table.Row{w.Id, w.Template,
						w.State.String(),
						w.CreatedAt.AsTime().Unix,
						w.UpdatedAt.AsTime().Unix})
				}
			}
			return nil
		},
	})
	return cmd
}
