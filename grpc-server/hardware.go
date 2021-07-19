package grpcserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tinkerbell/tink/db"
	"github.com/tinkerbell/tink/metrics"
	"github.com/tinkerbell/tink/protos/hardware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	conflictMACAddr = "conflicting hardware MAC address %v provided with hardware data/info"
	duplicateMAC    = "Duplicate MAC address found"
)

func (s *server) Push(ctx context.Context, in *hardware.PushRequest) (*hardware.Empty, error) {
	s.logger.Info("push")
	labels := prometheus.Labels{"method": "Push", "op": ""}
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	// must be a copy so deferred cacheInFlight.Dec matches the Inc
	labels = prometheus.Labels{"method": "Push", "op": ""}

	hw := in.GetData()
	if hw == nil {
		err := errors.New("expected data not to be nil")
		s.logger.Error(err)
		return &hardware.Empty{}, err
	}

	// we know hw is non-nil at this point, since we returned early above
	// if it was nil
	if hw.GetId() == "" {
		metrics.CacheTotals.With(labels).Inc()
		metrics.CacheErrors.With(labels).Inc()
		err := errors.New("id must be set to a UUID, got id: " + hw.Id)
		s.logger.Error(err)
		return &hardware.Empty{}, err
	}

	// normalize data prior to storing in the database
	normalizeHardwareData(hw)

	// validate the hardware data to avoid duplicate mac address
	err := s.validateHardwareData(ctx, hw)
	if err != nil {
		return &hardware.Empty{}, err
	}

	const msg = "inserting into DB"
	data, err := json.Marshal(hw)
	if err != nil {
		s.logger.Error(err)
	}

	labels["op"] = "insert"

	metrics.CacheTotals.With(labels).Inc()
	timer := prometheus.NewTimer(metrics.CacheDuration.With(labels))
	defer timer.ObserveDuration()

	s.logger.Info(msg)
	err = s.db.InsertIntoDB(ctx, string(data))
	s.logger.Info("done " + msg)
	if err != nil {
		metrics.CacheErrors.With(labels).Inc()
		l := s.logger
		if pqErr := db.Error(err); pqErr != nil {
			l = l.With("detail", pqErr.Detail, "where", pqErr.Where)
		}
		l.Error(err)
	}
	s.logger.With("id", hw.Id).Info("data pushed")

	s.watchLock.RLock()
	if ch := s.watch[hw.Id]; ch != nil {
		select {
		case ch <- string(data):
		default:
			metrics.WatchMissTotal.Inc()
		}
	}
	s.watchLock.RUnlock()
	s.logger.With("id", hw.Id).Info("skipping blocked watcher")

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
	if err := json.Unmarshal([]byte(j), hw); err != nil {
		return nil, err
	}
	return hw, nil
}

func (s *server) ByMAC(ctx context.Context, in *hardware.GetRequest) (*hardware.Hardware, error) {
	return s.by("ByMAC", func() (string, error) {
		return s.db.GetByMAC(ctx, in.Mac)
	})
}

func (s *server) ByIP(ctx context.Context, in *hardware.GetRequest) (*hardware.Hardware, error) {
	return s.by("ByIP", func() (string, error) {
		return s.db.GetByIP(ctx, in.Ip)
	})
}

// ByID implements hardware.ByID
func (s *server) ByID(ctx context.Context, in *hardware.GetRequest) (*hardware.Hardware, error) {
	return s.by("ByID", func() (string, error) {
		return s.db.GetByID(ctx, in.Id)
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
	err := s.db.GetAll(func(j []byte) error {
		hw := &hardware.Hardware{}
		if err := json.Unmarshal(j, hw); err != nil {
			return err
		}
		return stream.Send(hw)
	})
	if err != nil {
		metrics.CacheErrors.With(labels).Inc()
		return err
	}

	metrics.CacheHits.With(labels).Inc()
	return nil
}

func (s *server) DeprecatedWatch(in *hardware.GetRequest, stream hardware.HardwareService_DeprecatedWatchServer) error {
	l := s.logger.With("id", in.Id)

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
			if err := json.Unmarshal([]byte(j), hw); err != nil {
				return err
			}
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
	s.logger.Info("delete")
	labels := prometheus.Labels{"method": "Delete", "op": ""}
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	// must be a copy so deferred cacheInFlight.Dec matches the Inc
	labels = prometheus.Labels{"method": "Delete", "op": ""}

	if in.Id == "" {
		metrics.CacheTotals.With(labels).Inc()
		metrics.CacheErrors.With(labels).Inc()
		err := errors.New("id must be set to a UUID")
		s.logger.Error(err)
		return &hardware.Empty{}, err
	}

	s.logger.With("id", in.Id).Info("data deleted")

	labels["op"] = "delete"
	const msg = "deleting into DB"

	metrics.CacheTotals.With(labels).Inc()
	timer := prometheus.NewTimer(metrics.CacheDuration.With(labels))
	defer timer.ObserveDuration()

	s.logger.Info(msg)
	err := s.db.DeleteFromDB(ctx, in.Id)
	s.logger.Info("done " + msg)
	if err != nil {
		metrics.CacheErrors.With(labels).Inc()
		logger := s.logger
		if pqErr := db.Error(err); pqErr != nil {
			logger = s.logger.With("detail", pqErr.Detail, "where", pqErr.Where)
		}
		logger.Error(err)
	}

	s.watchLock.RLock()
	if ch := s.watch[in.Id]; ch != nil {
		select {
		case ch <- in.Id:
		default:
			metrics.WatchMissTotal.Inc()
			s.logger.With("id", in.Id).Info("skipping blocked watcher")
		}
	}
	s.watchLock.RUnlock()

	return &hardware.Empty{}, err
}

func (s *server) validateHardwareData(ctx context.Context, hw *hardware.Hardware) error {
	for _, iface := range hw.GetNetwork().GetInterfaces() {
		mac := iface.GetDhcp().GetMac()

		if data, _ := s.db.GetByMAC(ctx, mac); data != "" {
			s.logger.With("MAC", mac).Info(duplicateMAC)

			newhw := hardware.Hardware{}
			if err := json.Unmarshal([]byte(data), &newhw); err != nil {
				s.logger.Error(err, "Failed to unmarshal hardware data")
				return err
			}

			if newhw.Id == hw.Id {
				return nil
			}

			return fmt.Errorf(conflictMACAddr, mac)
		}
	}

	return nil
}

func normalizeHardwareData(hw *hardware.Hardware) {
	// Ensure MAC is stored as lowercase
	for _, iface := range hw.GetNetwork().GetInterfaces() {
		dhcp := iface.GetDhcp()
		if mac := dhcp.GetMac(); mac != "" {
			dhcp.Mac = strings.ToLower(mac)
		}
	}
}
