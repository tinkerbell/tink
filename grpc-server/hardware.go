package grpcserver

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tinkerbell/tink/db"
	"github.com/tinkerbell/tink/metrics"
	"github.com/tinkerbell/tink/protos/hardware"
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

	hw := in.Data
	if hw.Id == "" {
		metrics.CacheTotals.With(labels).Inc()
		metrics.CacheErrors.With(labels).Inc()
		err := errors.New("id must be set to a UUID, got id: " + hw.Id)
		logger.Error(err)
		return &hardware.Empty{}, err
	}

	// TODO: somewhere here validate json (if ip addr contains cidr, etc.)

	logger.With("id", hw.Id).Info("data pushed")

	var fn func() error
	msg := ""
	data, err := json.Marshal(hw)
	if err != nil {
		logger.Error(err)
	}
	if hw.Metadata.State != "deleted" {
		labels["op"] = "insert"
		msg = "inserting into DB"
		fn = func() error { return db.InsertIntoDB(ctx, s.db, string(data)) }
	} else {
		msg = "deleting from DB"
		labels["op"] = "delete"
		fn = func() error { return db.DeleteFromDB(ctx, s.db, hw.Id) }
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
	if ch := s.watch[hw.Id]; ch != nil {
		select {
		case ch <- string(data):
		default:
			metrics.WatchMissTotal.Inc()
			logger.With("id", hw.Id ).Info("skipping blocked watcher")
		}
	}
	s.watchLock.RUnlock()

	return &hardware.Empty{}, err
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
	hw := &hardware.Hardware{}
	json.Unmarshal([]byte(j), hw)
	return hw, nil
}

func (s *server) ByMAC(ctx context.Context, in *hardware.GetRequest) (*hardware.Hardware, error) {
	return s.by("ByMAC", func() (string, error) {
		return db.GetByMAC(ctx, s.db, in.Mac)
	})
}

func (s *server) ByIP(ctx context.Context, in *hardware.GetRequest) (*hardware.Hardware, error) {
	return s.by("ByIP", func() (string, error) {
		return db.GetByIP(ctx, s.db, in.Ip)
	})
}

// ByID implements hardware.ByID
func (s *server) ByID(ctx context.Context, in *hardware.GetRequest) (*hardware.Hardware, error) {
	return s.by("ByID", func() (string, error) {
		return db.GetByID(ctx, s.db, in.Id)
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
	err := db.GetAll(s.db, func(j []byte) error {
		hw := &hardware.Hardware{}
		json.Unmarshal(j, hw)
		return stream.Send(hw)
	})
	if err != nil {
		metrics.CacheErrors.With(labels).Inc()
		return err
	}

	metrics.CacheHits.With(labels).Inc()
	return nil
}

func (s *server) Watch(in *hardware.GetRequest, stream hardware.HardwareService_WatchServer) error {
	l := logger.With("id", in.Id)

	ch := make(chan string, 1)
	s.watchLock.Lock()
	old, ok := s.watch[in.Id]
	if ok {
		l.Info("evicting old watch")
		close(old)
	}
	s.watch[in.Id] = ch
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
		delete(s.watch, in.Id)
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
			json.Unmarshal([]byte(j), hw)
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

func (s *server) Delete(ctx context.Context, in *hardware.DeleteRequest) (*hardware.Empty, error) {
	logger.Info("delete")
	labels := prometheus.Labels{"method": "Delete", "op": ""}
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	// must be a copy so deferred cacheInFlight.Dec matches the Inc
	labels = prometheus.Labels{"method": "Delete", "op": ""}

	if in.ID == "" {
		metrics.CacheTotals.With(labels).Inc()
		metrics.CacheErrors.With(labels).Inc()
		err := errors.New("id must be set to a UUID")
		logger.Error(err)
		return &hardware.Empty{}, err
	}

	logger.With("id", in.ID).Info("data deleted")

	var fn func() error
	labels["op"] = "delete"
	msg := "deleting into DB"
	fn = func() error { return db.DeleteFromDB(ctx, s.db, in.ID) }

	metrics.CacheTotals.With(labels).Inc()
	timer := prometheus.NewTimer(metrics.CacheDuration.With(labels))
	defer timer.ObserveDuration()

	logger.Info(msg)
	err := fn()
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
	if ch := s.watch[in.ID]; ch != nil {
		select {
		case ch <- in.ID:
		default:
			metrics.WatchMissTotal.Inc()
			logger.With("id", in.ID).Info("skipping blocked watcher")
		}
	}
	s.watchLock.RUnlock()

	return &hardware.Empty{}, err
}
