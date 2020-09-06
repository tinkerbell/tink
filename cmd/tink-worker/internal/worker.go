package internal

import (
	"bufio"
	"context"
	sha "crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/packethost/pkg/log"
	"github.com/pkg/errors"
	pb "github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/grpc/status"
)

const (
	dataFile = "data"
	dataDir  = "/worker"

	errGetWfContext       = "failed to get workflow context"
	errGetWfActions       = "failed to get actions for workflow"
	errReportActionStatus = "failed to report action status"

	msgTurn = "it's turn for a different worker: %s"
)

var (
	workflowcontexts = map[string]*pb.WorkflowContext{}
	workflowDataSHA  = map[string]string{}
)

// WorkflowMetadata is the metadata related to workflow data
type WorkflowMetadata struct {
	WorkerID  string    `json:"workerID"`
	Action    string    `json:"actionName"`
	Task      string    `json:"taskName"`
	UpdatedAt time.Time `json:"updatedAt"`
	SHA       string    `json:"sha256"`
}

// Worker details provide all the context needed to run a
type Worker struct {
	client         pb.WorkflowSvcClient
	regConn        *RegistryConnDetails
	registryClient *client.Client
	logger         log.Logger
	registry       string
	retries        int
	retryInterval  time.Duration
	maxSize        int64
}

// NewWorker creates a new Worker, creating a new Docker registry client
func NewWorker(client pb.WorkflowSvcClient, regConn *RegistryConnDetails, logger log.Logger, registry string, retries int, retryInterval time.Duration, maxFileSize int64) *Worker {
	registryClient, err := regConn.NewClient()
	if err != nil {
		panic(err)
	}
	return &Worker{
		client:         client,
		regConn:        regConn,
		registryClient: registryClient,
		logger:         logger,
		registry:       registry,
		retries:        retries,
		retryInterval:  retryInterval,
		maxSize:        maxFileSize,
	}
}

func (w *Worker) captureLogs(ctx context.Context, id string) {
	reader, err := w.registryClient.ContainerLogs(ctx, id, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: false,
	})
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

