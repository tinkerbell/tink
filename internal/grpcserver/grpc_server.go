package grpcserver

import (
	"context"
	"net"

	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Registrar is an interface for registering APIs on a gRPC server.
type Registrar interface {
	Register(*grpc.Server)
}

// SetupGRPC opens a listener and serves a given Registrar's APIs on a gRPC server and returns the listener's address or an error.
func SetupGRPC(ctx context.Context, r Registrar, listenAddr string, errCh chan<- error) (string, error) {
	params := []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.UnaryInterceptor(grpcprometheus.UnaryServerInterceptor),
		grpc.StreamInterceptor(grpcprometheus.StreamServerInterceptor),
	}

	// register servers
	s := grpc.NewServer(params...)
	r.Register(s)
	reflection.Register(s)
	grpcprometheus.Register(s)

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
