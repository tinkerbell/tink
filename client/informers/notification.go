package informers

import (
	"encoding/json"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/tinkerbell/tink/protos/events"
)

// Notification represents an event notification
type Notification struct {
	ID           string      `json:"id,omitempty"`
	ResourceID   string      `json:"resource_id,omitempty"`
	ResourceType string      `json:"resource_type,omitempty"`
	EventType    string      `json:"event_type,omitempty"`
	Data         interface{} `json:"data,omitempty"`
	CreatedAt    *time.Time  `json:"created_at,omitempty"`
}

// ToEvent converts a notification into events.Event type
func (n Notification) ToEvent() (*events.Event, error) {
	d, err := json.Marshal(n.Data)
	if err != nil {
		return nil, err
	}

	createdAt, err := ptypes.TimestampProto(*n.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &events.Event{
		Id:           n.ID,
		ResourceId:   n.ResourceID,
		ResourceType: ResourceType(n.ResourceType),
		EventType:    EventType(n.EventType),
		Data:         d,
		CreatedAt:    createdAt,
	}, nil
}

// Prefix adds prefix to notification's resource and event type
func (n *Notification) Prefix() {
	const (
		resourceTypePrefix = "RESOURCE_TYPE_"
		eventTypePrefix    = "EVENT_TYPE_"
	)
	n.ResourceType = resourceTypePrefix + n.ResourceType
	n.EventType = eventTypePrefix + n.EventType
}

// Filter a notification based on given reducer.
func Filter(n *Notification, reducer func(n *Notification) bool) bool {
	return reducer(n)
}

// ResourceType returns events.ResourceType for a given key
func ResourceType(name string) events.ResourceType {
	switch name {
	case events.ResourceType_RESOURCE_TYPE_TEMPLATE.String():
		return events.ResourceType_RESOURCE_TYPE_TEMPLATE
	case events.ResourceType_RESOURCE_TYPE_HARDWARE.String():
		return events.ResourceType_RESOURCE_TYPE_HARDWARE
	case events.ResourceType_RESOURCE_TYPE_WORKFLOW.String():
		return events.ResourceType_RESOURCE_TYPE_WORKFLOW
	default:
		return events.ResourceType_RESOURCE_TYPE_UNKNOWN
	}
}

// EventType returns events.EventType for a given key
func EventType(name string) events.EventType {
	switch name {
	case events.EventType_EVENT_TYPE_CREATED.String():
		return events.EventType_EVENT_TYPE_CREATED
	case events.EventType_EVENT_TYPE_UPDATED.String():
		return events.EventType_EVENT_TYPE_UPDATED
	case events.EventType_EVENT_TYPE_DELETED.String():
		return events.EventType_EVENT_TYPE_DELETED
	default:
		return events.EventType_EVENT_TYPE_UNKNOWN
	}
}

// Reduce returns a closure to filter notifications.
func Reduce(req *events.WatchRequest) func(n *Notification) bool {
	return func(n *Notification) bool {
		if req.ResourceId != "" && n.ResourceID != req.ResourceId {
			return true
		}

		eType := EventType(n.EventType)
		for i, t := range req.EventTypes {
			if t == eType {
				break
			}
			if i == len(req.EventTypes)-1 && t != eType {
				return true
			}
		}

		rType := ResourceType(n.ResourceType)
		for i, t := range req.ResourceTypes {
			if t == rType {
				break
			}
			if i == len(req.ResourceTypes)-1 && t != rType {
				return true
			}
		}
		return false
	}
}
