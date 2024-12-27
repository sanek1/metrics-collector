package handlers

import (
	"github.com/sanek1/metrics-collector/internal/storage"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

type MetricStorage struct {
	Storage storage.Storage
	Logger  *l.ZapLogger
}
