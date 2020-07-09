package hardware

import (
	"encoding/json"
	"fmt"
	"github.com/tinkerbell/tink/util"
	"log"
	"os"

	"github.com/tinkerbell/tink/protos/hardware"

	"github.com/jedib0t/go-pretty/table"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

// SubCommands holds the sub commands for template command
// Example: tinkerbell template [subcommand]
var SubCommands []*cobra.Command

func verifyUUIDs(args []string) error {
	if len(args) < 1 {
		return errors.New("requires at least one id")
	}
	for _, arg := range args {
		if _, err := uuid.FromString(arg); err != nil {
			return fmt.Errorf("invalid uuid: %s", arg)
		}
	}
	return nil
}

func printOutput(data bool, hw *hardware.Hardware, input string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Field Name", "Value"})
	if !data {
		for _, iface := range hw.GetNetwork().GetInterfaces() {
			if iface.Dhcp.Ip.Address == input || iface.Dhcp.Mac == input {
				t.AppendRow(table.Row{"ID", hw.Id})
				t.AppendRow(table.Row{"MAC Address", iface.Dhcp.Mac})
				t.AppendRow(table.Row{"IP Address", iface.Dhcp.Ip.Address})
				t.AppendRow(table.Row{"Hostname", iface.Dhcp.Hostname})
			}
		}
		t.Render()
	} else {
		hwData, err := json.Marshal(util.HardwareWrapper{Hardware: hw})
		if err != nil {
			log.Fatal("Failed to marshal hardware data: ", err)
		}
		log.Println(string(hwData))
	}
}
