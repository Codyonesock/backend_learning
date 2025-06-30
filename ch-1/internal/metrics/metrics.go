// Package metrics initializes prometheu.\
package metrics

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ProducerMetrics captures producer events..
type ProducerMetrics struct {
	EventsConsumed  prometheus.Counter
	EventsPersisted prometheus.Counter
}

// NewProducerMetrics creates Metrics events.
func NewProducerMetrics() *ProducerMetrics {
	m := &ProducerMetrics{
		EventsConsumed: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "producer_events_consumed_total",
			Help:        "Number of events consumed from the stream",
			ConstLabels: nil,
		}),
		EventsPersisted: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "producer_events_persisted_total",
			Help:        "Number of events persisted to Redpanda",
			ConstLabels: nil,
		}),
	}
	prometheus.MustRegister(m.EventsConsumed, m.EventsPersisted)

	return m
}

// StartServer starts the /metrics endpoint for prometheus.
func StartServer(addr string) {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		// #nosec G114: Ignore timeouts for simplicity
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("metrics server error: %v\n", err)
		}
	}()
}
