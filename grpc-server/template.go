package grpcserver

import (
	"context"

	"github.com/packethost/cacher/pg"
	"github.com/packethost/rover/metrics"
	"github.com/packethost/rover/protos/template"
	"github.com/prometheus/client_golang/prometheus"
)

// Create implements template.Create
func (s *server) Create(ctx context.Context, in *template.WorkflowTemplate) (*template.Empty, error) {
	logger.Info("createtemplate")
	labels := prometheus.Labels{"method": "CreateTemplate", "op": ""}
	metrics.CacheInFlight.With(labels).Inc()
	defer metrics.CacheInFlight.With(labels).Dec()

	msg := ""
	labels["op"] = "createtemplate"
	msg = "creating a new Teamplate"
	fn := func() error { return pg.CreateTemplate(ctx, s.db, in.Name, in.Data) }

	metrics.CacheTotals.With(labels).Inc()
	timer := prometheus.NewTimer(metrics.CacheDuration.With(labels))
	defer timer.ObserveDuration()

	logger.Info(msg)
	err := fn()
	logger.Info("done " + msg)
	if err != nil {
		metrics.CacheErrors.With(labels).Inc()
		l := logger
		if pqErr := pg.Error(err); pqErr != nil {
			l = l.With("detail", pqErr.Detail, "where", pqErr.Where)
		}
		l.Error(err)
	}
	return &template.Empty{}, err
}
