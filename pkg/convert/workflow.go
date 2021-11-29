package convert

import (
	"fmt"
	"sort"

	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	protoworkflow "github.com/tinkerbell/tink/protos/workflow"
	"github.com/tinkerbell/tink/workflow"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func WorkflowYAMLToStatus(wf *workflow.Workflow) *v1alpha1.WorkflowStatus {
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
				Status:      protoworkflow.State_name[int32(protoworkflow.State_STATE_PENDING)],
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

func WorkflowCRDToProto(w *v1alpha1.Workflow) *protoworkflow.Workflow {
	if w == nil {
		return nil
	}
	v, ok := protoworkflow.State_value[w.Status.State]
	state := protoworkflow.State(v)
	if !ok {
		state = protoworkflow.State_STATE_PENDING
	}
	return &protoworkflow.Workflow{
		Id:        w.TinkID(),
		Template:  w.Spec.TemplateRef,
		State:     state,
		CreatedAt: timestamppb.New(w.CreationTimestamp.Time),
		DeletedAt: metav1ToTimestamppb(w.DeletionTimestamp),
	}
}

func WorkflowActionListCRDToProto(wf *v1alpha1.Workflow) *protoworkflow.WorkflowActionList {
	if wf == nil {
		return nil
	}
	resp := &protoworkflow.WorkflowActionList{
		ActionList: []*protoworkflow.WorkflowAction{},
	}
	for _, task := range wf.Status.Tasks {
		for _, action := range task.Actions {
			resp.ActionList = append(resp.ActionList, &protoworkflow.WorkflowAction{
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

func WorkflowProtoToCRD(w *protoworkflow.Workflow) *v1alpha1.Workflow {
	if w == nil {
		return nil
	}
	resp := &v1alpha1.Workflow{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Workflow",
			APIVersion: "tinkerbell.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				v1alpha1.WorkflowIDAnnotation: w.Id,
			},
			CreationTimestamp: metav1.NewTime(w.CreatedAt.AsTime()),
		},
		Spec:   v1alpha1.WorkflowSpec{},
		Status: v1alpha1.WorkflowStatus{},
	}

	if v, ok := protoworkflow.State_name[int32(w.State)]; ok {
		resp.Status.State = v
	}
	return resp
}
