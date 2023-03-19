package controller

import (
	"fmt"

	"github.com/tinkerbell/tink/api/v1alpha1"
	"github.com/tinkerbell/tink/internal/workflow"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

var schemeBuilder = runtime.NewSchemeBuilder(
	clientgoscheme.AddToScheme,
	v1alpha1.AddToScheme,
)

// DefaultScheme returns a scheme with all the types necessary for the tink controller.
func DefaultScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = schemeBuilder.AddToScheme(s)
	return s
}

// NewManager creates a new controller manager with tink controller controllers pre-registered.
// If opts.Scheme is nil, DefaultScheme() is used.
func NewManager(cfg *rest.Config, opts ctrl.Options) (ctrl.Manager, error) {
	if opts.Scheme == nil {
		opts.Scheme = DefaultScheme()
	}

	mgr, err := ctrl.NewManager(cfg, opts)
	if err != nil {
		return nil, fmt.Errorf("controller manager: %w", err)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return nil, fmt.Errorf("set up health check: %w", err)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return nil, fmt.Errorf("set up ready check: %w", err)
	}

	err = workflow.NewReconciler(mgr.GetClient()).SetupWithManager(mgr)
	if err != nil {
		return nil, fmt.Errorf("setup workflow reconciler: %w", err)
	}

	return mgr, nil
}
