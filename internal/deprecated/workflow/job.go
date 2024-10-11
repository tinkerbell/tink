package workflow

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	rufio "github.com/tinkerbell/rufio/api/v1alpha1"
	"github.com/tinkerbell/tink/api/v1alpha1"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type jobName string

const (
	jobNameNetboot  jobName = "netboot"
	jobNameISOMount jobName = "iso-mount"
	jobNameISOEject jobName = "iso-eject"
)

func (j jobName) String() string {
	return string(j)
}

// this function will update the Workflow status.
func (s *state) handleJob(ctx context.Context, actions []rufio.Action, name jobName) (reconcile.Result, error) {
	tracer := otel.Tracer("handleJob")
	var span trace.Span
	ctx, span = tracer.Start(ctx, "handleJob")
	defer span.End()
	// there are 3 phases. 1. Clean up existing 2. Create new 3. Track status
	// 1. clean up existing job if it wasn't already deleted
	if j := s.workflow.Status.BootOptions.Jobs[name.String()]; !j.ExistingJobDeleted {
		s.logger.InfoContext(ctx, "deleting existing job", "name", name)
		result, err := s.deleteExisting(ctx, name)
		if err != nil {
			return result, err
		}

		return result, nil
	}

	// 2. create a new job
	if uid := s.workflow.Status.BootOptions.Jobs[name.String()].UID; uid == "" {
		s.logger.InfoContext(ctx, "no uid found for job", "name", name)
		result, err := s.createJob(ctx, actions, name)
		if err != nil {
			s.workflow.Status.SetCondition(v1alpha1.WorkflowCondition{
				Type:    v1alpha1.NetbootJobSetupFailed,
				Status:  metav1.ConditionTrue,
				Reason:  "Error",
				Message: fmt.Sprintf("error creating job: %v", err),
				Time:    &metav1.Time{Time: metav1.Now().UTC()},
			})
			return result, err
		}
		s.workflow.Status.SetCondition(v1alpha1.WorkflowCondition{
			Type:    v1alpha1.NetbootJobSetupComplete,
			Status:  metav1.ConditionTrue,
			Reason:  "Created",
			Message: "job created",
			Time:    &metav1.Time{Time: metav1.Now().UTC()},
		})
		return result, nil
	}

	// 3. track status
	if !s.workflow.Status.BootOptions.Jobs[name.String()].Complete {
		s.logger.InfoContext(ctx, "tracking job", "name", name)
		// track status
		r, tState, err := s.trackRunningJob(ctx, name)
		if err != nil {
			s.workflow.Status.SetCondition(v1alpha1.WorkflowCondition{
				Type:    v1alpha1.NetbootJobFailed,
				Status:  metav1.ConditionTrue,
				Reason:  "Error",
				Message: err.Error(),
				Time:    &metav1.Time{Time: metav1.Now().UTC()},
			})
			return r, err
		}
		if tState == trackedStateComplete {
			s.workflow.Status.SetCondition(v1alpha1.WorkflowCondition{
				Type:    v1alpha1.NetbootJobComplete,
				Status:  metav1.ConditionTrue,
				Reason:  "Complete",
				Message: "job completed",
				Time:    &metav1.Time{Time: metav1.Now().UTC()},
			})
		}
		return r, nil
	}

	return reconcile.Result{Requeue: true}, nil
}

func (s *state) deleteExisting(ctx context.Context, name jobName) (reconcile.Result, error) {
	existingJob := &rufio.Job{ObjectMeta: metav1.ObjectMeta{Name: name.String(), Namespace: s.workflow.Namespace}}
	opts := []client.DeleteOption{
		client.GracePeriodSeconds(0),
		client.PropagationPolicy(metav1.DeletePropagationForeground),
	}
	if err := s.client.Delete(ctx, existingJob, opts...); client.IgnoreNotFound(err) != nil {
		return reconcile.Result{}, fmt.Errorf("error deleting job.bmc.tinkerbell.org object: %w", err)
	}

	jStatus := s.workflow.Status.BootOptions.Jobs[name.String()]
	jStatus.ExistingJobDeleted = true
	// if we delete an existing job, we need to remove any uid that was set.
	jStatus.UID = ""
	jStatus.Complete = false
	s.workflow.Status.BootOptions.Jobs[name.String()] = jStatus

	return reconcile.Result{Requeue: true}, nil
}

