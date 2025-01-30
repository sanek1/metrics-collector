package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"go.uber.org/zap"

	flags "github.com/sanek1/metrics-collector/internal/flags/server"
	m "github.com/sanek1/metrics-collector/internal/models"
)

func startDBConnection(ctx context.Context, opt *flags.ServerOptions) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, opt.DBPath)
	if err != nil {
		return nil, err
	}
	return pool, nil
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

func (s *DBStorage) updateMetrics(ctx context.Context, models []m.Metrics) error {
	batch := &pgx.Batch{}
	for _, model := range models {
		if model.MType == m.TypeCounter {
			batch.Queue("UPDATE metrics SET delta=delta+$1 WHERE m_type=$2 AND key=$3", model.Delta, model.MType, model.ID)
		} else if model.MType == m.TypeGauge {
			batch.Queue("UPDATE metrics SET value=$1 WHERE m_type=$2 AND key=$3", model.Value, model.MType, model.ID)
		}
	}

	// Execute batch query
	br := s.conn.SendBatch(ctx, batch)
	defer br.Close()

	// Commit or rollback transaction
	var err error
	if _, err := br.Exec(); err != nil {
		s.Logger.ErrorCtx(ctx, "failed to execute batch request:  %w"+err.Error(), zap.Error(err))
	}
	return err
}
func (s *DBStorage) insertMetric(ctx context.Context, models []m.Metrics) error {
	conn, err := s.conn.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	batch := &pgx.Batch{}
	for _, model := range models {
		batch.Queue("INSERT INTO metrics (key, m_type, value, delta) VALUES ($1, $2, $3, $4)", model.ID, model.MType, model.Value, model.Delta)
	}

	br := s.conn.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			s.Logger.ErrorCtx(ctx, "failed to execute batch request - "+err.Error(), zap.Error(err))
			return fmt.Errorf("failed to execute batch request: %w", err)
		}
	}
	return nil
}

func (s *DBStorage) getMetricsOnDBs(ctx context.Context, metrics ...m.Metrics) ([]*m.Metrics, error) {
	jsonData, _ := json.Marshal(metrics)
	s.Logger.DebugCtx(ctx, "METRICS "+string(jsonData))

	query, mTypes, args := CollectorQuery(ctx, metrics)

	rows, err := s.conn.Query(ctx, query, args...)
	if err != nil {
		s.Logger.ErrorCtx(ctx, "Failed to execute query", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	results := make([]*m.Metrics, 0, len(mTypes))
	for rows.Next() {
		metric := new(m.Metrics)
		var delta sql.NullInt64
		var value sql.NullFloat64

		if err := rows.Scan(&metric.ID, &metric.MType, &delta, &value); err != nil {
			s.Logger.ErrorCtx(ctx, "Failed to scan row", zap.Error(err))
			continue
		}

		if delta.Valid {
			metric.Delta = new(int64)
			*metric.Delta = delta.Int64
		} else {
			metric.Delta = nil
		}

		if value.Valid {
			metric.Value = new(float64)
			*metric.Value = value.Float64
		} else {
			metric.Value = nil
		}
		results = append(results, metric)
	}
	return results, nil
}

func (s *DBStorage) setMetrics(ctx context.Context, models ...m.Metrics) ([]*m.Metrics, error) {
	// filter duplicates and batches before saving
	models = FilterBatchesBeforeSaving(models)

	existingMetrics, err := s.getMetricsOnDBs(ctx, models...)
	if err != nil {
		return nil, err
	}
	// filter duplicates and sort before updating/inserting
	updatingBatch, insertingBatch := SortingBatchData(existingMetrics, models)

	if len(updatingBatch) != 0 {
		if err := s.updateMetrics(ctx, updatingBatch); err != nil {
			s.Logger.ErrorCtx(ctx, "failed to update metric", zap.Error(err))
			return nil, err
		}
	}
	if len(insertingBatch) != 0 {
		if err := s.insertMetric(ctx, insertingBatch); err != nil {
			s.Logger.ErrorCtx(ctx, "failed to insert metric", zap.Error(err))
			return nil, err
		}
	}

	metrics, err := s.getMetricsOnDBs(ctx, models...)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}
