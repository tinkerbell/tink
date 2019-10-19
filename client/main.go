package client

import (
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/packethost/rover/protos/template"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func connectGRPC(client string) {}

// ConnectGRPC returns a rover gRPC client
func ConnectGRPC() template.TemplateClient {
	c, err := new()
	if err != nil {
		panic(err)
	}
	return c
}

func new() (template.TemplateClient, error) {
	conn, err := getConnection()
	if err != nil {
		return nil, err
	}
	return template.NewTemplateClient(conn), nil
}

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
