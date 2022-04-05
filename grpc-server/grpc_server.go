package grpcserver

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GetCerts returns a TLS certificate, PEM bytes, and file modification time for a
// given path. An error is returned for any failure.
//
// The public key is expected to be named "bundle.pem" and the private key
// "server.pem".
func GetCerts(certsDir string) (*tls.Certificate, []byte, *time.Time, error) {
	certFile, err := os.Open(filepath.Join(certsDir, "bundle.pem"))
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to open TLS cert")
	}

	stat, err := certFile.Stat()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to stat TLS cert")
	}
	modT := stat.ModTime()
	certPEM, err := ioutil.ReadAll(certFile)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to read TLS cert")
	}
	err = certFile.Close()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to close TLS cert")
	}

	keyPEM, err := ioutil.ReadFile(filepath.Join(certsDir, "server-key.pem"))
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to read TLS key")
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to parse TLS file content")
	}

	return &cert, certPEM, &modT, nil
}

// Registrar is an interface for registering APIs on a gRPC server.
type Registrar interface {
	Register(*grpc.Server)
}

// SetupGRPC opens a listener and serves a given Registrar's APIs on a gRPC server
// and returns the listener's address or an error.
func SetupGRPC(ctx context.Context, r Registrar, listenAddr string, opts []grpc.ServerOption, errCh chan<- error) (serverAddr string, err error) {
	params := []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(grpc_prometheus.UnaryServerInterceptor, otelgrpc.UnaryServerInterceptor()),
		grpc_middleware.WithStreamServerChain(grpc_prometheus.StreamServerInterceptor, otelgrpc.StreamServerInterceptor()),
	}
	params = append(params, opts...)

	// register servers
	s := grpc.NewServer(params...)
	r.Register(s)
	reflection.Register(s)
	grpc_prometheus.Register(s)

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return "", errors.Wrap(err, "failed to listen")
	}

	go func(errChan chan<- error) {
		errChan <- s.Serve(lis)
	}(errCh)

	go func(ctx context.Context, s *grpc.Server) {
		<-ctx.Done()
		s.GracefulStop()
	}(ctx, s)

	return lis.Addr().String(), nil
}
