package server

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/tinkerbell/tink/api/v1alpha1"
	"github.com/tinkerbell/tink/internal/controller"
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

// +kubebuilder:rbac:groups=tinkerbell.org,resources=hardware;hardware/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=tinkerbell.org,resources=templates;templates/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=tinkerbell.org,resources=workflows;workflows/status,verbs=get;list;watch;update;patch

// NewKubeBackedServer returns a server that implements the Workflow server interface for a given kubeconfig.
func NewKubeBackedServer(logger logr.Logger, kubeconfig, apiserver, namespace string) (*KubernetesBackedServer, error) {
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

	namespace, _, err = ccfg.Namespace()
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return NewKubeBackedServerFromREST(logger, cfg, namespace)
}

// NewKubeBackedServerFromREST returns a server that implements the Workflow
// server interface with the given Kubernetes rest client and namespace.
func NewKubeBackedServerFromREST(logger logr.Logger, config *rest.Config, namespace string) (*KubernetesBackedServer, error) {
	clstr, err := cluster.New(config, func(opts *cluster.Options) {
		opts.Scheme = controller.DefaultScheme()
		opts.Logger = zapr.NewLogger(zap.NewNop())
		opts.Cache.DefaultNamespaces = map[string]cache.Config{
			namespace: {},
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

	return &KubernetesBackedServer{
		logger:     logger,
		ClientFunc: clstr.GetClient,
		namespace:  namespace,
		nowFunc:    time.Now,
	}, nil
}

// KubernetesBackedServer is a server that implements a workflow API.
type KubernetesBackedServer struct {
	logger     logr.Logger
	ClientFunc func() client.Client
	namespace  string

	nowFunc func() time.Time
}

// Register registers the service on the gRPC server.
func (s *KubernetesBackedServer) Register(server *grpc.Server) {
	proto.RegisterWorkflowServiceServer(server, s)
}
