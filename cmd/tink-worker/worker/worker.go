package worker

import (
	"context"
	sha "crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/packethost/pkg/log"
	"github.com/pkg/errors"
	pb "github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/grpc/status"
)

const (
	dataFile       = "data"
	defaultDataDir = "/worker"

	errGetWfContext       = "failed to get workflow context"
	errGetWfActions       = "failed to get actions for workflow"
	errReportActionStatus = "failed to report action status"

	msgTurn = "it's turn for a different worker: %s"
)

type loggingContext string

var loggingContextKey loggingContext = "logger"

var (
	workflowcontexts = map[string]*pb.WorkflowContext{}
	workflowDataSHA  = map[string]string{}
)

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
	CreateContainer(ctx context.Context, cmd []string, wfID string, action *pb.WorkflowAction, captureLogs, privileged bool) (string, error)
	StartContainer(ctx context.Context, id string) error
	WaitForContainer(ctx context.Context, id string) (pb.State, error)
	WaitForFailedContainer(ctx context.Context, id string, failedActionStatus chan pb.State)
	RemoveContainer(ctx context.Context, id string) error
	PullImage(ctx context.Context, image string) error
}

// Worker details provide all the context needed to run workflows.
type Worker struct {
	workerID         string
	logCapturer      LogCapturer
	containerManager ContainerManager
	tinkClient       pb.WorkflowServiceClient
	logger           log.Logger

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
	tinkClient pb.WorkflowServiceClient,
	containerManager ContainerManager,
	logCapturer LogCapturer,
	logger log.Logger,
	opts ...Option) *Worker {
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
func (w Worker) getLogger(ctx context.Context) *log.Logger {
	loggerIface := ctx.Value(loggingContextKey)
	if loggerIface == nil {
		return &w.logger
	}
	return loggerIface.(*log.Logger)
}

// execute executes a workflow action, optionally capturing logs.
func (w *Worker) execute(ctx context.Context, wfID string, action *pb.WorkflowAction) (pb.State, error) {
	l := w.getLogger(ctx).With("workflowID", wfID, "workerID", action.GetWorkerId(), "actionName", action.GetName(), "actionImage", action.GetImage())

	if err := w.containerManager.PullImage(ctx, action.GetImage()); err != nil {
		return pb.State_STATE_RUNNING, errors.Wrap(err, "pull image")
	}

	id, err := w.containerManager.CreateContainer(ctx, action.Command, wfID, action, w.captureLogs, w.createPrivileged)
	if err != nil {
		return pb.State_STATE_RUNNING, errors.Wrap(err, "create container")
	}

	l.With("containerID", id, "command", action.GetOnTimeout()).Info("container created")

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
		return pb.State_STATE_RUNNING, errors.Wrap(err, "start container")
	}

	if w.captureLogs {
		go w.logCapturer.CaptureLogs(ctx, id)
	}

	st, err := w.containerManager.WaitForContainer(timeCtx, id)
	l.With("status", st.String()).Info("wait container completed")

	// If we've made it this far, the container has successfully completed.
	// Everything after this is just cleanup.

	defer func() {
		if err := w.containerManager.RemoveContainer(ctx, id); err != nil {
			l.With("containerID", id).Error(err)
		}
		l.With("status", st.String()).Info("container removed")
	}()

	if err != nil {
		return st, errors.Wrap(err, "wait container")
	}

	l.With("status", st.String()).Info("container removed")

	if st == pb.State_STATE_SUCCESS {
		l.With("status", st).Info("action container exited with success", st)
		return st, nil
	}

	if st == pb.State_STATE_TIMEOUT && action.OnTimeout != nil {
		rst := w.executeReaction(ctx, st.String(), action.OnTimeout, wfID, action)
		l.With("status", rst).Info("action timeout")
	} else if action.OnFailure != nil {
		rst := w.executeReaction(ctx, st.String(), action.OnFailure, wfID, action)
		l.With("status", rst).Info("action failed")
	}

	l.Info(infoWaitFinished)
	if err != nil {
		l.Error(errors.Wrap(err, errFailedToWait))
	}

	l.With("status", st).Info("action container exited")
	return st, nil
}

