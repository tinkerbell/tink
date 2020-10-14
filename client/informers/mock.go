package informers

import "github.com/tinkerbell/tink/protos/events"

type watchRequestModifier func(*events.WatchRequest)

func watchRequest(m ...watchRequestModifier) *events.WatchRequest {
	req := &events.WatchRequest{
		EventTypes:    []events.EventType{},
		ResourceTypes: []events.ResourceType{},
	}
	for _, fn := range m {
		fn(req)
	}
	return req
}

func withAllEventTypes() watchRequestModifier {
	return func(e *events.WatchRequest) {
		e.EventTypes = []events.EventType{
			events.EventType_CREATED,
			events.EventType_UPDATED,
			events.EventType_DELETED,
		}
	}
}

func withEventTypes(ets []events.EventType) watchRequestModifier {
	return func(e *events.WatchRequest) {
		e.EventTypes = ets
	}
}

func withAllResourceTypes() watchRequestModifier {
	return func(e *events.WatchRequest) {
		e.ResourceTypes = []events.ResourceType{
			events.ResourceType_TEMPLATE,
			events.ResourceType_HARDWARE,
			events.ResourceType_WORKFLOW,
		}
	}
}

func withResourceTypes(rts []events.ResourceType) watchRequestModifier {
	return func(e *events.WatchRequest) {
		e.ResourceTypes = rts
	}
}

func withResourceID(id string) watchRequestModifier {
	return func(e *events.WatchRequest) {
		e.ResourceId = id
	}
}
