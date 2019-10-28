package target

import (
	"context"
	"fmt"
	"log"

	"github.com/packethost/rover/client"
	"github.com/packethost/rover/protos/target"
	"github.com/spf13/cobra"
)

// pushCmd represents the push command
var createTargets = &cobra.Command{
	Use:     "create",
	Short:   "create a target",
	Example: `rover target create '{"targets": {"machine1": {"mac_addr": "02:42:db:98:4b:1e"},"machine2": {"ipv4_addr": "192.168.1.5"}}}'`,
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
			fmt.Println("Target Created :", uuid)
		}
	},
}

func init() {
	SubCommands = append(SubCommands, createTargets)
}
