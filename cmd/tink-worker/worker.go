package main

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

	"github.com/sirupsen/logrus"
	pb "github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/grpc/status"
)

const (
	dataFile                 = "data"
	dataDir                  = "/worker"
	maxFileSize              = "MAX_FILE_SIZE" // in bytes
	defaultMaxFileSize int64 = 10485760        //10MB ~= 10485760Bytes
)

var (
	workflowcontexts = map[string]*pb.WorkflowContext{}
	workflowactions  = map[string]*pb.WorkflowActionList{}
	workflowDataSHA  = map[string]string{}
)

// WorkflowMetadata is the metadata related to workflow data
type WorkflowMetadata struct {
	WorkerID  string    `json:"worker-id"`
	Action    string    `json:"action-name"`
	Task      string    `json:"task-name"`
	UpdatedAt time.Time `json:"updated-at"`
	SHA       string    `json:"sha256"`
}

func processWorkflowActions(client pb.WorkflowSvcClient) error {
	workerID := os.Getenv("WORKER_ID")
	if workerID == "" {
		return fmt.Errorf("requried WORKER_NAME")
	}
	log = logger.WithField("worker_id", workerID)
	ctx := context.Background()
	for {
		err := fetchLatestContext(ctx, client, workerID)
		if err != nil {
			return err
		}

		if allWorkflowsFinished() {
			log.Infoln("All workflows finished")
			return nil
		}

		cli, err = initializeDockerClient()
		if err != nil {
			return err
		}

		for wfID, wfContext := range workflowcontexts {
			actions, ok := workflowactions[wfID]
			if !ok {
				return fmt.Errorf("Can't find actions for workflow %s", wfID)
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
						log.Infof("Workflow %s completed successfully\n", wfID)
						continue
					}
					nextAction = actions.GetActionList()[wfContext.GetCurrentActionIndex()+1]
					actionIndex = int(wfContext.GetCurrentActionIndex()) + 1
				case pb.ActionState_ACTION_FAILED:
					log.Infof("Workflow %s Failed\n", wfID)
					continue
				case pb.ActionState_ACTION_TIMEOUT:
					log.Infof("Workflow %s Timeout\n", wfID)
					continue
				default:
					log.Infof("Current context %s\n", wfContext)
					nextAction = actions.GetActionList()[wfContext.GetCurrentActionIndex()]
					actionIndex = int(wfContext.GetCurrentActionIndex())
				}
				if nextAction.GetWorkerId() == workerID {
					turn = true
				}
			}

			if turn {
				wfDir := dataDir + string(os.PathSeparator) + wfID
				if _, err := os.Stat(wfDir); os.IsNotExist(err) {
					err := os.Mkdir(wfDir, os.FileMode(0755))
					if err != nil {
						log.Fatal(err)
					}

					f := openDataFile(wfDir)
					_, err = f.Write([]byte("{}"))
					if err != nil {
						log.Fatal(err)
					}

					f.Close()
					if err != nil {
						log.Fatal(err)
					}
				}
				log.Printf("Starting with action %s\n", actions.GetActionList()[actionIndex])
			} else {
				log.Infof("Sleep for %d seconds\n", retryInterval)
				time.Sleep(retryInterval)
			}

			for turn {
				action := actions.GetActionList()[actionIndex]
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
						exitWithGrpcError(err)
					}
					log.WithField("action_name", actionStatus.ActionName).Infoln("Sent action status ", actionStatus.ActionStatus)
					log.Debugf("Sent action status %s\n", actionStatus)
				}

				// get workflow data
				getWorkflowData(ctx, client, wfID)

				// start executing the action
				start := time.Now()
				_, status, err := executeAction(ctx, actions.GetActionList()[actionIndex], wfID)
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
						log.WithFields(logrus.Fields{"action": action.GetName(), "Task": action.GetTaskName()}).Errorln("Action timed out")
						actionStatus.ActionStatus = pb.ActionState_ACTION_TIMEOUT
						actionStatus.Message = "Action Timed out"
					} else {
						log.WithFields(logrus.Fields{"action": action.GetName(), "Task": action.GetTaskName()}).Errorln("Action Failed")
						actionStatus.ActionStatus = pb.ActionState_ACTION_FAILED
						actionStatus.Message = "Action Failed"
					}
					rerr := reportActionStatus(ctx, client, actionStatus)
					if rerr != nil {
						exitWithGrpcError(rerr)
					}
					return err
				}

				actionStatus.ActionStatus = pb.ActionState_ACTION_SUCCESS
				actionStatus.Message = "Finished Execution Successfully"

				err = reportActionStatus(ctx, client, actionStatus)
				if err != nil {
					exitWithGrpcError(err)
				}
				log.Infof("Sent action status %s\n", actionStatus)

				// send workflow data, if updated
				updateWorkflowData(ctx, client, actionStatus)

				if len(actions.GetActionList()) == actionIndex+1 {
					log.Infoln("Reached to end of workflow")
					turn = false
					break
				}
				nextAction := actions.GetActionList()[actionIndex+1]
				if nextAction.GetWorkerId() != workerID {
					log.Debugf("Different worker has turn %s\n", nextAction.GetWorkerId())
					turn = false
				} else {
					actionIndex = actionIndex + 1
				}
			}
		}
	}
}

