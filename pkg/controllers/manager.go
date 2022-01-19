package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/zapr"
	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"knative.dev/pkg/logging"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	WorkerAddr = "status.tasks.workeraddr"
)

var (
	runtimescheme = runtime.NewScheme()
	options       = Options{}
)

func init() {
	_ = clientgoscheme.AddToScheme(runtimescheme)
	_ = v1alpha1.AddToScheme(runtimescheme)
}

// Options for running this binary.
type Options struct {
	MetricsPort     int
	HealthProbePort int
}

// GetControllerOptions returns a set of options used by the Tink controller.
// These options include leader election enabled.
func GetControllerOptions() controllerruntime.Options {
	return controllerruntime.Options{
		Logger:                 zapr.NewLogger(logging.FromContext(context.Background()).Desugar()),
		LeaderElection:         true,
		LeaderElectionID:       "tink-leader-election",
		Scheme:                 runtimescheme,
		MetricsBindAddress:     fmt.Sprintf(":%d", options.MetricsPort),
		HealthProbeBindAddress: fmt.Sprintf(":%d", options.HealthProbePort),
	}
}

// GetServerOptions returns a set of options used by the Tink API.
// These options include leader election disabled.
func GetServerOptions() controllerruntime.Options {
	return controllerruntime.Options{
		Logger:                 zapr.NewLogger(logging.FromContext(context.Background()).Desugar()),
		LeaderElection:         false,
		Scheme:                 runtimescheme,
		MetricsBindAddress:     fmt.Sprintf(":%d", options.MetricsPort),
		HealthProbeBindAddress: fmt.Sprintf(":%d", options.HealthProbePort),
	}
}

// NewManagerOrDie instantiates a controller manager.
func NewManagerOrDie(config *rest.Config, options controllerruntime.Options) Manager {
	m, err := NewManager(config, options)
	if err != nil {
		panic(err)
	}
	return m
}

// NewManager instantiates a controller manager.
func NewManager(config *rest.Config, options controllerruntime.Options) (Manager, error) {
	m, err := controllerruntime.NewManager(config, options)
	if err != nil {
		return nil, err
	}
	indexers := []struct {
		obj          client.Object
		field        string
		extractValue client.IndexerFunc
	}{
		{
			&v1alpha1.Workflow{},
			WorkerAddr,
			wokerIndexFunc,
		},
	}
	for _, indexer := range indexers {
		if err := m.GetFieldIndexer().IndexField(
			context.Background(),
			indexer.obj,
			indexer.field,
			indexer.extractValue,
		); err != nil {
			return nil, fmt.Errorf("failed to setup %s indexer, %w", indexer.field, err)
		}
	}
	return &GenericControllerManager{Manager: m}, nil
}

// GenericControllerManager is a manager.Manager that allows for registering of controllers.
type GenericControllerManager struct {
	manager.Manager
}

// RegisterControllers registers a set of controllers to the controller manager.
func (m *GenericControllerManager) RegisterControllers(ctx context.Context, controllers ...Controller) Manager {
	for _, c := range controllers {
		if err := c.Register(ctx, m); err != nil {
			panic(err)
		}
	}
	if err := m.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		panic(fmt.Sprintf("Failed to add readiness probe, %s", err.Error()))
	}
	return m
}

// workerIndex func returns a list of worker addresses from a workflow.
func wokerIndexFunc(obj client.Object) []string {
	wf, ok := obj.(*v1alpha1.Workflow)
	if !ok {
		return nil
	}
	resp := []string{}
	for _, task := range wf.Status.Tasks {
		if task.WorkerAddr != "" {
			resp = append(resp, task.WorkerAddr)
		}
	}
	return resp
}
