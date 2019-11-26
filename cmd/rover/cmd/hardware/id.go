// Copyright © 2018 packet.net

package hardware

import (
	"context"
	"fmt"
	"log"

	"github.com/packethost/rover/client"
	"github.com/packethost/rover/protos/hardware"
	"github.com/spf13/cobra"
)

// idCmd represents the id command
var idCmd = &cobra.Command{
	Use:     "id",
	Short:   "Get hardware by id",
	Example: "rover hardware id 224ee6ab-ad62-4070-a900-ed816444cec0 cb76ae54-93e9-401c-a5b2-d455bb3800b1",
	Args: func(_ *cobra.Command, args []string) error {
		return verifyUUIDs(args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		for _, id := range args {
			hw, err := client.HardwareClient.ByID(context.Background(), &hardware.GetRequest{ID: id})
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(hw.JSON)
		}
	},
}

func init() {
	SubCommands = append(SubCommands, idCmd)
}
