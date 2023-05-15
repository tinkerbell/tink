// Package event describes the event set and an interface for recording events. Events are
// generated as workflows execute.
package event

import "context"

// Name is a unique name identifying an event.
type Name string

// Event is a recordable event.
type Event interface {
	// GetName retrieves the event name.
	GetName() Name

	// Force events to reside in this package - see zz_known.go.
	isEventFromThisPackage()
}

// Recorder provides event recording methods.
type Recorder interface {
	RecordEvent(context.Context, Event) error
}
