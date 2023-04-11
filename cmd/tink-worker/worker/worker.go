package worker

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/tinkerbell/tink/internal/proto"
)

const (
	defaultDataDir = "/worker"

	errGetWfContext       = "failed to get workflow context"
	errGetWfActions       = "failed to get actions for workflow"
	errReportActionStatus = "failed to report action status"

	msgTurn = "it's turn for a different worker: %s"
)

type loggingContext string

var loggingContextKey loggingContext = "logger"

// WorkflowMetadata is the metadata related to workflow data.
type WorkflowMetadata struct {
	WorkerID  string    `json:"workerID"`
	Action    string    `json:"actionName"`
	Task      string    `json:"taskName"`
	UpdatedAt time.Time `json:"updatedAt"`
	SHA       string    `json:"sha256"`
}

// Option is a type for modifying a worker.
type Option func(*Worker)

// WithRetries adds custom retries to a worker.
func WithRetries(interval time.Duration, retries int) Option {
	return func(w *Worker) {
		w.retries = retries
		w.retryInterval = interval
	}
}

// WithDataDir changes the default directory for a worker.
func WithDataDir(dir string) Option {
	return func(w *Worker) {
		w.dataDir = dir
	}
}

// WithMaxFileSize changes the max file size for a worker.
func WithMaxFileSize(maxSize int64) Option {
	return func(w *Worker) {
		w.maxSize = maxSize
	}
}

// WithLogCapture enables capture of container logs.
func WithLogCapture(capture bool) Option {
	return func(w *Worker) {
		w.captureLogs = capture
	}
}

// WithPrivileged enables containers to be privileged.
func WithPrivileged(privileged bool) Option {
	return func(w *Worker) {
		w.createPrivileged = privileged
	}
}

// LogCapturer emits container logs.
type LogCapturer interface {
	CaptureLogs(ctx context.Context, containerID string)
}

// ContainerManager manages linux containers for Tinkerbell workers.
type ContainerManager interface {
	CreateContainer(ctx context.Context, cmd []string, wfID string, action *proto.WorkflowAction, captureLogs, privileged bool) (string, error)
	StartContainer(ctx context.Context, id string) error
	WaitForContainer(ctx context.Context, id string) (proto.State, error)
	WaitForFailedContainer(ctx context.Context, id string, failedActionStatus chan proto.State)
	RemoveContainer(ctx context.Context, id string) error
	PullImage(ctx context.Context, image string) error
}

// Worker details provide all the context needed to run workflows.
type Worker struct {
	workerID         string
	logCapturer      LogCapturer
	containerManager ContainerManager
	tinkClient       proto.WorkflowServiceClient
	logger           logr.Logger

	dataDir string
	maxSize int64

	createPrivileged bool
	captureLogs      bool

	retries       int
	retryInterval time.Duration
}

// NewWorker creates a new Worker, creating a new Docker registry client.
func NewWorker(
	workerID string,
	tinkClient proto.WorkflowServiceClient,
	containerManager ContainerManager,
	logCapturer LogCapturer,
	logger logr.Logger,
	opts ...Option,
) *Worker {
	w := &Worker{
		workerID:         workerID,
		dataDir:          defaultDataDir,
		containerManager: containerManager,
		logCapturer:      logCapturer,
		tinkClient:       tinkClient,
		logger:           logger,
		captureLogs:      false,
		createPrivileged: false,
		retries:          3,
		retryInterval:    time.Second * 3,
		maxSize:          1 << 20,
	}
	for _, opt := range opts {
		opt(w)
	}

	return w
}

// getLogger is a helper function to get logging out of a context, or use the default logger.
func (w Worker) getLogger(ctx context.Context) logr.Logger {
	loggerIface := ctx.Value(loggingContextKey)
	if loggerIface == nil {
		return w.logger
	}
	l, _ := loggerIface.(logr.Logger)
	return l
}

