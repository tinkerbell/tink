// Copyright © 2018 packet.net

package hardware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/hardware"
)

// ipCmd represents the ip command
var ipCmd = &cobra.Command{
	Use:     "ip",
	Short:   "get hardware by any associated ip",
	Example: "tink hardware ip 10.0.0.2 10.0.0.3",
	Args: func(_ *cobra.Command, args []string) error {
		for _, arg := range args {
			if net.ParseIP(arg) == nil {
				return fmt.Errorf("invalid ip: %s", arg)
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		for _, ip := range args {
			hw, err := client.HardwareClient.ByIP(context.Background(), &hardware.GetRequest{Ip: ip})
			if err != nil {
				log.Fatal(err)
			}
			b, err := json.Marshal(hw)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(b))
		}
	},
}

func init() {
	SubCommands = append(SubCommands, ipCmd)
}
