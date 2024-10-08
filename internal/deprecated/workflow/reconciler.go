package workflow

import (
	"context"
	serrors "errors"
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

	switch wflow.Status.State {
	case "":
		resp, err := r.processNewWorkflow(ctx, logger, wflow)

		return resp, serrors.Join(err, mergePatchsStatus(ctx, r.client, stored, wflow))
	case v1alpha1.WorkflowStateWaiting:
		// make sure any existing job is deleted
		if !wflow.Status.BootOptions.OneTimeNetboot.ExistingJobDeleted {
			rc, err := handleExistingJob(ctx, r.client, wflow)

			return rc, serrors.Join(err, mergePatchsStatus(ctx, r.client, stored, wflow))
		}

		// create a new job
		if wflow.Status.BootOptions.OneTimeNetboot.UID == "" && wflow.Status.BootOptions.OneTimeNetboot.ExistingJobDeleted {
			rc, err := handleJobCreation(ctx, r.client, wflow)

			return rc, serrors.Join(err, mergePatchsStatus(ctx, r.client, stored, wflow))
		}

		// check if the job is complete
		if !wflow.Status.BootOptions.OneTimeNetboot.Complete && wflow.Status.BootOptions.OneTimeNetboot.UID != "" && wflow.Status.BootOptions.OneTimeNetboot.ExistingJobDeleted {
			rc, err := handleJobComplete(ctx, r.client, wflow)

			return rc, serrors.Join(err, mergePatchsStatus(ctx, r.client, stored, wflow))
		}
	case v1alpha1.WorkflowStateRunning:
		r.processRunningWorkflow(wflow)
		// set the current action in the status
		ca := runningAction(wflow)
		if ca != "" && wflow.Status.CurrentAction != ca {
			wflow.Status.CurrentAction = ca
		}

		return reconcile.Result{}, mergePatchsStatus(ctx, r.client, stored, wflow)
	case v1alpha1.WorkflowStatePending, v1alpha1.WorkflowStateTimeout, v1alpha1.WorkflowStateFailed:
		return reconcile.Result{}, nil
	case v1alpha1.WorkflowStateSuccess:
		if wflow.Spec.BootOptions.ToggleAllowNetboot && !wflow.Status.HasCondition(v1alpha1.ToggleAllowNetbootFalse, metav1.ConditionTrue) {
			// handle updating hardware allowPXE to false
			if err := handleHardwareAllowPXE(ctx, r.client, wflow, nil, false); err != nil {
				stored.Status.SetCondition(v1alpha1.WorkflowCondition{
					Type:    v1alpha1.ToggleAllowNetbootFalse,
					Status:  metav1.ConditionTrue,
					Reason:  "Error",
					Message: fmt.Sprintf("error setting Allow PXE: %v", err),
					Time:    &metav1.Time{Time: metav1.Now().UTC()},
				})
				return reconcile.Result{}, serrors.Join(err, mergePatchsStatus(ctx, r.client, stored, wflow))
			}
			wflow.Status.SetCondition(v1alpha1.WorkflowCondition{
				Type:    v1alpha1.ToggleAllowNetbootFalse,
				Status:  metav1.ConditionTrue,
				Reason:  "Complete",
				Message: "set allowPXE to false",
				Time:    &metav1.Time{Time: metav1.Now().UTC()},
			})

			return reconcile.Result{}, mergePatchsStatus(ctx, r.client, stored, wflow)
		}
	}

	return reconcile.Result{}, nil
}

// mergePatchsStatus merges an updated Workflow with an original Workflow and patches the Status object via the client (cc).
func mergePatchsStatus(ctx context.Context, cc ctrlclient.Client, original, updated *v1alpha1.Workflow) error {
	// Patch any changes, regardless of errors
	if !equality.Semantic.DeepEqual(updated, original) {
		if err := cc.Status().Patch(ctx, updated, ctrlclient.MergeFrom(original)); err != nil {
			return fmt.Errorf("error patching status of workflow: %s, error: %w", updated.Name, err)
		}
	}
	return nil
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
			stored.Status.TemplateRendering = v1alpha1.TemplateRenderingFailed
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
		stored.Status.TemplateRendering = v1alpha1.TemplateRenderingFailed
		stored.Status.SetCondition(v1alpha1.WorkflowCondition{
			Type:    v1alpha1.TemplateRenderedSuccess,
			Status:  metav1.ConditionFalse,
			Reason:  "Error",
			Message: err.Error(),
			Time:    &metav1.Time{Time: metav1.Now().UTC()},
		})
		return reconcile.Result{}, err
	}

	var hardware v1alpha1.Hardware
	err := r.client.Get(ctx, ctrlclient.ObjectKey{Name: stored.Spec.HardwareRef, Namespace: stored.Namespace}, &hardware)
	if ctrlclient.IgnoreNotFound(err) != nil {
		logger.Error(err, "error getting Hardware object in processNewWorkflow function")
		stored.Status.TemplateRendering = v1alpha1.TemplateRenderingFailed
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
		stored.Status.TemplateRendering = v1alpha1.TemplateRenderingFailed
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

	data := make(map[string]interface{})
	for key, val := range stored.Spec.HardwareMap {
		data[key] = val
	}
	contract := toTemplateHardwareData(hardware)
	data["Hardware"] = contract

	tinkWf, err := renderTemplateHardware(stored.Name, ptr.StringValue(tpl.Spec.Data), data)
	if err != nil {
		stored.Status.TemplateRendering = v1alpha1.TemplateRenderingFailed
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
	stored.Status.TemplateRendering = v1alpha1.TemplateRenderingSuccessful
	stored.Status.SetCondition(v1alpha1.WorkflowCondition{
		Type:    v1alpha1.TemplateRenderedSuccess,
		Status:  metav1.ConditionTrue,
		Reason:  "Complete",
		Message: "template rendered successfully",
		Time:    &metav1.Time{Time: metav1.Now().UTC()},
	})

	// set hardware allowPXE if requested.
	if stored.Spec.BootOptions.ToggleAllowNetboot {
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
		stored.Status.SetCondition(v1alpha1.WorkflowCondition{
			Type:    v1alpha1.ToggleAllowNetbootTrue,
			Status:  metav1.ConditionTrue,
			Reason:  "Complete",
			Message: "set allowPXE to true",
			Time:    &metav1.Time{Time: metav1.Now().UTC()},
		})
	}

	// netboot the hardware if requested
	if stored.Spec.BootOptions.OneTimeNetboot {
		stored.Status.State = v1alpha1.WorkflowStateWaiting
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

func (r *Reconciler) processRunningWorkflow(stored *v1alpha1.Workflow) {
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
}
