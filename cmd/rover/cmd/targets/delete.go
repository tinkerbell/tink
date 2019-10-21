package targets

import (
	"context"
	"log"

	"github.com/packethost/rover/client"
	"github.com/packethost/rover/protos/targets"
	"github.com/spf13/cobra"
)

// idCmd represents the id command
var deleteCmd = &cobra.Command{
	Use:     "deleteByID",
	Short:   "Delete targets by id",
	Example: "rover targets deleteByID 224ee6ab-ad62-4070-a900-ed816444cec0 cb76ae54-93e9-401c-a5b2-d455bb3800b1",
	Args: func(_ *cobra.Command, args []string) error {
		return verifyUUIDs(args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		for _, id := range args {
			_, err := client.TargetClient.DeleteTargetByID(context.Background(), &targets.GetRequest{ID: id})
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	SubCommands = append(SubCommands, deleteCmd)
}