func fetchLatestContext(ctx context.Context, client pb.WorkflowSvcClient, workerID string) error {
	log.Infof("Fetching latest context for worker %s\n", workerID)
	res, err := client.GetWorkflowContexts(ctx, &pb.WorkflowContextRequest{WorkerId: workerID})
	if err != nil {
		return err
	}
	for _, wfContext := range res.GetWorkflowContexts() {
		workflowcontexts[wfContext.WorkflowId] = wfContext
		if _, ok := workflowactions[wfContext.WorkflowId]; !ok {
			wfActions, err := client.GetWorkflowActions(ctx, &pb.WorkflowActionsRequest{WorkflowId: wfContext.WorkflowId})
			if err != nil {
				return err
			}
			workflowactions[wfContext.WorkflowId] = wfActions
		}
	}
	return nil
}

func allWorkflowsFinished() bool {
	for wfID, wfContext := range workflowcontexts {
		actions := workflowactions[wfID]
		if wfContext.GetCurrentActionState() == pb.ActionState_ACTION_FAILED || wfContext.GetCurrentActionState() == pb.ActionState_ACTION_TIMEOUT {
			continue
		}
		if !(wfContext.GetCurrentActionState() == pb.ActionState_ACTION_SUCCESS && isLastAction(wfContext, actions)) {
			return false
		}
	}
	return true
}

func exitWithGrpcError(err error) {
	if err != nil {
		errStatus, _ := status.FromError(err)
		log.WithField("Error code : ", errStatus.Code()).Errorln(errStatus.Message())
		os.Exit(1)
	}
}

func isLastAction(wfContext *pb.WorkflowContext, actions *pb.WorkflowActionList) bool {
	return int(wfContext.GetCurrentActionIndex()) == len(actions.GetActionList())-1
}

func reportActionStatus(ctx context.Context, client pb.WorkflowSvcClient, actionStatus *pb.WorkflowActionStatus) error {
	var err error
	for r := 1; r <= retries; r++ {
		_, err = client.ReportActionStatus(ctx, actionStatus)
		if err != nil {
			log.Println("Report action status to server failed as : ", err)
			log.Printf("Retrying after %v seconds", retryInterval)
			<-time.After(retryInterval * time.Second)
			continue
		}
		return nil
	}
	return err
}

func getWorkflowData(ctx context.Context, client pb.WorkflowSvcClient, workflowID string) {
	res, err := client.GetWorkflowData(ctx, &pb.GetWorkflowDataRequest{WorkflowID: workflowID})
	if err != nil {
		log.Fatal(err)
	}

	if len(res.GetData()) != 0 {
		log.Debugf("Data received: %x", res.GetData())
		wfDir := dataDir + string(os.PathSeparator) + workflowID
		f := openDataFile(wfDir)
		defer f.Close()

		_, err := f.Write(res.GetData())
		if err != nil {
			log.Fatal(err)
		}
		h := sha.New()
		workflowDataSHA[workflowID] = base64.StdEncoding.EncodeToString(h.Sum(res.Data))
	}
}

func updateWorkflowData(ctx context.Context, client pb.WorkflowSvcClient, actionStatus *pb.WorkflowActionStatus) {
	wfDir := dataDir + string(os.PathSeparator) + actionStatus.GetWorkflowId()
	f := openDataFile(wfDir)
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	if isValidDataFile(f, data) {
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
	meta := WorkflowMetadata{
		WorkerID:  st.GetWorkerId(),
		Action:    st.GetActionName(),
		Task:      st.GetTaskName(),
		UpdatedAt: time.Now(),
		SHA:       checksum,
	}
	metadata, err := json.Marshal(meta)
	if err != nil {
		log.Fatal(err)
	}

	log.Debugf("Sending updated data: %v\n", string(data))
	_, err = client.UpdateWorkflowData(ctx, &pb.UpdateWorkflowDataRequest{
		WorkflowID: st.GetWorkflowId(),
		Data:       data,
		Metadata:   metadata,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func openDataFile(wfDir string) *os.File {
	f, err := os.OpenFile(wfDir+string(os.PathSeparator)+dataFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return f
}

func isValidDataFile(f *os.File, data []byte) bool {
	var dataMap map[string]interface{}
	err := json.Unmarshal(data, &dataMap)
	if err != nil {
		log.Print(err)
		return false
	}

	stat, err := f.Stat()
	if err != nil {
		log.Print(err)
		return false
	}

	val := os.Getenv(maxFileSize)
	if val != "" {
		maxSize, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			log.Print(err)
		}
		return stat.Size() <= maxSize
	}
	return stat.Size() <= defaultMaxFileSize
}
