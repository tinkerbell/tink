// Copyright © 2018 packet.net

package hardware

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/hardware"
)

var data bool

// ipCmd represents the ip command.
func NewGetByIPCmd() *cobra.Command {
	cmd := &cobra.Command{
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
				if hw.GetId() == "" {
					log.Fatal("IP address not found in the database ", ip)
				}
				printOutput(data, hw, ip)
			}
		},
	}
	flags := cmd.Flags()
	flags.BoolVarP(&data, "details", "d", false, "provide the complete hardware details in json format")
	return cmd
}
