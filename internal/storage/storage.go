package storage

import (
	"context"

	m "github.com/sanek1/metrics-collector/internal/models"
)

type Storage interface {
	SetGauge(ctx context.Context, metric m.Metrics) bool         // Set the value of the gauge
	SetCounter(ctx context.Context, metric m.Metrics) m.Metrics  // Set the value of the counter
	GetAllMetrics() []string                                     // Get all metrics
	GetMetrics(metricType, metricName string) (*m.Metrics, bool) // Get the value of the metric
	SaveToFile(fname string) error                               // Save the metric to a file
	LoadFromFile(fname string) error                             // Save the metric to a file
}
