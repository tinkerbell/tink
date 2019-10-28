package client

import (
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/packethost/rover/protos/target"
	"github.com/packethost/rover/protos/template"
	"github.com/packethost/rover/protos/workflow"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// gRPC clients
var (
	TemplateClient template.TemplateClient
	TargetClient   target.TargetClient
	WorkflowClient workflow.WorkflowSvcClient
)

func getConnection() (*grpc.ClientConn, error) {
	certURL := os.Getenv("ROVER_CERT_URL")
	if certURL == "" {
		return nil, errors.New("undefined ROVER_CERT_URL")
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

	grpcAuthority := os.Getenv("ROVER_GRPC_AUTHORITY")
	if grpcAuthority == "" {
		return nil, errors.New("undefined ROVER_GRPC_AUTHORITY")
	}
	creds := credentials.NewClientTLSFromCert(cp, "")
	conn, err := grpc.Dial(grpcAuthority, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, errors.Wrap(err, "connect to rover server")
	}
	return conn, nil
}

func init() {
	conn, err := getConnection()
	if err != nil {
		log.Fatal(err)
	}
	TemplateClient = template.NewTemplateClient(conn)
	TargetClient = target.NewTargetClient(conn)
	WorkflowClient = workflow.NewWorkflowSvcClient(conn)
}
