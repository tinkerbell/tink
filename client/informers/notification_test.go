package informers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tinkerbell/tink/protos/events"
)

func TestFilter(t *testing.T) {
	testCases := map[string]struct {
		notification       Notification
		req                *events.WatchRequest
		filterNotification bool
	}{
		"different ResourceID": {
			notification: Notification{
				ResourceID: "4b2ef275-735e-472f-8a3e-3d394712cf79",
			},
			req: watchRequest(
				withResourceID("87794e7a-653c-46c2-8a48-195d0ee6eafa"),
			),
			filterNotification: true,
		},
		"different EventType": {
			notification: Notification{
				EventType: "CREATED",
			},
			req: watchRequest(
				withEventTypes([]events.EventType{events.EventType_EVENT_TYPE_DELETED}),
			),
			filterNotification: true,
		},
		"different ResourceType": {
			notification: Notification{
				ResourceType: "WORKFLOW",
			},
			req: watchRequest(
				withResourceTypes([]events.ResourceType{events.ResourceType_RESOURCE_TYPE_TEMPLATE}),
			),
			filterNotification: true,
		},
		"same ResourceID different EventType": {
			notification: Notification{
				ResourceID: "4b2ef275-735e-472f-8a3e-3d394712cf79",
				EventType:  "DELETED",
			},
			req: watchRequest(
				withResourceID("4b2ef275-735e-472f-8a3e-3d394712cf79"),
				withEventTypes([]events.EventType{events.EventType_EVENT_TYPE_CREATED}),
			),
			filterNotification: true,
		},
		"same ResourceType different EventType": {
			notification: Notification{
				EventType:    "DELETED",
				ResourceType: "TEMPLATE",
			},
			req: watchRequest(
				withEventTypes([]events.EventType{events.EventType_EVENT_TYPE_CREATED}),
				withResourceTypes([]events.ResourceType{events.ResourceType_RESOURCE_TYPE_HARDWARE}),
			),
			filterNotification: true,
		},
		"same EventType different ResourceType": {
			notification: Notification{
				EventType:    "UPDATED",
				ResourceType: "TEMPLATE",
			},
			req: watchRequest(
				withEventTypes([]events.EventType{events.EventType_EVENT_TYPE_UPDATED}),
				withResourceTypes([]events.ResourceType{events.ResourceType_RESOURCE_TYPE_HARDWARE}),
			),
			filterNotification: true,
		},
		"same ResourceID and EventType": {
			notification: Notification{
				ResourceID:   "4b2ef275-735e-472f-8a3e-3d394712cf79",
				EventType:    "CREATED",
				ResourceType: "TEMPLATE",
			},
			req: watchRequest(
				withResourceID("4b2ef275-735e-472f-8a3e-3d394712cf79"),
				withEventTypes([]events.EventType{events.EventType_EVENT_TYPE_CREATED}),
				withResourceTypes([]events.ResourceType{events.ResourceType_RESOURCE_TYPE_TEMPLATE}),
			),
			filterNotification: false,
		},
		"same ResourceTypes and EventTypes": {
			notification: Notification{
				EventType:    "CREATED",
				ResourceType: "HARDWARE",
			},
			req: watchRequest(
				withEventTypes([]events.EventType{
					events.EventType_EVENT_TYPE_CREATED,
					events.EventType_EVENT_TYPE_DELETED,
				}),
				withResourceTypes([]events.ResourceType{
					events.ResourceType_RESOURCE_TYPE_HARDWARE,
					events.ResourceType_RESOURCE_TYPE_WORKFLOW,
				}),
			),
			filterNotification: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tc.notification.Prefix()
			result := Filter(&tc.notification, Reduce(tc.req))
			assert.Equal(t, tc.filterNotification, result)
		})
	}
}

func TestToEvent(t *testing.T) {
	t.Run("unmarshal notification", func(t *testing.T) {
		notification := Notification{
			ID:           "87794e7a-653c-46c2-8a48-195d0ee6eafa",
			ResourceID:   "4b2ef275-735e-472f-8a3e-3d394712cf79",
			EventType:    events.EventType_EVENT_TYPE_CREATED.String(),
			ResourceType: events.ResourceType_RESOURCE_TYPE_TEMPLATE.String(),
			Data:         nil,
			CreatedAt:    &time.Time{},
		}
		_, err := notification.ToEvent()
		assert.NoError(t, err)
	})
}
