package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// This is inspired by the Prometheus operator metrics collector
// https://github.com/prometheus-operator/prometheus-operator/blob/2038d841001f6094b79339d394b3d59b3d72cd75/pkg/operator/operator.go

type MetricRegistry struct {
	reg prometheus.Registerer

	reconcileCounter           prometheus.Counter
	reconcileErrorsCounter     prometheus.Counter
	reconcileDurationHistogram prometheus.Histogram
	ready                      prometheus.Gauge
}

// NewMetrics initializes operator metrics and registers them with the given registerer.
// All metrics have a "controller=<name>" label.
func NewMetrics(name string, r prometheus.Registerer) *MetricRegistry {
	reg := prometheus.WrapRegistererWith(prometheus.Labels{"controller": name}, r)
	m := MetricRegistry{
		reg: reg,
		reconcileCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "nifikop_operator_reconcile_operations_total",
			Help: "Total number of reconcile operations",
		}),
		reconcileDurationHistogram: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "nifikop_operator_reconcile_duration_seconds",
			Help:    "Histogram of reconcile operations",
			Buckets: []float64{.1, .5, 1, 5, 10},
		}),
		reconcileErrorsCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "nifikop_operator_reconcile_errors_total",
			Help: "Number of errors that occurred during reconcile operations",
		}),
		ready: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nifikop_operator_ready",
			Help: "1 when the controller is ready to reconcile resources, 0 otherwise",
		}),
	}

	m.reg.MustRegister(
		m.reconcileCounter,
		m.reconcileDurationHistogram,
		m.reconcileErrorsCounter,
		m.ready,
	)

	return &m
}

// Ready returns a gauge to track whether the controller is ready or not.
func (m *MetricRegistry) Ready() prometheus.Gauge {
	return m.ready
}

// ReconcileCounter returns a counter to track attempted reconciliations.
func (m *MetricRegistry) ReconcileCounter() prometheus.Counter {
	return m.reconcileCounter
}

// ReconcileDurationHistogram returns a histogram to track the duration of reconciliations.
func (m *MetricRegistry) ReconcileDurationHistogram() prometheus.Histogram {
	return m.reconcileDurationHistogram
}

// ReconcileErrorsCounter returns a counter to track reconciliation errors.
func (m *MetricRegistry) ReconcileErrorsCounter() prometheus.Counter {
	return m.reconcileErrorsCounter
}
