package client

import (
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/tinkerbell/tink/protos/events"
	"github.com/tinkerbell/tink/protos/hardware"
	"github.com/tinkerbell/tink/protos/template"
	"github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// gRPC clients
var (
	TemplateClient template.TemplateServiceClient
	WorkflowClient workflow.WorkflowServiceClient
	HardwareClient hardware.HardwareServiceClient
	EventsClient   events.EventsServiceClient
)

// FullClient aggregates all the grpc clients available from Tinkerbell Server
type FullClient struct {
	TemplateClient template.TemplateServiceClient
	WorkflowClient workflow.WorkflowServiceClient
	HardwareClient hardware.HardwareServiceClient
	EventsClient   events.EventsServiceClient
}

// NewFullClientFromGlobal is a dirty hack that returns a FullClient using the
// global variables exposed by the client package. Globals should be avoided
// and we will deprecated them at some point replacing this function with
// NewFullClient. If you are strating a new project please use the last one
func NewFullClientFromGlobal() (*FullClient, error) {
	// This is required because we use init() too often, even more in the
	// CLI and based on where you are sometime the clients are not initialised
	if TemplateClient == nil {
		err := Setup()
		if err != nil {
			panic(err)
		}
	}
	return &FullClient{
		TemplateClient: TemplateClient,
		WorkflowClient: WorkflowClient,
		HardwareClient: HardwareClient,
		EventsClient:   EventsClient,
	}, nil
}

// NewFullClient returns a FullClient. A structure that contains all the
// clients made available from tink-server. This is the function you should use
// instead of NewFullClientFromGlobal that will be deprecated soon
func NewFullClient(conn grpc.ClientConnInterface) *FullClient {
	return &FullClient{
		TemplateClient: template.NewTemplateServiceClient(conn),
		WorkflowClient: workflow.NewWorkflowServiceClient(conn),
		HardwareClient: hardware.NewHardwareServiceClient(conn),
		EventsClient:   events.NewEventsServiceClient(conn),
	}
}

// GetConnection returns a gRPC client connection
func GetConnection() (*grpc.ClientConn, error) {
	certURL := os.Getenv("TINKERBELL_CERT_URL")
	if certURL == "" {
		return nil, errors.New("undefined TINKERBELL_CERT_URL")
	}
	resp, err := http.Get(certURL)
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

	grpcAuthority := os.Getenv("TINKERBELL_GRPC_AUTHORITY")
	if grpcAuthority == "" {
		return nil, errors.New("undefined TINKERBELL_GRPC_AUTHORITY")
	}
	creds := credentials.NewClientTLSFromCert(cp, "")
	conn, err := grpc.Dial(grpcAuthority, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, errors.Wrap(err, "connect to tinkerbell server")
	}
	return conn, nil
}

// Setup : create a connection to server
func Setup() error {
	conn, err := GetConnection()
	if err != nil {
		return err
	}
	TemplateClient = template.NewTemplateServiceClient(conn)
	WorkflowClient = workflow.NewWorkflowServiceClient(conn)
	HardwareClient = hardware.NewHardwareServiceClient(conn)
	EventsClient = events.NewEventsServiceClient(conn)
	return nil
}

// TinkHardwareClient creates a new hardware client
func TinkHardwareClient() (hardware.HardwareServiceClient, error) {
	conn, err := GetConnection()
	if err != nil {
		log.Fatal(err)
	}
	return hardware.NewHardwareServiceClient(conn), nil
}

// TinkWorkflowClient creates a new workflow client
func TinkWorkflowClient() (workflow.WorkflowServiceClient, error) {
	conn, err := GetConnection()
	if err != nil {
		log.Fatal(err)
	}
	return workflow.NewWorkflowServiceClient(conn), nil
}