func (w *Worker) execute(ctx context.Context, wfID string, action *pb.WorkflowAction) (pb.ActionState, error) {
	l := w.logger.With("workflowID", wfID, "workerID", action.GetWorkerId(), "actionName", action.GetName(), "actionImage", action.GetImage())

	cli := w.registryClient
	err := w.regConn.pullImage(ctx, cli, action.GetImage())
	if err != nil {
		return pb.ActionState_ACTION_IN_PROGRESS, errors.Wrap(err, "DOCKER PULL")
	}
	id, err := w.createContainer(ctx, action.Command, wfID, action)
	if err != nil {
		return pb.ActionState_ACTION_IN_PROGRESS, errors.Wrap(err, "DOCKER CREATE")
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

	err = startContainer(timeCtx, l, cli, id)
	if err != nil {
		return pb.ActionState_ACTION_IN_PROGRESS, errors.Wrap(err, "DOCKER RUN")
	}

	failedActionStatus := make(chan pb.ActionState)

	// capturing logs of action container in a go-routine
	go w.captureLogs(ctx, id)

	status, waitErr := waitContainer(timeCtx, cli, id)
	defer func() {
		if removalErr := removeContainer(ctx, l, cli, id); removalErr != nil {
			l.With("containerID", id).Error(removalErr)
		}
	}()

	if waitErr != nil {
		return status, errors.Wrap(waitErr, "DOCKER_WAIT")
	}

	l.With("status", status.String()).Info("container removed")
	if status != pb.ActionState_ACTION_SUCCESS {
		if status == pb.ActionState_ACTION_TIMEOUT && action.OnTimeout != nil {
			id, err = w.createContainer(ctx, action.OnTimeout, wfID, action)
			if err != nil {
				l.Error(errors.Wrap(err, errCreateContainer))
			}
			l.With("containerID", id, "status", status.String(), "command", action.GetOnTimeout()).Info("container created")
			failedActionStatus := make(chan pb.ActionState)
			go w.captureLogs(ctx, id)
			go waitFailedContainer(ctx, l, cli, id, failedActionStatus)
			err = startContainer(ctx, l, cli, id)
			if err != nil {
				l.Error(errors.Wrap(err, errFailedToRunCmd))
			}
			onTimeoutStatus := <-failedActionStatus
			l.With("status", onTimeoutStatus).Info("action timeout")
		} else {
			if action.OnFailure != nil {
				id, err = w.createContainer(ctx, action.OnFailure, wfID, action)
				if err != nil {
					l.Error(errors.Wrap(err, errFailedToRunCmd))
				}
				l.With("containerID", id, "actionStatus", status.String(), "command", action.GetOnFailure()).Info("container created")
				go w.captureLogs(ctx, id)
				go waitFailedContainer(ctx, l, cli, id, failedActionStatus)
				err = startContainer(ctx, l, cli, id)
				if err != nil {
					l.Error(errors.Wrap(err, errFailedToRunCmd))
				}
				onFailureStatus := <-failedActionStatus
				l.With("status", onFailureStatus).Info("action failed")
			}
		}
		l.Info(infoWaitFinished)
		if err != nil {
			l.Error(errors.Wrap(err, errFailedToWait))
		}
	}
	l.With("status", status).Info("action container exited")
	return status, nil
}

// ProcessWorkflowActions gets all Workflow contexts and processes their actions
func (w *Worker) ProcessWorkflowActions(ctx context.Context, workerID string) error {
	l := w.logger.With("workerID", workerID)

	for {
		res, err := w.client.GetWorkflowContexts(ctx, &pb.WorkflowContextRequest{WorkerId: workerID})
		if err != nil {
			return errors.Wrap(err, errGetWfContext)
		}
		for wfContext, err := res.Recv(); err == nil && wfContext != nil; wfContext, err = res.Recv() {
			wfID := wfContext.GetWorkflowId()
			l = l.With("workflowID", wfID)
			actions, err := w.client.GetWorkflowActions(ctx, &pb.WorkflowActionsRequest{WorkflowId: wfID})
			if err != nil {
				return errors.Wrap(err, errGetWfActions)
			}

			turn := false
			actionIndex := 0
			var nextAction *pb.WorkflowAction
			if wfContext.GetCurrentAction() == "" {
				if actions.GetActionList()[0].GetWorkerId() == workerID {
					actionIndex = 0
					turn = true
				}
			} else {
				switch wfContext.GetCurrentActionState() {
				case pb.ActionState_ACTION_SUCCESS:
					if isLastAction(wfContext, actions) {
						continue
					}
					nextAction = actions.GetActionList()[wfContext.GetCurrentActionIndex()+1]
					actionIndex = int(wfContext.GetCurrentActionIndex()) + 1
				case pb.ActionState_ACTION_FAILED:
					continue
				case pb.ActionState_ACTION_TIMEOUT:
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
				if nextAction.GetWorkerId() == workerID {
					turn = true
				}
			}

			if turn {
				wfDir := dataDir + string(os.PathSeparator) + wfID
				l := l.With("actionName", actions.GetActionList()[actionIndex].GetName(),
					"taskName", actions.GetActionList()[actionIndex].GetTaskName(),
				)
				if _, err := os.Stat(wfDir); os.IsNotExist(err) {
					err := os.Mkdir(wfDir, os.FileMode(0755))
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

					f.Close()
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
				if wfContext.GetCurrentActionState() != pb.ActionState_ACTION_IN_PROGRESS {
					actionStatus := &pb.WorkflowActionStatus{
						WorkflowId:   wfID,
						TaskName:     action.GetTaskName(),
						ActionName:   action.GetName(),
						ActionStatus: pb.ActionState_ACTION_IN_PROGRESS,
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
				getWorkflowData(ctx, l, w.client, workerID, wfID)

				// start executing the action
				start := time.Now()
				status, err := w.execute(ctx, wfID, action)
				elapsed := time.Since(start)

				actionStatus := &pb.WorkflowActionStatus{
					WorkflowId: wfID,
					TaskName:   action.GetTaskName(),
					ActionName: action.GetName(),
					Seconds:    int64(elapsed.Seconds()),
					WorkerId:   action.GetWorkerId(),
				}

				if err != nil || status != pb.ActionState_ACTION_SUCCESS {
					if status == pb.ActionState_ACTION_TIMEOUT {
						actionStatus.ActionStatus = pb.ActionState_ACTION_TIMEOUT
					} else {
						actionStatus.ActionStatus = pb.ActionState_ACTION_FAILED
					}
					l.With("actionStatus", actionStatus.ActionStatus.String())
					l.Error(err)
					if reportErr := w.reportActionStatus(ctx, actionStatus); reportErr != nil {
						exitWithGrpcError(reportErr, l)
					}
					delete(workflowcontexts, wfID)
					return err
				}

				actionStatus.ActionStatus = pb.ActionState_ACTION_SUCCESS
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
					turn = false
					break
				}
				nextAction := actions.GetActionList()[actionIndex+1]
				if nextAction.GetWorkerId() != workerID {
					l.Debug(fmt.Sprintf(msgTurn, nextAction.GetWorkerId()))
					turn = false
				} else {
					actionIndex = actionIndex + 1
				}
			}
		}
		// sleep before asking for new workflows
		<-time.After(w.retryInterval * time.Second)
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
	l := w.logger.With("workflowID", actionStatus.GetWorkflowId,
		"workerID", actionStatus.GetWorkerId(),
		"actionName", actionStatus.GetActionName(),
		"taskName", actionStatus.GetTaskName(),
	)
	var err error
	for r := 1; r <= w.retries; r++ {
		_, err = w.client.ReportActionStatus(ctx, actionStatus)
		if err != nil {
			l.Error(errors.Wrap(err, errReportActionStatus))
			<-time.After(w.retryInterval * time.Second)

			continue
		}
		return nil
	}
	return err
}

func getWorkflowData(ctx context.Context, logger log.Logger, client pb.WorkflowSvcClient, workerID, workflowID string) {
	l := logger.With("workflowID", workflowID,
		"workerID", workerID,
	)
	res, err := client.GetWorkflowData(ctx, &pb.GetWorkflowDataRequest{WorkflowID: workflowID})
	if err != nil {
		l.Error(err)
	}

	if len(res.GetData()) != 0 {
		wfDir := dataDir + string(os.PathSeparator) + workflowID
		f := openDataFile(wfDir, l)
		defer f.Close()

		_, err := f.Write(res.GetData())
		if err != nil {
			l.Error(err)
		}
		h := sha.New()
		workflowDataSHA[workflowID] = base64.StdEncoding.EncodeToString(h.Sum(res.Data))
	}
}

func (w *Worker) updateWorkflowData(ctx context.Context, actionStatus *pb.WorkflowActionStatus) {
	l := w.logger.With("workflowID", actionStatus.GetWorkflowId,
		"workerID", actionStatus.GetWorkerId(),
		"actionName", actionStatus.GetActionName(),
		"taskName", actionStatus.GetTaskName(),
	)

	wfDir := dataDir + string(os.PathSeparator) + actionStatus.GetWorkflowId()
	f := openDataFile(wfDir, l)
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		l.Error(err)
	}

	if isValidDataFile(f, w.maxSize, data, l) {
		h := sha.New()
		if _, ok := workflowDataSHA[actionStatus.GetWorkflowId()]; !ok {
			checksum := base64.StdEncoding.EncodeToString(h.Sum(data))
			workflowDataSHA[actionStatus.GetWorkflowId()] = checksum
			sendUpdate(ctx, w.logger, w.client, actionStatus, data, checksum)
		} else {
			newSHA := base64.StdEncoding.EncodeToString(h.Sum(data))
			if !strings.EqualFold(workflowDataSHA[actionStatus.GetWorkflowId()], newSHA) {
				sendUpdate(ctx, w.logger, w.client, actionStatus, data, newSHA)
			}
		}
	}
}

func sendUpdate(ctx context.Context, logger log.Logger, client pb.WorkflowSvcClient, st *pb.WorkflowActionStatus, data []byte, checksum string) {
	l := logger.With("workflowID", st.GetWorkflowId,
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

	_, err = client.UpdateWorkflowData(ctx, &pb.UpdateWorkflowDataRequest{
		WorkflowID: st.GetWorkflowId(),
		Data:       data,
		Metadata:   metadata,
	})
	if err != nil {
		l.Error(err)
		os.Exit(1)
	}
}

func openDataFile(wfDir string, l log.Logger) *os.File {
	f, err := os.OpenFile(wfDir+string(os.PathSeparator)+dataFile, os.O_RDWR|os.O_CREATE, 0644)
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
