package storage

import (
	"context"
	"time"

	m "github.com/sanek1/metrics-collector/internal/models"
)

type Storage interface {
	SetGauge(ctx context.Context, models ...m.Metrics) ([]*m.Metrics, error)          // Set the value of the gauge
	SetCounter(ctx context.Context, models ...m.Metrics) ([]*m.Metrics, error)        // Set the value of the counter
	GetAllMetrics() []string                                                          // Get all metrics
	GetMetrics(ctx context.Context, metricType, metricName string) (*m.Metrics, bool) // Get the value of the metric
}

type DatabaseStorage interface {
	PingIsOk() bool
	EnsureMetricsTableExists(ctx context.Context) error
}

type FileStorage interface {
	SaveToFile(fname string) error   // Save the metric to a file
	LoadFromFile(fname string) error // Save the metric to a file
	PeriodicallySaveBackUp(ctx context.Context, filename string, restore bool, interval time.Duration)
}
