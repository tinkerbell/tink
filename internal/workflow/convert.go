package workflow

import (
	"fmt"
	"sort"

	"github.com/tinkerbell/tink/api/v1alpha1"
	"github.com/tinkerbell/tink/internal/proto"
)

func ToWorkflowContext(wf *v1alpha1.Workflow) *proto.WorkflowContext {
	if wf == nil {
		return nil
	}
	return &proto.WorkflowContext{
		WorkflowId:           wf.GetName(),
		CurrentWorker:        wf.GetCurrentWorker(),
		CurrentTask:          wf.GetCurrentTask(),
		CurrentAction:        wf.GetCurrentAction(),
		CurrentActionIndex:   int64(wf.GetCurrentActionIndex()),
		CurrentActionState:   proto.State(proto.State_value[string(wf.GetCurrentActionState())]),
		TotalNumberOfActions: int64(wf.GetTotalNumberOfActions()),
	}
}

func YAMLToStatus(wf *Workflow) *v1alpha1.WorkflowStatus {
	if wf == nil {
		return nil
	}
	tasks := []v1alpha1.Task{}
	for _, task := range wf.Tasks {
		actions := []v1alpha1.Action{}
		for _, action := range task.Actions {
			actions = append(actions, v1alpha1.Action{
				Name:        action.Name,
				Image:       action.Image,
				Timeout:     action.Timeout,
				Command:     action.Command,
				Volumes:     action.Volumes,
				Status:      v1alpha1.WorkflowState(proto.State_name[int32(proto.State_STATE_PENDING)]),
				Environment: action.Environment,
				Pid:         action.Pid,
			})
		}
		tasks = append(tasks, v1alpha1.Task{
			Name:        task.Name,
			WorkerAddr:  task.WorkerAddr,
			Volumes:     task.Volumes,
			Environment: task.Environment,
			Actions:     actions,
		})
	}
	return &v1alpha1.WorkflowStatus{
		GlobalTimeout: int64(wf.GlobalTimeout),
		Tasks:         tasks,
	}
}

func ActionListCRDToProto(wf *v1alpha1.Workflow) *proto.WorkflowActionList {
	if wf == nil {
		return nil
	}
	resp := &proto.WorkflowActionList{
		ActionList: []*proto.WorkflowAction{},
	}
	for _, task := range wf.Status.Tasks {
		for _, action := range task.Actions {
			resp.ActionList = append(resp.ActionList, &proto.WorkflowAction{
				TaskName: task.Name,
				Name:     action.Name,
				Image:    action.Image,
				Timeout:  action.Timeout,
				Command:  action.Command,
				WorkerId: task.WorkerAddr,
				Volumes:  append(task.Volumes, action.Volumes...),
				// TODO: (micahhausler) Dedupe task volume targets overridden in the action volumes?
				//   Also not sure how Docker handles nested mounts (ex: "/foo:/foo" and "/bar:/foo/bar")
				Environment: func(env map[string]string) []string {
					resp := []string{}
					merged := map[string]string{}
					for k, v := range env {
						merged[k] = v
					}
					for k, v := range action.Environment {
						merged[k] = v
					}
					for k, v := range merged {
						resp = append(resp, fmt.Sprintf("%s=%s", k, v))
					}
					sort.Strings(resp)
					return resp
				}(task.Environment),
				Pid: action.Pid,
			})
		}
	}
	return resp
}
