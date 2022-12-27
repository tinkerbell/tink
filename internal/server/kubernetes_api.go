package server

import (
	"context"
	"time"

	"github.com/packethost/pkg/log"
	"github.com/tinkerbell/tink/pkg/controllers"
	pb "github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/grpc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:rbac:groups=tinkerbell.org,resources=hardware;hardware/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=tinkerbell.org,resources=templates;templates/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=tinkerbell.org,resources=workflows;workflows/status,verbs=get;list;watch;update;patch

// NewKubeBackedServer returns a server that implements the Workflow server interface for a given kubeconfig.
func NewKubeBackedServer(logger log.Logger, kubeconfig, apiserver, namespace string) (*KubernetesBackedServer, error) {
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
	return NewKubeBackedServerFromREST(logger, cfg, namespace), nil
}

// NewKubeBackedServerFromREST returns a server that implements the Workflow
// server interface with the given Kubernetes rest client and namespace.
func NewKubeBackedServerFromREST(logger log.Logger, config *rest.Config, namespace string) *KubernetesBackedServer {
	options := controllers.GetServerOptions()
	options.Namespace = namespace
	manager := controllers.NewManagerOrDie(config, options)
	go func() {
		err := manager.Start(context.Background())
		if err != nil {
			logger.Error(err, "Error starting manager")
		}
	}()
	return &KubernetesBackedServer{
		logger:     logger,
		ClientFunc: manager.GetClient,
		namespace:  namespace,
		nowFunc:    time.Now,
	}
}

// KubernetesBackedServer is a server that implements a workflow API.
type KubernetesBackedServer struct {
	logger     log.Logger
	ClientFunc func() client.Client
	namespace  string

	nowFunc func() time.Time
}

// Register registers the service on the gRPC server.
func (s *KubernetesBackedServer) Register(server *grpc.Server) {
	pb.RegisterWorkflowServiceServer(server, s)
}
