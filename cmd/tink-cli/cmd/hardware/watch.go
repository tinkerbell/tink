// Copyright © 2018 packet.net

package hardware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client/informers"
	"github.com/tinkerbell/tink/protos/events"
	"github.com/tinkerbell/tink/protos/hardware"
	"google.golang.org/protobuf/types/known/timestamppb"
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
		then := time.Now().Local().Add(time.Duration(int64(-5) * int64(time.Minute)))
		stdoutLock := sync.Mutex{}
		for _, id := range args {
			go func(id string) {
				req := &events.WatchRequest{
					ResourceId: id,
					EventTypes: []events.EventType{
						events.EventType_EVENT_TYPE_CREATED,
						events.EventType_EVENT_TYPE_UPDATED,
						events.EventType_EVENT_TYPE_DELETED,
					},
					ResourceTypes:   []events.ResourceType{events.ResourceType_RESOURCE_TYPE_HARDWARE},
					WatchEventsFrom: timestamppb.New(then),
				}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				informer := informers.New()
				err := informer.Start(ctx, req, func(e *events.Event) error {
					var encodedData string
					var jsonData map[string]interface{}

					if er := json.Unmarshal(e.Data, &jsonData); er == nil {
						encodedData = base64.StdEncoding.EncodeToString(e.Data)
					} else {
						encodedData = strings.Trim(string(e.Data), "\"")
					}

					d, err := base64.StdEncoding.DecodeString(encodedData)
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
					stdoutLock.Lock()
					fmt.Printf("%s\n\n", string(hw))
					stdoutLock.Unlock()
					return nil
				})
				if err != nil && err != io.EOF {
					cancel()
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
