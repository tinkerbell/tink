package main

import (
	"context"
	sha "crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	pb "github.com/packethost/rover/protos/workflow"

	"google.golang.org/grpc/status"
)

const (
	dataFile                 = "/workflow/data"
	maxFileSize              = "MAX_FILE_SIZE" // in bytes
	defaultMaxFileSize int64 = 10485760        //10MB ~= 10485760Bytes
)

var (
	workflowcontexts = map[string]*pb.WorkflowContext{}
	workflowactions  = map[string]*pb.WorkflowActionList{}
	workflowDataSHA  = map[string]string{}
)

func initializeWorker(client pb.WorkflowSvcClient) error {
	workerID := os.Getenv("WORKER_ID")
	if workerID == "" {
		return fmt.Errorf("requried WORKER_NAME")
	}

	ctx := context.Background()
	for {
		err := fetchLatestContext(ctx, client, workerID)
		if err != nil {
			return err
		}

		if allWorkflowsFinished() {
			fmt.Println("All workflows finished")
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
					// send updated workflow data
					updateWorkflowData(ctx, client, wfContext)

					if isLastAction(wfContext, actions) {
						fmt.Printf("Workflow %s completed successfully\n", wfID)
						continue
					}
					nextAction = actions.GetActionList()[wfContext.GetCurrentActionIndex()+1]
					actionIndex = int(wfContext.GetCurrentActionIndex()) + 1
				case pb.ActionState_ACTION_FAILED:
					fmt.Printf("Workflow %s Failed\n", wfID)
					continue
				case pb.ActionState_ACTION_TIMEOUT:
					fmt.Printf("Workflow %s Timeout\n", wfID)
					continue
				default:
					fmt.Printf("Current context %s\n", wfContext)
					nextAction = actions.GetActionList()[wfContext.GetCurrentActionIndex()]
					actionIndex = int(wfContext.GetCurrentActionIndex())
				}
				if nextAction.GetWorkerId() == workerID {
					turn = true
				}
			}

			if turn {
				fmt.Printf("Starting with action %s\n", actions.GetActionList()[actionIndex])
			} else {
				fmt.Printf("Sleep for %d seconds\n", retryInterval)
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
					fmt.Printf("Sent action status %s\n", actionStatus)
				}

				// get workflow data
				getWorkflowData(ctx, client, wfID)

				// start executing the action
				start := time.Now()
				message, status, err := executeAction(ctx, actions.GetActionList()[actionIndex])
				elapsed := time.Since(start)

				actionStatus := &pb.WorkflowActionStatus{
					WorkflowId: wfID,
					TaskName:   action.GetTaskName(),
					ActionName: action.GetName(),
					Seconds:    int64(elapsed.Seconds()),
					WorkerId:   action.GetWorkerId(),
				}

				if err != nil || status != 0 {
					if status == pb.ActionState_ACTION_TIMEOUT {
						fmt.Printf("Action \"%s\" from task \"%s\" timeout\n", action.GetName(), action.GetTaskName())
						actionStatus.ActionStatus = pb.ActionState_ACTION_TIMEOUT
					} else {
						fmt.Printf("Action \"%s\" from task \"%s\" failed\n", action.GetName(), action.GetTaskName())
						actionStatus.ActionStatus = pb.ActionState_ACTION_FAILED
					}
					actionStatus.Message = message
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
				fmt.Printf("Sent action status %s\n", actionStatus)

				if len(actions.GetActionList()) == actionIndex+1 {
					fmt.Printf("Reached to end of workflow\n")
					turn = false
					break
				}
				nextAction := actions.GetActionList()[actionIndex+1]
				if nextAction.GetWorkerId() != workerID {
					fmt.Printf("Different worker has turn %s\n", nextAction.GetWorkerId())
					turn = false
				} else {
					actionIndex = actionIndex + 1
				}
			}
		}
	}
}

func fetchLatestContext(ctx context.Context, client pb.WorkflowSvcClient, workerID string) error {
	fmt.Printf("Fetching latest context for worker %s\n", workerID)
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
		fmt.Println(errStatus.Message())
		fmt.Println(errStatus.Code())
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
			log.Println(err)
			log.Printf("Retrying after %v seconds", retryInterval)
			<-time.After(retryInterval * time.Second)
			continue
		}
		return nil
	}
	return err
}

func getWorkflowData(ctx context.Context, client pb.WorkflowSvcClient, workflowID string) {
	log.Println("Start Getting ephemeral data")
	res, err := client.GetWorkflowData(ctx, &pb.GetWorkflowDataRequest{WorkflowID: workflowID})
	if err != nil {
		log.Fatal(err)
	}

	f := openDataFile()
	defer f.Close()
	if len(res.Data) == 0 {
		f.Write([]byte("{}"))
	} else {
		h := sha.New()
		workflowDataSHA[workflowID] = base64.StdEncoding.EncodeToString(h.Sum(res.Data))
		f.Write(res.Data)
	}
}

func updateWorkflowData(ctx context.Context, client pb.WorkflowSvcClient, workflowCtx *pb.WorkflowContext) {
	log.Println("Starting  updateWorkflowData")
	f := openDataFile()
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	if isValidDataFile(f, data) {
		h := sha.New()
		newSHA := base64.StdEncoding.EncodeToString(h.Sum(data))
		if !strings.EqualFold(workflowDataSHA[workflowCtx.GetWorkflowId()], newSHA) {
			log.Println("Start sending ephemeral data to rover server")
			_, err := client.UpdateWorkflowData(ctx, &pb.UpdateWorkflowDataRequest{
				WorkflowID: workflowCtx.GetWorkflowId(),
				Data:       data,
				ActionName: workflowCtx.GetCurrentAction(),
				WorkerID:   workflowCtx.GetCurrentWorker(),
			})
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func openDataFile() *os.File {
	f, err := os.OpenFile(dataFile, os.O_RDWR|os.O_CREATE, 0644)
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
