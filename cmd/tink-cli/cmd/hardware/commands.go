package hardware

import (
	"encoding/json"
	"fmt"
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
		hwData := formatHardwareForPrint(hw)
		log.Println(string(hwData))
	}
}

// formatHardwareForPush formats a hardware string with its metadata field converted from map to string
func formatHardwareForPush(data string) []byte {
	hwJSON := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &hwJSON)
	if err != nil {
		log.Println(err)
	}

	if _, ok := hwJSON["metadata"]; ok {
		metadata, err := json.Marshal(hwJSON["metadata"])
		if err != nil {
			log.Println(err)
		}
		hwJSON["metadata"] = string(metadata)
	}
	b, err := json.Marshal(hwJSON)
	if err != nil {
		log.Println(err)
	}
	return b
}

// formatHardwareForPrint returns hardware as a string
// converts its metadata field from a json formatted string to a map
func formatHardwareForPrint(hw *hardware.Hardware) string {
	hwJSON := make(map[string]interface{})
	hwByte, err := json.Marshal(hw)
	if err != nil {
		log.Println(err)
	}
	err = json.Unmarshal(hwByte, &hwJSON) // create hardware
	if err != nil {
		log.Println(err)
	}
	if hw.Metadata != "" {
		metadata := make(map[string]interface{})
		err = json.Unmarshal([]byte(hw.Metadata), &metadata) // metadata is now a map
		if err != nil {
			log.Println(err)
		}
		hwJSON["metadata"] = metadata
	}
	b, err := json.Marshal(hwJSON)
	if err != nil {
		log.Fatal("Failed to marshal hardware data", err)
	}
	return string(b)
}
