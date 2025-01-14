package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	m "github.com/sanek1/metrics-collector/internal/models"
	l "github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/zap"
)

type DBStorage struct {
	conn   *sql.DB
	Logger *l.ZapLogger
}

func NewDBStorage(c *sql.DB, logger *l.ZapLogger) *DBStorage {
	return &DBStorage{
		conn:   c,
		Logger: logger,
	}
}

func (s *DBStorage) IsOK() bool {
	if s.conn == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := s.conn.PingContext(ctx); err != nil {
		return false
	}
	return true
}

func (s *DBStorage) GetAllMetrics() []string {
	var res []string
	rows, err := s.conn.Query("SELECT key, m_type, delta, value FROM metrics")
	if err != nil {
		s.Logger.ErrorCtx(context.Background(), "failed to get all metrics from database", zap.Error(err))
		return nil
	}
	if rows.Err() != nil {
		s.Logger.ErrorCtx(context.Background(), "failed to get all metrics from database", zap.Error(rows.Err()))
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		var metric m.Metrics
		if err := rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value); err != nil {
			s.Logger.ErrorCtx(context.Background(), "failed to scan metric from database", zap.Error(err))
			continue
		}
		res = append(res, metric.ID)
	}
	return res
}

func (s *DBStorage) GetMetrics(ctx context.Context, mType, id string) (*m.Metrics, bool) {
	var metric m.Metrics
	metric.ID = id
	metric.MType = mType

	models, err := s.getMetricsOnDBs(ctx, metric)
	if err != nil {
		return nil, false
	}
	if len(models) == 0 {
		return nil, false
	}
	model := models[0]

	return model, true
}

func (s *DBStorage) updateMetric(ctx context.Context, models []m.Metrics) error {
	stmt1, err := s.conn.PrepareContext(ctx, `
       UPDATE metrics SET delta=delta+$1 WHERE m_type=$2 AND key=$3
    `)
	if err != nil {
		s.Logger.ErrorCtx(ctx, "failed to prepare statement for updating metrics: %w", zap.Error(err))
		return fmt.Errorf("failed to prepare statement for updating metrics: %w", err)
	}
	stmt2, err := s.conn.PrepareContext(ctx, `
       UPDATE metrics SET value=$1 WHERE m_type=$2 AND key=$3
    `)
	if err != nil {
		s.Logger.ErrorCtx(ctx, "failed to prepare statement for updating metrics: %w", zap.Error(err))
		return fmt.Errorf("failed to prepare statement for updating metrics: %w", err)
	}
	defer stmt1.Close()
	defer stmt2.Close()

	for _, model := range models {
		if model.MType == "counter" {
			_, err = stmt1.ExecContext(ctx, model.Delta, model.MType, model.ID)
			if err != nil {
				s.Logger.ErrorCtx(ctx, "failed to update metric counter"+model.ID+
					model.ID+" with Delta "+fmt.Sprintf("%d", model.Delta)+
					"err "+err.Error(),
					zap.String("id", model.ID))
				return fmt.Errorf("failed to update metric counter: %w", err)
			}
		}
		if model.MType == "gauge" {
			_, err = stmt2.ExecContext(ctx, model.Value, model.MType, model.ID)
			if err != nil {
				s.Logger.ErrorCtx(ctx, "failed to update metric gauge: %w", zap.String("id", model.ID))
				return fmt.Errorf("failed to update metric gauge: %w", err)
			}
		}
	}
	return nil
}

func (s *DBStorage) insertMetric(ctx context.Context, models []m.Metrics) error {
	stmt, err := s.conn.PrepareContext(ctx, `
       INSERT INTO metrics ("key", m_type, value, delta) VALUES ($1, $2, $3, $4)
    `)
	if err != nil {
		s.Logger.ErrorCtx(ctx, "failed to prepare statement for inserting metrics: %w", zap.Error(err))
		return fmt.Errorf("failed to prepare statement for inserting metrics: %w", err)
	}
	defer stmt.Close()

	for _, model := range models {
		_, err = stmt.ExecContext(ctx, model.ID, model.MType, model.Value, model.Delta)
		if err != nil {
			s.Logger.ErrorCtx(ctx, "failed to insert metric: %w", zap.Error(err))
			return fmt.Errorf("failed to insert metric: %w", err)
		}
	}
	return nil
}

func (s *DBStorage) getMetricsOnDBs(ctx context.Context, metrics ...m.Metrics) ([]*m.Metrics, error) {
	query, mTypes, args := CollectorQuery(ctx, metrics)
	rows, err := s.conn.QueryContext(ctx, query, args...)
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
		metric.Delta = &delta.Int64
		metric.Value = &value.Float64
		results = append(results, metric)
	}

	if err := rows.Err(); err != nil {
		s.Logger.ErrorCtx(ctx, "Failed to iterate over rows", zap.Error(err))
		return nil, err
	}

	return results, nil
}

func (s *DBStorage) SetCounter(ctx context.Context, models ...m.Metrics) ([]*m.Metrics, error) {
	return s.ensureMetrics(ctx, models...)
}

func (s *DBStorage) SetGauge(ctx context.Context, models ...m.Metrics) ([]*m.Metrics, error) {
	return s.ensureMetrics(ctx, models...)
}

func (s *DBStorage) ensureMetrics(ctx context.Context, models ...m.Metrics) ([]*m.Metrics, error) {
	// check if table exists
	if err := s.EnsureMetricsTableExists(ctx); err != nil {
		s.Logger.ErrorCtx(ctx, "failed to ensure Metrics table exists", zap.Error(err))
		return nil, err
	}
	// filter duplicates and batches before saving
	models = FilterBatchesBeforeSaving(models)

	existingMetrics, err := s.getMetricsOnDBs(ctx, models...)
	if err != nil {
		return nil, err
	}
	// filter duplicates and sort before updating/inserting
	updatingBatch, insertingBatch := SortingBatchData(existingMetrics, models)

	if len(updatingBatch) != 0 {
		if err := s.updateMetric(ctx, updatingBatch); err != nil {
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

func (s *DBStorage) LoadFromFile(fname string) error {
	return fmt.Errorf("LoadFromFile is not supported for DBStorage")
}

func (s *DBStorage) SaveToFile(fname string) error {
	return fmt.Errorf("SaveToFile is not supported for DBStorage")
}
