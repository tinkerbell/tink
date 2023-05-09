package event

// We want to force events to reside in this package so its clear what events are usable
// with by agent code. We achieve this using a compile time check that ensures all events
// implement an unexported method on the Event interface which is the interface passed around
// by event handling code.
//
// This source file should not contain methods other than the isEventFromThisPackage().
//
// This code is hand written.

func (ActionStarted) isEventFromThisPackage()   {}
func (ActionSucceeded) isEventFromThisPackage() {}
func (ActionFailed) isEventFromThisPackage()    {}

func (WorkflowRejected) isEventFromThisPackage() {}
