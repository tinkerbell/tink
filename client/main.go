package client

import (
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"os"

<<<<<<< HEAD
	"github.com/packethost/cacher/protos/targets"
	"github.com/packethost/cacher/protos/template"
	"github.com/packethost/cacher/protos/workflow"
=======
	"github.com/packethost/rover/protos/targets"
	"github.com/packethost/rover/protos/template"
	"github.com/packethost/rover/protos/workflow"
>>>>>>> 77bd68e... Added targets CLI
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// gRPC clients
var (
	TemplateClient template.TemplateClient
	TargetClient   targets.TargetClient
	WorkflowClient workflow.WorkflowClient
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
	WorkflowClient = workflow.NewWorkflowClient(conn)
	TargetClient = targets.NewTargetClient(conn)
}
