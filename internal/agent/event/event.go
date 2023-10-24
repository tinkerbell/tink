// Package event describes the event set and an interface for recording events. Events are
// generated as workflows execute.
package event

import "context"

// Name is a unique name identifying an event.
type Name string

// Event is an event generated during execution of a Workflow. Each event in the event package
// implements this interface. Consumers may type switch the Event to the appropriate type for
// event handling.
//
// E.g.
//
//	switch ev.(type) {
//	case event.ActionStarted:
//		// Handle ActionStarted event.
//	default:
//		// Unsupported event.
//	}
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
