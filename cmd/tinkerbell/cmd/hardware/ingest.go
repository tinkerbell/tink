// Copyright © 2018 packet.net

package hardware

import (
	"context"
	"fmt"
	"log"

	"github.com/packethost/tinkerbell/client"
	"github.com/packethost/tinkerbell/protos/hardware"
	"github.com/spf13/cobra"
)

// ingestCmd represents the ingest command
var ingestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Trigger tinkerbell to ingest",
	Long:  "This command only signals tinkerbell to ingest if it has not already done so.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ingest called")
		_, err := client.HardwareClient.Ingest(context.Background(), &hardware.Empty{})
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	SubCommands = append(SubCommands, ingestCmd)

}
