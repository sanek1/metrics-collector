package storage

import (
	"context"
	"database/sql"
	"fmt"

	m "github.com/sanek1/metrics-collector/internal/models"
	l "github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/zap"
)

func SetLog(ctx context.Context, ms *MetricsStorage, model *m.Metrics, name string) {
	ms.Logger.InfoCtx(ctx, name, zap.String(name, fmt.Sprintf("model%s", formatMetric(*model))))
}

func GetStorage(useDatabase bool, conn *sql.DB, logger *l.ZapLogger) MetricStorage {
	if useDatabase && conn != nil {
		return NewDBStorage(conn, logger)
	}
	return NewMetricsStorage(logger)
}
