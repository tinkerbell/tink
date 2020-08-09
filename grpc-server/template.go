package grpcserver

import (
	"context"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	uuid "github.com/satori/go.uuid"
	"github.com/tinkerbell/tink/db"
	"github.com/tinkerbell/tink/metrics"
	"github.com/tinkerbell/tink/protos/template"
)

// CreateTemplate implements template.CreateTemplate
func (s *server) CreateTemplate(ctx context.Context, in *template.WorkflowTemplate) (*template.CreateResponse, error) {
	logger.Info("createtemplate")
	labels := prometheus.Labels{"method": "CreateTemplate", "op": ""}
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	msg := ""
	labels["op"] = "createtemplate"
	msg = "creating a new Template"
	id := uuid.NewV4()
	fn := func() error { return s.db.CreateTemplate(ctx, in.Name, in.Data, id) }

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
		return &template.CreateResponse{}, err
	}
	return &template.CreateResponse{Id: id.String()}, err
}

// GetTemplate implements template.GetTemplate
func (s *server) GetTemplate(ctx context.Context, in *template.GetRequest) (*template.WorkflowTemplate, error) {
	logger.Info("gettemplate")
	labels := prometheus.Labels{"method": "GetTemplate", "op": ""}
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	msg := ""
	labels["op"] = "get"
	msg = "getting a template"

	fn := func() (string, string, error) { return s.db.GetTemplate(ctx, in.Id) }
	metrics.CacheTotals.With(labels).Inc()
	timer := prometheus.NewTimer(metrics.CacheDuration.With(labels))
	defer timer.ObserveDuration()

	logger.Info(msg)
	n, d, err := fn()
	logger.Info("done " + msg)
	if err != nil {
		metrics.CacheErrors.With(labels).Inc()
		l := logger
		if pqErr := db.Error(err); pqErr != nil {
			l = l.With("detail", pqErr.Detail, "where", pqErr.Where)
		}
		l.Error(err)
	}
	return &template.WorkflowTemplate{Id: in.Id, Name: n, Data: d}, err
}

// DeleteTemplate implements template.DeleteTemplate
func (s *server) DeleteTemplate(ctx context.Context, in *template.GetRequest) (*template.Empty, error) {
	logger.Info("deletetemplate")
	labels := prometheus.Labels{"method": "DeleteTemplate", "op": ""}
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	msg := ""
	labels["op"] = "delete"
	msg = "deleting a template"
	fn := func() error { return s.db.DeleteTemplate(ctx, in.Id) }

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
	return &template.Empty{}, err
}

// ListTemplates implements template.ListTemplates
func (s *server) ListTemplates(_ *template.Empty, stream template.Template_ListTemplatesServer) error {
	logger.Info("listtemplates")
	labels := prometheus.Labels{"method": "ListTemplates", "op": "list"}
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
	err := s.db.ListTemplates(func(id, n string, crTime, upTime *timestamp.Timestamp) error {
		return stream.Send(&template.WorkflowTemplate{Id: id, Name: n, CreatedAt: crTime, UpdatedAt: upTime})
	})

	if err != nil {
		metrics.CacheErrors.With(labels).Inc()
		return err
	}

	metrics.CacheHits.With(labels).Inc()
	return nil
}

// UpdateTemplate implements template.UpdateTemplate
func (s *server) UpdateTemplate(ctx context.Context, in *template.WorkflowTemplate) (*template.Empty, error) {
	logger.Info("updatetemplate")
	labels := prometheus.Labels{"method": "UpdateTemplate", "op": ""}
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	msg := ""
	labels["op"] = "updatetemplate"
	msg = "updating a template"
	fn := func() error { return s.db.UpdateTemplate(ctx, in.Name, in.Data, uuid.FromStringOrNil(in.Id)) }

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
	return &template.Empty{}, err
}
