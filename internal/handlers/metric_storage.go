package handlers

import (
	storage "github.com/sanek1/metrics-collector/internal/storage/server"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

type Storage struct {
	Storage storage.Storage
	Logger  *l.ZapLogger
}
