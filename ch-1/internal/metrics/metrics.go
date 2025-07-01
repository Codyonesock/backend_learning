// Package metrics initializes prometheu.\
package metrics

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ProducerMetrics captures producer events.
type ProducerMetrics struct {
	EventsConsumed  prometheus.Counter
	EventsPersisted prometheus.Counter
}

// NewProducerMetrics creates metrics events.
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

// ConsumerMetrics captures consumer events.
type ConsumerMetrics struct {
	EventsConsumed         prometheus.Counter
	EventsProcessedSuccess prometheus.Counter
	EventsProcessedFailed  prometheus.Counter
}

// NewConsumerMetrics creates metrics events.
func NewConsumerMetrics() *ConsumerMetrics {
	m := &ConsumerMetrics{
		EventsConsumed: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "consumer_events_consumed_total",
			Help:        "Number of events consumed from Redpanda",
			ConstLabels: nil,
		}),
		EventsProcessedSuccess: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "consumer_events_processed_success_total",
			Help:        "Number of events processed successfully",
			ConstLabels: nil,
		}),
		EventsProcessedFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "consumer_events_processed_failed_total",
			Help:        "Number of events that failed to be processed",
			ConstLabels: nil,
		}),
	}
	prometheus.MustRegister(
		m.EventsConsumed,
		m.EventsProcessedSuccess,
		m.EventsProcessedFailed,
	)

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
