package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
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
	failed = "failed"
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
	case v1alpha1.WorkflowStateRunning:
		resp = r.processRunningWorkflow(ctx, wflow)
		// set the current action in the status
		ca := runningAction(wflow)
		if ca != "" && wflow.Status.CurrentAction != ca {
			wflow.Status.CurrentAction = ca
		}
	case v1alpha1.WorkflowStatePreparing:
		// make sure any existing job is deleted
		if !wflow.Status.Job.ExistingJobDeleted {
			rc, err := handleExistingJob(ctx, r.client, wflow)
			// Patch any changes, regardless of errors
			if !equality.Semantic.DeepEqual(wflow, stored) {
				if perr := r.client.Status().Patch(ctx, wflow, ctrlclient.MergeFrom(stored)); perr != nil {
					err = fmt.Errorf("error patching workflow %s, %w", wflow.Name, perr)
				}
			}
			return rc, err
		}

		// create a new job
		if wflow.Status.Job.UID == "" && wflow.Status.Job.ExistingJobDeleted {
			rc, err := handleJobCreation(ctx, r.client, wflow)
			// Patch any changes, regardless of errors
			if !equality.Semantic.DeepEqual(wflow, stored) {
				if perr := r.client.Status().Patch(ctx, wflow, ctrlclient.MergeFrom(stored)); perr != nil {
					err = fmt.Errorf("error patching workflow %s, %w", wflow.Name, perr)
				}
			}
			return rc, err
		}

		// check if the job is complete
		if !wflow.Status.Job.Complete && wflow.Status.Job.UID != "" && wflow.Status.Job.ExistingJobDeleted {
			rc, err := handleJobComplete(ctx, r.client, wflow)
			// Patch any changes, regardless of errors
			if !equality.Semantic.DeepEqual(wflow, stored) {
				if perr := r.client.Status().Patch(ctx, wflow, ctrlclient.MergeFrom(stored)); perr != nil {
					err = fmt.Errorf("error patching workflow %s, %w", wflow.Name, perr)
				}
			}
			return rc, err
		}
	case v1alpha1.WorkflowStateSuccess:
		if wflow.Spec.BootOpts.ToggleAllowNetboot && !wflow.Status.HasCondition(v1alpha1.ToggleAllowNetbootFalse, metav1.ConditionTrue) {
			// handle updating hardware allowPXE to false
			wflow.Status.SetCondition(v1alpha1.WorkflowCondition{
				Type:    v1alpha1.ToggleAllowNetbootFalse,
				Status:  metav1.ConditionTrue,
				Reason:  "Complete",
				Message: "setting allowPXE to false",
				Time:    &metav1.Time{Time: metav1.Now().UTC()},
			})
			if gerr := handleHardwareAllowPXE(ctx, r.client, wflow, nil, false); gerr != nil {
				stored.Status.SetCondition(v1alpha1.WorkflowCondition{
					Type:    v1alpha1.ToggleAllowNetbootFalse,
					Status:  metav1.ConditionTrue,
					Reason:  "Error",
					Message: fmt.Sprintf("error setting Allow PXE: %v", gerr),
					Time:    &metav1.Time{Time: metav1.Now().UTC()},
				})
				err = gerr
			}
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

func runningAction(wf *v1alpha1.Workflow) string {
	for _, task := range wf.Status.Tasks {
		for _, action := range task.Actions {
			if action.Status == v1alpha1.WorkflowStateRunning {
				return action.Name
			}
		}
	}

	return ""
}

func (r *Reconciler) processNewWorkflow(ctx context.Context, logger logr.Logger, stored *v1alpha1.Workflow) (reconcile.Result, error) {
	tpl := &v1alpha1.Template{}
	if err := r.client.Get(ctx, ctrlclient.ObjectKey{Name: stored.Spec.TemplateRef, Namespace: stored.Namespace}, tpl); err != nil {
		if errors.IsNotFound(err) {
			// Throw an error to raise awareness and take advantage of immediate requeue.
			logger.Error(err, "error getting Template object in processNewWorkflow function")
			stored.Status.TemplateRendering = failed
			stored.Status.SetCondition(v1alpha1.WorkflowCondition{
				Type:    v1alpha1.TemplateRenderedSuccess,
				Status:  metav1.ConditionFalse,
				Reason:  "Error",
				Message: "template not found",
				Time:    &metav1.Time{Time: metav1.Now().UTC()},
			})
			return reconcile.Result{}, fmt.Errorf(
				"no template found: name=%v; namespace=%v",
				stored.Spec.TemplateRef,
				stored.Namespace,
			)
		}
		stored.Status.TemplateRendering = failed
		stored.Status.SetCondition(v1alpha1.WorkflowCondition{
			Type:    v1alpha1.TemplateRenderedSuccess,
			Status:  metav1.ConditionFalse,
			Reason:  "Error",
			Message: err.Error(),
			Time:    &metav1.Time{Time: metav1.Now().UTC()},
		})
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
		stored.Status.TemplateRendering = failed
		stored.Status.SetCondition(v1alpha1.WorkflowCondition{
			Type:    v1alpha1.TemplateRenderedSuccess,
			Status:  metav1.ConditionFalse,
			Reason:  "Error",
			Message: fmt.Sprintf("error getting hardware: %v", err),
			Time:    &metav1.Time{Time: metav1.Now().UTC()},
		})
		return reconcile.Result{}, err
	}

	if stored.Spec.HardwareRef != "" && errors.IsNotFound(err) {
		logger.Error(err, "hardware not found in processNewWorkflow function")
		stored.Status.TemplateRendering = failed
		stored.Status.SetCondition(v1alpha1.WorkflowCondition{
			Type:    v1alpha1.TemplateRenderedSuccess,
			Status:  metav1.ConditionFalse,
			Reason:  "Error",
			Message: fmt.Sprintf("hardware not found: %v", err),
			Time:    &metav1.Time{Time: metav1.Now().UTC()},
		})
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
		stored.Status.TemplateRendering = failed
		stored.Status.SetCondition(v1alpha1.WorkflowCondition{
			Type:    v1alpha1.TemplateRenderedSuccess,
			Status:  metav1.ConditionFalse,
			Reason:  "Error",
			Message: fmt.Sprintf("error rendering template: %v", err),
			Time:    &metav1.Time{Time: metav1.Now().UTC()},
		})
		return reconcile.Result{}, err
	}

	// populate Task and Action data
	stored.Status = *YAMLToStatus(tinkWf)
	stored.Status.TemplateRendering = "successful"
	stored.Status.SetCondition(v1alpha1.WorkflowCondition{
		Type:    v1alpha1.TemplateRenderedSuccess,
		Status:  metav1.ConditionTrue,
		Reason:  "Complete",
		Message: "template rendered successfully",
		Time:    &metav1.Time{Time: metav1.Now().UTC()},
	})

	// set hardware allowPXE if requested.
	if stored.Spec.BootOpts.ToggleAllowNetboot {
		stored.Status.SetCondition(v1alpha1.WorkflowCondition{
			Type:    v1alpha1.ToggleAllowNetbootTrue,
			Status:  metav1.ConditionTrue,
			Reason:  "Complete",
			Message: "setting allowPXE to true",
			Time:    &metav1.Time{Time: metav1.Now().UTC()},
		})
		if err := handleHardwareAllowPXE(ctx, r.client, stored, &hardware, true); err != nil {
			stored.Status.SetCondition(v1alpha1.WorkflowCondition{
				Type:    v1alpha1.ToggleAllowNetbootTrue,
				Status:  metav1.ConditionFalse,
				Reason:  "Error",
				Message: fmt.Sprintf("error setting allowPXE to true: %v", err),
				Time:    &metav1.Time{Time: metav1.Now().UTC()},
			})
			return reconcile.Result{}, err
		}
	}

	// netboot the hardware if requested
	if stored.Spec.BootOpts.OneTimeNetboot {
		stored.Status.State = v1alpha1.WorkflowStatePreparing
		return reconcile.Result{Requeue: true}, err
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
