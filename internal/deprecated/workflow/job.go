package workflow

import (
	"context"
	"fmt"
	"time"

	rufio "github.com/tinkerbell/rufio/api/v1alpha1"
	"github.com/tinkerbell/tink/api/v1alpha1"
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
	// there are 3 phases. 1. Clean up existing 2. Create new 3. Track status
	// 1. clean up existing job if it wasn't already deleted
	if j := s.workflow.Status.BootOptions.Jobs[name.String()]; !j.ExistingJobDeleted {
		result, err := s.deleteExisting(ctx, name)
		if err != nil {
			return result, err
		}
		jStatus := s.workflow.Status.BootOptions.Jobs[name.String()]
		jStatus.ExistingJobDeleted = true
		s.workflow.Status.BootOptions.Jobs[name.String()] = jStatus
		return result, nil
	}

	// 2. create a new job
	if uid := s.workflow.Status.BootOptions.Jobs[name.String()].UID; uid == "" {
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

	return reconcile.Result{Requeue: true}, nil
}

// This function will update the Workflow status.
func (s *state) createJob(ctx context.Context, actions []rufio.Action, name jobName) (reconcile.Result, error) {
	// create a new job
	// The assumption is that the UID is not set. UID checking is not handled here.
	// 1. look up if there's an existing job with the same name, if so update the status with the UID and return
	// 2. if there's no existing job, create a new job, update the status with the UID, and return

	rj := &rufio.Job{}
	if err := s.client.Get(ctx, client.ObjectKey{Name: name.String(), Namespace: s.workflow.Namespace}, rj); err == nil {
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
		return reconcile.Result{}, fmt.Errorf("hardware %s does not have a BMC", s.hardware.Name)
	}

	if err := create(ctx, s.client, name.String(), s.hardware, s.workflow.Namespace, actions); err != nil {
		return reconcile.Result{}, fmt.Errorf("error creating job: %w", err)
	}

	// update the status with the UID
	jStatus := s.workflow.Status.BootOptions.Jobs[name.String()]
	jStatus.UID = rj.GetUID()
	s.workflow.Status.BootOptions.Jobs[name.String()] = jStatus

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
	// track status
	// get the job
	rj := &rufio.Job{}
	if err := s.client.Get(ctx, client.ObjectKey{Name: name.String(), Namespace: s.workflow.Namespace}, rj); err != nil {
		return reconcile.Result{}, trackedStateError, fmt.Errorf("error getting job: %w", err)
	}
	if rj.HasCondition(rufio.JobFailed, rufio.ConditionTrue) {
		// job failed
		return reconcile.Result{}, trackedStateFailed, fmt.Errorf("job failed")
	}
	if rj.HasCondition(rufio.JobCompleted, rufio.ConditionTrue) {
		// job completed
		jStatus := s.workflow.Status.BootOptions.Jobs[name.String()]
		jStatus.Complete = true
		s.workflow.Status.BootOptions.Jobs[name.String()] = jStatus

		return reconcile.Result{}, trackedStateComplete, nil
	}
	// still running
	time.Sleep(s.backoff.NextBackOff())
	return reconcile.Result{Requeue: true}, trackedStateRunning, nil
}

func create(ctx context.Context, cc client.Client, name string, hw *v1alpha1.Hardware, ns string, tasks []rufio.Action) error {
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
