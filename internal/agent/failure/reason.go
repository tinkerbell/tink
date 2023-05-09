package failure

import "errors"

// Reason extracts a failure reason from err.  err has a reason if it satisfies the failure reason
// interface:
//
//	interface {
//		FailureReason() string
//	}
//
// If err does not have a reason or FailureReason() returns an empty string, ReasonUnknown is
// returned.
func Reason(err error) (string, bool) {
	fr, ok := err.(interface {
		FailureReason() string
	})

	if !ok || fr.FailureReason() == "" {
		return "", false
	}

	return fr.FailureReason(), true
}

// WithReason decorates err with reason. The reason can be extracted using Reason().
func WithReason(err error, reason string) error {
	return withReason{err, reason}
}

// NewReason creates a new error using message and wraps it with reason. The reason can be
// extracted using Reason().
func NewReason(message, reason string) error {
	return WithReason(errors.New(message), reason)
}

type withReason struct {
	error
	reason string
}

func (e withReason) FailureReason() string {
	return e.reason
}
