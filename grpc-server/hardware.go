package grpcserver

import (
	"context"
	"encoding/json"
	"time"

	"github.com/tinkerbell/tink/db"
	"github.com/tinkerbell/tink/metrics"
	"github.com/tinkerbell/tink/protos/hardware"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) Push(ctx context.Context, in *hardware.PushRequest) (*hardware.Empty, error) {
	logger.Info("push")
	labels := prometheus.Labels{"method": "Push", "op": ""}
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	// must be a copy so deferred cacheInFlight.Dec matches the Inc
	labels = prometheus.Labels{"method": "Push", "op": ""}

	var h struct {
		ID    string
		State string
	}
	err := json.Unmarshal([]byte(in.Data), &h)
	if err != nil {
		metrics.CacheTotals.With(labels).Inc()
		metrics.CacheErrors.With(labels).Inc()
		err = errors.Wrap(err, "unmarshal json")
		logger.Error(err)
		return &hardware.Empty{}, err
	}

	if h.ID == "" {
		metrics.CacheTotals.With(labels).Inc()
		metrics.CacheErrors.With(labels).Inc()
		err = errors.New("id must be set to a UUID")
		logger.Error(err)
		return &hardware.Empty{}, err
	}

	logger.With("id", h.ID).Info("data pushed")

	var fn func() error
	msg := ""
	if h.State != "deleted" {
		labels["op"] = "insert"
		msg = "inserting into DB"
		fn = func() error { return db.InsertIntoDB(ctx, s.db, in.Data) }
	} else {
		msg = "deleting from DB"
		labels["op"] = "delete"
		fn = func() error { return db.DeleteFromDB(ctx, s.db, h.ID) }
	}

	metrics.CacheTotals.With(labels).Inc()
	timer := prometheus.NewTimer(metrics.CacheDuration.With(labels))
	defer timer.ObserveDuration()

	logger.Info(msg)
	err = fn()
	logger.Info("done " + msg)
	if err != nil {
		metrics.CacheErrors.With(labels).Inc()
		l := logger
		if pqErr := db.Error(err); pqErr != nil {
			l = l.With("detail", pqErr.Detail, "where", pqErr.Where)
		}
		l.Error(err)
	}

	s.watchLock.RLock()
	if ch := s.watch[h.ID]; ch != nil {
		select {
		case ch <- in.Data:
		default:
			metrics.WatchMissTotal.Inc()
			logger.With("id", h.ID).Info("skipping blocked watcher")
		}
	}
	s.watchLock.RUnlock()

	return &hardware.Empty{}, err
}

func (s *server) Ingest(ctx context.Context, in *hardware.Empty) (*hardware.Empty, error) {
	logger.Info("ingest")
	labels := prometheus.Labels{"method": "Ingest", "op": ""}
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	logger.Info("Ingest called but is deprecated")

	return &hardware.Empty{}, nil
}

func (s *server) by(method string, fn func() (string, error)) (*hardware.Hardware, error) {
	labels := prometheus.Labels{"method": method, "op": "get"}

	metrics.CacheTotals.With(labels).Inc()
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	timer := prometheus.NewTimer(metrics.CacheDuration.With(labels))
	defer timer.ObserveDuration()
	j, err := fn()
	if err != nil {
		metrics.CacheErrors.With(labels).Inc()
		return &hardware.Hardware{}, err
	}

	if j == "" {
		s.dbLock.RLock()
		ready := s.dbReady
		s.dbLock.RUnlock()
		if !ready {
			metrics.CacheStalls.With(labels).Inc()
			return &hardware.Hardware{}, errors.New("DB is not ready")
		}
	}

	metrics.CacheHits.With(labels).Inc()
	return &hardware.Hardware{JSON: j}, nil
}

func (s *server) ByMAC(ctx context.Context, in *hardware.GetRequest) (*hardware.Hardware, error) {
	return s.by("ByMAC", func() (string, error) {
		return db.GetByMAC(ctx, s.db, in.MAC)
	})
}

func (s *server) ByIP(ctx context.Context, in *hardware.GetRequest) (*hardware.Hardware, error) {
	return s.by("ByIP", func() (string, error) {
		return db.GetByIP(ctx, s.db, in.IP)
	})
}

// ByID implements hardware.ByID
func (s *server) ByID(ctx context.Context, in *hardware.GetRequest) (*hardware.Hardware, error) {
	return s.by("ByID", func() (string, error) {
		return db.GetByID(ctx, s.db, in.ID)
	})
}

// ALL implements hardware.All
func (s *server) All(_ *hardware.Empty, stream hardware.HardwareService_AllServer) error {
	labels := prometheus.Labels{"method": "All", "op": "get"}

	metrics.CacheTotals.With(labels).Inc()
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	s.dbLock.RLock()
	ready := s.dbReady
	s.dbLock.RUnlock()
	if !ready {
		metrics.CacheStalls.With(labels).Inc()
		return errors.New("DB is not ready")
	}

	timer := prometheus.NewTimer(metrics.CacheDuration.With(labels))
	defer timer.ObserveDuration()
	err := db.GetAll(s.db, func(j string) error {
		return stream.Send(&hardware.Hardware{JSON: j})
	})
	if err != nil {
		metrics.CacheErrors.With(labels).Inc()
		return err
	}

	metrics.CacheHits.With(labels).Inc()
	return nil
}

func (s *server) Watch(in *hardware.GetRequest, stream hardware.HardwareService_WatchServer) error {
	l := logger.With("id", in.ID)

	ch := make(chan string, 1)
	s.watchLock.Lock()
	old, ok := s.watch[in.ID]
	if ok {
		l.Info("evicting old watch")
		close(old)
	}
	s.watch[in.ID] = ch
	s.watchLock.Unlock()

	labels := prometheus.Labels{"method": "Watch", "op": "push"}
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	disconnect := true
	defer func() {
		if !disconnect {
			return
		}
		s.watchLock.Lock()
		delete(s.watch, in.ID)
		s.watchLock.Unlock()
		close(ch)
	}()

	hw := &hardware.Hardware{}
	for {
		select {
		case <-s.quit:
			l.Info("server is shutting down")
			return status.Error(codes.OK, "server is shutting down")
		case <-stream.Context().Done():
			l.Info("client disconnected")
			return status.Error(codes.OK, "client disconnected")
		case j, ok := <-ch:
			if !ok {
				disconnect = false
				l.Info("we are being evicted, goodbye")
				// ch was replaced and already closed
				return status.Error(codes.Unknown, "evicted")
			}

			hw.Reset()
			hw.JSON = j
			err := stream.Send(hw)
			if err != nil {
				metrics.CacheErrors.With(labels).Inc()
				err = errors.Wrap(err, "stream send")
				l.Error(err)
				return err
			}
		}
	}
}

// Cert returns the public cert that can be served to clients
func (s *server) Cert() []byte {
	return s.cert
}

// ModTime returns the modified-time of the grpc cert
func (s *server) ModTime() time.Time {
	return s.modT
}
