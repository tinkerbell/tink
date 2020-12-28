package template

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/template"
)

// table headers
var (
	id        = "Template ID"
	name      = "Template Name"
	createdAt = "Created At"
	updatedAt = "Updated At"
)

var (
	quiet bool
	t     table.Writer
)

// listCmd represents the list subcommand for template command
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "list all saved templates",
	Example: "tink template list",
	Args: func(c *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("%v takes no arguments", c.UseLine())
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if quiet {
			listTemplates()
			return
		}
		t = table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{id, name, createdAt, updatedAt})
		listTemplates()
		t.Render()
	},
}

func listTemplates() {
	list, err := client.TemplateClient.ListTemplates(context.Background(), &template.ListRequest{
		FilterBy: &template.ListRequest_Name{
			Name: "*",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	var tmp *template.WorkflowTemplate
	for tmp, err = list.Recv(); err == nil && tmp.Name != ""; tmp, err = list.Recv() {
		printOutput(tmp)
	}

	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
}

func printOutput(tmp *template.WorkflowTemplate) {
	if quiet {
		fmt.Println(tmp.Id)
	} else {
		cr := tmp.CreatedAt
		up := tmp.UpdatedAt
		t.AppendRows([]table.Row{
			{tmp.Id, tmp.Name, time.Unix(cr.Seconds, 0), time.Unix(up.Seconds, 0)},
		})
	}
}

func addListFlags() {
	flags := listCmd.Flags()
	flags.BoolVarP(&quiet, "quiet", "q", false, "only display template IDs")
}

func init() {
	addListFlags()
	SubCommands = append(SubCommands, listCmd)
}
