package workflow

import (
	"context"
	"time"

	tinkv1 "github.com/tinkerbell/tink/api/v1alpha2"
	"github.com/tinkerbell/tink/internal/workflow/internal"
	"k8s.io/apimachinery/pkg/api/errors"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciler reconciles Workflow instances.
type Reconciler struct {
	client  client.Client
	nowFunc func() time.Time
}

// NewReconciler creates a Reconciler instance.
func NewReconciler(clnt client.Client) *Reconciler {
	return &Reconciler{
		client:  clnt,
		nowFunc: time.Now,
	}
}

// +kubebuilder:rbac:groups=tinkerbell.org,resources=hardware;hardware/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=tinkerbell.org,resources=templates;templates/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=tinkerbell.org,resources=workflows;workflows/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=tinkerbell.org,resources=workflows;workflows/finalizers,verbs=update

func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (result reconcile.Result, rerr error) {
	logger := ctrl.LoggerFrom(ctx)
	logger.Info("Reconciling")

	wrkflw := &tinkv1.Workflow{}
	if err := r.client.Get(ctx, req.NamespacedName, wrkflw); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Workflow not found; discontinuing reconciliation")
		}
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// TODO(chrisdoherty)
	if !wrkflw.DeletionTimestamp.IsZero() {
		return reconcile.Result{}, nil
	}

	rc := internal.ReconciliationContext{
		Client:   r.client,
		Log:      logger,
		Workflow: wrkflw.DeepCopy(),
	}

	// Always attempt to patch.
	defer func() {
		if err := r.client.Status().Patch(ctx, rc.Workflow, client.MergeFrom(wrkflw)); err != nil {
			rerr = kerrors.NewAggregate([]error{rerr, err})
		}
	}()

	return rc.Reconcile(ctx)
}

func (r *Reconciler) SetupWithManager(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tinkv1.Workflow{}).
		Complete(r)
}
