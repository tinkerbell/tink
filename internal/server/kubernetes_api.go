package server

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/tinkerbell/tink/api/v1alpha1"
	"github.com/tinkerbell/tink/internal/deprecated/controller"
	"github.com/tinkerbell/tink/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

// Option for setting optional KubernetesBackedServer fields.
type Option func(*KubernetesBackedServer)

// KubernetesBackedServer is a server that implements a workflow API.
type KubernetesBackedServer struct {
	logger                 logr.Logger
	ClientFunc             func() client.Client
	namespace              string
	AutoCapMode            AutoCapMode
	AutoEnrollmentTemplate string

	nowFunc func() time.Time
}

func WithAutoCapMode(mode AutoCapMode) Option {
	return func(k *KubernetesBackedServer) {
		k.AutoCapMode = mode
	}
}

func WithAutoEnrollmentTemplate(name string) Option {
	return func(k *KubernetesBackedServer) {
		k.AutoEnrollmentTemplate = name
	}
}

// +kubebuilder:rbac:groups=tinkerbell.org,resources=hardware;hardware/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=tinkerbell.org,resources=templates;templates/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=tinkerbell.org,resources=workflows;workflows/status,verbs=get;list;watch;update;patch

// NewKubeBackedServer returns a server that implements the Workflow server interface for a given kubeconfig.
func NewKubeBackedServer(logger logr.Logger, kubeconfig, apiserver, namespace string, opts ...Option) (*KubernetesBackedServer, error) {
	ccfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{
			ClusterInfo: clientcmdapi.Cluster{
				Server: apiserver,
			},
			Context: clientcmdapi.Context{
				Namespace: namespace,
			},
		},
	)

	cfg, err := ccfg.ClientConfig()
	if err != nil {
		return nil, err
	}

	return NewKubeBackedServerFromREST(logger, cfg, namespace, opts...)
}

// NewKubeBackedServerFromREST returns a server that implements the Workflow
// server interface with the given Kubernetes rest client and namespace.
func NewKubeBackedServerFromREST(logger logr.Logger, config *rest.Config, namespace string, opts ...Option) (*KubernetesBackedServer, error) {
	clstr, err := cluster.New(config, func(opts *cluster.Options) {
		opts.Scheme = controller.DefaultScheme()
		opts.Logger = zapr.NewLogger(zap.NewNop())
		if namespace != "" {
			opts.Cache.DefaultNamespaces = map[string]cache.Config{
				namespace: {},
			}
		}
	})
	if err != nil {
		return nil, fmt.Errorf("init client: %w", err)
	}

	err = clstr.GetFieldIndexer().IndexField(
		context.Background(),
		&v1alpha1.Workflow{},
		workflowByNonTerminalState,
		workflowByNonTerminalStateFunc,
	)
	if err != nil {
		return nil, fmt.Errorf("setup %s index: %w", workflowByNonTerminalState, err)
	}

	go func() {
		err := clstr.Start(context.Background())
		if err != nil {
			logger.Error(err, "Error starting cluster")
		}
	}()

	k := &KubernetesBackedServer{
		logger:     logger,
		ClientFunc: clstr.GetClient,
		nowFunc:    time.Now,
	}
	for _, opt := range opts {
		opt(k)
	}

	return k, nil
}

// Register registers the service on the gRPC server.
func (s *KubernetesBackedServer) Register(server *grpc.Server) {
	proto.RegisterWorkflowServiceServer(server, s)
}
