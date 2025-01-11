package app

import (
	"context"
	"database/sql"
	"log"
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
	store         storage.Storage
}

func New(opt *flags.ServerOptions, useDatabase bool) *App {
	// init zap logger
	logger, err := logging.NewZapLogger(zap.ErrorLevel)
	if err != nil {
		panic(err)
	}
	l := startLogger()
	// init db connection
	var conn *sql.DB
	if useDatabase {
		conn, err = startDB(opt)
		if err != nil {
			logger.InfoCtx(context.Background(), opt.DBPath, zap.Error(err))
			logger.ErrorCtx(context.Background(), "Error connecting to database", zap.Error(err))
		}
	}

	fs := storage.NewMetricsStorage(l)
	s := storage.GetStorage(useDatabase, conn, l)
	ctrl := c.New(fs, s, conn, l)

	return &App{
		controller:    ctrl,
		addr:          opt.FlagRunAddr,
		storeInterval: opt.StoreInterval,
		path:          opt.Path,
		restore:       opt.Restore,
		logger:        logger,
		store:         s,
	}
}

func startDB(opt *flags.ServerOptions) (*sql.DB, error) {
	db, err := sql.Open("pgx", opt.DBPath)

	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	err = db.PingContext(context.Background())
	if err != nil {
		return nil, err
	}
	return db, nil
}

func startLogger() *logging.ZapLogger {
	ctx := context.Background()
	l, err := logging.NewZapLogger(zap.InfoLevel)

	if err != nil {
		log.Panic(err)
	}

	_ = l.WithContextFields(ctx,
		zap.String("app", "logging"))

	defer l.Sync()
	return l
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
	go a.controller.PeriodicallySaveBackUp(ctx, a.path, a.restore, time.Duration(a.storeInterval)*time.Second)
	return server.ListenAndServe()
}