// executeReaction executes special case OnTimeout/OnFailure actions.
func (w *Worker) executeReaction(ctx context.Context, reaction string, cmd []string, wfID string, action *pb.WorkflowAction) pb.State {
	l := w.getLogger(ctx)
	id, err := w.containerManager.CreateContainer(ctx, cmd, wfID, action, w.captureLogs, w.createPrivileged)
	if err != nil {
		l.Error(errors.Wrap(err, errFailedToRunCmd))
	}
	l.With("containerID", id, "actionStatus", reaction, "command", cmd).Info("container created")

	if w.captureLogs {
		go w.logCapturer.CaptureLogs(ctx, id)
	}

	st := make(chan pb.State)

	go w.containerManager.WaitForFailedContainer(ctx, id, st)
	err = w.containerManager.StartContainer(ctx, id)
	if err != nil {
		l.Error(errors.Wrap(err, errFailedToRunCmd))
	}

	return <-st
}

// ProcessWorkflowActions gets all Workflow contexts and processes their actions.
func (w *Worker) ProcessWorkflowActions(ctx context.Context) error {
	l := w.logger.With("workerID", w.workerID)

	for {
		res, err := w.tinkClient.GetWorkflowContexts(ctx, &pb.WorkflowContextRequest{WorkerId: w.workerID})
		if err != nil {
			return errors.Wrap(err, errGetWfContext)
		}
		for wfContext, err := res.Recv(); err == nil && wfContext != nil; wfContext, err = res.Recv() {
			wfID := wfContext.GetWorkflowId()
			l = l.With("workflowID", wfID)
			ctx := context.WithValue(ctx, loggingContextKey, &l)

			actions, err := w.tinkClient.GetWorkflowActions(ctx, &pb.WorkflowActionsRequest{WorkflowId: wfID})
			if err != nil {
				return errors.Wrap(err, errGetWfActions)
			}

			turn := false
			actionIndex := 0
			var nextAction *pb.WorkflowAction
			if wfContext.GetCurrentAction() == "" {
				if actions.GetActionList()[0].GetWorkerId() == w.workerID {
					actionIndex = 0
					turn = true
				}
			} else {
				switch wfContext.GetCurrentActionState() {
				case pb.State_STATE_SUCCESS:
					if isLastAction(wfContext, actions) {
						continue
					}
					nextAction = actions.GetActionList()[wfContext.GetCurrentActionIndex()+1]
					actionIndex = int(wfContext.GetCurrentActionIndex()) + 1
				case pb.State_STATE_FAILED:
					continue
				case pb.State_STATE_TIMEOUT:
					continue
				default:
					nextAction = actions.GetActionList()[wfContext.GetCurrentActionIndex()]
					actionIndex = int(wfContext.GetCurrentActionIndex())
				}
				l := l.With(
					"currentWorker", wfContext.GetCurrentWorker(),
					"currentTask", wfContext.GetCurrentTask(),
					"currentAction", wfContext.GetCurrentAction(),
					"currentActionIndex", strconv.FormatInt(wfContext.GetCurrentActionIndex(), 10),
					"currentActionState", wfContext.GetCurrentActionState(),
					"totalNumberOfActions", wfContext.GetTotalNumberOfActions(),
				)
				l.Info("current context")
				if nextAction.GetWorkerId() == w.workerID {
					turn = true
				}
			}

			if turn {
				wfDir := filepath.Join(w.dataDir, wfID)
				l := l.With("actionName", actions.GetActionList()[actionIndex].GetName(),
					"taskName", actions.GetActionList()[actionIndex].GetTaskName(),
				)
				if _, err := os.Stat(wfDir); os.IsNotExist(err) {
					err := os.MkdirAll(wfDir, os.FileMode(0o755))
					if err != nil {
						l.Error(err)
						os.Exit(1)
					}

					f := openDataFile(wfDir, l)
					_, err = f.Write([]byte("{}"))
					if err != nil {
						l.Error(err)
						os.Exit(1)
					}

					err = f.Close()
					if err != nil {
						l.Error(err)
						os.Exit(1)
					}
				}
				l.Info("starting with action")
			}

			for turn {
				action := actions.GetActionList()[actionIndex]
				l := l.With("actionName", action.GetName(),
					"taskName", action.GetTaskName(),
				)
				ctx := context.WithValue(ctx, loggingContextKey, &l)
				if wfContext.GetCurrentActionState() != pb.State_STATE_RUNNING {
					actionStatus := &pb.WorkflowActionStatus{
						WorkflowId:   wfID,
						TaskName:     action.GetTaskName(),
						ActionName:   action.GetName(),
						ActionStatus: pb.State_STATE_RUNNING,
						Seconds:      0,
						Message:      "Started execution",
						WorkerId:     action.GetWorkerId(),
					}
					err := w.reportActionStatus(ctx, actionStatus)
					if err != nil {
						exitWithGrpcError(err, l)
					}
					l.With("duration", strconv.FormatInt(actionStatus.Seconds, 10)).Info("sent action status")
				}

				// get workflow data
				w.getWorkflowData(ctx, wfID)

				// start executing the action
				start := time.Now()
				st, err := w.execute(ctx, wfID, action)
				elapsed := time.Since(start)

				actionStatus := &pb.WorkflowActionStatus{
					WorkflowId: wfID,
					TaskName:   action.GetTaskName(),
					ActionName: action.GetName(),
					Seconds:    int64(elapsed.Seconds()),
					WorkerId:   action.GetWorkerId(),
				}

				if err != nil || st != pb.State_STATE_SUCCESS {
					if st == pb.State_STATE_TIMEOUT {
						actionStatus.ActionStatus = pb.State_STATE_TIMEOUT
					} else {
						actionStatus.ActionStatus = pb.State_STATE_FAILED
					}
					l = l.With("actionStatus", actionStatus.ActionStatus.String())
					l.Error(err)
					if reportErr := w.reportActionStatus(ctx, actionStatus); reportErr != nil {
						exitWithGrpcError(reportErr, l)
					}
					delete(workflowcontexts, wfID)
					return err
				}

				actionStatus.ActionStatus = pb.State_STATE_SUCCESS
				actionStatus.Message = "finished execution successfully"

				err = w.reportActionStatus(ctx, actionStatus)
				if err != nil {
					exitWithGrpcError(err, l)
				}
				l.Info("sent action status")

				// send workflow data, if updated
				w.updateWorkflowData(ctx, actionStatus)

				if len(actions.GetActionList()) == actionIndex+1 {
					l.Info("reached to end of workflow")
					delete(workflowcontexts, wfID)
					turn = false // nolint:wastedassign // assigned to turn, but reassigned without using the value
					break
				}

				nextAction := actions.GetActionList()[actionIndex+1]
				if nextAction.GetWorkerId() != w.workerID {
					l.Debug(fmt.Sprintf(msgTurn, nextAction.GetWorkerId()))
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

func exitWithGrpcError(err error, l log.Logger) {
	if err != nil {
		errStatus, _ := status.FromError(err)
		l.With("errorCode", errStatus.Code()).Error(err)
		os.Exit(1)
	}
}

func isLastAction(wfContext *pb.WorkflowContext, actions *pb.WorkflowActionList) bool {
	return int(wfContext.GetCurrentActionIndex()) == len(actions.GetActionList())-1
}

func (w *Worker) reportActionStatus(ctx context.Context, actionStatus *pb.WorkflowActionStatus) error {
	l := w.getLogger(ctx).With("workflowID", actionStatus.GetWorkflowId(),
		"workerID", actionStatus.GetWorkerId(),
		"actionName", actionStatus.GetActionName(),
		"taskName", actionStatus.GetTaskName(),
	)
	var err error
	for r := 1; r <= w.retries; r++ {
		l.Info("reporting Action Status")
		_, err = w.tinkClient.ReportActionStatus(ctx, actionStatus)
		if err != nil {
			l.Error(errors.Wrap(err, errReportActionStatus))
			<-time.After(w.retryInterval)

			continue
		}
		return nil
	}
	return err
}

func (w *Worker) getWorkflowData(ctx context.Context, workflowID string) {
	l := w.getLogger(ctx).With("workflowID", workflowID,
		"workerID", w.workerID,
	)
	res, err := w.tinkClient.GetWorkflowData(ctx, &pb.GetWorkflowDataRequest{WorkflowId: workflowID})
	if err != nil {
		l.Error(err)
	}

	if len(res.GetData()) != 0 {
		wfDir := filepath.Join(w.dataDir, workflowID)
		f := openDataFile(wfDir, l)
		defer func() {
			if err := f.Close(); err != nil {
				l.With("file", f.Name()).Error(err)
			}
		}()

		_, err := f.Write(res.GetData())
		if err != nil {
			l.Error(err)
		}
		h := sha.New()
		workflowDataSHA[workflowID] = base64.StdEncoding.EncodeToString(h.Sum(res.Data))
	}
}

func (w *Worker) updateWorkflowData(ctx context.Context, actionStatus *pb.WorkflowActionStatus) {
	l := w.getLogger(ctx).With("workflowID", actionStatus.GetWorkflowId,
		"workerID", actionStatus.GetWorkerId(),
		"actionName", actionStatus.GetActionName(),
		"taskName", actionStatus.GetTaskName(),
	)

	wfDir := filepath.Join(w.dataDir, actionStatus.GetWorkflowId())
	f := openDataFile(wfDir, l)
	defer func() {
		if err := f.Close(); err != nil {
			l.With("file", f.Name()).Error(err)
		}
	}()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		l.Error(err)
	}

	if isValidDataFile(f, w.maxSize, data, l) {
		h := sha.New()
		if _, ok := workflowDataSHA[actionStatus.GetWorkflowId()]; !ok {
			checksum := base64.StdEncoding.EncodeToString(h.Sum(data))
			workflowDataSHA[actionStatus.GetWorkflowId()] = checksum
			w.sendUpdate(ctx, actionStatus, data, checksum)
		} else {
			newSHA := base64.StdEncoding.EncodeToString(h.Sum(data))
			if !strings.EqualFold(workflowDataSHA[actionStatus.GetWorkflowId()], newSHA) {
				w.sendUpdate(ctx, actionStatus, data, newSHA)
			}
		}
	}
}

func (w *Worker) sendUpdate(ctx context.Context, st *pb.WorkflowActionStatus, data []byte, checksum string) {
	l := w.getLogger(ctx).With("workflowID", st.GetWorkflowId,
		"workerID", st.GetWorkerId(),
		"actionName", st.GetActionName(),
		"taskName", st.GetTaskName(),
	)
	meta := WorkflowMetadata{
		WorkerID:  st.GetWorkerId(),
		Action:    st.GetActionName(),
		Task:      st.GetTaskName(),
		UpdatedAt: time.Now(),
		SHA:       checksum,
	}
	metadata, err := json.Marshal(meta)
	if err != nil {
		l.Error(err)
		os.Exit(1)
	}

	_, err = w.tinkClient.UpdateWorkflowData(ctx, &pb.UpdateWorkflowDataRequest{
		WorkflowId: st.GetWorkflowId(),
		Data:       data,
		Metadata:   metadata,
	})
	if err != nil {
		l.Error(err)
		os.Exit(1)
	}
}

func openDataFile(wfDir string, l log.Logger) *os.File {
	f, err := os.OpenFile(filepath.Clean(wfDir+string(os.PathSeparator)+dataFile), os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		l.Error(err)
		os.Exit(1)
	}
	return f
}

func isValidDataFile(f *os.File, maxSize int64, data []byte, l log.Logger) bool {
	var dataMap map[string]interface{}
	err := json.Unmarshal(data, &dataMap)
	if err != nil {
		l.Error(err)
		return false
	}

	stat, err := f.Stat()
	if err != nil {
		l.Error(err)
		return false
	}

	return stat.Size() <= maxSize
}
