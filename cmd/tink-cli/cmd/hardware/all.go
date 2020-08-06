package hardware

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/hardware"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all known hardware",
	Run: func(cmd *cobra.Command, args []string) {

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"ID", "MAC Address", "IP address", "Hostname"})

		list, err := client.HardwareClient.All(context.Background(), &hardware.Empty{})
		if err != nil {
			log.Fatal(err)
		}

		var hw *hardware.Hardware
		for hw, err = list.Recv(); err == nil && hw != nil; hw, err = list.Recv() {
			for _, iface := range hw.GetNetwork().GetInterfaces() {
				t.AppendRow(table.Row{hw.Id, iface.Dhcp.Mac, iface.Dhcp.Ip.Address, iface.Dhcp.Hostname})
			}
		}
		if err != nil && err != io.EOF {
			log.Fatal(err)
		} else {
			t.Render()
		}
	},
}

func init() {
	SubCommands = append(SubCommands, listCmd)
}
