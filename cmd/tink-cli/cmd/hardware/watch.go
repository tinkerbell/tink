// Copyright © 2018 packet.net

package hardware

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client/informers"
	"github.com/tinkerbell/tink/protos/events"
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
				req := &events.WatchRequest{
					ResourceId: id,
					EventTypes: []events.EventType{
						events.EventType_CREATED,
						events.EventType_UPDATED,
						events.EventType_DELETED,
					},
					ResourceTypes: []events.ResourceType{events.ResourceType_HARDWARE},
				}
				informer := informers.NewInformer()
				err := informer.Start(cmd.Context(), req, func(e *events.Event) error {
					stdoutLock.Lock()
					d, err := base64.StdEncoding.DecodeString(strings.Trim(string(e.Data), "\""))
					if err != nil {
						log.Fatal(err)
					}

					hd := &struct {
						Data *hardware.Hardware
					}{}

					err = json.Unmarshal(d, hd)
					if err != nil {
						log.Fatal(err)
					}

					hw, err := json.Marshal(hd.Data)
					if err != nil {
						log.Fatal(err)
					}
					fmt.Println(string(hw))
					stdoutLock.Unlock()
					return nil
				})
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
