// Copyright © 2018 packet.net

package hardware

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/hardware"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:     "watch",
	Short:   "register to watch an id for any changes",
	Example: "tink hardware watch 224ee6ab-ad62-4070-a900-ed816444cec0 cb76ae54-93e9-401c-a5b2-d455bb3800b1",
	Args: func(_ *cobra.Command, args []string) error {
		return verifyUUIDs(args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		stdoutLock := sync.Mutex{}
		for _, id := range args {
			go func(id string) {
				stream, err := client.HardwareClient.Watch(context.Background(), &hardware.GetRequest{Id: id})
				if err != nil {
					log.Fatal(err)
				}

				var hw *hardware.Hardware
				err = nil
				for hw, err = stream.Recv(); err == nil && hw != nil; hw, err = stream.Recv() {
					stdoutLock.Lock()
					b, err := json.Marshal(hw)
					if err != nil {
						log.Fatal(err)
					}
					fmt.Println(string(b))
					stdoutLock.Unlock()
				}
				if err != nil && err != io.EOF {
					log.Fatal(err)
				}
			}(id)
		}
		select {}
	},
}

func init() {
	watchCmd.Flags().String("id", "", "id of the hardware")
	SubCommands = append(SubCommands, watchCmd)
}
