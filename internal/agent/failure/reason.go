package failure

import "errors"

// ReasonUnknown is returned when Reason() is called on an error without a reason.
const ReasonUnknown = "Unknown"

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

// WrapWithReason decorates err with reason. The reason can be extracted using Reason().
func WrapWithReason(err error, reason string) error {
	return withReason{err, reason}
}

// WithReason creates a new error using message and wraps it with reason. The reason can be
// extracted using Reason().
func WithReason(message, reason string) error {
	return WrapWithReason(errors.New(message), reason)
}

type withReason struct {
	error
	reason string
}

func (e withReason) FailureReason() string {
	return e.reason
}
