package grpcserver

import (
	"context"
	"crypto/tls"
	"database/sql"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/packethost/pkg/log"
	"github.com/packethost/tinkerbell/db"
	"github.com/packethost/tinkerbell/metrics"
	"github.com/packethost/tinkerbell/protos/hardware"
	"github.com/packethost/tinkerbell/protos/target"
	"github.com/packethost/tinkerbell/protos/template"
	"github.com/packethost/tinkerbell/protos/workflow"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	logger         log.Logger
	grpcListenAddr = ":42113"
)

// Server is the gRPC server for rover
type server struct {
	cert []byte
	modT time.Time

	db   *sql.DB
	quit <-chan struct{}

	dbLock  sync.RWMutex
	dbReady bool

	watchLock sync.RWMutex
	watch     map[string]chan string
}

// SetupGRPC setup and return a gRPC server
func SetupGRPC(ctx context.Context, log log.Logger, facility string, errCh chan<- error) ([]byte, time.Time) {
	params := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	}
	logger = log
	metrics.SetupMetrics(facility, logger)
	db := db.ConnectDB(logger)
	server := &server{
		db:      db,
		dbReady: true,
	}
	if cert := os.Getenv("ROVER_TLS_CERT"); cert != "" {
		server.cert = []byte(cert)
		server.modT = time.Now()
	} else {
		tlsCert, certPEM, modT := getCerts(facility, logger)
		params = append(params, grpc.Creds(credentials.NewServerTLSFromCert(&tlsCert)))
		server.cert = certPEM
		server.modT = modT
	}

	// register servers
	s := grpc.NewServer(params...)
	template.RegisterTemplateServer(s, server)
	target.RegisterTargetServer(s, server)
	workflow.RegisterWorkflowSvcServer(s, server)
	hardware.RegisterHardwareServiceServer(s, server)

	grpc_prometheus.Register(s)

	go func() {
		logger.Info("serving grpc")
		lis, err := net.Listen("tcp", grpcListenAddr)
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

	certsDir := os.Getenv("ROVER_CERTS_DIR")
	if certsDir == "" {
		certsDir = "/certs/" + facility
	}
	if !strings.HasSuffix(certsDir, "/") {
		certsDir += "/"
	}

	certFile, err := os.Open(certsDir + "bundle.pem")
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
	keyPEM, err := ioutil.ReadFile(certsDir + "server-key.pem")
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
