package template

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

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
		return nil
	},
	Run: func(c *cobra.Command, args []string) {
		for _, arg := range args {
			var req = template.GetRequest{}
			// Parse arg[0] to see if it is a UUID
			if _, err := uuid.Parse(arg); err == nil {
				// UUID
				req = template.GetRequest{
					GetBy: &template.GetRequest_Id{
						Id: arg,
					},
				}
			} else {
				// String (Name)
				req = template.GetRequest{
					GetBy: &template.GetRequest_Name{
						Name: arg,
					},
				}
			}
			t, err := client.TemplateClient.GetTemplate(context.Background(), &req)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(t.Data))
		}
	},
}

type getTemplate struct {
	get.Options
}

func (h *getTemplate) RetrieveByID(ctx context.Context, cl *client.FullClient, requestedID string) (interface{}, error) {
	return cl.TemplateClient.GetTemplate(context.Background(), &template.GetRequest{
		GetBy: &template.GetRequest_Id{
			Id: requestedID,
		},
	})
}

func (h *getTemplate) RetrieveByName(ctx context.Context, cl *client.FullClient, requestName string) (interface{}, error) {
	return cl.TemplateClient.GetTemplate(context.Background(), &template.GetRequest{
		GetBy: &template.GetRequest_Name{
			Name: requestName,
		},
	})
}

func (h *getTemplate) RetrieveData(ctx context.Context, cl *client.FullClient) ([]interface{}, error) {
	list, err := cl.TemplateClient.ListTemplates(ctx, &template.ListRequest{
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
}

func (h *getTemplate) PopulateTable(data []interface{}, t table.Writer) error {
	for _, v := range data {
		if tmp, ok := v.(*template.WorkflowTemplate); ok {
			t.AppendRow(table.Row{tmp.Id, tmp.Name,
				tmp.CreatedAt.AsTime().Format(time.RFC3339),
				tmp.UpdatedAt.AsTime().Format(time.RFC3339)})
		}
	}
	return nil
}

func NewGetOptions() get.Options {
	h := getTemplate{}
	return get.Options{
		Headers:        []string{"ID", "Name", "Created At", "Updated At"},
		RetrieveByID:   h.RetrieveByID,
		RetrieveByName: h.RetrieveByName,
		RetrieveData:   h.RetrieveData,
		PopulateTable:  h.PopulateTable,
	}
}
