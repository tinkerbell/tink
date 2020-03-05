package framework

import (
	"context"

	"github.com/packethost/rover/client"
	"github.com/packethost/rover/protos/workflow"
)

// CreateWorkflow : create workflow
func CreateWorkflow(template string, target string) (string, error) {
	req := workflow.CreateRequest{Template: template, Target: target}
	res, err := client.WorkflowClient.CreateWorkflow(context.Background(), &req)
	if err != nil {
		return "", err
	}
	return res.Id, nil
}

// GetCurrentStatus : get the current status of workflow from server
func GetCurrentStatus(ctx context.Context, wfID string, status chan workflow.ActionState) {
	req := workflow.GetRequest{Id: wfID}
	wf, err := client.WorkflowClient.GetWorkflowContext(ctx, &req)
	if err != nil {
		log.Errorln("This is in Getting status ERROR: ", err)
		status <- workflow.ActionState_ACTION_FAILED
	}
	if wf.CurrentActionState == workflow.ActionState_ACTION_FAILED {
		status <- workflow.ActionState_ACTION_FAILED
	} else if wf.CurrentActionState == workflow.ActionState_ACTION_TIMEOUT {
		status <- workflow.ActionState_ACTION_TIMEOUT
	}
	currProgress := calWorkflowProgress(wf.CurrentActionIndex, wf.TotalNumberOfActions, wf.CurrentActionState)
	if currProgress == 100 && wf.CurrentActionState == workflow.ActionState_ACTION_SUCCESS {
		status <- workflow.ActionState_ACTION_SUCCESS
	}
}

func calWorkflowProgress(cur int64, total int64, state workflow.ActionState) int64 {
	if total == 0 || (cur == 0 && state != workflow.ActionState_ACTION_SUCCESS) {
		return 0
	}
	var taskCompleted int64
	if state == workflow.ActionState_ACTION_SUCCESS {
		taskCompleted = cur + 1
	} else {
		taskCompleted = cur
	}
	progress := (taskCompleted * 100) / total
	return progress
}
