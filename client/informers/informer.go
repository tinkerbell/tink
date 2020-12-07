package informers

import (
	"context"
	"io"

	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/events"
)

// Informer is the base informer
type Informer interface {
	Start(ctx context.Context, req *events.WatchRequest, fn func(e *events.Event) error) error
}

type sharedInformer struct {
	eventsCh chan *events.Event
	errCh    chan error
}

// New returns an instance of event informer
func New() Informer {
	return &sharedInformer{
		eventsCh: make(chan *events.Event),
		errCh:    make(chan error),
	}
}

func (s *sharedInformer) Start(ctx context.Context, req *events.WatchRequest, fn func(e *events.Event) error) error {
	defer close(s.errCh)
	stream, err := client.EventsClient.Watch(ctx, req)
	if err != nil {
		return err
	}

	go processEvents(s.eventsCh, s.errCh, fn)

	var event *events.Event
	for event, err = stream.Recv(); err == nil && event != nil; event, err = stream.Recv() {
		if err == io.EOF {
			return nil
		}
		s.eventsCh <- event
	}
	if err != nil {
		return err
	}
	close(s.eventsCh)
	return <-s.errCh
}

func processEvents(eventsCh <-chan *events.Event, errCh chan<- error, fn func(e *events.Event) error) {
	for event := range eventsCh {
		err := fn(event)
		if err != nil {
			errCh <- err
			return
		}
	}
}
