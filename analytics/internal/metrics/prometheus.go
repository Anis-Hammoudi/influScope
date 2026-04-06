package metrics

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusMetrics struct {
	registry            *prometheus.Registry
	engagementRequests  *prometheus.CounterVec
	calculationDuration prometheus.Histogram
}

func NewPrometheusMetrics() *PrometheusMetrics {
	reg := prometheus.NewRegistry()

	pm := &PrometheusMetrics{
		registry: reg,
		engagementRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "analytics_engagement_requests_total",
				Help: "Total number of engagement calculation requests received",
			},
			[]string{"platform"},
		),
		calculationDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "analytics_calculation_duration_seconds",
				Help:    "Time taken to calculate engagement rate",
				Buckets: prometheus.DefBuckets,
			},
		),
	}

	reg.MustRegister(pm.engagementRequests)
	reg.MustRegister(pm.calculationDuration)

	return pm
}

// StartTimer returns a closure that observes the duration when called via defer
func (m *PrometheusMetrics) StartTimer() func() {
	timer := prometheus.NewTimer(m.calculationDuration)
	return func() { timer.ObserveDuration() }
}

func (m *PrometheusMetrics) IncEngagementRequest(platform string) {
	m.engagementRequests.WithLabelValues(platform).Inc()
}

func (m *PrometheusMetrics) StartServer(addr string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}))

	log.Printf("Metrics server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Printf("Metrics server stopped: %v", err)
	}
}
