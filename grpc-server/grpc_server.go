package grpcserver

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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

// Server is the gRPC server for tinkerbell
type server struct {
	cert []byte
	modT time.Time

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
	GRPCAuthority string
	DB            *db.TinkDB
}

// SetupGRPC setup and return a gRPC server
func SetupGRPC(ctx context.Context, logger log.Logger, config *ConfigGRPCServer, errCh chan<- error) ([]byte, time.Time) {
	params := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	}
	metrics.SetupMetrics(config.Facility, logger)
	server := &server{
		db:      config.DB,
		dbReady: true,
		logger:  logger,
	}
	if cert := config.TLSCert; cert != "" {
		server.cert = []byte(cert)
		server.modT = time.Now()
	} else {
		tlsCert, certPEM, modT := getCerts(config.Facility, logger)
		params = append(params, grpc.Creds(credentials.NewServerTLSFromCert(&tlsCert)))
		server.cert = certPEM
		server.modT = modT
	}

	// register servers
	s := grpc.NewServer(params...)
	template.RegisterTemplateServiceServer(s, server)
	workflow.RegisterWorkflowServiceServer(s, server)
	hardware.RegisterHardwareServiceServer(s, server)
	reflection.Register(s)

	grpc_prometheus.Register(s)

	go func() {
		lis, err := net.Listen("tcp", config.GRPCAuthority)
		if err != nil {
			err = errors.Wrap(err, "failed to listen")
			logger.Error(err)
			panic(err)
		}

		errCh <- s.Serve(lis)
	}()

	go func() {
		<-ctx.Done()
		s.GracefulStop()
	}()
	return server.cert, server.modT
}

func getCerts(facility string, logger log.Logger) (tls.Certificate, []byte, time.Time) {
	var (
		certPEM []byte
		modT    time.Time
	)

	certsDir := os.Getenv("TINKERBELL_CERTS_DIR")
	if certsDir == "" {
		certsDir = "/certs/" + facility
	}
	if !strings.HasSuffix(certsDir, "/") {
		certsDir += "/"
	}

	certFile, err := os.Open(filepath.Clean(certsDir + "bundle.pem"))
	if err != nil {
		err = errors.Wrap(err, "failed to open TLS cert")
		logger.Error(err)
		panic(err)
	}

	if stat, err := certFile.Stat(); err != nil {
		err = errors.Wrap(err, "failed to stat TLS cert")
		logger.Error(err)
		panic(err)
	} else {
		modT = stat.ModTime()
	}

	certPEM, err = ioutil.ReadAll(certFile)
	if err != nil {
		err = errors.Wrap(err, "failed to read TLS cert")
		logger.Error(err)
		panic(err)
	}
	keyPEM, err := ioutil.ReadFile(filepath.Clean(certsDir + "server-key.pem"))
	if err != nil {
		err = errors.Wrap(err, "failed to read TLS key")
		logger.Error(err)
		panic(err)
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		err = errors.Wrap(err, "failed to ingest TLS files")
		logger.Error(err)
		panic(err)
	}
	return cert, certPEM, modT
}
