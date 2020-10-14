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
				EventType: events.EventType_CREATED.String(),
			},
			req: watchRequest(
				withEventTypes([]events.EventType{events.EventType_DELETED}),
			),
			filterNotification: true,
		},
		"different ResourceType": {
			notification: Notification{
				ResourceType: events.ResourceType_WORKFLOW.String(),
			},
			req: watchRequest(
				withResourceTypes([]events.ResourceType{events.ResourceType_TEMPLATE}),
			),
			filterNotification: true,
		},
		"same ResourceID different EventType": {
			notification: Notification{
				ResourceID: "4b2ef275-735e-472f-8a3e-3d394712cf79",
				EventType:  events.EventType_DELETED.String(),
			},
			req: watchRequest(
				withResourceID("4b2ef275-735e-472f-8a3e-3d394712cf79"),
				withEventTypes([]events.EventType{events.EventType_CREATED}),
			),
			filterNotification: true,
		},
		"same ResourceType different EventType": {
			notification: Notification{
				EventType:    events.EventType_DELETED.String(),
				ResourceType: events.ResourceType_TEMPLATE.String(),
			},
			req: watchRequest(
				withEventTypes([]events.EventType{events.EventType_CREATED}),
				withResourceTypes([]events.ResourceType{events.ResourceType_HARDWARE}),
			),
			filterNotification: true,
		},
		"same EventType different ResourceType": {
			notification: Notification{
				EventType:    events.EventType_UPDATED.String(),
				ResourceType: events.ResourceType_TEMPLATE.String(),
			},
			req: watchRequest(
				withEventTypes([]events.EventType{events.EventType_UPDATED}),
				withResourceTypes([]events.ResourceType{events.ResourceType_HARDWARE}),
			),
			filterNotification: true,
		},
		"same ResourceID and EventType": {
			notification: Notification{
				ResourceID:   "4b2ef275-735e-472f-8a3e-3d394712cf79",
				EventType:    events.EventType_CREATED.String(),
				ResourceType: events.ResourceType_TEMPLATE.String(),
			},
			req: watchRequest(
				withResourceID("4b2ef275-735e-472f-8a3e-3d394712cf79"),
				withEventTypes([]events.EventType{events.EventType_CREATED}),
				withResourceTypes([]events.ResourceType{events.ResourceType_TEMPLATE}),
			),
			filterNotification: false,
		},
		"same ResourceTypes and EventTypes": {
			notification: Notification{
				EventType:    events.EventType_CREATED.String(),
				ResourceType: events.ResourceType_HARDWARE.String(),
			},
			req: watchRequest(
				withEventTypes([]events.EventType{
					events.EventType_CREATED,
					events.EventType_DELETED,
				}),
				withResourceTypes([]events.ResourceType{
					events.ResourceType_HARDWARE,
					events.ResourceType_WORKFLOW,
				}),
			),
			filterNotification: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := tc.notification.Filter(
				tc.req.ResourceId,
				MapEventType(tc.req.EventTypes),
				MapResourceType(tc.req.ResourceTypes),
			)
			assert.Equal(t, tc.filterNotification, result)
		})
	}
}

func TestUnmarshal(t *testing.T) {
	t.Run("unmarshal notification", func(t *testing.T) {
		event := &events.Event{}
		notification := Notification{
			ID:           "87794e7a-653c-46c2-8a48-195d0ee6eafa",
			ResourceID:   "4b2ef275-735e-472f-8a3e-3d394712cf79",
			EventType:    events.EventType_CREATED.String(),
			ResourceType: events.ResourceType_TEMPLATE.String(),
			Data:         nil,
			CreatedAt:    &time.Time{},
		}
		err := notification.Unmarshal(event)
		assert.NoError(t, err)
	})
}
