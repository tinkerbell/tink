package controller

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Controller is an interface implemented by Karpenter custom resources.
type Controller interface {
	// Reconcile hands a hydrated kubernetes resource to the controller for
	// reconciliation. Any changes made to the resource's status are persisted
	// after Reconcile returns, even if it returns an error.
	reconcile.Reconciler

	// Register will register the controller with the manager
	Register(context.Context, manager.Manager) error
}

// Manager manages a set of controllers and webhooks.
type Manager interface {
	manager.Manager
	RegisterControllers(context.Context, ...Controller) Manager
}
