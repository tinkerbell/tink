package main

import (
	"context"
	"fmt"
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
		fetchLatestContext(ctx, client, workerID)
		if allWorkflowsFinished() {
			fmt.Println("All workflows finished")
			return nil
		}
		for wfID, wfContext := range workflowcontexts {
			actions, ok := workflowactions[wfID]
			if !ok {
				return fmt.Errorf("Can't find actions for workflow %s", wfID)
			}

			turn := false
			actionIndex := 0
			if wfContext.GetCurrentAction() == "" {
				if actions.GetActionList()[0].GetWorkerId() == workerID {
					actionIndex = 0
					turn = true
				}
			} else {
				if wfContext.GetCurrentActionState() == pb.ActionState_ACTION_SUCCESS && isLastAction(wfContext, actions) {
					fmt.Printf("Worflow %s completed\n", wfID)
					continue
				}
				if wfContext.GetCurrentActionState() != pb.ActionState_ACTION_SUCCESS {
					fmt.Printf("Current context %s\n", wfContext)
					fmt.Println("Sleep for 1 second...")
					time.Sleep(1 * time.Second)
					continue
				}
				nextAction := actions.GetActionList()[wfContext.GetCurrentActionIndex()+1]
				if nextAction.GetWorkerId() == workerID {
					turn = true
					actionIndex = int(wfContext.GetCurrentActionIndex()) + 1
				}
			}

			if turn {
				fmt.Printf("Starting with action %s\n", actions.GetActionList()[actionIndex])
			} else {
				fmt.Println("Sleep for 2 seconds..")
				time.Sleep(2 * time.Second)
			}

			for turn {
				action := actions.GetActionList()[actionIndex]
				actionStatus := &pb.WorkflowActionStatus{
					WorkflowId:   wfID,
					TaskName:     action.GetTaskName(),
					ActionName:   action.GetName(),
					ActionStatus: pb.ActionState_ACTION_IN_PROGRESS,
					Seconds:      0,
					Message:      "Started execution",
				}
				_, err := client.ReportActionStatus(ctx, actionStatus)
				if err != nil {
					exitWithGrpcError(err)
				}
				fmt.Printf("Sent action status %s\n", actionStatus)

				// start executing the action
				executeAction(ctx, actions.GetActionList()[actionIndex])

				actionStatus = &pb.WorkflowActionStatus{
					WorkflowId:   wfID,
					TaskName:     action.GetTaskName(),
					ActionName:   action.GetName(),
					ActionStatus: pb.ActionState_ACTION_SUCCESS,
					Seconds:      2,
					Message:      "Finished execution",
				}
				_, err = client.ReportActionStatus(ctx, &pb.WorkflowActionStatus{
					WorkflowId:   wfID,
					TaskName:     action.GetTaskName(),
					ActionName:   action.GetName(),
					ActionStatus: pb.ActionState_ACTION_SUCCESS,
					Seconds:      2,
					Message:      "Finished execution",
				})
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

func fetchLatestContext(ctx context.Context, client pb.RoverClient, workerID string) {
	fmt.Printf("Fetching latest context for worker %s\n", workerID)
	res, err := client.GetWorkflowContexts(ctx, &pb.WorkflowContextRequest{WorkerId: workerID})
	if err != nil {
		exitWithGrpcError(err)
	}
	for _, wfContext := range res.GetWorkflowContexts() {
		workflowcontexts[wfContext.WorkflowId] = wfContext
		if _, ok := workflowactions[wfContext.WorkflowId]; !ok {
			wfActions, err := client.GetWorkflowActions(ctx, &pb.WorkflowActionsRequest{WorkflowId: wfContext.WorkflowId})
			if err != nil {
				exitWithGrpcError(err)
			}
			workflowactions[wfContext.WorkflowId] = wfActions
		}
	}
}

func allWorkflowsFinished() bool {
	for wfID, wfContext := range workflowcontexts {
		actions := workflowactions[wfID]
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
