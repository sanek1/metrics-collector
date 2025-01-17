package app

import (
	"context"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	c "github.com/sanek1/metrics-collector/internal/controller/server"
	flags "github.com/sanek1/metrics-collector/internal/flags/server"
	storage "github.com/sanek1/metrics-collector/internal/storage/server"
	"github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/zap"
)

type App struct {
	controller    *c.Controller
	addr          string
	storeInterval int64
	path          string
	restore       bool
	logger        *logging.ZapLogger
	storage       storage.Storage
}

func New(opt *flags.ServerOptions, useDatabase bool) *App {
	// init zap logger
	logger, err := logging.NewZapLogger(zap.ErrorLevel)
	if err != nil {
		panic(err)
	}
	//l := startLogger()
	s := storage.GetStorage(useDatabase, opt, logger)
	if useDatabase {
		if _, ok := s.(*storage.DBStorage); !ok {
			logger.InfoCtx(context.Background(), opt.DBPath, zap.Error(err))
			logger.ErrorCtx(context.Background(), "Error connecting to database", zap.Error(err))
		}

	}

	fs := storage.NewMetricsStorage(logger)
	ctrl := c.NewController(fs, s, logger)

	return &App{
		controller:    ctrl,
		addr:          opt.FlagRunAddr,
		storeInterval: opt.StoreInterval,
		path:          opt.Path,
		restore:       opt.Restore,
		logger:        logger,
		storage:       s,
	}
}

func (a *App) Run() error {
	ctx := context.Background()
	server := &http.Server{
		Addr:              a.addr,
		Handler:           a.controller.Router(),
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}
	a.logger.InfoCtx(ctx, "Running server", zap.String("address%s", a.addr))

	if fs, ok := a.storage.(storage.FileStorage); ok {
		go fs.PeriodicallySaveBackUp(ctx, a.path, a.restore, time.Duration(a.storeInterval)*time.Second)
	}
	if dbs, ok := a.storage.(storage.DatabaseStorage); ok {
		// check if table exists
		if err := dbs.EnsureMetricsTableExists(ctx); err != nil {
			a.logger.ErrorCtx(ctx, "failed to ensure Metrics table exists", zap.Error(err))
		}
	}

	return server.ListenAndServe()
}

// func startLogger() *logging.ZapLogger {
// 	ctx := context.Background()
// 	l, err := logging.NewZapLogger(zap.InfoLevel)

// 	if err != nil {
// 		log.Panic(err)
// 	}

// 	_ = l.WithContextFields(ctx,
// 		zap.String("app", "logging"))

// 	defer l.Sync()
// 	return l
// }
