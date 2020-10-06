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

// Filter a notification based on given event and resource types
func (n Notification) Filter(resourceID string, eventTypes map[events.EventType]struct{}, resourceTypes map[events.ResourceType]struct{}) bool {
	if resourceID != "" && n.ResourceID != resourceID {
		return true
	}

	eType := GetEventType(n.EventType)
	if _, ok := eventTypes[eType]; !ok {
		return true
	}

	rType := GetResourceType(n.ResourceType)
	if _, ok := resourceTypes[rType]; !ok {
		return true
	}
	return false
}

// Unmarshal converts a notification into events.Event type
func (n Notification) Unmarshal(e *events.Event) error {
	d, err := json.Marshal(n.Data)
	if err != nil {
		return err
	}

	createdAt, err := ptypes.TimestampProto(*n.CreatedAt)
	if err != nil {
		return err
	}

	e.Id = n.ID
	e.ResourceId = n.ResourceID
	e.ResourceType = GetResourceType(n.ResourceType)
	e.EventType = GetEventType(n.EventType)
	e.Data = d
	e.CreatedAt = createdAt
	return nil
}

// GetResourceType returns events.ResourceType for a given key
func GetResourceType(name string) events.ResourceType {
	switch name {
	case events.ResourceType_TEMPLATE.String():
		return events.ResourceType_TEMPLATE
	case events.ResourceType_HARDWARE.String():
		return events.ResourceType_HARDWARE
	case events.ResourceType_WORKFLOW.String():
		return events.ResourceType_WORKFLOW
	default:
		return events.ResourceType_UNKNOWN_RESOURCE
	}
}

// GetEventType returns events.EventType for a given key
func GetEventType(name string) events.EventType {
	switch name {
	case events.EventType_CREATED.String():
		return events.EventType_CREATED
	case events.EventType_UPDATED.String():
		return events.EventType_UPDATED
	case events.EventType_DELETED.String():
		return events.EventType_DELETED
	default:
		return events.EventType_UNKNOWN_EVENT
	}
}

// MapEventType converts a []events.EventType to a map[events.EventType]struct{}
func MapEventType(eventTypes []events.EventType) map[events.EventType]struct{} {
	mapETs := map[events.EventType]struct{}{}
	for _, eventType := range eventTypes {
		mapETs[eventType] = struct{}{}
	}
	return mapETs
}

// MapResourceType converts a []events.ResourceType to a map[events.ResourceType]struct{}
func MapResourceType(resourceTypes []events.ResourceType) map[events.ResourceType]struct{} {
	mapRTs := map[events.ResourceType]struct{}{}
	for _, resourceType := range resourceTypes {
		mapRTs[resourceType] = struct{}{}
	}
	return mapRTs
}
