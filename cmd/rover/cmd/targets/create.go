// Copyright © 2018 packet.net

package targets

import (
	"context"
	"fmt"
	"log"

	"github.com/packethost/cacher/protos/targets"
	"github.com/packethost/rover/client"
	"github.com/spf13/cobra"
)

// pushCmd represents the push command
var createTargets = &cobra.Command{
	Use:     "create",
	Short:   "create targets to cacher",
	Example: `rover target create '{targets": {"machine1": {"mac_addr": "98:67:f5:86:0"},"machine2": {"ip_addr": "192.168.1.5"}}}'`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, j := range args {
			err := isValidData([]byte(j))
			if err != nil {
				log.Fatal(err)
			}
			uuid, err := client.TargetClient.CreateTargets(context.Background(), &targets.PushRequest{Data: j})
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
