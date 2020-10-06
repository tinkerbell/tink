package event

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client/informers"
	"github.com/tinkerbell/tink/protos/events"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	ignoreRTAll = "ignoring \"all\" since specific resource type(s) have been provided"
	ignoreETAll = "ignoring \"all\" since specific event type(s) have been provided"
)

var (
	resourceTypes, eventTypes []string
	resourceID                string
	from                      int

	allResourceTypes = []events.ResourceType{
		events.ResourceType_TEMPLATE,
		events.ResourceType_HARDWARE,
		events.ResourceType_WORKFLOW,
	}

	allEventTypes = []events.EventType{
		events.EventType_CREATED,
		events.EventType_UPDATED,
		events.EventType_DELETED,
	}

	eventKeys = map[string]string{
		"create": events.EventType_CREATED.String(),
		"update": events.EventType_UPDATED.String(),
		"delete": events.EventType_DELETED.String(),
	}
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "watch for event(s) on given resource(s)",
	Example: `tink event watch [flags]
tink event watch --resource-type workflow --resource-type hardware --event-type create`,
	Run: func(cmd *cobra.Command, args []string) {
		stdoutLock := sync.Mutex{}

		req := &events.WatchRequest{
			EventTypes:    []events.EventType{},
			ResourceTypes: []events.ResourceType{},
		}
		processFlags(req)

		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		informer := informers.NewInformer()
		err := informer.Start(ctx, req, func(e *events.Event) error {
			event, _ := json.Marshal(e)
			stdoutLock.Lock()
			var prettyJSON bytes.Buffer
			err := json.Indent(&prettyJSON, event, "", "\t")
			if err != nil {
				return err
			}
			fmt.Println(prettyJSON.String())
			stdoutLock.Unlock()
			return nil
		})
		if err != nil && err != io.EOF {
			cancel()
			log.Fatal(err)
		}
	},
}

func processFlags(req *events.WatchRequest) {
	if resourceID != "" {
		if _, err := uuid.Parse(resourceID); err != nil {
			log.Fatalf("invalid uuid: %s", resourceID)
		}
		req.ResourceId = resourceID
	}

	if len(resourceTypes) == 0 || (len(resourceTypes) == 1 && strings.EqualFold(resourceTypes[0], "all")) {
		req.ResourceTypes = allResourceTypes
	} else {
		for _, rt := range resourceTypes {
			if strings.EqualFold(rt, "all") {
				fmt.Println(ignoreRTAll)
				continue
			}
			req.ResourceTypes = append(req.ResourceTypes, informers.GetResourceType(strings.ToUpper(rt)))
		}
	}

	if len(eventTypes) == 0 || (len(eventTypes) == 1 && strings.EqualFold(eventTypes[0], "all")) {
		req.EventTypes = allEventTypes
	} else {
		for _, et := range eventTypes {
			if strings.EqualFold(et, "all") {
				fmt.Println(ignoreETAll)
				continue
			}
			if key, ok := eventKeys[et]; ok {
				req.EventTypes = append(req.EventTypes, informers.GetEventType(key))
			}

		}
	}
	then := time.Now().Local().Add(time.Duration(int64(-from) * int64(time.Minute)))
	req.WatchEventsFrom = timestamppb.New(then)
}

func addFlags() {
	flags := watchCmd.PersistentFlags()
	flags.StringVarP(&resourceID, "resource-id", "i", "", "resource ID to watch for")
	flags.StringSliceVarP(&resourceTypes, "resource-type", "r", nil, "resource types to watch for [hardware, template, workflow] or \"all\"")
	flags.StringSliceVarP(&eventTypes, "event-type", "e", nil, "events to watch for on a given resource [create, update, delete] or \"all\"")
	flags.IntVarP(&from, "from", "m", 5, "include past events for given time from now (in minutes)")
}

func init() {
	addFlags()
	SubCommands = append(SubCommands, watchCmd)
}
