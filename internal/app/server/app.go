// Package app представляет собой основную библиотеку приложения
package app

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	sc "github.com/sanek1/metrics-collector/internal/controller/server"
	sf "github.com/sanek1/metrics-collector/internal/flags/server"
	"github.com/sanek1/metrics-collector/internal/grpcserver"
	ss "github.com/sanek1/metrics-collector/internal/storage/server"
	"github.com/sanek1/metrics-collector/pkg/logging"
)

type App struct {
	options     *sf.ServerOptions
	useDatabase bool
}

func New(opt *sf.ServerOptions, useDatabase bool) *App {
	return &App{
		options:     opt,
		useDatabase: useDatabase,
	}
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()
	l, err := logging.NewZapLogger(zap.InfoLevel)
	if err != nil {
		panic(err)
	}
	ctx = l.WithContextFields(ctx, zap.String("app", "Server"))

	storage := ss.GetStorage(a.useDatabase, a.options, l)
	if a.useDatabase {
		if _, ok := storage.(*ss.DBStorage); !ok {
			l.ErrorCtx(context.Background(), "Error connecting to database", zap.Error(err))
		}
	}

	if a.options.EnableGRPC {
		go func() {
			if err := grpcserver.RunGRPCServer(a.options.GRPCAddress, storage); err != nil {
				l.ErrorCtx(ctx, "gRPC server failed", zap.Error(err))
			}
		}()
	}

	fs := ss.NewMetricsStorage(l)
	ctrl := sc.NewController(fs, storage, a.options, l)

	if fs, ok := storage.(ss.FileStorage); ok {
		go fs.PeriodicallySaveBackUp(ctx, a.options.Path, a.options.Restore, time.Duration(a.options.StoreInterval)*time.Second)
	}
	if dbs, ok := storage.(ss.DatabaseStorage); ok {
		if err = dbs.EnsureMetricsTableExists(ctx); err != nil {
			l.ErrorCtx(ctx, "failed to ensure Metrics table exists", zap.Error(err))
		}
	}

	server := &http.Server{
		Addr:              a.options.FlagRunAddr,
		Handler:           ctrl.Router(),
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	l.InfoCtx(ctx, "Running server"+a.options.FlagRunAddr, zap.String("address%s", a.options.FlagRunAddr))

	err = server.ListenAndServe()
	if err != nil {
		l.FatalCtx(ctx, "Failed to start server", zap.Error(err))
	}

	<-ctx.Done()
	l.InfoCtx(ctx, "get signal to stop server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		l.FatalCtx(ctx, "Failed to shutdown server", zap.Error(err))
	}

	l.InfoCtx(ctx, "Server gracefully stopped")

	return nil
}
