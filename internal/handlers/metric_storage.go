package handlers

import (
	"github.com/sanek1/metrics-collector/internal/storage"
)

type MetricStorage struct {
	Storage storage.Storage
}
