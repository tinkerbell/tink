package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	"github.com/tinkerbell/tink/pkg/controllers"
	"github.com/tinkerbell/tink/pkg/convert"
	protoworkflow "github.com/tinkerbell/tink/protos/workflow"
	tinkworkflow "github.com/tinkerbell/tink/workflow"
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

func (c *Controller) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
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
		resp, err = c.processNewWorkflow(ctx, wflow)
	case protoworkflow.State_name[int32(protoworkflow.State_STATE_RUNNING)]:
		resp = c.processRunningWorkflow(ctx, wflow)
	}

	// Patch any changes, regardless of errors
	if !equality.Semantic.DeepEqual(wflow, stored) {
		if perr := c.kubeClient.Status().Patch(ctx, wflow, client.MergeFrom(stored)); perr != nil {
			err = fmt.Errorf("error patching workflow %s, %w", wflow.Name, perr)
		}
	}
	return resp, err
}

func (c *Controller) processNewWorkflow(ctx context.Context, stored *v1alpha1.Workflow) (reconcile.Result, error) {
	tpl := &v1alpha1.Template{}
	if err := c.kubeClient.Get(ctx, client.ObjectKey{Name: stored.Spec.TemplateRef, Namespace: stored.Namespace}, tpl); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return controllers.RetryIfError(ctx, err)
	}

	tinkWf, _, err := tinkworkflow.RenderTemplateHardware(stored.Name, ptr.StringValue(tpl.Spec.Data), stored.Spec.HardwareMap)
	if err != nil {
		return reconcile.Result{}, err
	}

	// populate Task and Action data
	stored.Status = *convert.WorkflowYAMLToStatus(tinkWf)

	stored.Status.State = protoworkflow.State_name[int32(protoworkflow.State_STATE_PENDING)]
	return reconcile.Result{}, nil
}

func (c *Controller) processRunningWorkflow(_ context.Context, stored *v1alpha1.Workflow) reconcile.Result {
	// Check for global timeout expiration
	if c.nowFunc().After(stored.GetStartTime().Add(time.Duration(stored.Status.GlobalTimeout) * time.Second)) {
		stored.Status.State = protoworkflow.State_name[int32(protoworkflow.State_STATE_TIMEOUT)]
	}

	// check for any running actions that may have timed out
	for ti, task := range stored.Status.Tasks {
		for ai, action := range task.Actions {
			// A running workflow task action has timed out
			if action.Status == protoworkflow.State_name[int32(protoworkflow.State_STATE_RUNNING)] &&
				action.StartedAt != nil &&
				c.nowFunc().After(action.StartedAt.Add(time.Duration(action.Timeout)*time.Second)) {
				// Set fields on the timed out action
				stored.Status.Tasks[ti].Actions[ai].Status = protoworkflow.State_name[int32(protoworkflow.State_STATE_TIMEOUT)]
				stored.Status.Tasks[ti].Actions[ai].Message = "Action timed out"
				stored.Status.Tasks[ti].Actions[ai].Seconds = int64(c.nowFunc().Sub(action.StartedAt.Time).Seconds())
				// Mark the workflow as timed out
				stored.Status.State = protoworkflow.State_name[int32(protoworkflow.State_STATE_TIMEOUT)]
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
