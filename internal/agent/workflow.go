package agent

import (
	"context"
	"regexp"
	"time"

	"github.com/tinkerbell/tink/internal/agent/event"
	"github.com/tinkerbell/tink/internal/agent/failure"
	"github.com/tinkerbell/tink/internal/agent/workflow"
)

// ReasonRuntimeError is the default reason used when no reason is provided by the runtime.
const ReasonRuntimeError = "RuntimeError"

// ReasonInvalid indicates a reason provided by the runtime was invalid.
const ReasonInvalid = "InvalidReason"

// Consistent logging keys.
const (
	logEventKey  = "event"
	logErrorKey  = "error"
	logReasonKey = "reason"
)

// validReasonRegex defines the regex for a valid action failure reason.
var validReasonRegex = regexp.MustCompile(`^[a-zA-Z]+$`)

// run executes the workflow using the runtime configured on a.
func (a *Agent) run(ctx context.Context, wflw workflow.Workflow, events event.Recorder) error {
	logger := a.Log.WithValues("workflow", wflw)
	logger.Info("Starting workflow")

	for _, action := range wflw.Actions {
		logger := logger.WithValues("action_id", action.ID, "action_name", action.Name)

		start := time.Now()
		logger.Info("Starting action")

		a.recordNonTerminatingEvent(ctx, events, event.ActionStarted{
			ActionID:   action.ID,
			WorkflowID: wflw.ID,
		})

		if err := a.Runtime.Run(ctx, action); err != nil {
			reason := ReasonRuntimeError
			if r, ok := failure.Reason(err); ok {
				reason = r
				if !validReasonRegex.MatchString(reason) {
					logger.Info(
						"Received invalid reason for action failure; using InvalidReason instead",
						logReasonKey, reason,
					)
					reason = ReasonInvalid
				}
			}

			logger.Info("Failed to run action; terminating workflow",
				logErrorKey, err,
				logReasonKey, reason,
				"duration", time.Since(start).String(),
			)
			a.recordTerminatingEvent(ctx, events, event.ActionFailed{
				ActionID:   action.ID,
				WorkflowID: wflw.ID,
				Reason:     reason,
				Message:    err.Error(),
			})
			return nil
		}

		a.recordNonTerminatingEvent(ctx, events, event.ActionSucceeded{
			ActionID:   action.ID,
			WorkflowID: wflw.ID,
		})

		logger.Info("Finished action", "duration", time.Since(start).String())
	}

	logger.Info("Finished workflow")

	return nil
}

func (a *Agent) recordNonTerminatingEvent(ctx context.Context, events event.Recorder, e event.Event) {
	if err := events.RecordEvent(ctx, e); err != nil {
		a.Log.Info(
			"Failed to record event; continuing workflow",
			logEventKey, e,
			logErrorKey, err,
		)
	}
}

func (a *Agent) recordTerminatingEvent(ctx context.Context, events event.Recorder, e event.Event) {
	if err := events.RecordEvent(ctx, e); err != nil {
		a.Log.Info("Failed to record event", logEventKey, e, logErrorKey, err)
	}
}
