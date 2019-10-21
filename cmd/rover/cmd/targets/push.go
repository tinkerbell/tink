// Copyright © 2018 packet.net

package targets

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/packethost/cacher/protos/targets"
	"github.com/packethost/rover/client"
	"github.com/spf13/cobra"
)

// pushCmd represents the push command
var createTargets = &cobra.Command{
	Use:     "push",
	Short:   "Push targets to cacher",
	Example: `rover targets push '{targets": {"machine1": "mac_addr=98:67:f5:86","machine2": "ip_addr=192.168.1.5"}}'`,
	Args: func(_ *cobra.Command, args []string) error {
		s := struct {
			targets map[string]string
		}{}
		for _, arg := range args {
			if json.NewDecoder(strings.NewReader(arg)).Decode(&s) != nil {
				return fmt.Errorf("invalid json: %s", arg)
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		for _, j := range args {
			uuid, err := client.TargetClient.CreateTargets(context.Background(), &targets.PushRequest{Data: j})
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("UUID :", uuid)
		}
	},
}

func init() {
	SubCommands = append(SubCommands, createTargets)
}
