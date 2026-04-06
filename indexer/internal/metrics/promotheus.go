package metrics

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusMetrics struct {
	profilesIndexed prometheus.Counter
	indexingErrors  prometheus.Counter
}

func NewPrometheusMetrics() *PrometheusMetrics {
	pm := &PrometheusMetrics{
		profilesIndexed: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "influencers_indexed_total",
				Help: "Total number of profiles successfully saved to Elasticsearch",
			},
		),
		indexingErrors: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "indexer_errors_total",
				Help: "Total number of failed indexing attempts",
			},
		),
	}

	prometheus.MustRegister(pm.profilesIndexed)
	prometheus.MustRegister(pm.indexingErrors)
	return pm
}

func (m *PrometheusMetrics) IncIndexed() { m.profilesIndexed.Inc() }
func (m *PrometheusMetrics) IncError()   { m.indexingErrors.Inc() }

func (m *PrometheusMetrics) StartServer(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Printf("Metrics server stopped: %v", err)
	}
}
