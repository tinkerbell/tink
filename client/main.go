package client

import (
	"github.com/packethost/pkg/env"
	"github.com/pkg/errors"
	"github.com/tinkerbell/tink/protos/workflow"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// gRPC clients.
var (
	WorkflowClient workflow.WorkflowServiceClient
)

// FullClient aggregates all the gRPC clients available from Tinkerbell Server.
type FullClient struct {
	WorkflowClient workflow.WorkflowServiceClient
}

// NewFullClient returns a FullClient. A structure that contains all the
// clients made available from tink-server.
func NewFullClient(conn grpc.ClientConnInterface) *FullClient {
	return &FullClient{
		WorkflowClient: workflow.NewWorkflowServiceClient(conn),
	}
}

func NewClientConn(authority string, tls bool) (*grpc.ClientConn, error) {
	var creds grpc.DialOption
	if tls {
		creds = grpc.WithTransportCredentials(credentials.NewTLS(nil))
	} else {
		creds = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	conn, err := grpc.Dial(authority,
		creds,
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "connect to tinkerbell server")
	}
	return conn, nil
}

// GetConnection returns a gRPC client connection.
func GetConnection() (*grpc.ClientConn, error) {
	authority := env.Get("TINKERBELL_GRPC_AUTHORITY")
	if authority == "" {
		return nil, errors.New("undefined TINKERBELL_GRPC_AUTHORITY")
	}

	tls := env.Bool("TINKERBELL_TLS", true)
	return NewClientConn(authority, tls)
}

// Setup : create a connection to server.
func Setup() error {
	conn, err := GetConnection()
	if err != nil {
		return err
	}
	WorkflowClient = workflow.NewWorkflowServiceClient(conn)
	return nil
}

// TinkWorkflowClient creates a new workflow client.
func TinkWorkflowClient() (workflow.WorkflowServiceClient, error) {
	conn, err := GetConnection()
	if err != nil {
		return nil, err
	}
	return workflow.NewWorkflowServiceClient(conn), nil
}

// TinkFullClient creates a new full client.
func TinkFullClient() (FullClient, error) {
	conn, err := GetConnection()
	if err != nil {
		return FullClient{}, err
	}
	return FullClient{
		WorkflowClient: workflow.NewWorkflowServiceClient(conn),
	}, nil
}
