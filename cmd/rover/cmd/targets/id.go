package targets

import (
	"context"
	"fmt"
	"log"

	"github.com/packethost/rover/client"
	"github.com/packethost/cacher/protos/targets"
	"github.com/spf13/cobra"
)

// idCmd represents the id command
var idCmd = &cobra.Command{
	Use:     "id",
	Short:   "Get targets by id",
	Example: "rover targets id 224ee6ab-ad62-4070-a900-ed816444cec0 cb76ae54-93e9-401c-a5b2-d455bb3800b1",
	Args: func(_ *cobra.Command, args []string) error {
		return verifyUUIDs(args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		for _, id := range args {
			tr, err := client.TargetClient.TargetByID(context.Background(), &targets.GetRequest{ID: id})
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(tr.JSON)
		}
	},
}

func init() {
	SubCommands = append(SubCommands, idCmd)
}
