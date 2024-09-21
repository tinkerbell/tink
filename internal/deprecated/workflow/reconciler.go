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

func (r *Reconciler) SetupWithManager(mgr manager.Manager) error {
	return ctrl.
		NewControllerManagedBy(mgr).
		For(&v1alpha1.Workflow{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=tinkerbell.org,resources=hardware;hardware/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=tinkerbell.org,resources=templates;templates/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=tinkerbell.org,resources=workflows;workflows/status,verbs=get;list;watch;update;patch;delete
// +kubebuilder:rbac:groups=bmc.tinkerbell.org,resources=job;job/status,verbs=get;list;watch;delete;create

// Reconcile handles Workflow objects. This includes Template rendering, optional Hardware allowPXE toggling, and optional Hardware one-time netbooting.
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
		if wflow.Status.TimeStarted == nil {
			wflow.Status.TimeStarted = &metav1.Time{Time: metav1.Now().UTC()}
		}
	case v1alpha1.WorkflowStateRunning:
		resp = r.processRunningWorkflow(ctx, wflow)
	case v1alpha1.WorkflowStatePreparing:
		if isNetbootJobSetupFinished(wflow.Status.Conditions) {
			println("81")
			existingJob := &rufio.Job{}
			jobName := fmt.Sprintf(bmcJobName, wflow.Spec.HardwareRef)
			if err := r.client.Get(ctx, ctrlclient.ObjectKey{Name: jobName, Namespace: wflow.Namespace}, existingJob); err != nil {
				return reconcile.Result{}, fmt.Errorf("error getting one time netboot job: %w", err)
			}
			if existingJob.HasCondition(rufio.JobFailed, rufio.ConditionTrue) {
				println("88")
				wflow.Status.SetCondition(v1alpha1.WorkflowCondition{
					Type:    v1alpha1.NetbootJobFailed,
					Status:  metav1.ConditionTrue,
					Reason:  "Error",
					Message: "one time netboot job failed",
					Time:    &metav1.Time{Time: metav1.Now().UTC()},
				}, false)
				return reconcile.Result{}, fmt.Errorf("one time netboot job failed")
			}
			if existingJob.HasCondition(rufio.JobCompleted, rufio.ConditionTrue) {
				println("94")
				wflow.Status.SetCondition(v1alpha1.WorkflowCondition{
					Type:    v1alpha1.NetbootJobComplete,
					Status:  metav1.ConditionTrue,
					Reason:  "Complete",
					Message: "one time netboot job completed",
					Time:    &metav1.Time{Time: metav1.Now().UTC()},
				}, false)
				wflow.Status.State = v1alpha1.WorkflowStatePending
			} else {
				println("99")
				wflow.Status.SetCondition(v1alpha1.WorkflowCondition{
					Type:    v1alpha1.NetbootJobRunning,
					Status:  metav1.ConditionTrue,
					Reason:  "Running",
					Message: "netboot job running",
					Time:    &metav1.Time{Time: metav1.Now().UTC()},
				}, true)
				resp.Requeue = true
			}
		}
	case v1alpha1.WorkflowStateSuccess:
		if wflow.Spec.BootOpts.ToggleAllowNetboot && !wflow.Status.HasCondition(v1alpha1.ToggleAllowNetbootComplete, metav1.ConditionTrue) {
			// handle updating hardware allowPXE to false
			wflow.Status.SetCondition(v1alpha1.WorkflowCondition{
				Type:    v1alpha1.ToggleAllowNetbootComplete,
				Status:  metav1.ConditionTrue,
				Reason:  "Complete",
				Message: "setting allowPXE to false",
				Time:    &metav1.Time{Time: metav1.Now().UTC()},
			}, false)
			if err := handleHardwareAllowPXE(ctx, r.client, wflow, nil, false); err != nil {
				println("102")
				return reconcile.Result{}, err
			}
		}
		if wflow.Status.TimeCompleted == nil {
			wflow.Status.TimeCompleted = &metav1.Time{Time: metav1.Now().UTC()}
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

// handleHardwareAllowPXE sets the allowPXE field on the hardware interfaces to true before a workflow runs and false after a workflow completes successfully.
// If hardware is nil then it will be retrieved using the client.
func handleHardwareAllowPXE(ctx context.Context, client ctrlclient.Client, stored *v1alpha1.Workflow, hardware *v1alpha1.Hardware, allowPXE bool) error {
	if stored == nil {
		return fmt.Errorf("cannot handle hardware allowPXE without a workflow")
	}

	if hardware == nil {
		hardware = &v1alpha1.Hardware{}
		if err := client.Get(ctx, ctrlclient.ObjectKey{Name: stored.Spec.HardwareRef, Namespace: stored.Namespace}, hardware); err != nil {
			stored.Status.SetCondition(v1alpha1.WorkflowCondition{
				Type:    v1alpha1.ToggleAllowNetbootFailed,
				Status:  metav1.ConditionTrue,
				Reason:  "Error",
				Message: fmt.Sprintf("error getting hardware: %v", err),
				Time:    &metav1.Time{Time: metav1.Now().UTC()},
			}, false)
			return fmt.Errorf("hardware not found: name=%v; namespace=%v, error: %w", stored.Spec.HardwareRef, stored.Namespace, err)
		}
	}

	for _, iface := range hardware.Spec.Interfaces {
		iface.Netboot.AllowPXE = ptr.Bool(allowPXE)
	}

	if err := client.Update(ctx, hardware); err != nil {
		stored.Status.SetCondition(v1alpha1.WorkflowCondition{
			Type:    v1alpha1.ToggleAllowNetbootFailed,
			Status:  metav1.ConditionTrue,
			Reason:  "Error",
			Message: fmt.Sprintf("error setting allowPXE: %v", err),
			Time:    &metav1.Time{Time: metav1.Now().UTC()},
		}, false)
		return err
	}

	return nil
}

func isNetbootJobSetupFinished(cs []v1alpha1.WorkflowCondition) bool {
	for _, c := range cs {
		if c.Type == v1alpha1.NetbootJobSetupComplete && c.Status == metav1.ConditionTrue {
			return true
		}
	}
	return false
}

func handleOneTimeNetboot(ctx context.Context, client ctrlclient.Client, hw *v1alpha1.Hardware, stored *v1alpha1.Workflow) (reconcile.Result, error) {
	if hw.Spec.BMCRef == nil {
		return reconcile.Result{}, fmt.Errorf("hardware %s does not have a BMC, cannot perform one time netboot", hw.Name)
	}

	if !isNetbootJobSetupFinished(stored.Status.Conditions) || stored.Status.State == v1alpha1.WorkflowStatePreparing {
		println("170")
		existingJob := &rufio.Job{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf(bmcJobName, hw.Name), Namespace: hw.Namespace}}
		opts := []ctrlclient.DeleteOption{
			ctrlclient.GracePeriodSeconds(0),
			ctrlclient.PropagationPolicy(metav1.DeletePropagationForeground),
		}
		err := client.Delete(ctx, existingJob, opts...)
		if err == nil {
			println("189")
			stored.Status.SetCondition(v1alpha1.WorkflowCondition{
				Type:    v1alpha1.NetbootJobSetupComplete,
				Status:  metav1.ConditionTrue,
				Reason:  "Deleted",
				Message: "existing job deleted",
				Time:    &metav1.Time{Time: metav1.Now().UTC()},
			}, false)
			return reconcile.Result{}, nil
		}
		if !errors.IsNotFound(err) {
			println("185")
			stored.Status.SetCondition(v1alpha1.WorkflowCondition{
				Type:    v1alpha1.NetbootJobSetupFailed,
				Status:  metav1.ConditionTrue,
				Reason:  "Error",
				Message: fmt.Sprintf("error deleting existing job: %v", err),
				Time:    &metav1.Time{Time: metav1.Now().UTC()},
			}, false)
			return reconcile.Result{}, fmt.Errorf("error deleting existing job.bmc.tinkerbell.org object for netbooting machine: %w", err)
		}

		name := fmt.Sprintf(bmcJobName, hw.Name)
		ns := hw.Namespace
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
		if err := client.Create(ctx, job); err != nil {
			println("248")
			stored.Status.SetCondition(v1alpha1.WorkflowCondition{
				Type:    v1alpha1.NetbootJobSetupFailed,
				Status:  metav1.ConditionTrue,
				Reason:  "Error",
				Message: fmt.Sprintf("error creating job: %v", err),
				Time:    &metav1.Time{Time: metav1.Now().UTC()},
			}, false)
			return reconcile.Result{}, fmt.Errorf("error creating job.bmc.tinkerbell.org object for netbooting machine: %w", err)
		}
		println("252")
		stored.Status.SetCondition(v1alpha1.WorkflowCondition{
			Type:    v1alpha1.NetbootJobSetupComplete,
			Status:  metav1.ConditionTrue,
			Reason:  "Created",
			Message: "job created",
			Time:    &metav1.Time{Time: metav1.Now().UTC()},
		}, false)
		stored.Status.State = v1alpha1.WorkflowStatePreparing
	}

	println("338")
	return reconcile.Result{}, nil
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
	if stored.Spec.BootOpts.ToggleAllowNetboot {
		stored.Status.SetCondition(v1alpha1.WorkflowCondition{
			Type:    v1alpha1.ToggleAllowNetbootRunning,
			Status:  metav1.ConditionTrue,
			Reason:  "Started",
			Message: "setting allowPXE to true",
			Time:    &metav1.Time{Time: metav1.Now().UTC()},
		}, false)
		if err := handleHardwareAllowPXE(ctx, r.client, stored, &hardware, true); err != nil {
			return reconcile.Result{}, err
		}
	}

	// netboot the hardware if requested
	if stored.Spec.BootOpts.OneTimeNetboot {
		return handleOneTimeNetboot(ctx, r.client, &hardware, stored)
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
