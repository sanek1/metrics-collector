package models

import (
	"context"
)

type MetricsStorage interface {
	SetGauge(ctx context.Context, metrics ...Metrics) ([]*Metrics, error)
	SetCounter(ctx context.Context, metrics ...Metrics) ([]*Metrics, error)
	GetMetrics(ctx context.Context, metricType, metricName string) (*Metrics, bool)
	GetAllMetrics() []string
	Ping(ctx context.Context) error
}
