package server

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tinkerbell/tink/api/v1alpha1"
	"github.com/tinkerbell/tink/internal/proto"
	"github.com/tinkerbell/tink/internal/workflow"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errInvalidWorkflowID     = "invalid workflow id"
	errInvalidTaskName       = "invalid task name"
	errInvalidActionName     = "invalid action name"
	errInvalidTaskReported   = "reported task name does not match the current action details"
	errInvalidActionReported = "reported action name does not match the current action details"
)

func getWorkflowContext(wf v1alpha1.Workflow) *proto.WorkflowContext {
	return &proto.WorkflowContext{
		WorkflowId:           wf.Name,
		CurrentWorker:        wf.GetCurrentWorker(),
		CurrentTask:          wf.GetCurrentTask(),
		CurrentAction:        wf.GetCurrentAction(),
		CurrentActionIndex:   int64(wf.GetCurrentActionIndex()),
		CurrentActionState:   proto.State(proto.State_value[string(wf.GetCurrentActionState())]),
		TotalNumberOfActions: int64(wf.GetTotalNumberOfActions()),
	}
}

func (s *KubernetesBackedServer) getCurrentAssignedNonTerminalWorkflowsForWorker(ctx context.Context, workerID string) ([]v1alpha1.Workflow, error) {
	stored := &v1alpha1.WorkflowList{}
	err := s.ClientFunc().List(ctx, stored, &client.MatchingFields{
		workflowByNonTerminalState: workerID,
	})
	if err != nil {
		return nil, err
	}
	wfs := []v1alpha1.Workflow{}
	for _, wf := range stored.Items {
		// If the current assigned or running action is assigned to the requested worker, include it
		if wf.Status.Tasks[wf.GetCurrentTaskIndex()].WorkerAddr == workerID {
			wfs = append(wfs, wf)
		}
	}
	return wfs, nil
}

func (s *KubernetesBackedServer) getWorkflowByName(ctx context.Context, workflowID, namespace string) (*v1alpha1.Workflow, error) {
	wflw := &v1alpha1.Workflow{}
	err := s.ClientFunc().Get(ctx, types.NamespacedName{Name: workflowID, Namespace: namespace}, wflw)
	if err != nil {
		s.logger.Error(err, "get client", "workflow", workflowID)
		return nil, err
	}
	return wflw, nil
}

// The following APIs are used by the worker.

func (s *KubernetesBackedServer) GetWorkflowContexts(req *proto.WorkflowContextRequest, stream proto.WorkflowService_GetWorkflowContextsServer) error {
	if req.GetWorkerId() == "" {
		return status.Errorf(codes.InvalidArgument, errInvalidWorkflowID)
	}
	wflows, err := s.getCurrentAssignedNonTerminalWorkflowsForWorker(stream.Context(), req.WorkerId)
	if err != nil {
		return err
	}
	for _, wf := range wflows {
		if err := stream.Send(getWorkflowContext(wf)); err != nil {
			return err
		}
	}
	return nil
}

func (s *KubernetesBackedServer) GetWorkflowActions(ctx context.Context, req *proto.WorkflowActionsRequest) (*proto.WorkflowActionList, error) {
	wfID := req.GetWorkflowId()
	if wfID == "" {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidWorkflowID)
	}
	wf, err := s.getWorkflowByName(ctx, wfID, s.namespace)
	if err != nil {
		return nil, err
	}
	return workflow.ActionListCRDToProto(wf), nil
}

