package event

import (
	"context"
)

// Name is a unique name identifying an event.
type Name string

// Event is a recordable event.
type Event interface {
	// GetName retrieves the event name.
	GetName() Name

	// Force events to reside in this package - see zz_known.go.
	isEventFromThisPackage()
}

// Recorder records events generated from running a Workflow.
type Recorder interface {
	RecordEvent(context.Context, Event) error
}
