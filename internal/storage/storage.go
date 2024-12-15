package storage

import (
	m "github.com/sanek1/metrics-collector/internal/validation"
)

type Storage interface {
	SetGauge(m.Metrics) bool                                     // Set the value of the gauge
	SetCounter(m.Metrics) m.Metrics                              // Set the value of the counter
	GetAllMetrics() []string                                     // Get all metrics
	GetMetrics(metricType, metricName string) (*m.Metrics, bool) // Get the value of the metric
}
