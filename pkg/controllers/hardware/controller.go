package hardware

import (
	"context"
	"fmt"
	"time"

	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	"github.com/tinkerbell/tink/pkg/controllers"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Controller is a type for managing Hardwares.
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

func (c *Controller) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	stored := &v1alpha1.Hardware{}
	if err := c.kubeClient.Get(ctx, req.NamespacedName, stored); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return controllers.RetryIfError(ctx, err)
	}
	if !stored.DeletionTimestamp.IsZero() {
		return reconcile.Result{}, nil
	}
	hw := stored.DeepCopy()

	resp := c.reconcileDiskData(ctx, hw)

	// Patch any changes, regardless of errors
	if !equality.Semantic.DeepEqual(hw, stored) {
		if perr := c.kubeClient.Patch(ctx, hw, client.MergeFrom(stored)); perr != nil {
			return reconcile.Result{}, fmt.Errorf("error patching hardware %s, %w", hw.Name, perr)
		}
	}

	return resp, nil
}

func (c *Controller) reconcileDiskData(_ context.Context, hardware *v1alpha1.Hardware) reconcile.Result {
	if hardware.Spec.Disks == nil {
		foundDisks := make([]v1alpha1.Disk, 0)

		if hardware.Spec.Metadata != nil &&
			hardware.Spec.Metadata.Instance != nil &&
			hardware.Spec.Metadata.Instance.Storage != nil {
			for _, disk := range hardware.Spec.Metadata.Instance.Storage.Disks {
				foundDisks = append(foundDisks, v1alpha1.Disk{Device: disk.Device})
			}
		}

		hardware.Spec.Disks = foundDisks
	}

	return reconcile.Result{}
}

func (c *Controller) Register(_ context.Context, m manager.Manager) error {
	return controllerruntime.
		NewControllerManagedBy(m).
		For(&v1alpha1.Hardware{}).
		Complete(c)
}
