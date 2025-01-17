package storage

import (
	"context"
	"fmt"

	flags "github.com/sanek1/metrics-collector/internal/flags/server"
	m "github.com/sanek1/metrics-collector/internal/models"
	l "github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/zap"
)

func SetLog(ctx context.Context, ms *MetricsStorage, model *m.Metrics, name string) {
	ms.Logger.InfoCtx(ctx, name, zap.String(name, fmt.Sprintf("model%s", formatMetric(*model))))
}

func GetStorage(useDatabase bool, opt *flags.ServerOptions, logger *l.ZapLogger) Storage {
	if useDatabase {
		return NewDBStorage(opt, logger)
	}

	return NewMetricsStorage(logger)
}
