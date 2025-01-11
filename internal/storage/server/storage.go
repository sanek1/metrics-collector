package storage

import (
	"context"

	m "github.com/sanek1/metrics-collector/internal/models"
)

type MetricStorage interface {
	Storage
	FileStorage
}

type Storage interface {
	SetGauge(ctx context.Context, metric m.Metrics) (m.Metrics, error)                // Set the value of the gauge
	SetCounter(ctx context.Context, metric m.Metrics) (m.Metrics, error)              // Set the value of the counter
	GetAllMetrics() []string                                                          // Get all metrics
	GetMetrics(ctx context.Context, metricType, metricName string) (*m.Metrics, bool) // Get the value of the metric
}

type FileStorage interface {
	SaveToFile(fname string) error   // Save the metric to a file
	LoadFromFile(fname string) error // Save the metric to a file
}
