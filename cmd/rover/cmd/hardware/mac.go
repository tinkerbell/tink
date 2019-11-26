// Copyright © 2018 packet.net

package hardware

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/packethost/rover/client"
	"github.com/packethost/rover/protos/hardware"
	"github.com/spf13/cobra"
)

// macCmd represents the mac command
var macCmd = &cobra.Command{
	Use:     "mac",
	Short:   "Get hardware by any associated mac",
	Example: "rover hardware mac 00:00:00:00:00:01 00:00:00:00:00:02",
	Args: func(_ *cobra.Command, args []string) error {
		for _, arg := range args {
			if _, err := net.ParseMAC(arg); err != nil {
				return fmt.Errorf("invalid mac: %s", arg)
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		for _, mac := range args {
			hw, err := client.HardwareClient.ByMAC(context.Background(), &hardware.GetRequest{MAC: mac})
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(hw.JSON)
		}
	},
}

func init() {
	SubCommands = append(SubCommands, macCmd)
}
