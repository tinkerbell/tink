package metrics

import (
	"github.com/packethost/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus Metrics
var (
	CacheDuration prometheus.ObserverVec
	CacheErrors   *prometheus.CounterVec
	CacheHits     *prometheus.CounterVec
	CacheInFlight *prometheus.GaugeVec
	CacheStalls   *prometheus.CounterVec
	CacheTotals   *prometheus.CounterVec

	ingestCount    *prometheus.CounterVec
	ingestErrors   *prometheus.CounterVec
	ingestDuration *prometheus.GaugeVec

	watchMissTotal prometheus.Counter
)

// SetupMetrics sets the defaults for metrics
func SetupMetrics(facility string, logger log.Logger) {
	curryLabels := prometheus.Labels{
		"service":  "rover",
		"facility": facility,
	}

	CacheDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "cache_ops_duration_seconds",
		Help:    "Duration of cache operations",
		Buckets: prometheus.LinearBuckets(.01, .05, 10),
	}, []string{"service", "facility", "method", "op"}).MustCurryWith(curryLabels)
	CacheErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_ops_errors_total",
		Help: "Number of cache errors.",
	}, []string{"service", "facility", "method", "op"}).MustCurryWith(curryLabels)
	CacheHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_hit_total",
		Help: "Number of cache hits.",
	}, []string{"service", "facility", "method", "op"}).MustCurryWith(curryLabels)
	CacheInFlight = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cache_ops_current_total",
		Help: "Number of in flight cache requests.",
	}, []string{"service", "facility", "method", "op"}).MustCurryWith(curryLabels)
	CacheStalls = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_stall_total",
		Help: "Number of cache stalled due to DB.",
	}, []string{"service", "facility", "method", "op"}).MustCurryWith(curryLabels)
	CacheTotals = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_ops_total",
		Help: "Number of cache ops.",
	}, []string{"service", "facility", "method", "op"}).MustCurryWith(curryLabels)

	logger.Info("initializing label values")
	var labels []prometheus.Labels

	labels = []prometheus.Labels{
		{"method": "Push", "op": ""},
		{"method": "Ingest", "op": ""},
	}
	initCounterLabels(CacheErrors, labels)
	initGaugeLabels(CacheInFlight, labels)
	initCounterLabels(CacheStalls, labels)
	initCounterLabels(CacheTotals, labels)
	labels = []prometheus.Labels{
		{"method": "Push", "op": "insert"},
		{"method": "Push", "op": "delete"},
	}
	initObserverLabels(CacheDuration, labels)
	initCounterLabels(CacheHits, labels)

	labels = []prometheus.Labels{
		{"method": "ByMAC", "op": "get"},
		{"method": "ByIP", "op": "get"},
		{"method": "ByID", "op": "get"},
		{"method": "All", "op": "get"},
		{"method": "Ingest", "op": ""},
		{"method": "Watch", "op": "get"},
		{"method": "Watch", "op": "push"},
	}
	initCounterLabels(CacheErrors, labels)
	initGaugeLabels(CacheInFlight, labels)
	initCounterLabels(CacheStalls, labels)
	initCounterLabels(CacheTotals, labels)
	initObserverLabels(CacheDuration, labels)
	initCounterLabels(CacheHits, labels)

	ingestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ingest_op_count_total",
		Help: "Number of attempts made to ingest facility data.",
	}, []string{"service", "facility", "method", "op"}).MustCurryWith(curryLabels)
	ingestDuration = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ingest_op_duration_seconds",
		Help: "Duration of successful ingestion actions while attempting to ingest facility data.",
	}, []string{"service", "facility", "method", "op"}).MustCurryWith(curryLabels)
	ingestErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ingest_error_count_total",
		Help: "Number of errors occurred attempting to ingest facility data.",
	}, []string{"service", "facility", "method", "op"}).MustCurryWith(curryLabels)
	labels = []prometheus.Labels{
		{"method": "Ingest", "op": ""},
		{"method": "Ingest", "op": "fetch"},
		{"method": "Ingest", "op": "copy"},
	}
	initCounterLabels(ingestCount, labels)
	initGaugeLabels(ingestDuration, labels)
	initCounterLabels(ingestErrors, labels)

	watchMissTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "watch_miss_count_total",
		Help: "Number of missed updates due to a blocked channel.",
	})
}

func initObserverLabels(m prometheus.ObserverVec, l []prometheus.Labels) {
	for _, labels := range l {
		m.With(labels)
	}
}

func initGaugeLabels(m *prometheus.GaugeVec, l []prometheus.Labels) {
	for _, labels := range l {
		m.With(labels)
	}
}

func initCounterLabels(m *prometheus.CounterVec, l []prometheus.Labels) {
	for _, labels := range l {
		m.With(labels)
	}
}
