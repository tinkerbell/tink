package server

import (
	"sync"
	"time"

	"github.com/packethost/pkg/log"
	"github.com/tinkerbell/tink/db"
	"github.com/tinkerbell/tink/protos/hardware"
	"github.com/tinkerbell/tink/protos/template"
	"github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/grpc"
)

const (
	errInvalidWorkerID       = "invalid worker id"
	errInvalidWorkflowID     = "invalid workflow id"
	errInvalidTaskName       = "invalid task name"
	errInvalidActionName     = "invalid action name"
	errInvalidTaskReported   = "reported task name does not match the current action details"
	errInvalidActionReported = "reported action name does not match the current action details"

	msgReceivedStatus   = "received action status: %s"
	msgCurrentWfContext = "current workflow context"
	msgSendWfContext    = "send workflow context: %s"
)

// DBServerOption is a type for modifying a DBServer.
type DBServerOption func(*DBServer) error

// WithCerts sets a certificate mod time and material on a server.
func WithCerts(modTime time.Time, publicCertPEM []byte) DBServerOption {
	return func(s *DBServer) error {
		s.modT = modTime
		s.cert = publicCertPEM
		return nil
	}
}

// DBServer is a gRPC Server for database-backed Tinkerbell.
type DBServer struct {
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

// NewServer returns a new Tinkerbell server.
func NewDBServer(l log.Logger, database db.Database, opts ...DBServerOption) (*DBServer, error) {
	ts := &DBServer{
		db:      database,
		logger:  l,
		dbReady: true,
	}
	for _, opt := range opts {
		if err := opt(ts); err != nil {
			return nil, err
		}
	}

	return ts, nil
}

// Register registers Template, Workflow, and Hardware APIs on a gRPC server.
func (s *DBServer) Register(gserver *grpc.Server) {
	template.RegisterTemplateServiceServer(gserver, s)
	workflow.RegisterWorkflowServiceServer(gserver, s)
	hardware.RegisterHardwareServiceServer(gserver, s)
}
