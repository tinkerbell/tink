package grpcserver

import (
	"io"

	"github.com/tinkerbell/tink/client/informers"
	"github.com/tinkerbell/tink/client/listener"
	"github.com/tinkerbell/tink/protos/events"
)

func (s *server) Watch(req *events.WatchRequest, stream events.EventsService_WatchServer) error {
	err := s.db.Events(req, func(n informers.Notification) error {
		event, err := n.ToEvent()
		if err != nil {
			return err
		}
		return stream.Send(event)
	})
	if err != nil && err != io.EOF {
		logger.Error(err)
		return err
	}

	return listener.Listen(req, func(e *events.Event) error {
		err := stream.Send(e)
		if err != nil {
			logger.With("eventTypes", req.EventTypes, "resourceTypes", req.ResourceTypes).Info("events stream closed")
			return listener.RemoveHandlers(req)
		}
		return nil
	})
}
