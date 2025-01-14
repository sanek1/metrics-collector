package storage

import (
	"context"
	"database/sql"
	"strings"

	"github.com/lib/pq"
	m "github.com/sanek1/metrics-collector/internal/models"
	"go.uber.org/zap"
)

func (s DBStorage) EnsureMetricsTableExists(ctx context.Context) error {
	var metrics string
	err := s.conn.QueryRowContext(ctx, `SELECT tablename FROM pg_tables WHERE schemaname='public' AND tablename='metrics'`).Scan(&metrics)
	if err == sql.ErrNoRows {
		_, err := s.conn.ExecContext(ctx, `
		CREATE TABLE metrics (
			id serial PRIMARY KEY,
			key text,
			m_type text,
			delta bigint,
			value double precision
		)
	`)
		if err != nil {
			s.Logger.ErrorCtx(ctx, "failed to create Metrics table", zap.Error(err))
			return err
		}
		s.Logger.InfoCtx(ctx, "created Metrics table")
	}
	return nil
}

func FilterBatchesBeforeSaving(metrics []m.Metrics) []m.Metrics {
	seen := make(map[string]m.Metrics)
	for _, model := range metrics {
		key := model.ID + ":" + model.MType
		if existingMetric, ok := seen[key]; ok {
			switch strings.ToLower(model.MType) {
			case "gauge":
				seen[key] = model
			case "counter":
				if model.Delta != nil && existingMetric.Delta != nil {
					*existingMetric.Delta += *model.Delta
				} else {
					seen[key] = model
				}
			}
		} else {
			seen[key] = model
		}
	}

	result := make([]m.Metrics, 0, len(seen))
	for _, metric := range seen {
		result = append(result, metric)
	}

	return result
}

func SortingBatchData(existingMetrics []*m.Metrics, metrics []m.Metrics) (updatingBatch, insertingBatch []m.Metrics) {
	updatingBatch = make([]m.Metrics, 0, len(existingMetrics))
	insertingBatch = make([]m.Metrics, 0, len(metrics)-len(existingMetrics))

	for _, m := range metrics {
		found := false
		for _, r := range existingMetrics {
			if m.ID == r.ID && m.MType == r.MType {
				found = true
				break
			}
		}

		if found {
			updatingBatch = append(updatingBatch, m)
		} else {
			insertingBatch = append(insertingBatch, m)
		}
	}
	return updatingBatch, insertingBatch
}

func CollectorQuery(ctx context.Context, metrics []m.Metrics) (query string, mTypes []string, args []interface{}) {
	mTypes = make([]string, 0, len(metrics))
	keys := make([]string, 0, len(metrics))
	for _, metric := range metrics {
		mTypes = append(mTypes, metric.MType)
		keys = append(keys, metric.ID)
	}

	query = `
	SELECT key, m_type, delta, value
	FROM metrics
	WHERE m_type = ANY($1) AND key = ANY($2)
  `
	args = []interface{}{pq.Array(mTypes), pq.Array(keys)}
	return query, mTypes, args
}
