// Package log sets up a shared logger that can be used by all packages run under one binary.
//
// This package wraps zap very lightly so zap best practices apply here too, namely use `With` for KV pairs to add context to a line.
// The lack of a wide gamut of logging levels is by design.
// The intended use case for each of the levels are:
//   Error:
//     Logs a message as an error, may also have external side effects such as posting to rollbar, sentry or alerting directly.
//   Info:
//     Used for production.
//     Context should all be in K=V pairs so they can be useful to ops and future-you-at-3am.
//   Debug:
//     Meant for developer use *during development*.
package log

import (
	"os"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/packethost/pkg/log/internal/rollbar"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	// zap.LevelFlag adds an option to the default set of command line flags as part of the flag pacakge.
	logLevel = zap.LevelFlag("log-level", zap.InfoLevel, "Log level. one of ERROR, INFO, or DEBUG")
)

// Logger is a wrapper around zap.SugaredLogger
type Logger struct {
	service string
	s       *zap.SugaredLogger
}

func configureLogger(l *zap.Logger, service string) (Logger, func(), error) {
	l = l.With(zap.String("service", service))

	rollbarClean := rollbar.Setup(l.Sugar().With("pkg", "log"), service)
	cleanup := func() {
		rollbarClean()
		l.Sync()
	}

	return Logger{service: service, s: l.Sugar()}.AddCallerSkip(1), cleanup, nil
}

// Init initializes the logging system and sets the "service" key to the provided argument.
// This func should only be called once and after flag.Parse() has been called otherwise leveled logging will not be configured correctly.
func Init(service string) (Logger, func(), error) {
	var config zap.Config
	if os.Getenv("DEBUG") != "" {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}
	// We expect that errors will already log the stacktrace from pkg/errors functionality as errorVerbose context
	// key
	config.DisableStacktrace = true

	if os.Getenv("LOG_DISCARD_LOGS") != "" {
		config.OutputPaths = nil
		config.ErrorOutputPaths = nil
	}

	config.Level = zap.NewAtomicLevelAt(*logLevel)

	l, err := config.Build()
	if err != nil {
		return Logger{}, nil, errors.Wrap(err, "failed to build logger config")
	}

	return configureLogger(l, service)
}

// Error is used to log an error, the error will be forwared to rollbar and/or other external services.
// All the values of arg are stringified and concatenated without any strings.
// If no args are provided err.Error() is used as the log message.
func (l Logger) Error(err error, args ...interface{}) {
	rollbar.Notify(err, args)
	if len(args) == 0 {
		args = append(args, err)
	}
	l.s.With("error", err).Error(args...)
}

// Info is used to log message in production, only simple strings should be given in the args.
// Context should be added as K=V pairs using the `With` method.
// All the values of arg are stringified and concatenated without any strings.
func (l Logger) Info(args ...interface{}) {
	l.s.Info(args...)
}

// Debug is used to log messages in development, not even for lab.
// No one cares what you pass to Debug.
// All the values of arg are stringified and concatenated without any strings.
func (l Logger) Debug(args ...interface{}) {
	l.s.Debug(args...)
}

// With is used to add context to the logger, a new logger copy with the new K=V pairs as context is returned.
func (l Logger) With(args ...interface{}) Logger {
	return Logger{service: l.service, s: l.s.With(args...)}
}

// AddCallerSkip increases the number of callers skipped by caller annotation.
// When building wrappers around the Logger, supplying this option prevents Logger from always reporting the wrapper code as the caller.
func (l Logger) AddCallerSkip(skip int) Logger {
	s := l.s.Desugar().WithOptions(zap.AddCallerSkip(skip)).Sugar()
	return Logger{service: l.service, s: s}
}

// Package returns a copy of the logger with the "pkg" set to the argument.
// It should be called before the original Logger has had any keys set to values, otherwise confusion may ensue.
func (l Logger) Package(pkg string) Logger {
	return Logger{service: l.service, s: l.s.With("pkg", pkg)}
}

// GRPCLoggers returns server side logging middleware for gRPC servers
func (l Logger) GRPCLoggers() (grpc.StreamServerInterceptor, grpc.UnaryServerInterceptor) {
	logger := l.s.Desugar()
	return grpc_zap.StreamServerInterceptor(logger), grpc_zap.UnaryServerInterceptor(logger)
}
