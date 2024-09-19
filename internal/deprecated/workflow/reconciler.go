package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	rufio "github.com/tinkerbell/rufio/api/v1alpha1"
	"github.com/tinkerbell/tink/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	bmcJobName = "tink-controller-%s-one-time-netboot"
)

// Reconciler is a type for managing Workflows.
type Reconciler struct {
	client  ctrlclient.Client
	nowFunc func() time.Time
}

func NewReconciler(client ctrlclient.Client) *Reconciler {
	return &Reconciler{
		client:  client,
		nowFunc: time.Now,
	}
}

// +kubebuilder:rbac:groups=tinkerbell.org,resources=hardware;hardware/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=tinkerbell.org,resources=templates;templates/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=tinkerbell.org,resources=workflows;workflows/status,verbs=get;list;watch;update;patch;delete
// +kubebuilder:rbac:groups=bmc.tinkerbell.org,resources=job;job/status,verbs=get;delete;create;watch

func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := ctrl.LoggerFrom(ctx)
	logger.Info("Reconciling")

	stored := &v1alpha1.Workflow{}
	if err := r.client.Get(ctx, req.NamespacedName, stored); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	if !stored.DeletionTimestamp.IsZero() {
		return reconcile.Result{}, nil
	}
	wflow := stored.DeepCopy()

	var (
		resp reconcile.Result
		err  error
	)
	switch wflow.Status.State {
	case "":
		resp, err = r.processNewWorkflow(ctx, logger, wflow)
	case v1alpha1.WorkflowStateRunning:
		resp = r.processRunningWorkflow(ctx, wflow)
	case v1alpha1.WorkflowStatePreparing:
		if wflow.Status.OneTimeNetboot.CreationStatus == nil {
			return reconcile.Result{Requeue: true}, nil
		}
		if !wflow.Status.OneTimeNetboot.CreationStatus.IsSuccess() {
			return reconcile.Result{Requeue: true}, nil
		}
		existingJob := &rufio.Job{}
		jobName := fmt.Sprintf(bmcJobName, wflow.Spec.HardwareRef)
		if err := r.client.Get(ctx, ctrlclient.ObjectKey{Name: jobName, Namespace: wflow.Namespace}, existingJob); err != nil {
			return reconcile.Result{}, fmt.Errorf("error getting one time netboot job: %w", err)
		}
		if existingJob.HasCondition(rufio.JobFailed, rufio.ConditionTrue) {
			return reconcile.Result{}, fmt.Errorf("one time netboot job failed")
		}
		if existingJob.HasCondition(rufio.JobCompleted, rufio.ConditionTrue) {
			wflow.Status.State = v1alpha1.WorkflowStatePending
		} else {
			return reconcile.Result{Requeue: true}, nil
		}
	case v1alpha1.WorkflowStateSuccess:
		// handle updating hardware allowPXE to false
		var hw v1alpha1.Hardware
		err := r.client.Get(ctx, ctrlclient.ObjectKey{Name: wflow.Spec.HardwareRef, Namespace: wflow.Namespace}, &hw)
		if err != nil && !errors.IsNotFound(err) {
			logger.Error(err, "error getting Hardware object for WorkflowStateSuccess processing")
			return reconcile.Result{}, err
		}
		if err := r.handleHardwareAllowPXE(ctx, wflow, &hw); err != nil {
			return resp, err
		}
	default:
		return resp, nil
	}

	// Patch any changes, regardless of errors
	if !equality.Semantic.DeepEqual(wflow, stored) {
		if perr := r.client.Status().Patch(ctx, wflow, ctrlclient.MergeFrom(stored)); perr != nil {
			err = fmt.Errorf("error patching workflow %s, %w", wflow.Name, perr)
		}
	}
	return resp, err
}

