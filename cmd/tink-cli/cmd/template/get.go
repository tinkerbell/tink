package template

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/google/uuid"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/get"
	"github.com/tinkerbell/tink/protos/template"
)

// getCmd represents the get subcommand for template command
var GetCmd = &cobra.Command{
	Use:                   "get [id]",
	Short:                 "get a template",
	Example:               "tink template get [id]",
	DisableFlagsInUseLine: true,
	Args: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("%v requires an argument", c.UseLine())
		}
		for _, arg := range args {
			if _, err := uuid.Parse(arg); err != nil {
				return fmt.Errorf("invalid uuid: %s", arg)
			}
		}
		return nil
	},
	Run: func(c *cobra.Command, args []string) {
		for _, arg := range args {
			req := template.GetRequest{
				GetBy: &template.GetRequest_Id{
					Id: arg,
				},
			}
			t, err := client.TemplateClient.GetTemplate(context.Background(), &req)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(t.Data))
		}
	},
}

func NewGetTemplateCommand(cl *client.FullClient) *cobra.Command {
	return get.NewGetCommand(get.Options{
		Headers: []string{"ID", "Name", "Created At", "Updated At"},
		RetrieveData: func(ctx context.Context) ([]interface{}, error) {
			list, err := cl.TemplateClient.ListTemplates(context.Background(), &template.ListRequest{
				FilterBy: &template.ListRequest_Name{
					Name: "*",
				},
			})
			if err != nil {
				return nil, err
			}

			data := []interface{}{}
			var tmp *template.WorkflowTemplate
			for tmp, err = list.Recv(); err == nil && tmp.Name != ""; tmp, err = list.Recv() {
				data = append(data, tmp)
			}
			if err != nil && err != io.EOF {
				return nil, err
			}
			return data, nil
		},
		PopulateTable: func(data []interface{}, t table.Writer) error {
			for _, v := range data {
				if tmp, ok := v.(*template.WorkflowTemplate); ok {
					t.AppendRow(table.Row{tmp.Id, tmp.Name, tmp.CreatedAt.AsTime().Unix(), tmp.UpdatedAt.AsTime().Unix()})
				}
			}
			return nil
		},
	})
}
