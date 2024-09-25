package workflow

import (
	"context"
	"fmt"
	"time"

	rufio "github.com/tinkerbell/rufio/api/v1alpha1"
	"github.com/tinkerbell/tink/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	bmcJobName = "tink-controller-%s-one-time-netboot"
)

// handleExistingJob ensures that an existing job.bmc.tinkerbell.org is removed.
func handleExistingJob(ctx context.Context, cc client.Client, wf *v1alpha1.Workflow) (reconcile.Result, error) {
	if wf.Status.Job.ExistingJobDeleted {
		return reconcile.Result{}, nil
	}
	name := fmt.Sprintf(bmcJobName, wf.Spec.HardwareRef)
	namespace := wf.Namespace
	if err := cc.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &rufio.Job{}); (err != nil && !errors.IsNotFound(err)) || err == nil {
		existingJob := &rufio.Job{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
		opts := []client.DeleteOption{
			client.GracePeriodSeconds(0),
			client.PropagationPolicy(metav1.DeletePropagationForeground),
		}
		if err := cc.Delete(ctx, existingJob, opts...); err != nil {
			return reconcile.Result{}, fmt.Errorf("error deleting job.bmc.tinkerbell.org object: %w", err)
		}
		return reconcile.Result{Requeue: true}, nil
	}
	wf.Status.Job.ExistingJobDeleted = true

	return reconcile.Result{Requeue: true}, nil
}

