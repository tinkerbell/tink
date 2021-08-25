// Copyright © 2018 packet.net

package hardware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/pkg"
	"github.com/tinkerbell/tink/protos/hardware"
)

// idCmd represents the id command
func NewGetByIDCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "id",
		Short:   "get hardware by id",
		Example: "tink hardware id 224ee6ab-ad62-4070-a900-ed816444cec0 cb76ae54-93e9-401c-a5b2-d455bb3800b1",
		Args: func(_ *cobra.Command, args []string) error {
			return verifyUUIDs(args)
		},
		Run: func(cmd *cobra.Command, args []string) {
			for _, id := range args {
				hw, err := client.HardwareClient.ByID(context.Background(), &hardware.GetRequest{Id: id})
				if err != nil {
					log.Fatal(err)
				}
				b, err := json.Marshal(pkg.HardwareWrapper{Hardware: hw})
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(string(b))
			}
		},
	}
}