// Modifies a workflow for a given workflowContext.
func (s *KubernetesBackedServer) modifyWorkflowState(wf *v1alpha1.Workflow, wfContext *proto.WorkflowContext) error {
	if wf == nil {
		return errors.New("no workflow provided")
	}
	if wfContext == nil {
		return errors.New("no workflow context provided")
	}
	var (
		taskIndex   = -1
		actionIndex = -1
	)

	for ti, task := range wf.Status.Tasks {
		if wfContext.CurrentTask == task.Name {
			taskIndex = ti
			for ai, action := range task.Actions {
				if action.Name == wfContext.CurrentAction && wfContext.CurrentActionIndex == int64(ai) {
					actionIndex = ai
					goto cont
				}
			}
		}
	}
cont:

	if taskIndex < 0 {
		return errors.New("task not found")
	}
	if actionIndex < 0 {
		return errors.New("action not found")
	}
	wf.Status.Tasks[taskIndex].Actions[actionIndex].Status = v1alpha1.WorkflowState(proto.State_name[int32(wfContext.CurrentActionState)])

	switch wfContext.CurrentActionState {
	case proto.State_STATE_RUNNING:
		// Workflow is running, so set the start time to now
		wf.Status.State = v1alpha1.WorkflowState(proto.State_name[int32(wfContext.CurrentActionState)])
		wf.Status.Tasks[taskIndex].Actions[actionIndex].StartedAt = func() *metav1.Time {
			t := metav1.NewTime(s.nowFunc())
			return &t
		}()
	case proto.State_STATE_FAILED, proto.State_STATE_TIMEOUT:
		// Handle terminal statuses by updating the workflow state and time
		wf.Status.State = v1alpha1.WorkflowState(proto.State_name[int32(wfContext.CurrentActionState)])
		if wf.Status.Tasks[taskIndex].Actions[actionIndex].StartedAt != nil {
			wf.Status.Tasks[taskIndex].Actions[actionIndex].Seconds = int64(s.nowFunc().Sub(wf.Status.Tasks[taskIndex].Actions[actionIndex].StartedAt.Time).Seconds())
		}
	case proto.State_STATE_SUCCESS:
		// Handle a success by marking the task as complete
		if wf.Status.Tasks[taskIndex].Actions[actionIndex].StartedAt != nil {
			wf.Status.Tasks[taskIndex].Actions[actionIndex].Seconds = int64(s.nowFunc().Sub(wf.Status.Tasks[taskIndex].Actions[actionIndex].StartedAt.Time).Seconds())
		}
		// Mark success on last action success
		if wfContext.CurrentActionIndex+1 == wfContext.TotalNumberOfActions {
			wf.Status.State = v1alpha1.WorkflowState(proto.State_name[int32(wfContext.CurrentActionState)])
		}
	case proto.State_STATE_PENDING:
		// This is probably a client bug?
		return errors.New("no update requested")
	}
	return nil
}

func validateActionStatusRequest(req *proto.WorkflowActionStatus) error {
	if req.GetWorkflowId() == "" {
		return status.Errorf(codes.InvalidArgument, errInvalidWorkflowID)
	}
	if req.GetTaskName() == "" {
		return status.Errorf(codes.InvalidArgument, errInvalidTaskName)
	}
	if req.GetActionName() == "" {
		return status.Errorf(codes.InvalidArgument, errInvalidActionName)
	}
	return nil
}

func getWorkflowContextForRequest(req *proto.WorkflowActionStatus, wf *v1alpha1.Workflow) *proto.WorkflowContext {
	wfContext := getWorkflowContext(*wf)
	wfContext.CurrentWorker = req.GetWorkerId()
	wfContext.CurrentTask = req.GetTaskName()
	wfContext.CurrentActionState = req.GetActionStatus()
	wfContext.CurrentActionIndex = int64(wf.GetCurrentActionIndex())
	return wfContext
}

func (s *KubernetesBackedServer) ReportActionStatus(ctx context.Context, req *proto.WorkflowActionStatus) (*proto.Empty, error) {
	err := validateActionStatusRequest(req)
	if err != nil {
		return nil, err
	}
	wfID := req.GetWorkflowId()
	l := s.logger.WithValues("actionName", req.GetActionName(), "status", req.GetActionStatus(), "workflowID", req.GetWorkflowId(), "taskName", req.GetTaskName(), "worker", req.WorkerId)

	wf, err := s.getWorkflowByName(ctx, wfID, s.namespace)
	if err != nil {
		l.Error(err, "get workflow")
		return nil, status.Errorf(codes.InvalidArgument, errInvalidWorkflowID)
	}
	if req.GetTaskName() != wf.GetCurrentTask() {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidTaskReported)
	}
	if req.GetActionName() != wf.GetCurrentAction() {
		return nil, status.Errorf(codes.InvalidArgument, errInvalidActionReported)
	}

	wfContext := getWorkflowContextForRequest(req, wf)
	err = s.modifyWorkflowState(wf, wfContext)
	if err != nil {
		l.Error(err, "modify workflow state")
		return nil, status.Errorf(codes.InvalidArgument, errInvalidWorkflowID)
	}
	l.Info("updating workflow in Kubernetes")
	err = s.ClientFunc().Status().Update(ctx, wf)
	if err != nil {
		l.Error(err, "applying update to workflow")
		return nil, status.Errorf(codes.InvalidArgument, errInvalidWorkflowID)
	}
	return &proto.Empty{}, nil
}
