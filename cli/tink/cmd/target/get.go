package target

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/target"
)

// idCmd represents the id command
var getCmd = &cobra.Command{
	Use:     "get",
	Short:   "get a target",
	Example: "tink target get 224ee6ab-ad62-4070-a900-ed816444cec0 cb76ae54-93e9-401c-a5b2-d455bb3800b1",
	Args: func(_ *cobra.Command, args []string) error {
		return verifyUUIDs(args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		for _, id := range args {
			tr, err := client.TargetClient.TargetByID(context.Background(), &target.GetRequest{ID: id})
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(tr.JSON)
		}
	},
}

func init() {
	SubCommands = append(SubCommands, getCmd)
}
