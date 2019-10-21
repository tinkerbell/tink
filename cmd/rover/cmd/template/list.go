package template

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/table"
	"github.com/packethost/rover/client"
	"github.com/packethost/rover/protos/template"
	"github.com/spf13/cobra"
)

// table headers
var (
	id         = "Template ID"
	name       = "Template Name"
	insertedAt = "Inserted At"
	deletedAt  = "Deleted At"
)

// listCmd represents the list subcommand for template command
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "list all saved templates",
	Example: "rover template list",
	Args: func(c *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("%v takes no arguments", c.UseLine())
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{id, name, insertedAt, deletedAt})
		listTemplates(cmd, t)
		t.Render()
	},
}

func listTemplates(cmd *cobra.Command, t table.Writer) {
	list, err := client.TemplateClient.List(context.Background(), &template.Empty{})
	if err != nil {
		log.Fatal(err)
	}

	var tmp *template.WorkflowTemplate
	err = nil
	for tmp, err = list.Recv(); err == nil && tmp.Name != ""; tmp, err = list.Recv() {
		in := *tmp.InsertedAt
		del := *tmp.DeletedAt
		if del.Seconds > 0 {
			t.AppendRows([]table.Row{
				{tmp.Id, tmp.Name, time.Unix(in.Seconds, 0), time.Unix(del.Seconds, 0)},
			})
		} else {
			t.AppendRows([]table.Row{
				{tmp.Id, tmp.Name, time.Unix(in.Seconds, 0), ""},
			})
		}
	}

	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
}

func init() {
	listCmd.DisableFlagsInUseLine = true
	SubCommands = append(SubCommands, listCmd)
}
