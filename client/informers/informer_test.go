package informers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tinkerbell/tink/protos/events"
)

func TestProcessEvents(t *testing.T) {
	var eventsCh chan *events.Event
	var counter int
	testCases := map[string]struct {
		eventGenerator  func()
		fn              func(e *events.Event) error
		expectedCounter int
	}{
		"no error": {
			eventGenerator: func() {
				for i := 0; i < 5; i++ {
					eventsCh <- &events.Event{}
				}
			},
			fn: func(e *events.Event) error {
				counter++
				return nil
			},
			expectedCounter: 5,
		},
		"error": {
			eventGenerator: func() {
				for i := 0; i <= 5; i++ {
					if i == 3 {
						break
					}
					eventsCh <- &events.Event{}
				}
			},
			fn: func(e *events.Event) error {
				if counter == 2 {
					return errors.New("event processing error")
				}
				counter++
				return nil
			},
			expectedCounter: 2,
		},
	}

	for name, tc := range testCases {
		eventsCh = make(chan *events.Event)
		counter = 0

		t.Run(name, func(t *testing.T) {
			go processEvents(eventsCh, nil, tc.fn)
			tc.eventGenerator()
			assert.Equal(t, tc.expectedCounter, counter)
		})
	}
}
