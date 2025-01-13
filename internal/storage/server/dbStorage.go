package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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
				s.Logger.ErrorCtx(ctx, "failed to update metric counter"+model.ID, zap.String("id", model.ID))
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
	if len(metrics) == 0 {
		return nil, nil
	}

	mTypes := make([]string, 0, len(metrics))
	keys := make([]string, 0, len(metrics))
	for _, metric := range metrics {
		mTypes = append(mTypes, metric.MType)
		keys = append(keys, metric.ID)
	}

	mTypePlaceholders := make([]string, len(mTypes))
	keyPlaceholders := make([]string, len(keys))
	for i := range mTypePlaceholders {
		mTypePlaceholders[i] = fmt.Sprintf("$%d", i+1)
	}
	for i := range keyPlaceholders {
		keyPlaceholders[i] = fmt.Sprintf("$%d", len(mTypePlaceholders)+i+1)
	}

	query := fmt.Sprintf(`
        SELECT key, m_type, delta, value 
        FROM metrics 
        WHERE m_type IN (%s) AND key IN (%s)
    `,
		strings.Join(mTypePlaceholders, ","),
		strings.Join(keyPlaceholders, ","),
	)

	args := make([]interface{}, 0, len(mTypes)+len(keys))
	for _, v := range mTypes {
		args = append(args, v)
	}
	for _, v := range keys {
		args = append(args, v)
	}

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
		if delta.Valid {
			metric.Delta = &delta.Int64
		}
		if value.Valid {
			metric.Value = &value.Float64
		}
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
	if err := s.ensureMetricsTableExists(ctx); err != nil {
		s.Logger.ErrorCtx(ctx, "failed to ensure Metrics table exists", zap.Error(err))
		return nil, err
	}
	res, err := s.getMetricsOnDBs(ctx, models...)
	if err != nil {
		return nil, err
	}
	if len(res) != 0 {
		if err := s.updateMetric(ctx, models); err != nil {
			s.Logger.ErrorCtx(ctx, "failed to update metric", zap.Error(err))
			return nil, err
		}
	} else {
		if err := s.insertMetric(ctx, models); err != nil {
			s.Logger.ErrorCtx(ctx, "failed to insert metric", zap.Error(err))
			return nil, err
		}
	}
	res, err = s.getMetricsOnDBs(ctx, models...)
	if err != nil {
		return nil, err
	}
	if len(models) < 1 {
		s.Logger.InfoCtx(ctx, "failed to find inserted/updated metric", zap.Error(err))
		return nil, fmt.Errorf("metric not found")
	}
	return res, nil
}

func (s *DBStorage) LoadFromFile(fname string) error {
	return fmt.Errorf("LoadFromFile is not supported for DBStorage")
}

func (s *DBStorage) SaveToFile(fname string) error {
	return fmt.Errorf("SaveToFile is not supported for DBStorage")
}
