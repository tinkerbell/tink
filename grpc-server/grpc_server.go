package grpcserver

import (
	"context"
	"net"
	"sync"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/packethost/pkg/log"
	"github.com/pkg/errors"
	"github.com/tinkerbell/tink/db"
	"github.com/tinkerbell/tink/metrics"
	"github.com/tinkerbell/tink/protos/hardware"
	"github.com/tinkerbell/tink/protos/template"
	"github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

// Server is the gRPC server for tinkerbell.
type server struct {
	db   db.Database
	quit <-chan struct{}

	dbLock  sync.RWMutex
	dbReady bool

	watchLock sync.RWMutex
	watch     map[string]chan string

	logger log.Logger
}

type ConfigGRPCServer struct {
	Facility      string
	TLSCert       string
	TLSKey        string
	GRPCAuthority string
	DB            *db.TinkDB
}

// SetupGRPC setup and return a gRPC server.
func SetupGRPC(ctx context.Context, logger log.Logger, config *ConfigGRPCServer, errCh chan<- error) error {
	creds, err := credentials.NewServerTLSFromFile(config.TLSCert, config.TLSKey)
	if err != nil {
		return err
	}

	params := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.Creds(creds),
	}

	metrics.SetupMetrics(config.Facility, logger)
	server := &server{
		db:      config.DB,
		dbReady: true,
		logger:  logger,
	}

	// register servers
	s := grpc.NewServer(params...)
	template.RegisterTemplateServiceServer(s, server)
	workflow.RegisterWorkflowServiceServer(s, server)
	hardware.RegisterHardwareServiceServer(s, server)
	reflection.Register(s)

	grpc_prometheus.Register(s)

	lis, err := net.Listen("tcp", config.GRPCAuthority)
	if err != nil {
		return errors.Wrap(err, "failed to listen")
	}

	go func() {
		errCh <- s.Serve(lis)
	}()

	go func() {
		<-ctx.Done()
		s.GracefulStop()
	}()

	return nil
}
