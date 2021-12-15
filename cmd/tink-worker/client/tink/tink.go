package tink

import (
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ObtainServerCreds obtains tink-server credentials from the TLS certificate found at certURL.
func ObtainServerCreds(certURL string) (credentials.TransportCredentials, error) {
	// fetch the cert
	if certURL == "" {
		return nil, errors.New("certURL cannot be empty")
	}
	resp, err := http.Get(certURL)
	if err != nil {
		errMsg := fmt.Sprintf("unable to fetch cert from %s", certURL)
		return nil, errors.New(errMsg)
	}
	defer resp.Body.Close()

	// read the cert
	certs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("unable to read the cert data")
	}

	// parse the cert
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(certs)
	if !ok {
		return nil, errors.New("unable to parse the cert data")
	}

	// generate credentials from the certPool
	creds := credentials.NewClientTLSFromCert(certPool, "")

	return creds, err
}

// EstablishServerConnection returns a GRPC client connection to tink-server.
func EstablishServerConnection(grpcAuthority string, creds credentials.TransportCredentials) (*grpc.ClientConn, error) {
	// use the cert creds to connect to the server
	if grpcAuthority == "" {
		return nil, errors.New("grpcAuthority cannot be empty")
	}

	conn, err := grpc.Dial(grpcAuthority, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, errors.New("connect to tinkerbell server")
	}

	return conn, nil
}
