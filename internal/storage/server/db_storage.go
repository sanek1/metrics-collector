package storage

import (
	"context"
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	flags "github.com/sanek1/metrics-collector/internal/flags/server"
	m "github.com/sanek1/metrics-collector/internal/models"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

type DBStorage struct {
	conn   *pgxpool.Pool
	Logger *l.ZapLogger
}

const (
	selectAllMetricsQuery = "SELECT key, m_type, delta, value FROM metrics"
)

func NewDBStorage(opt *flags.ServerOptions, logger *l.ZapLogger) *DBStorage {
	ctx := context.Background()
	dbConnection, err := startDBConnection(ctx, opt)
	if err != nil {
		logger.ErrorCtx(context.Background(), "Error connecting to database", zap.Error(err))
	}
	return &DBStorage{
		conn:   dbConnection,
		Logger: logger,
	}
}

func (s *DBStorage) SetGauge(ctx context.Context, models ...m.Metrics) ([]*m.Metrics, error) {
	return s.SetMetrics(ctx, models...)
}
func (s *DBStorage) SetCounter(ctx context.Context, models ...m.Metrics) ([]*m.Metrics, error) {
	return s.SetMetrics(ctx, models...)
}

func (s *DBStorage) GetAllMetrics() []string {
	var res []string
	ctx := context.Background()
	rows, err := s.conn.Query(ctx, selectAllMetricsQuery)
	if err != nil {
		s.Logger.ErrorCtx(ctx, "failed to get all metrics from database", zap.Error(err))
		return nil
	}
	if rows.Err() != nil {
		s.Logger.ErrorCtx(ctx, "failed to get all metrics from database", zap.Error(rows.Err()))
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		var metric m.Metrics
		if err := rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value); err != nil {
			s.Logger.ErrorCtx(ctx, "failed to scan metric from database", zap.Error(err))
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

	models, err := s.GetMetricsOnDBs(ctx, metric)
	if err != nil || len(models) == 0 {
		return nil, false
	}
	model := models[0]
	return model, true
}

func (s *DBStorage) PingIsOk() bool {
	if s.conn == nil {
		return false
	}
	return s.conn.Ping(context.Background()) == nil
}

func (s *DBStorage) EnsureMetricsTableExists(ctx context.Context) error {
	var exists bool
	err := s.conn.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pg_catalog.pg_tables WHERE tablename = 'metrics')`).Scan(&exists)
	if err != nil {
		s.Logger.ErrorCtx(ctx, "failed to check if Metrics table exists", zap.Error(err))
		return err
	}

	if !exists {
		db, err := sql.Open("postgres", s.conn.Config().ConnString())
		if err != nil {
			s.Logger.ErrorCtx(ctx, "failed to acquire connection: %w", zap.Error(err))
		}
		driver, err := pgx.WithInstance(db, &pgx.Config{})
		if err != nil {
			s.Logger.ErrorCtx(ctx, "failed to create migration driver %w", zap.Error(err))
		}

		migration, err := migrate.NewWithDatabaseInstance(
			"file:../../internal/storage/migrations",
			"MetricStore",
			driver,
		)
		if err != nil {
			s.Logger.ErrorCtx(ctx, "failed to create migration instance %w", zap.Error(err))
		}

		if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
			return err
		}
		s.Logger.InfoCtx(ctx, "created Metrics table")
	}
	return nil
}
