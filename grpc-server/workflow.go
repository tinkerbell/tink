package grpcserver

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tinkerbell/tink/db"
	"github.com/tinkerbell/tink/metrics"
	"github.com/tinkerbell/tink/protos/workflow"
	wkf "github.com/tinkerbell/tink/workflow"
)

const errFailedToGetTemplate = "failed to get template with ID %s"

// CreateWorkflow implements workflow.CreateWorkflow
func (s *server) CreateWorkflow(ctx context.Context, in *workflow.CreateRequest) (*workflow.CreateResponse, error) {
	s.logger.Info("createworkflow")
	labels := prometheus.Labels{"method": "CreateWorkflow", "op": ""}
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	const msg = "creating a new workflow"
	labels["op"] = "createworkflow"
	id, err := uuid.NewUUID()
	if err != nil {
		return &workflow.CreateResponse{}, err
	}

	metrics.CacheTotals.With(labels).Inc()
	timer := prometheus.NewTimer(metrics.CacheDuration.With(labels))
	defer timer.ObserveDuration()

	s.logger.Info(msg)
	fields := map[string]string{
		"id": in.GetTemplate(),
	}
	_, _, templateData, err := s.db.GetTemplate(ctx, fields, false)
	if err != nil {
		return &workflow.CreateResponse{}, errors.Wrapf(err, errFailedToGetTemplate, in.GetTemplate())
	}
	data, err := wkf.RenderTemplate(in.GetTemplate(), templateData, []byte(in.Hardware))

	if err != nil {
		metrics.CacheErrors.With(labels).Inc()
		s.logger.Error(err)
		return &workflow.CreateResponse{}, err
	}

	wf := db.Workflow{
		ID:       id.String(),
		Template: in.Template,
		Hardware: in.Hardware,
		State:    workflow.State_value[workflow.State_STATE_PENDING.String()],
	}
	err = s.db.CreateWorkflow(ctx, wf, data, id)
	if err != nil {
		metrics.CacheErrors.With(labels).Inc()
		l := s.logger
		if pqErr := db.Error(err); pqErr != nil {
			l = l.With("detail", pqErr.Detail, "where", pqErr.Where)
		}
		l.Error(err)
		return &workflow.CreateResponse{}, err
	}

	l := s.logger.With("workflowID", id.String())
	l.Info("done " + msg)
	return &workflow.CreateResponse{Id: id.String()}, err
}

// GetWorkflow implements workflow.GetWorkflow
func (s *server) GetWorkflow(ctx context.Context, in *workflow.GetRequest) (*workflow.Workflow, error) {
	s.logger.Info("getworkflow")
	labels := prometheus.Labels{"method": "GetWorkflow", "op": ""}
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	const msg = "getting a workflow"
	labels["op"] = "get"

	metrics.CacheTotals.With(labels).Inc()
	timer := prometheus.NewTimer(metrics.CacheDuration.With(labels))
	defer timer.ObserveDuration()

	s.logger.Info(msg)
	w, err := s.db.GetWorkflow(ctx, in.Id)
	if err != nil {
		metrics.CacheErrors.With(labels).Inc()
		l := s.logger
		if pqErr := db.Error(err); pqErr != nil {
			l = l.With("detail", pqErr.Detail, "where", pqErr.Where)
		}
		l.Error(err)
		return &workflow.Workflow{}, err
	}
	fields := map[string]string{
		"id": w.Template,
	}
	_, _, templateData, err := s.db.GetTemplate(ctx, fields, true)
	if err != nil {
		return &workflow.Workflow{}, errors.Wrapf(err, errFailedToGetTemplate, w.Template)
	}
	data, err := wkf.RenderTemplate(w.Template, templateData, []byte(w.Hardware))
	if err != nil {
		return &workflow.Workflow{}, err
	}

	wf := &workflow.Workflow{
		Id:       w.ID,
		Template: w.Template,
		Hardware: w.Hardware,
		State:    getWorkflowState(s.db, ctx, in.Id),
		Data:     data,
	}
	l := s.logger.With("workflowID", w.ID)
	l.Info("done " + msg)
	return wf, err
}

// DeleteWorkflow implements workflow.DeleteWorkflow
func (s *server) DeleteWorkflow(ctx context.Context, in *workflow.GetRequest) (*workflow.Empty, error) {
	s.logger.Info("deleteworkflow")
	labels := prometheus.Labels{"method": "DeleteWorkflow", "op": ""}
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	const msg = "deleting a workflow"
	labels["op"] = "delete"
	l := s.logger.With("workflowID", in.GetId())

	metrics.CacheTotals.With(labels).Inc()
	timer := prometheus.NewTimer(metrics.CacheDuration.With(labels))
	defer timer.ObserveDuration()

	l.Info(msg)
	err := s.db.DeleteWorkflow(ctx, in.Id, workflow.State_value[workflow.State_STATE_RUNNING.String()])
	if err != nil {
		metrics.CacheErrors.With(labels).Inc()
		l := s.logger
		if pqErr := db.Error(err); pqErr != nil {
			l = l.With("detail", pqErr.Detail, "where", pqErr.Where)
		}
		l.Error(err)
	}
	l.Info("done " + msg)
	return &workflow.Empty{}, err
}