// execute executes a workflow action, optionally capturing logs.
func (w *Worker) execute(ctx context.Context, wfID string, action *proto.WorkflowAction) (proto.State, error) {
	l := w.getLogger(ctx).WithValues("workflowID", wfID, "workerID", action.GetWorkerId(), "actionName", action.GetName(), "actionImage", action.GetImage())

	if err := w.containerManager.PullImage(ctx, action.GetImage()); err != nil {
		return proto.State_STATE_RUNNING, errors.Wrap(err, "pull image")
	}

	id, err := w.containerManager.CreateContainer(ctx, action.Command, wfID, action, w.captureLogs, w.createPrivileged)
	if err != nil {
		return proto.State_STATE_RUNNING, errors.Wrap(err, "create container")
	}

	l.Info("container created", "containerID", id, "command", action.Command)

	var timeCtx context.Context
	var cancel context.CancelFunc

	if action.Timeout > 0 {
		timeCtx, cancel = context.WithTimeout(ctx, time.Duration(action.Timeout)*time.Second)
	} else {
		timeCtx, cancel = context.WithTimeout(ctx, 1*time.Hour)
	}
	defer cancel()

	err = w.containerManager.StartContainer(timeCtx, id)
	if err != nil {
		return proto.State_STATE_RUNNING, errors.Wrap(err, "start container")
	}

	if w.captureLogs {
		go w.logCapturer.CaptureLogs(ctx, id)
	}

	st, err := w.containerManager.WaitForContainer(timeCtx, id)
	l.Info("wait container completed", "status", st.String())

	// If we've made it this far, the container has successfully completed.
	// Everything after this is just cleanup.

	defer func() {
		if err := w.containerManager.RemoveContainer(ctx, id); err != nil {
			l.Error(err, "remove container", "containerID", id)
		}
		l.Info("container removed", "status", st.String())
	}()

	if err != nil {
		return st, errors.Wrap(err, "wait container")
	}

	if st == proto.State_STATE_SUCCESS {
		l.Info("action container exited with success", "status", st)
		return st, nil
	}

	if st == proto.State_STATE_TIMEOUT && action.OnTimeout != nil {
		rst := w.executeReaction(ctx, st.String(), action.OnTimeout, wfID, action)
		l.Info("action timeout", "status", rst)
	} else if action.OnFailure != nil {
		rst := w.executeReaction(ctx, st.String(), action.OnFailure, wfID, action)
		l.Info("action failed", "status", rst)
	}

	l.Info(infoWaitFinished)
	if err != nil {
		l.Error(err, errFailedToWait)
	}

	l.Info("action container exited", "status", st)
	return st, nil
}

// executeReaction executes special case OnTimeout/OnFailure actions.
func (w *Worker) executeReaction(ctx context.Context, reaction string, cmd []string, wfID string, action *proto.WorkflowAction) proto.State {
	l := w.getLogger(ctx)
	id, err := w.containerManager.CreateContainer(ctx, cmd, wfID, action, w.captureLogs, w.createPrivileged)
	if err != nil {
		l.Error(err, errFailedToRunCmd)
	}
	l.Info("container created", "containerID", id, "actionStatus", reaction, "command", cmd)

	if w.captureLogs {
		go w.logCapturer.CaptureLogs(ctx, id)
	}

	st := make(chan proto.State)

	go w.containerManager.WaitForFailedContainer(ctx, id, st)
	err = w.containerManager.StartContainer(ctx, id)
	if err != nil {
		l.Error(err, errFailedToRunCmd)
	}

	return <-st
}

