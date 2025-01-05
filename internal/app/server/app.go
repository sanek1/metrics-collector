package app

import (
	"context"
	"log"
	"net/http"
	"time"

	c "github.com/sanek1/metrics-collector/internal/controller/server"
	storage "github.com/sanek1/metrics-collector/internal/storage/server"
	"github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/zap"
)

type App struct {
	controller *c.Controller
	addr       string

	storeInterval int64
	path          string
	restore       bool
	logger        *logging.ZapLogger

	store storage.Storage
}

func New(addr string, storeInterval int64, path string, restore bool) *App {
	// init zap logger
	logger, err := logging.NewZapLogger(zap.InfoLevel)
	if err != nil {
		panic(err)
	}
	l := startLogger()

	// init storage
	s := storage.NewMetricsStorage(l)

	ctrl := c.New(s, l)

	return &App{
		controller: ctrl,
		addr:       addr,

		storeInterval: storeInterval,
		path:          path,
		restore:       restore,
		logger:        logger,

		store: s,
	}
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
	a.logger.InfoCtx(context.Background(), "Running server", zap.String("address%s", a.addr))
	go a.controller.PeriodicallySaveBackUp(ctx, a.path, a.restore, time.Duration(a.storeInterval)*time.Second)
	return server.ListenAndServe()
}
