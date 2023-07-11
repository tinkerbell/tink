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

// validReasonRegex defines the regex for a valid action failure reason.
var validReasonRegex = regexp.MustCompile(`^[a-zA-Z]+$`)

// run executes the workflow using the runtime configured on agent.
func (agent *Agent) run(ctx context.Context, wflw workflow.Workflow, events event.Recorder) {
	log := agent.Log.WithValues("workflow_id", wflw.ID)

	workflowStart := time.Now()
	log.Info("Starting workflow")

	for _, action := range wflw.Actions {
		log := log.WithValues("action_id", action.ID, "action_name", action.Name)

		actionStart := time.Now()
		log.Info("Starting action")

		started := event.ActionStarted{
			ActionID:   action.ID,
			WorkflowID: wflw.ID,
		}
		if err := events.RecordEvent(ctx, started); err != nil {
			log.Error(err, "Record action start event")
			return
		}

		if err := agent.Runtime.Run(ctx, action); err != nil {
			reason := extractReason(log, err)

			// We consider newlines in the failure message invalid because it upsets formatting.
			// The failure message is vital to easy debugability so we force the string into
			// something we're happy with and communicate that.
			message := strings.ReplaceAll(err.Error(), "\n", `\n`)

			log.Info("Action failed; terminating workflow",
				"error", err,
				"reason", reason,
				"duration", time.Since(actionStart).String(),
			)

			failed := event.ActionFailed{
				ActionID:   action.ID,
				WorkflowID: wflw.ID,
				Reason:     reason,
				Message:    message,
			}
			if err := events.RecordEvent(ctx, failed); err != nil {
				log.Error(err, "Record failed action event", "event", failed)
			}

			return
		}

		succeed := event.ActionSucceeded{
			ActionID:   action.ID,
			WorkflowID: wflw.ID,
		}
		if err := events.RecordEvent(ctx, succeed); err != nil {
			log.Error(err, "Record succeeded action event")
			return
		}

		log.Info("Finished action", "duration", time.Since(actionStart).String())
	}

	log.Info("Finished workflow", "duration", time.Since(workflowStart).String())
}

func extractReason(log logr.Logger, err error) string {
	reason := ReasonRuntimeError
	if r, ok := failure.Reason(err); ok {
		reason = r
		if !validReasonRegex.MatchString(reason) {
			log.Info(
				"Received invalid reason for action failure; using InvalidReason",
				"invalid_reason", reason,
			)
			reason = ReasonInvalid
		}
	}
	return reason
}
