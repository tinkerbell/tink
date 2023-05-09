package agent

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/go-logr/logr"
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
	logErrorKey  = "error"
	logReasonKey = "reason"
)

// validReasonRegex defines the regex for a valid action failure reason.
var validReasonRegex = regexp.MustCompile(`^[a-zA-Z]+$`)

// run executes the workflow using the runtime configured on the Agent.
func (agent *Agent) run(ctx context.Context, wflw workflow.Workflow, events event.Recorder) error {
	logger := agent.Log.WithValues("workflow", wflw)
	logger.Info("Starting workflow")

	for _, action := range wflw.Actions {
		logger := logger.WithValues("action_id", action.ID, "action_name", action.Name)

		start := time.Now()
		logger.Info("Starting action")

		events.RecordEvent(ctx, event.ActionStarted{
			ActionID:   action.ID,
			WorkflowID: wflw.ID,
		})

		if err := agent.Runtime.Run(ctx, action); err != nil {
			reason := extractReason(logger, err)

			// We consider newlines in the failure message invalid because it upsets formatting.
			// The failure message is vital to easy debugability so we force the string into
			// something we're happy with and communicate that.
			message := strings.ReplaceAll(err.Error(), "\n", `\n`)

			logger.Info("Action failed - terminating workflow",
				logErrorKey, err,
				logReasonKey, reason,
				"duration", time.Since(start).String(),
			)
			events.RecordEvent(ctx, event.ActionFailed{
				ActionID:   action.ID,
				WorkflowID: wflw.ID,
				Reason:     reason,
				Message:    message,
			})
			return nil
		}

		events.RecordEvent(ctx, event.ActionSucceeded{
			ActionID:   action.ID,
			WorkflowID: wflw.ID,
		})

		logger.Info("Finished action", "duration", time.Since(start).String())
	}

	logger.Info("Finished workflow")

	return nil
}

func extractReason(logger logr.Logger, err error) string {
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
	return reason
}
