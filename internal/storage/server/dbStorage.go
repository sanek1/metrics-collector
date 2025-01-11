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

func (ds *DBStorage) IsOK() bool {
	if ds.conn == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := ds.conn.PingContext(ctx); err != nil {
		return false
	}
	return true
}

func (s DBStorage) ensureMetricsTableExists(ctx context.Context) error {
	var metrics string
	err := s.conn.QueryRowContext(ctx, `SELECT tablename FROM pg_tables WHERE schemaname='public' AND tablename='metrics'`).Scan(&metrics)
	if err == sql.ErrNoRows {
		_, err := s.conn.ExecContext(ctx, `
		CREATE TABLE metrics (
			id serial PRIMARY KEY,
			key text,
			m_type text,
			delta integer,
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

func (s *DBStorage) GetAllMetrics() []string {
	var metrics []string
	rows, err := s.conn.Query("SELECT key, m_type, delta, value FROM metrics")
	if err != nil {
		s.Logger.ErrorCtx(context.Background(), "failed to get all metrics from database", zap.Error(err))
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		var metric m.Metrics
		if err := rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value); err != nil {
			s.Logger.ErrorCtx(context.Background(), "failed to scan metric from database", zap.Error(err))
			continue
		}
		metrics = append(metrics, metric.ID)
	}
	return metrics
}

func (s *DBStorage) GetMetrics(ctx context.Context, mType, id string) (*m.Metrics, bool) {
	var metric m.Metrics
	metric.ID = id
	metric.MType = mType

	model, found := s.GetMetricByTypeAndID(ctx, metric.MType, metric.ID)
	if !found {
		return nil, false
	}
	return model, true
}

func (s *DBStorage) insertMetric(ctx context.Context, model m.Metrics) error {
	stmt, err := s.conn.PrepareContext(ctx, `
	 INSERT INTO metrics ("key", m_type, value, delta) VALUES ($1, $2, $3, $4)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement for inserting metric: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, model.ID, model.MType, model.Value, model.Delta)
	if err != nil {
		return fmt.Errorf("failed to insert metric: %w", err)
	}
	return nil
}

func (s *DBStorage) updateMetric(ctx context.Context, model m.Metrics) error {
	stmt, err := s.conn.PrepareContext(ctx, `
	 UPDATE metrics SET value=$1, delta=delta+$2 WHERE m_type=$3 AND key=$4
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement for updating metric: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, model.Value, model.Delta, model.MType, model.ID)
	if err != nil {
		return fmt.Errorf("failed to update metric: %w", err)
	}
	return nil
}

func (s *DBStorage) GetMetricByTypeAndID(ctx context.Context, mType, id string) (*m.Metrics, bool) {
	var metric m.Metrics
	row := s.conn.QueryRowContext(ctx, `
	 SELECT key, m_type, delta, value FROM metrics WHERE m_type=$1 AND key=$2
	`, mType, id)

	err := row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
	if err == sql.ErrNoRows {
		return nil, false
	} else if err != nil {
		s.Logger.ErrorCtx(ctx, "failed to query Metric by type and ID", zap.Error(err))
		return nil, false
	}
	return &metric, true
}

func (s *DBStorage) SetCounter(ctx context.Context, model m.Metrics) (m.Metrics, error) {
	return s.ensureMetrics(ctx, model)
}

func (s *DBStorage) SetGauge(ctx context.Context, model m.Metrics) (m.Metrics, error) {
	return s.ensureMetrics(ctx, model)
}

func (s *DBStorage) ensureMetrics(ctx context.Context, model m.Metrics) (m.Metrics, error) {
	if err := s.ensureMetricsTableExists(ctx); err != nil {
		return m.Metrics{}, err
	}
	_, found := s.GetMetricByTypeAndID(ctx, model.MType, model.ID)
	if found {
		if err := s.updateMetric(ctx, model); err != nil {
			return m.Metrics{}, err
		}
	} else {
		if err := s.insertMetric(ctx, model); err != nil {
			return m.Metrics{}, err
		}
	}
	res, found := s.GetMetricByTypeAndID(ctx, model.MType, model.ID)
	if !found {
		return m.Metrics{}, fmt.Errorf("metric not found")
	}
	return *res, nil
}
