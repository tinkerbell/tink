package rollbar

import (
	"fmt"
	"os"
	"runtime"

	"github.com/pkg/errors"
	rollbar "github.com/rollbar/rollbar-go"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

func Setup(l *zap.SugaredLogger, service string) func() {
	log = l

	token := os.Getenv("ROLLBAR_TOKEN")
	if token == "" {
		log.Panicw("required envvar is unset", "envvar", "ROLLBAR_TOKEN")
	}
	rollbar.SetToken(token)

	env := os.Getenv("PACKET_ENV")
	if env == "" {
		log.Panicw("required envvar is unset", "envvar", "PACKET_ENV")
	}
	rollbar.SetEnvironment(env)

	v := os.Getenv("PACKET_VERSION")
	if v == "" {
		log.Panicw("required envvar is unset", "envvar", "PACKET_VERSION")
	}
	rollbar.SetCodeVersion(v)
	rollbar.SetServerRoot("/" + service)

	enable := true
	if os.Getenv("ROLLBAR_DISABLE") != "" {
		enable = false
	}
	rollbar.SetEnabled(enable)

	return rollbar.Wait
}

// rError exists to implement rollbar.CauseStacker so that rollbar can have stack info.
// see https://github.com/rollbar/rollbar-go/blob/v1.0.2/doc.go#L64
type rError struct {
	err error
}

func (e rError) Error() string {
	return e.err.Error()
}

func (e rError) Cause() error {
	return e.err
}

// logInternalError is a helper to log errors through zap and to rollbar if we run into an error while logging a client's error.
// We can use rollbar.ErrorWithExtras here because the stack trace rollbar collects will be of where error is.
// This handles the so called "error while logging error" case.
func logInternalError(err error, ctx map[string]interface{}) {
	l := log.With("error", err)
	if len(ctx) != 0 {
		fields := make([]interface{}, 0, len(ctx)*2)
		for k, v := range ctx {
			fields = append(fields, k)
			fields = append(fields, v)
		}
		l = l.With(fields...)
	}
	ctx["errorVerbose"] = fmt.Sprintf("%+v", err)
	l.Error(err)
	// 1 level of stack frames are skipped, because we don't want care to have logInternalError show up
	rollbar.ErrorWithStackSkipWithExtras(rollbar.ERR, err, 1, ctx)
}

// Stack converts a github.com/pkg/errors Error stack into a rollbar stack
func (e rError) Stack() rollbar.Stack {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}
	type causer interface {
		Cause() error
	}
	ctx := map[string]interface{}{}

	err := e.Cause()
	var st stackTracer
	var ok bool
	// try to find if there's a stackTracer in the stack of errors
	// WithMessage does not add a stack so New->WithMessage->WM->...->WM means we need to unwrap until we get to the
	// New'd one.
	for {
		st, ok = err.(stackTracer)
		if ok {
			break
		}
		cause, ok := err.(causer)
		if !ok {
			ctx["cause"] = e.Cause()
			logInternalError(errors.New("cause does not implement StackTracer"), ctx)
			return nil
		}
		err = cause.Cause()
	}

	stack := st.StackTrace()
	rStack := rollbar.Stack(make([]rollbar.Frame, len(stack)))

	for i, frame := range stack {
		// From pkg/error's docs
		//
		// Frame represents a program counter inside a stack frame.
		// For historical reasons if Frame is interpreted as a uintptr
		// its value represents the program counter + 1.
		// type Frame uintptr
		frame -= 1

		f := runtime.FuncForPC(uintptr(frame))
		rStack[i].Method = f.Name()
		rStack[i].Filename, rStack[i].Line = f.FileLine(uintptr(frame))
	}

	return rStack
}

func Notify(err error, args ...interface{}) {
	rErr := rError{err: err}
	rollbar.Error(rErr)
}
