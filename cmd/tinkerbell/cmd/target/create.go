package target

import (
	"context"
	"fmt"
	"log"

	"github.com/packethost/tinkerbell/client"
	"github.com/packethost/tinkerbell/protos/target"
	"github.com/spf13/cobra"
)

// pushCmd represents the push command
var createTargets = &cobra.Command{
	Use:     "create",
	Short:   "create a target",
	Example: `tinkerbell target create '{"targets": {"machine1": {"mac_addr": "02:42:db:98:4b:1e"},"machine2": {"ipv4_addr": "192.168.1.5"}}}'`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, j := range args {
			err := isValidData([]byte(j))
			if err != nil {
				log.Fatal(err)
			}
			uuid, err := client.TargetClient.CreateTargets(context.Background(), &target.PushRequest{Data: j})
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Created Target:", uuid.Uuid)
		}
	},
}

func init() {
	SubCommands = append(SubCommands, createTargets)
}
