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

func withEventTypes(ets []events.EventType) watchRequestModifier {
	return func(e *events.WatchRequest) {
		e.EventTypes = ets
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