func handleJobCreation(ctx context.Context, cc client.Client, wf *v1alpha1.Workflow) (reconcile.Result, error) {
	if wf.Status.Job.UID == "" && wf.Status.Job.ExistingJobDeleted {
		existingJob := &rufio.Job{}
		if err := cc.Get(ctx, client.ObjectKey{Name: fmt.Sprintf(bmcJobName, wf.Spec.HardwareRef), Namespace: wf.Namespace}, existingJob); err == nil {
			wf.Status.Job.UID = existingJob.GetUID()
			return reconcile.Result{Requeue: true}, nil
		}
		hw := &v1alpha1.Hardware{ObjectMeta: metav1.ObjectMeta{Name: wf.Spec.HardwareRef, Namespace: wf.Namespace}}
		if gerr := cc.Get(ctx, client.ObjectKey{Name: wf.Spec.HardwareRef, Namespace: wf.Namespace}, hw); gerr != nil {
			return reconcile.Result{}, fmt.Errorf("error getting hardware %s: %w", wf.Spec.HardwareRef, gerr)
		}
		if hw.Spec.BMCRef == nil {
			return reconcile.Result{}, fmt.Errorf("hardware %s does not have a BMC, cannot perform one time netboot", hw.Name)
		}
		if err := createNetbootJob(ctx, cc, hw, wf.Namespace); err != nil {
			wf.Status.SetCondition(v1alpha1.WorkflowCondition{
				Type:    v1alpha1.NetbootJobSetupFailed,
				Status:  metav1.ConditionTrue,
				Reason:  "Error",
				Message: fmt.Sprintf("error creating job: %v", err),
				Time:    &metav1.Time{Time: metav1.Now().UTC()},
			})
			return reconcile.Result{}, fmt.Errorf("error creating job.bmc.tinkerbell.org object: %w", err)
		}
		wf.Status.SetCondition(v1alpha1.WorkflowCondition{
			Type:    v1alpha1.NetbootJobSetupComplete,
			Status:  metav1.ConditionTrue,
			Reason:  "Created",
			Message: "job created",
			Time:    &metav1.Time{Time: metav1.Now().UTC()},
		})

		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

func handleJobComplete(ctx context.Context, cc client.Client, wf *v1alpha1.Workflow) (reconcile.Result, error) {
	if !wf.Status.Job.Complete && wf.Status.Job.UID != "" && wf.Status.Job.ExistingJobDeleted {
		existingJob := &rufio.Job{}
		jobName := fmt.Sprintf(bmcJobName, wf.Spec.HardwareRef)
		if err := cc.Get(ctx, client.ObjectKey{Name: jobName, Namespace: wf.Namespace}, existingJob); err != nil {
			return reconcile.Result{}, fmt.Errorf("error getting one time netboot job: %w", err)
		}
		if existingJob.HasCondition(rufio.JobFailed, rufio.ConditionTrue) {
			wf.Status.SetCondition(v1alpha1.WorkflowCondition{
				Type:    v1alpha1.NetbootJobFailed,
				Status:  metav1.ConditionTrue,
				Reason:  "Error",
				Message: "one time netboot job failed",
				Time:    &metav1.Time{Time: metav1.Now().UTC()},
			})
			return reconcile.Result{}, fmt.Errorf("one time netboot job failed")
		}
		if existingJob.HasCondition(rufio.JobCompleted, rufio.ConditionTrue) {
			wf.Status.SetCondition(v1alpha1.WorkflowCondition{
				Type:    v1alpha1.NetbootJobComplete,
				Status:  metav1.ConditionTrue,
				Reason:  "Complete",
				Message: "one time netboot job completed",
				Time:    &metav1.Time{Time: metav1.Now().UTC()},
			})
			wf.Status.State = v1alpha1.WorkflowStatePending
			wf.Status.Job.Complete = true
			return reconcile.Result{Requeue: true}, nil
		}
		if !wf.Status.HasCondition(v1alpha1.NetbootJobRunning, metav1.ConditionTrue) {
			wf.Status.SetCondition(v1alpha1.WorkflowCondition{
				Type:    v1alpha1.NetbootJobRunning,
				Status:  metav1.ConditionTrue,
				Reason:  "Running",
				Message: "one time netboot job running",
				Time:    &metav1.Time{Time: metav1.Now().UTC()},
			})
		}
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	return reconcile.Result{}, nil
}

func createNetbootJob(ctx context.Context, cc client.Client, hw *v1alpha1.Hardware, ns string) error {
	name := fmt.Sprintf(bmcJobName, hw.Name)
	efiBoot := func() bool {
		for _, iface := range hw.Spec.Interfaces {
			if iface.DHCP != nil && iface.DHCP.UEFI {
				return true
			}
		}
		return false
	}()
	job := &rufio.Job{
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
			Tasks: []rufio.Action{
				{
					PowerAction: rufio.PowerHardOff.Ptr(),
				},
				{
					OneTimeBootDeviceAction: &rufio.OneTimeBootDeviceAction{
						Devices: []rufio.BootDevice{
							rufio.PXE,
						},
						EFIBoot: efiBoot,
					},
				},
				{
					PowerAction: rufio.PowerOn.Ptr(),
				},
			},
		},
	}
	if err := cc.Create(ctx, job); err != nil {
		return fmt.Errorf("error creating job.bmc.tinkerbell.org object for netbooting machine: %w", err)
	}

	return nil
}

// handleHardwareAllowPXE sets the allowPXE field on the hardware interfaces to true before a workflow runs and false after a workflow completes successfully.
// If hardware is nil then it will be retrieved using the client.
func handleHardwareAllowPXE(ctx context.Context, cc client.Client, stored *v1alpha1.Workflow, hardware *v1alpha1.Hardware, allowPXE bool) error {
	if hardware == nil && stored != nil {
		hardware = &v1alpha1.Hardware{}
		if err := cc.Get(ctx, client.ObjectKey{Name: stored.Spec.HardwareRef, Namespace: stored.Namespace}, hardware); err != nil {
			return fmt.Errorf("hardware not found: name=%v; namespace=%v, error: %w", stored.Spec.HardwareRef, stored.Namespace, err)
		}
	} else if stored == nil {
		return fmt.Errorf("workflow and hardware cannot both be nil")
	}

	for _, iface := range hardware.Spec.Interfaces {
		iface.Netboot.AllowPXE = ptr.Bool(allowPXE)
	}

	if err := cc.Update(ctx, hardware); err != nil {
		return fmt.Errorf("error updating allow pxe: %w", err)
	}

	return nil
}
