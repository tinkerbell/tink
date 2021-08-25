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

func NewGetByMACCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mac",
		Short:   "get hardware by any associated mac",
		Example: "tink hardware mac 00:00:00:00:00:01 00:00:00:00:00:02",
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
				hw, err := client.HardwareClient.ByMAC(context.Background(), &hardware.GetRequest{Mac: mac})
				if err != nil {
					log.Fatal(err)
				}
				if hw.GetId() == "" {
					log.Fatal("MAC address not found in the database ", mac)
				}
				printOutput(data, hw, mac)
			}
		},
	}
	flags := cmd.Flags()
	flags.BoolVarP(&data, "details", "d", false, "provide the complete hardware details in json format")
	return cmd
}