func (r *Reconciler) handleHardwareAllowPXE(ctx context.Context, stored *v1alpha1.Workflow, hardware *v1alpha1.Hardware) error {
	// We need to set allowPXE to true before a workflow runs.
	// We need to set allowPXE to false after a workflow completes successfully.

	// before workflow case
	if stored.Status.ToggleHardware == nil || (stored.Status.ToggleHardware != nil && stored.Status.ToggleHardware.Status == "" && stored.Status.State == "" || stored.Status.State == v1alpha1.WorkflowStatePending) {
		status := &v1alpha1.Status{Status: v1alpha1.StatusSuccess, Message: "allowPXE set to true"}
		for _, iface := range hardware.Spec.Interfaces {
			iface.Netboot.AllowPXE = ptr.Bool(true)
		}
		if err := r.client.Update(ctx, hardware); err != nil {
			status.Status = v1alpha1.StatusFailure
			stored.Status.ToggleHardware = status
			return err
		}
		stored.Status.ToggleHardware = status
	}

	// after workflow case
	if stored.Status.State == v1alpha1.WorkflowStateSuccess {
		status := &v1alpha1.Status{Status: v1alpha1.StatusSuccess, Message: "allowPXE set to false"}
		for _, iface := range hardware.Spec.Interfaces {
			iface.Netboot.AllowPXE = ptr.Bool(false)
		}
		if err := r.client.Update(ctx, hardware); err != nil {
			status.Status = v1alpha1.StatusFailure
			stored.Status.ToggleHardware = status
			return err
		}
		stored.Status.ToggleHardware = status
	}

	return nil
}