// ListWorkflows implements workflow.ListWorkflows
func (s *server) ListWorkflows(_ *workflow.Empty, stream workflow.WorkflowService_ListWorkflowsServer) error {
	s.logger.Info("listworkflows")
	labels := prometheus.Labels{"method": "ListWorkflows", "op": "list"}
	metrics.CacheTotals.With(labels).Inc()
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	s.dbLock.RLock()
	ready := s.dbReady
	s.dbLock.RUnlock()
	if !ready {
		metrics.CacheStalls.With(labels).Inc()
		return errors.New("DB is not ready")
	}

	timer := prometheus.NewTimer(metrics.CacheDuration.With(labels))
	defer timer.ObserveDuration()
	err := s.db.ListWorkflows(func(w db.Workflow) error {
		wf := &workflow.Workflow{
			Id:        w.ID,
			Template:  w.Template,
			Hardware:  w.Hardware,
			CreatedAt: w.CreatedAt,
			UpdatedAt: w.UpdatedAt,
			State:     getWorkflowState(s.db, stream.Context(), w.ID),
		}
		return stream.Send(wf)
	})

	if err != nil {
		metrics.CacheErrors.With(labels).Inc()
		return err
	}

	metrics.CacheHits.With(labels).Inc()
	return nil
}

func (s *server) GetWorkflowContext(ctx context.Context, in *workflow.GetRequest) (*workflow.WorkflowContext, error) {
	s.logger.Info("GetworkflowContext")
	labels := prometheus.Labels{"method": "GetWorkflowContext", "op": ""}
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	const msg = "getting a workflow context"
	labels["op"] = "get"

	metrics.CacheTotals.With(labels).Inc()
	timer := prometheus.NewTimer(metrics.CacheDuration.With(labels))
	defer timer.ObserveDuration()

	s.logger.Info(msg)
	w, err := s.db.GetWorkflowContexts(ctx, in.Id)
	if err != nil {
		metrics.CacheErrors.With(labels).Inc()
		l := s.logger
		if pqErr := db.Error(err); pqErr != nil {
			l = l.With("detail", pqErr.Detail, "where", pqErr.Where)
		}
		l.Error(err)
		return &workflow.WorkflowContext{}, err
	}
	wf := &workflow.WorkflowContext{
		WorkflowId:           w.WorkflowId,
		CurrentWorker:        w.CurrentWorker,
		CurrentTask:          w.CurrentTask,
		CurrentAction:        w.CurrentAction,
		CurrentActionIndex:   w.CurrentActionIndex,
		CurrentActionState:   workflow.State(w.CurrentActionState),
		TotalNumberOfActions: w.TotalNumberOfActions,
	}
	l := s.logger.With(
		"workflowID", wf.GetWorkflowId(),
		"currentWorker", wf.GetCurrentWorker(),
		"currentTask", wf.GetCurrentTask(),
		"currentAction", wf.GetCurrentAction(),
		"currentActionIndex", strconv.FormatInt(wf.GetCurrentActionIndex(), 10),
		"currentActionState", wf.GetCurrentActionState(),
		"totalNumberOfActions", wf.GetTotalNumberOfActions(),
	)
	l.Info("done " + msg)
	return wf, err
}

// ShowWorflowevents  implements workflow.ShowWorflowEvents
func (s *server) ShowWorkflowEvents(req *workflow.GetRequest, stream workflow.WorkflowService_ShowWorkflowEventsServer) error {
	s.logger.Info("List workflows Events")
	labels := prometheus.Labels{"method": "ShowWorkflowEvents", "op": "list"}
	metrics.CacheTotals.With(labels).Inc()
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	s.dbLock.RLock()
	ready := s.dbReady
	s.dbLock.RUnlock()
	if !ready {
		metrics.CacheStalls.With(labels).Inc()
		return errors.New("DB is not ready")
	}

	timer := prometheus.NewTimer(metrics.CacheDuration.With(labels))
	defer timer.ObserveDuration()
	err := s.db.ShowWorkflowEvents(req.Id, func(w *workflow.WorkflowActionStatus) error {
		wfs := &workflow.WorkflowActionStatus{
			WorkerId:     w.WorkerId,
			TaskName:     w.TaskName,
			ActionName:   w.ActionName,
			ActionStatus: workflow.State(w.ActionStatus),
			Seconds:      w.Seconds,
			Message:      w.Message,
			CreatedAt:    w.CreatedAt,
		}
		return stream.Send(wfs)
	})

	if err != nil {
		metrics.CacheErrors.With(labels).Inc()
		return err
	}
	s.logger.Info("done listing workflows events")
	metrics.CacheHits.With(labels).Inc()
	return nil
}

// This function will provide the workflow state on the basis of the state of the actions
// For e.g. : If an action has Failed or Timeout then the workflow state will also be
// considered as Failed/Timeout. And If an action is successful then the workflow state
// will be considered as Running until the last action of the workflow is executed successfully.

func getWorkflowState(db db.Database, ctx context.Context, id string) workflow.State {
	wfCtx, _ := db.GetWorkflowContexts(ctx, id)
	if wfCtx.CurrentActionState != workflow.State_STATE_SUCCESS {
		return wfCtx.CurrentActionState
	} else {
		if wfCtx.GetCurrentActionIndex() == wfCtx.GetTotalNumberOfActions()-1 {
			return workflow.State_STATE_SUCCESS
		} else {
			return workflow.State_STATE_RUNNING
		}
	}
}
