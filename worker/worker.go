package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/packethost/rover/protos/rover"
	"google.golang.org/grpc/status"
)

var (
	workflowcontexts = map[string]*pb.WorkflowContext{}
	workflowactions  = map[string]*pb.WorkflowActionList{}
)

func initializeWorker(client pb.RoverClient) error {
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
					if status == 2 {
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

func fetchLatestContext(ctx context.Context, client pb.RoverClient, workerID string) error {
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

func reportActionStatus(ctx context.Context, client pb.RoverClient, actionStatus *pb.WorkflowActionStatus) error {
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
