// Copyright © 2018 packet.net

package hardware

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/packethost/tinkerbell/client"
	"github.com/packethost/tinkerbell/protos/hardware"
	"github.com/spf13/cobra"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:     "watch",
	Short:   "Register to watch an id for any changes",
	Example: "rover hardware watch 224ee6ab-ad62-4070-a900-ed816444cec0 cb76ae54-93e9-401c-a5b2-d455bb3800b1",
	Args: func(_ *cobra.Command, args []string) error {
		return verifyUUIDs(args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		stdoutLock := sync.Mutex{}
		for _, id := range args {
			go func(id string) {
				stream, err := client.HardwareClient.Watch(context.Background(), &hardware.GetRequest{ID: id})
				if err != nil {
					log.Fatal(err)
				}

				var hw *hardware.Hardware
				err = nil
				for hw, err = stream.Recv(); err == nil && hw != nil; hw, err = stream.Recv() {
					stdoutLock.Lock()
					fmt.Println(hw.JSON)
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
