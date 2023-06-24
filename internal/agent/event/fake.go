package event

import "context"

// NoopRecorder retrieves a nooping fake recorder.
func NoopRecorder() *RecorderMock {
	return &RecorderMock{
		RecordEventFunc: func(context.Context, Event) error { return nil },
	}
}