// This function will update the Workflow status.
func (s *state) createJob(ctx context.Context, actions []rufio.Action, name jobName) (reconcile.Result, error) {
	tracer := otel.Tracer("createJob")
	var span trace.Span
	ctx, span = tracer.Start(ctx, "createJob")
	defer span.End()
	// create a new job
	// The assumption is that the UID is not set. UID checking is not handled here.
	// 1. look up if there's an existing job with the same name, if so update the status with the UID and return
	// 2. if there's no existing job, create a new job, update the status with the UID, and return

	rj := &rufio.Job{}
	if err := s.client.Get(ctx, client.ObjectKey{Name: name.String(), Namespace: s.workflow.Namespace}, rj); err == nil {
		s.logger.InfoContext(ctx, "job already exists", "name", name)
		if !rj.DeletionTimestamp.IsZero() {
			s.logger.InfoContext(ctx, "job is being deleted", "name", name)
			return reconcile.Result{Requeue: true}, nil
		}
		//TODO(jacobweinstock): job exists means that the job name and uid from the status are the same. 
		// get the UID and update the status
		jStatus := s.workflow.Status.BootOptions.Jobs[name.String()]
		jStatus.UID = rj.GetUID()
		s.workflow.Status.BootOptions.Jobs[name.String()] = jStatus

		return reconcile.Result{Requeue: true}, nil
	}

	// create a new job
	if s.hardware == nil {
		return reconcile.Result{}, fmt.Errorf("hardware is nil")
	}
	if s.hardware.Spec.BMCRef == nil {
		return reconcile.Result{}, fmt.Errorf("hardware %q does not have a BMC", s.hardware.Name)
	}

	if err := create(ctx, s.logger, s.client, name.String(), s.hardware, s.workflow.Namespace, actions); err != nil {
		return reconcile.Result{}, fmt.Errorf("error creating job: %w", err)
	}
	s.logger.InfoContext(ctx, "job created", "name", name)

	return reconcile.Result{Requeue: true}, nil
}

type trackedState string

var (
	trackedStateComplete trackedState = "complete"
	trackedStateRunning  trackedState = "running"
	trackedStateError    trackedState = "error"
	trackedStateFailed   trackedState = "failed"
)

// This function will update the Workflow status.
func (s *state) trackRunningJob(ctx context.Context, name jobName) (reconcile.Result, trackedState, error) {
	tracer := otel.Tracer("trackRunningJob")
	var span trace.Span
	ctx, span = tracer.Start(ctx, "trackRunningJob")
	defer span.End()
	// track status
	// get the job
	rj := &rufio.Job{}
	if err := s.client.Get(ctx, client.ObjectKey{Name: name.String(), Namespace: s.workflow.Namespace}, rj); err != nil {
		return reconcile.Result{}, trackedStateError, fmt.Errorf("error getting job: %w", err)
	}
	if rj.HasCondition(rufio.JobFailed, rufio.ConditionTrue) {
		s.logger.InfoContext(ctx, "job failed", "name", name)
		// job failed
		return reconcile.Result{}, trackedStateFailed, fmt.Errorf("job failed")
	}
	if rj.HasCondition(rufio.JobCompleted, rufio.ConditionTrue) {
		s.logger.InfoContext(ctx, "job completed", "name", name)
		// job completed
		jStatus := s.workflow.Status.BootOptions.Jobs[name.String()]
		jStatus.Complete = true
		s.workflow.Status.BootOptions.Jobs[name.String()] = jStatus

		return reconcile.Result{}, trackedStateComplete, nil
	}
	// still running
	s.logger.InfoContext(ctx, "job still running", "name", name)
	time.Sleep(s.backoff.NextBackOff())
	return reconcile.Result{Requeue: true}, trackedStateRunning, nil
}

func create(ctx context.Context, logger *slog.Logger, cc client.Client, name string, hw *v1alpha1.Hardware, ns string, tasks []rufio.Action) error {
	tracer := otel.Tracer("create")
	var span trace.Span
	ctx, span = tracer.Start(ctx, "create")
	defer span.End()
	logger.InfoContext(ctx, "creating job", "name", name)
	if err := cc.Create(ctx, &rufio.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Annotations: map[string]string{
				"tink-controller-auto-created": "true",
			},
			Labels: map[string]string{
				"tink-controller-auto-created": "true",
			},
		},
		Spec: rufio.JobSpec{
			MachineRef: rufio.MachineRef{
				Name:      hw.Spec.BMCRef.Name,
				Namespace: ns,
			},
			Tasks: tasks,
		},
	}); err != nil {
		return fmt.Errorf("error creating job.bmc.tinkerbell.org object for netbooting machine: %w", err)
	}

	return nil
}
