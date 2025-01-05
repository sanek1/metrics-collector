package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/sanek1/metrics-collector/internal/handlers"
	"github.com/sanek1/metrics-collector/internal/routing"
	storage "github.com/sanek1/metrics-collector/internal/storage/server"
	l "github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/zap"
)

type Controller struct {
	metricStorage handlers.Storage
	router        http.Handler
}

func New(logger *l.ZapLogger) *Controller {
	metricStorage := handlers.Storage{
		Storage: storage.NewMetricsStorage(logger),
		Logger:  logger,
	}
	r := routing.New(logger)

	return &Controller{
		metricStorage: metricStorage,
		router:        r.InitRouting(metricStorage),
	}
}

func (c *Controller) Router() http.Handler {
	return c.router
}

func (c *Controller) PeriodicallySaveBackUp(ctx context.Context, filename string, restore bool, interval time.Duration) {
	ctx = c.metricStorage.Logger.WithContextFields(ctx,
		zap.String("app", "logging"),
		zap.String("service", "main"))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	if restore {
		err := c.metricStorage.Storage.LoadFromFile(filename)
		if err != nil {
			c.metricStorage.Logger.ErrorCtx(ctx, "Error loading metrics from file")
		}
	}

	for range ticker.C {
		err := c.metricStorage.Storage.SaveToFile(filename)
		c.metricStorage.Logger.InfoCtx(ctx, "saving to file was successful")
		if err != nil {
			c.metricStorage.Logger.ErrorCtx(ctx, "Error saving metrics to file"+err.Error())
		}
	}

	for {
		select {
		case <-ticker.C:
			err := c.metricStorage.Storage.SaveToFile(filename)
			c.metricStorage.Logger.InfoCtx(ctx, "saving to file was successful")
			if err != nil {
				c.metricStorage.Logger.ErrorCtx(ctx, "Error saving metrics to file"+err.Error())
			}

		case <-ctx.Done():
			c.metricStorage.Logger.InfoCtx(ctx, "Backup process stopped.")
			return
		}
	}
}
