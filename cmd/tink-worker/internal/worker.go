package internal

import (
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

	"github.com/packethost/pkg/log"
	"github.com/pkg/errors"
	pb "github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/grpc/status"
)

const (
	dataFile                 = "data"
	dataDir                  = "/worker"
	maxFileSize              = "MAX_FILE_SIZE" // in bytes
	defaultMaxFileSize int64 = 10485760        //10MB ~= 10485760Bytes

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

func processWorkflowActions(client pb.WorkflowSvcClient) error {
	workerID := os.Getenv("WORKER_ID")
	if workerID == "" {
		return errors.New("required WORKER_ID")
	}

	ctx := context.Background()
	var err error
	cli, err = initializeDockerClient()
	if err != nil {
		return err
	}
	for {
		l := logger.With("workerID", workerID)
		res, err := client.GetWorkflowContexts(ctx, &pb.WorkflowContextRequest{WorkerId: workerID})
		if err != nil {
			return errors.Wrap(err, errGetWfContext)
		}
		for wfContext, err := res.Recv(); err == nil && wfContext != nil; wfContext, err = res.Recv() {
			wfID := wfContext.GetWorkflowId()
			l = l.With("workflowID", wfID)
			actions, err := client.GetWorkflowActions(ctx, &pb.WorkflowActionsRequest{WorkflowId: wfID})
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

					err := reportActionStatus(ctx, client, actionStatus)
					if err != nil {
						exitWithGrpcError(err, l)
					}
					l.With("duration", strconv.FormatInt(actionStatus.Seconds, 10)).Info("sent action status")
				}

				// get workflow data
				getWorkflowData(ctx, client, workerID, wfID)

				// start executing the action
				start := time.Now()
				status, err := executeAction(ctx, actions.GetActionList()[actionIndex], wfID)
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
					rerr := reportActionStatus(ctx, client, actionStatus)
					if rerr != nil {
						exitWithGrpcError(rerr, l)
					}
					delete(workflowcontexts, wfID)
					return err
				}

				actionStatus.ActionStatus = pb.ActionState_ACTION_SUCCESS
				actionStatus.Message = "finished execution successfully"

				err = reportActionStatus(ctx, client, actionStatus)
				if err != nil {
					exitWithGrpcError(err, l)
				}
				l.Info("sent action status")

				// send workflow data, if updated
				updateWorkflowData(ctx, client, actionStatus)

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
		// sleep for 3 seconds before asking for new workflows
		time.Sleep(retryInterval * time.Second)
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

func reportActionStatus(ctx context.Context, client pb.WorkflowSvcClient, actionStatus *pb.WorkflowActionStatus) error {
	l := logger.With("workflowID", actionStatus.GetWorkflowId,
		"workerID", actionStatus.GetWorkerId(),
		"actionName", actionStatus.GetActionName(),
		"taskName", actionStatus.GetTaskName(),
	)
	var err error
	for r := 1; r <= retries; r++ {
		_, err = client.ReportActionStatus(ctx, actionStatus)
		if err != nil {
			l.Error(errors.Wrap(err, errReportActionStatus))
			l.With("default", retryIntervalDefault).Info("RETRY_INTERVAL not set")
			<-time.After(retryInterval * time.Second)
			continue
		}
		return nil
	}
	return err
}

func getWorkflowData(ctx context.Context, client pb.WorkflowSvcClient, workerID, workflowID string) {
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

func updateWorkflowData(ctx context.Context, client pb.WorkflowSvcClient, actionStatus *pb.WorkflowActionStatus) {
	l := logger.With("workflowID", actionStatus.GetWorkflowId,
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

	if isValidDataFile(f, data, l) {
		h := sha.New()
		if _, ok := workflowDataSHA[actionStatus.GetWorkflowId()]; !ok {
			checksum := base64.StdEncoding.EncodeToString(h.Sum(data))
			workflowDataSHA[actionStatus.GetWorkflowId()] = checksum
			sendUpdate(ctx, client, actionStatus, data, checksum)
		} else {
			newSHA := base64.StdEncoding.EncodeToString(h.Sum(data))
			if !strings.EqualFold(workflowDataSHA[actionStatus.GetWorkflowId()], newSHA) {
				sendUpdate(ctx, client, actionStatus, data, newSHA)
			}
		}
	}
}

func sendUpdate(ctx context.Context, client pb.WorkflowSvcClient, st *pb.WorkflowActionStatus, data []byte, checksum string) {
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

func isValidDataFile(f *os.File, data []byte, l log.Logger) bool {
	var dataMap map[string]interface{}
	err := json.Unmarshal(data, &dataMap)
	if err != nil {
		l.Error(err)
		return false
	}

	stat, err := f.Stat()
	if err != nil {
		logger.Error(err)
		return false
	}

	val := os.Getenv(maxFileSize)
	if val != "" {
		maxSize, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			logger.Error(err)
		}
		return stat.Size() <= maxSize
	}
	return stat.Size() <= defaultMaxFileSize
}