func (r *Reconciler) processNewWorkflow(ctx context.Context, logger logr.Logger, stored *v1alpha1.Workflow) (reconcile.Result, error) {
	tpl := &v1alpha1.Template{}
	if err := r.client.Get(ctx, ctrlclient.ObjectKey{Name: stored.Spec.TemplateRef, Namespace: stored.Namespace}, tpl); err != nil {
		if errors.IsNotFound(err) {
			// Throw an error to raise awareness and take advantage of immediate requeue.
			logger.Error(err, "error getting Template object in processNewWorkflow function")
			return reconcile.Result{}, fmt.Errorf(
				"no template found: name=%v; namespace=%v",
				stored.Spec.TemplateRef,
				stored.Namespace,
			)
		}
		return reconcile.Result{}, err
	}

	data := make(map[string]interface{})
	for key, val := range stored.Spec.HardwareMap {
		data[key] = val
	}

	var hardware v1alpha1.Hardware
	err := r.client.Get(ctx, ctrlclient.ObjectKey{Name: stored.Spec.HardwareRef, Namespace: stored.Namespace}, &hardware)
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "error getting Hardware object in processNewWorkflow function")
		return reconcile.Result{}, err
	}

	if stored.Spec.HardwareRef != "" && errors.IsNotFound(err) {
		logger.Error(err, "hardware not found in processNewWorkflow function")
		return reconcile.Result{}, fmt.Errorf(
			"hardware not found: name=%v; namespace=%v",
			stored.Spec.HardwareRef,
			stored.Namespace,
		)
	}

	if err == nil {
		contract := toTemplateHardwareData(hardware)
		data["Hardware"] = contract
	}

	tinkWf, err := renderTemplateHardware(stored.Name, ptr.StringValue(tpl.Spec.Data), data)
	if err != nil {
		return reconcile.Result{}, err
	}

	// populate Task and Action data
	stored.Status = *YAMLToStatus(tinkWf)

	// set hardware allowPXE if requested.
	if stored.Spec.BootOpts.ToggleHardware {
		// We need to set allowPXE to true before a workflow runs.
		// We need to set allowPXE to false after a workflow completes successfully.

		if err := r.handleHardwareAllowPXE(ctx, stored, &hardware); err != nil {
			return reconcile.Result{}, err
		}
	}

	// netboot the hardware if requested
	if stored.Spec.BootOpts.OneTimeNetboot { //nolint:nestif // Will work on this complexity.
		// check if the hardware has a bmcRef
		if hardware.Spec.BMCRef == nil {
			return reconcile.Result{}, fmt.Errorf("hardware %s does not have a BMC, cannot perform one time netboot", hardware.Name)
		}

		// check if an existing job.bmc.tinkerbell.org object exists, if so delete it.
		if !stored.Status.OneTimeNetboot.DeletionStatus.IsSuccess() && !stored.Status.OneTimeNetboot.CreationStatus.IsSuccess() {
			existingJob := &rufio.Job{}
			jobName := fmt.Sprintf(bmcJobName, hardware.Name)
			if err := r.client.Get(ctx, ctrlclient.ObjectKey{Name: jobName, Namespace: hardware.Namespace}, existingJob); err == nil {
				opts := []ctrlclient.DeleteOption{
					ctrlclient.GracePeriodSeconds(0),
					ctrlclient.PropagationPolicy(metav1.DeletePropagationForeground),
				}
				if err := r.client.Delete(ctx, existingJob, opts...); err != nil {
					return reconcile.Result{}, fmt.Errorf("error deleting existing job.bmc.tinkerbell.org object for netbooting machine: %w", err)
				}
				stored.Status.OneTimeNetboot.DeletionStatus = &v1alpha1.Status{Status: v1alpha1.StatusSuccess, Message: "previous existing one time netboot job deleted"}

				return reconcile.Result{Requeue: true}, nil
			} else if errors.IsNotFound(err) {
				stored.Status.OneTimeNetboot.DeletionStatus = &v1alpha1.Status{Status: v1alpha1.StatusSuccess, Message: "no existing one time netboot job found"}
			} else {
				return reconcile.Result{Requeue: true}, err
			}
		}

		if !stored.Status.OneTimeNetboot.CreationStatus.IsSuccess() && stored.Status.OneTimeNetboot.DeletionStatus.IsSuccess() {
			// create a job.bmc.tinkerbell.org object to netboot the hardware
			name := fmt.Sprintf(bmcJobName, hardware.Name)
			ns := hardware.Namespace
			efiBoot := func() bool {
				for _, iface := range hardware.Spec.Interfaces {
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
						Name:      hardware.Spec.BMCRef.Name,
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
			if err := r.client.Create(ctx, job); err != nil {
				return reconcile.Result{}, fmt.Errorf("error creating job.bmc.tinkerbell.org object for netbooting machine: %w", err)
			}
			stored.Status.OneTimeNetboot.CreationStatus = &v1alpha1.Status{Status: v1alpha1.StatusSuccess, Message: "one time netboot job created"}
			// block until the job completes. This is needed as there can be a race condition if the Hardware is already running
			// a Tink Worker.
			stored.Status.State = v1alpha1.WorkflowStatePreparing
			return reconcile.Result{Requeue: true}, nil
		}
		return reconcile.Result{Requeue: true}, nil
	}

	stored.Status.State = v1alpha1.WorkflowStatePending
	return reconcile.Result{}, nil
}

// templateHardwareData defines the data exposed for a Hardware instance to a Template.
type templateHardwareData struct {
	Disks      []string
	Interfaces []v1alpha1.Interface
	UserData   string
	Metadata   v1alpha1.HardwareMetadata
	VendorData string
}

// toTemplateHardwareData converts a Hardware instance of templateHardwareData for use in template
// rendering.
func toTemplateHardwareData(hardware v1alpha1.Hardware) templateHardwareData {
	var contract templateHardwareData
	for _, disk := range hardware.Spec.Disks {
		contract.Disks = append(contract.Disks, disk.Device)
	}
	if len(hardware.Spec.Interfaces) > 0 {
		contract.Interfaces = hardware.Spec.Interfaces
	}
	if hardware.Spec.UserData != nil {
		contract.UserData = ptr.StringValue(hardware.Spec.UserData)
	}
	if hardware.Spec.Metadata != nil {
		contract.Metadata = *hardware.Spec.Metadata
	}
	if hardware.Spec.VendorData != nil {
		contract.VendorData = ptr.StringValue(hardware.Spec.VendorData)
	}
	return contract
}

func (r *Reconciler) processRunningWorkflow(_ context.Context, stored *v1alpha1.Workflow) reconcile.Result { //nolint:unparam // This is the way controller runtime works.
	// Check for global timeout expiration
	if r.nowFunc().After(stored.GetStartTime().Add(time.Duration(stored.Status.GlobalTimeout) * time.Second)) {
		stored.Status.State = v1alpha1.WorkflowStateTimeout
	}

	// check for any running actions that may have timed out
	for ti, task := range stored.Status.Tasks {
		for ai, action := range task.Actions {
			// A running workflow task action has timed out
			if action.Status == v1alpha1.WorkflowStateRunning && action.StartedAt != nil &&
				r.nowFunc().After(action.StartedAt.Add(time.Duration(action.Timeout)*time.Second)) {
				// Set fields on the timed out action
				stored.Status.Tasks[ti].Actions[ai].Status = v1alpha1.WorkflowStateTimeout
				stored.Status.Tasks[ti].Actions[ai].Message = "Action timed out"
				stored.Status.Tasks[ti].Actions[ai].Seconds = int64(r.nowFunc().Sub(action.StartedAt.Time).Seconds())
				// Mark the workflow as timed out
				stored.Status.State = v1alpha1.WorkflowStateTimeout
			}
		}
	}

	return reconcile.Result{}
}

func (r *Reconciler) SetupWithManager(mgr manager.Manager) error {
	return ctrl.
		NewControllerManagedBy(mgr).
		For(&v1alpha1.Workflow{}).
		Complete(r)
}
