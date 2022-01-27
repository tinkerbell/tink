package client

import (
	"log"

	"github.com/packethost/pkg/env"
	"github.com/pkg/errors"
	"github.com/tinkerbell/tink/protos/hardware"
	"github.com/tinkerbell/tink/protos/template"
	"github.com/tinkerbell/tink/protos/workflow"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// gRPC clients.
var (
	TemplateClient template.TemplateServiceClient
	WorkflowClient workflow.WorkflowServiceClient
	HardwareClient hardware.HardwareServiceClient
)

// FullClient aggregates all the gRPC clients available from Tinkerbell Server.
type FullClient struct {
	TemplateClient template.TemplateServiceClient
	WorkflowClient workflow.WorkflowServiceClient
	HardwareClient hardware.HardwareServiceClient
}

// NewFullClient returns a FullClient. A structure that contains all the
// clients made available from tink-server.
func NewFullClient(conn grpc.ClientConnInterface) *FullClient {
	return &FullClient{
		TemplateClient: template.NewTemplateServiceClient(conn),
		WorkflowClient: workflow.NewWorkflowServiceClient(conn),
		HardwareClient: hardware.NewHardwareServiceClient(conn),
	}
}

type ConnOptions struct {
	GRPCAuthority string
	TLS           bool
}

func NewClientConn(opt *ConnOptions) (*grpc.ClientConn, error) {
	var creds grpc.DialOption
	if opt.TLS {
		creds = grpc.WithTransportCredentials(credentials.NewTLS(nil))
	} else {
		creds = grpc.WithInsecure()
	}

	conn, err := grpc.Dial(opt.GRPCAuthority,
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
	opts := ConnOptions{
		GRPCAuthority: env.Get("TINKERBELL_GRPC_AUTHORITY"),
		TLS:           env.Bool("TINKERBELL_TLS", true),
	}

	if opts.GRPCAuthority == "" {
		return nil, errors.New("undefined TINKERBELL_GRPC_AUTHORITY")
	}

	return NewClientConn(&opts)
}

// Setup : create a connection to server.
func Setup() error {
	conn, err := GetConnection()
	if err != nil {
		return err
	}
	TemplateClient = template.NewTemplateServiceClient(conn)
	WorkflowClient = workflow.NewWorkflowServiceClient(conn)
	HardwareClient = hardware.NewHardwareServiceClient(conn)
	return nil
}

// TinkHardwareClient creates a new hardware client.
func TinkHardwareClient() (hardware.HardwareServiceClient, error) {
	conn, err := GetConnection()
	if err != nil {
		log.Fatal(err)
	}
	return hardware.NewHardwareServiceClient(conn), nil
}

// TinkTemplateClient creates a new hardware client.
func TinkTemplateClient() (template.TemplateServiceClient, error) {
	conn, err := GetConnection()
	if err != nil {
		log.Fatal(err)
	}
	return template.NewTemplateServiceClient(conn), nil
}

// TinkWorkflowClient creates a new workflow client.
func TinkWorkflowClient() (workflow.WorkflowServiceClient, error) {
	conn, err := GetConnection()
	if err != nil {
		log.Fatal(err)
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
		HardwareClient: hardware.NewHardwareServiceClient(conn),
		TemplateClient: template.NewTemplateServiceClient(conn),
		WorkflowClient: workflow.NewWorkflowServiceClient(conn),
	}, nil
}