// ProcessWorkflowActions gets all Workflow contexts and processes their actions.
func (w *Worker) ProcessWorkflowActions(ctx context.Context) error {
	l := w.logger.WithValues("workerID", w.workerID)
	l.Info("starting to process workflow actions")

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		res, err := w.tinkClient.GetWorkflowContexts(ctx, &proto.WorkflowContextRequest{WorkerId: w.workerID})
		if err != nil {
			l.Error(err, errGetWfContext)
			<-time.After(w.retryInterval)
			continue
		}
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			wfContext, err := res.Recv()
			if err != nil || wfContext == nil {
				if !errors.Is(err, io.EOF) {
					l.Info(err.Error())
				}
				<-time.After(w.retryInterval)
				break
			}
			wfID := wfContext.GetWorkflowId()
			l = l.WithValues("workflowID", wfID)
			ctx := context.WithValue(ctx, loggingContextKey, l)

			actions, err := w.tinkClient.GetWorkflowActions(ctx, &proto.WorkflowActionsRequest{WorkflowId: wfID})
			if err != nil {
				l.Error(err, errGetWfActions)
				continue
			}

			turn := false
			actionIndex := 0
			var nextAction *proto.WorkflowAction
			if wfContext.GetCurrentAction() == "" {
				if actions.GetActionList()[0].GetWorkerId() == w.workerID {
					actionIndex = 0
					turn = true
				}
			} else {
				switch wfContext.GetCurrentActionState() {
				case proto.State_STATE_SUCCESS:
					if isLastAction(wfContext, actions) {
						continue
					}
					nextAction = actions.GetActionList()[wfContext.GetCurrentActionIndex()+1]
					actionIndex = int(wfContext.GetCurrentActionIndex()) + 1
				case proto.State_STATE_FAILED:
					continue
				case proto.State_STATE_TIMEOUT:
					continue
				default:
					nextAction = actions.GetActionList()[wfContext.GetCurrentActionIndex()]
					actionIndex = int(wfContext.GetCurrentActionIndex())
				}
				if nextAction.GetWorkerId() == w.workerID {
					turn = true
				}
			}

			for turn {
				l.Info("starting action")
				action := actions.GetActionList()[actionIndex]
				l := l.WithValues(
					"actionName", action.GetName(),
					"taskName", action.GetTaskName(),
				)
				ctx := context.WithValue(ctx, loggingContextKey, l)
				if wfContext.GetCurrentActionState() != proto.State_STATE_RUNNING {
					actionStatus := &proto.WorkflowActionStatus{
						WorkflowId:   wfID,
						TaskName:     action.GetTaskName(),
						ActionName:   action.GetName(),
						ActionStatus: proto.State_STATE_RUNNING,
						Seconds:      0,
						Message:      "Started execution",
						WorkerId:     action.GetWorkerId(),
					}
					w.reportActionStatus(ctx, l, actionStatus)
					l.Info("sent action status", "status", actionStatus.ActionStatus, "duration", strconv.FormatInt(actionStatus.Seconds, 10))
				}

				// start executing the action
				start := time.Now()
				st, err := w.execute(ctx, wfID, action)
				elapsed := time.Since(start)

				actionStatus := &proto.WorkflowActionStatus{
					WorkflowId: wfID,
					TaskName:   action.GetTaskName(),
					ActionName: action.GetName(),
					Seconds:    int64(elapsed.Seconds()),
					WorkerId:   action.GetWorkerId(),
				}

				if err != nil || st != proto.State_STATE_SUCCESS {
					if st == proto.State_STATE_TIMEOUT {
						actionStatus.ActionStatus = proto.State_STATE_TIMEOUT
					} else {
						actionStatus.ActionStatus = proto.State_STATE_FAILED
					}
					l = l.WithValues("actionStatus", actionStatus.ActionStatus.String())
					l.Error(err, "execute workflow")
					w.reportActionStatus(ctx, l, actionStatus)
					break
				}

				actionStatus.ActionStatus = proto.State_STATE_SUCCESS
				actionStatus.Message = "finished execution successfully"
				w.reportActionStatus(ctx, l, actionStatus)
				l.Info("sent action status")

				if len(actions.GetActionList()) == actionIndex+1 {
					l.Info("reached to end of workflow")
					break
				}

				nextAction := actions.GetActionList()[actionIndex+1]
				if nextAction.GetWorkerId() != w.workerID {
					l.Info(fmt.Sprintf(msgTurn, nextAction.GetWorkerId()))
					turn = false
				} else {
					actionIndex++
				}
			}
		}
		// sleep before asking for new workflows
		<-time.After(w.retryInterval)
	}
}

func isLastAction(wfContext *proto.WorkflowContext, actions *proto.WorkflowActionList) bool {
	return int(wfContext.GetCurrentActionIndex()) == len(actions.GetActionList())-1
}

// reportActionStatus reports the status of an action to the Tinkerbell server and retries forever on error.
func (w *Worker) reportActionStatus(ctx context.Context, l logr.Logger, actionStatus *proto.WorkflowActionStatus) {
	for {
		l.Info("reporting Action Status")
		_, err := w.tinkClient.ReportActionStatus(ctx, actionStatus)
		if err != nil {
			l.Error(err, errReportActionStatus)
			<-time.After(w.retryInterval)

			continue
		}
		return
	}
}
