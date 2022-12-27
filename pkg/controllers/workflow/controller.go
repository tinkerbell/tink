package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/tinkerbell/tink/internal/convert"
	"github.com/tinkerbell/tink/internal/workflow"
	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	"github.com/tinkerbell/tink/pkg/controllers"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"knative.dev/pkg/ptr"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Controller is a type for managing Workflows.
type Controller struct {
	kubeClient client.Client
	nowFunc    func() time.Time
}

func NewController(kubeClient client.Client) *Controller {
	return &Controller{
		kubeClient: kubeClient,
		nowFunc:    time.Now,
	}
}

// +kubebuilder:rbac:groups=tinkerbell.org,resources=hardware;hardware/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=tinkerbell.org,resources=templates;templates/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=tinkerbell.org,resources=workflows;workflows/status,verbs=get;list;watch;update;patch;delete

func (c *Controller) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := controllerruntime.LoggerFrom(ctx)
	logger.Info("Reconciling")

	stored := &v1alpha1.Workflow{}
	if err := c.kubeClient.Get(ctx, req.NamespacedName, stored); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return controllers.RetryIfError(ctx, err)
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
		resp, err = c.processNewWorkflow(ctx, logger, wflow)
	case v1alpha1.WorkflowStateRunning:
		resp = c.processRunningWorkflow(ctx, wflow)
	default:
		return resp, nil
	}

	// Patch any changes, regardless of errors
	if !equality.Semantic.DeepEqual(wflow, stored) {
		if perr := c.kubeClient.Status().Patch(ctx, wflow, client.MergeFrom(stored)); perr != nil {
			err = fmt.Errorf("error patching workflow %s, %w", wflow.Name, perr)
		}
	}
	return resp, err
}

func (c *Controller) processNewWorkflow(ctx context.Context, logger logr.Logger, stored *v1alpha1.Workflow) (reconcile.Result, error) {
	tpl := &v1alpha1.Template{}
	if err := c.kubeClient.Get(ctx, client.ObjectKey{Name: stored.Spec.TemplateRef, Namespace: stored.Namespace}, tpl); err != nil {
		if errors.IsNotFound(err) {
			// Throw an error to raise awareness and take advantage of immediate requeue.
			logger.Error(err, "error getting Template object in processNewWorkflow function")
			return reconcile.Result{}, fmt.Errorf(
				"no template found: name=%v; namespace=%v",
				stored.Spec.TemplateRef,
				stored.Namespace,
			)
		}
		return controllers.RetryIfError(ctx, err)
	}

	data := make(map[string]interface{})
	for key, val := range stored.Spec.HardwareMap {
		data[key] = val
	}

	var hardware v1alpha1.Hardware
	err := c.kubeClient.Get(ctx, client.ObjectKey{Name: stored.Spec.HardwareRef, Namespace: stored.Namespace}, &hardware)
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

	tinkWf, _, err := workflow.RenderTemplateHardware(stored.Name, ptr.StringValue(tpl.Spec.Data), data)
	if err != nil {
		return reconcile.Result{}, err
	}

	// populate Task and Action data
	stored.Status = *convert.WorkflowYAMLToStatus(tinkWf)

	stored.Status.State = v1alpha1.WorkflowStatePending
	return reconcile.Result{}, nil
}

// templateHardwareData defines the data exposed for a Hardware instance to a Template.
type templateHardwareData struct {
	Disks []string
}

// toTemplateHardwareData converts a Hardware instance of templateHardwareData for use in template
// rendering.
func toTemplateHardwareData(hardware v1alpha1.Hardware) templateHardwareData {
	var contract templateHardwareData
	for _, disk := range hardware.Spec.Disks {
		contract.Disks = append(contract.Disks, disk.Device)
	}
	return contract
}

func (c *Controller) processRunningWorkflow(_ context.Context, stored *v1alpha1.Workflow) reconcile.Result {
	// Check for global timeout expiration
	if c.nowFunc().After(stored.GetStartTime().Add(time.Duration(stored.Status.GlobalTimeout) * time.Second)) {
		stored.Status.State = v1alpha1.WorkflowStateTimeout
	}

	// check for any running actions that may have timed out
	for ti, task := range stored.Status.Tasks {
		for ai, action := range task.Actions {
			// A running workflow task action has timed out
			if action.Status == v1alpha1.WorkflowStateRunning && action.StartedAt != nil &&
				c.nowFunc().After(action.StartedAt.Add(time.Duration(action.Timeout)*time.Second)) {
				// Set fields on the timed out action
				stored.Status.Tasks[ti].Actions[ai].Status = v1alpha1.WorkflowStateTimeout
				stored.Status.Tasks[ti].Actions[ai].Message = "Action timed out"
				stored.Status.Tasks[ti].Actions[ai].Seconds = int64(c.nowFunc().Sub(action.StartedAt.Time).Seconds())
				// Mark the workflow as timed out
				stored.Status.State = v1alpha1.WorkflowStateTimeout
			}
		}
	}

	return reconcile.Result{}
}

func (c *Controller) Register(_ context.Context, m manager.Manager) error {
	return controllerruntime.
		NewControllerManagedBy(m).
		For(&v1alpha1.Workflow{}).
		Complete(c)
}
