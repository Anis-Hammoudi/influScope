package metrics

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusMetrics struct {
	registry        *prometheus.Registry
	profilesIndexed prometheus.Counter
	indexingErrors  prometheus.Counter
}

func NewPrometheusMetrics() *PrometheusMetrics {
	reg := prometheus.NewRegistry() // Create a local registry

	pm := &PrometheusMetrics{
		registry: reg,
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

	// Register metrics to the LOCAL registry, not the global one
	reg.MustRegister(pm.profilesIndexed)
	reg.MustRegister(pm.indexingErrors)
	return pm
}

func (m *PrometheusMetrics) IncIndexed() { m.profilesIndexed.Inc() }
func (m *PrometheusMetrics) IncError()   { m.indexingErrors.Inc() }

func (m *PrometheusMetrics) StartServer(addr string) {
	// Create a dedicated mux so we don't pollute the global http.DefaultServeMux
	mux := http.NewServeMux()

	// Expose only our local registry
	mux.Handle("/metrics", promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}))

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Printf("Metrics server stopped: %v", err)
	}
}
