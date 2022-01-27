package client

import (
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"

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
	CertURL       string
	GRPCAuthority string
	TLS           bool
}

// This function is bad and ideally should be removed, but for now it moves all the bad into one place.
// This is the legacy of packethost/cacher running behind an ingress that couldn't terminate TLS on behalf
// of GRPC. All of this functionality should be ripped out in favor of either using trusted certificates
// or moving the establishment of trust in the certificate out to the environment (or running in no-tls mode
// e.g. for development.)
func grpcCredentialFromCertEndpoint(url string) (credentials.TransportCredentials, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "fetch cert")
	}
	defer resp.Body.Close()

	certs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read cert")
	}

	cp := x509.NewCertPool()
	ok := cp.AppendCertsFromPEM(certs)
	if !ok {
		return nil, errors.Wrap(err, "parse cert")
	}

	return credentials.NewClientTLSFromCert(cp, ""), nil
}

func NewClientConn(opt *ConnOptions) (*grpc.ClientConn, error) {
	method := grpc.WithInsecure()
	if opt.TLS {
		creds, err := grpcCredentialFromCertEndpoint(opt.CertURL)
		if err != nil {
			return nil, err
		}
		method = grpc.WithTransportCredentials(creds)
	}
	conn, err := grpc.Dial(opt.GRPCAuthority,
		method,
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
		CertURL:       env.Get("TINKERBELL_CERT_URL"),
		GRPCAuthority: env.Get("TINKERBELL_GRPC_AUTHORITY"),
		TLS:           env.Bool("TINKERBELL_TLS", true),
	}

	if opts.GRPCAuthority == "" {
		return nil, errors.New("undefined TINKERBELL_GRPC_AUTHORITY")
	}

	if opts.TLS {
		if opts.CertURL == "" {
			return nil, errors.New("undefined TINKERBELL_CERT_URL")
		}
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
