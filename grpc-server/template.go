package grpcserver

import (
	"context"
	"strings"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
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

	const msg = "creating a new Template"
	labels["op"] = "createtemplate"
	id, _ := uuid.NewUUID()

	metrics.CacheTotals.With(labels).Inc()
	timer := prometheus.NewTimer(metrics.CacheDuration.With(labels))
	defer timer.ObserveDuration()

	logger.Info(msg)
	err := s.db.CreateTemplate(ctx, in.Name, in.Data, id)
	if err != nil {
		metrics.CacheErrors.With(labels).Inc()
		l := logger
		if pqErr := db.Error(err); pqErr != nil {
			l = l.With("detail", pqErr.Detail, "where", pqErr.Where)
		}
		l.Error(err)
		return &template.CreateResponse{}, err
	}
	logger.Info("done " + msg)
	return &template.CreateResponse{Id: id.String()}, err
}

// GetTemplate implements template.GetTemplate
func (s *server) GetTemplate(ctx context.Context, in *template.GetRequest) (*template.WorkflowTemplate, error) {
	logger.Info("gettemplate")
	labels := prometheus.Labels{"method": "GetTemplate", "op": ""}
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	const msg = "getting a template"
	labels["op"] = "get"

	metrics.CacheTotals.With(labels).Inc()
	timer := prometheus.NewTimer(metrics.CacheDuration.With(labels))
	defer timer.ObserveDuration()

	logger.Info(msg)
	fields := map[string]string{
		"id":   in.GetId(),
		"name": in.GetName(),
	}
	id, n, d, err := s.db.GetTemplate(ctx, fields)
	logger.Info("done " + msg)
	if err != nil {
		metrics.CacheErrors.With(labels).Inc()
		l := logger
		if pqErr := db.Error(err); pqErr != nil {
			l = l.With("detail", pqErr.Detail, "where", pqErr.Where)
		}
		l.Error(err)
	}
	return &template.WorkflowTemplate{Id: id, Name: n, Data: d}, err
}

// DeleteTemplate implements template.DeleteTemplate
func (s *server) DeleteTemplate(ctx context.Context, in *template.GetRequest) (*template.Empty, error) {
	logger.Info("deletetemplate")
	labels := prometheus.Labels{"method": "DeleteTemplate", "op": ""}
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	const msg = "deleting a template"
	labels["op"] = "delete"

	metrics.CacheTotals.With(labels).Inc()
	timer := prometheus.NewTimer(metrics.CacheDuration.With(labels))
	defer timer.ObserveDuration()

	logger.Info(msg)
	err := s.db.DeleteTemplate(ctx, in.GetId())
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
func (s *server) ListTemplates(in *template.ListRequest, stream template.TemplateService_ListTemplatesServer) error {
	logger.Info("listtemplates")
	labels := prometheus.Labels{"method": "ListTemplates", "op": "list"}
	metrics.CacheTotals.With(labels).Inc()
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	filter := "%" // default filter will match everything
	if in.GetName() != "" {
		filter = strings.ReplaceAll(in.GetName(), "*", "%") // replace '*' with psql '%' wildcard
	}

	s.dbLock.RLock()
	ready := s.dbReady
	s.dbLock.RUnlock()
	if !ready {
		metrics.CacheStalls.With(labels).Inc()
		return errors.New("DB is not ready")
	}

	timer := prometheus.NewTimer(metrics.CacheDuration.With(labels))
	defer timer.ObserveDuration()
	err := s.db.ListTemplates(filter, func(id, n string, crTime, upTime *timestamp.Timestamp) error {
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

	const msg = "updating a template"
	labels["op"] = "updatetemplate"

	metrics.CacheTotals.With(labels).Inc()
	timer := prometheus.NewTimer(metrics.CacheDuration.With(labels))
	defer timer.ObserveDuration()

	logger.Info(msg)
	err := s.db.UpdateTemplate(ctx, in.Name, in.Data, uuid.MustParse(in.Id))
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
