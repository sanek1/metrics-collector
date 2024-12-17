package handlers

import (
	"github.com/sanek1/metrics-collector/internal/storage"
	"go.uber.org/zap"
)

type MetricStorage struct {
	Storage storage.Storage
	Logger  *zap.SugaredLogger
}
