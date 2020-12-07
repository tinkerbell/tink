package mock

import (
	"github.com/tinkerbell/tink/client/informers"
	"github.com/tinkerbell/tink/protos/events"
)

// Events fetches events for a given time frame, and
// sends them to over the stream
func (d DB) Events(req *events.WatchRequest, fn func(n informers.Notification) error) error {
	return nil
}
