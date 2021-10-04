package hardware

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	hwpb "github.com/tinkerbell/tink/protos/hardware"
)

var (
	quiet bool
	t     table.Writer
)

func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all known hardware",
		Run: func(cmd *cobra.Command, args []string) {
			if quiet {
				listHardware()
				return
			}
			t = table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"ID", "MAC Address", "IP address", "Hostname"})
			listHardware()
			t.Render()
		},
	}
	flags := cmd.Flags()
	flags.BoolVarP(&quiet, "quiet", "q", false, "only display hardware IDs")
	return cmd
}

func listHardware() {
	list, err := client.HardwareClient.All(context.Background(), &hwpb.Empty{})
	if err != nil {
		log.Fatal(err)
	}

	var hw *hwpb.Hardware
	for hw, err = list.Recv(); err == nil && hw != nil; hw, err = list.Recv() {
		for _, iface := range hw.GetNetwork().GetInterfaces() {
			if quiet {
				fmt.Println(hw.Id)
			} else {
				t.AppendRow(table.Row{hw.Id, iface.Dhcp.Mac, iface.Dhcp.Ip.Address, iface.Dhcp.Hostname})
			}
		}
	}
	if err != nil && !errors.Is(err, io.EOF) {
		log.Fatal(err)
	}
}
