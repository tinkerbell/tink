package event

import (
	"fmt"
)

// IncompatibleError indicates an event was received that.
type IncompatibleError struct {
	Event Event
}

func (e IncompatibleError) Error() string {
	return fmt.Sprintf("incompatible event: %v", e.Event.GetName())
}
