// Copyright © 2018 packet.net

package hardware

import (
	"context"
	"log"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/hardware"
)

// deleteCmd represents the id command
var deleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "delete hardware by id",
	Example: "tink hardware delete 224ee6ab-ad62-4070-a900-ed816444cec0 cb76ae54-93e9-401c-a5b2-d455bb3800b1",
	Args: func(_ *cobra.Command, args []string) error {
		return verifyUUIDs(args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		for _, id := range args {
			_, err := client.HardwareClient.Delete(context.Background(), &hardware.DeleteRequest{ID: id})
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Hardware data with id", id, "deleted successfully")
		}
	},
}

func init() {
	SubCommands = append(SubCommands, deleteCmd)
}
