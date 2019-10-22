package targets

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jedib0t/go-pretty/table"
	"github.com/packethost/cacher/protos/targets"
	"github.com/packethost/rover/client"
	"github.com/spf13/cobra"
)

// listCmd represents the list subcommand for target command
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "list all targets which are no deleted",
	Example: "rover target list",
	Args: func(c *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("%v takes no arguments", c.UseLine())
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Target Id", "Target Data"})
		listTargets(cmd, t)
		t.Render()
	},
}

func listTargets(cmd *cobra.Command, t table.Writer) {
	list, err := client.TargetClient.ListTargets(context.Background(), &targets.Empty{})
	if err != nil {
		log.Fatal(err)
	}

	var tmp *targets.TargetList
	err = nil
	for tmp, err = list.Recv(); err == nil && tmp.Data != ""; tmp, err = list.Recv() {
		t.AppendRows([]table.Row{
			{tmp.ID, tmp.Data},
		})
	}

	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
}

func init() {
	listCmd.DisableFlagsInUseLine = true
	SubCommands = append(SubCommands, listCmd)
}
