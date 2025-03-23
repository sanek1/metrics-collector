package storage

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	flags "github.com/sanek1/metrics-collector/internal/flags/server"
	m "github.com/sanek1/metrics-collector/internal/models"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

func (ms *MetricsStorage) SetLog(ctx context.Context, model *m.Metrics) {
	ctx = ms.Logger.WithContextFields(ctx,
		zap.String("Type", model.MType),
		zap.String("Id", model.ID),
		zap.Any("Data", model))

	jsonData, err := json.Marshal(model)
	if err != nil {
		ms.Logger.ErrorCtx(ctx, "Failed to marshal metrics", zap.Error(err))
		return
	}
	ms.Logger.InfoCtx(ctx, string(jsonData))
}

func GetStorage(useDatabase bool, opt *flags.ServerOptions, logger *l.ZapLogger) Storage {
	if useDatabase {
		return NewDBStorage(opt, logger)
	}
	return NewMetricsStorage(logger)
}
