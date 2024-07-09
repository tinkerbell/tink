package client

import (
	"crypto/tls"

	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func NewClientConn(authority string, tlsEnabled bool, tlsInsecure bool) (*grpc.ClientConn, error) {
	var creds grpc.DialOption
	if tlsEnabled { // #nosec G402
		creds = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: tlsInsecure}))
	} else {
		creds = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	conn, err := grpc.Dial(authority, creds, grpc.WithStatsHandler(otelgrpc.NewClientHandler()))
	if err != nil {
		return nil, errors.Wrap(err, "dial tinkerbell server")
	}

	return conn, nil
}
