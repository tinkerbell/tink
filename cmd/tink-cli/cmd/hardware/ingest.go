// Copyright © 2018 packet.net

package hardware

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/hardware"
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
