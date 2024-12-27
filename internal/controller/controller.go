package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/sanek1/metrics-collector/internal/handlers"
	"github.com/sanek1/metrics-collector/internal/routing"
	"github.com/sanek1/metrics-collector/internal/storage"
	l "github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/zap"
)

type Controller struct {
	metricStorage handlers.MetricStorage
	router        http.Handler
}

func New(logger *l.ZapLogger) *Controller {
	storageImpl := storage.NewMemoryStorage(logger)
	metricStorage := handlers.MetricStorage{
		Storage: storageImpl,
		Logger:  storageImpl.Logger,
	}
	r := routing.New(storageImpl.Logger)

	return &Controller{
		metricStorage: metricStorage,
		router:        r.InitRouting(metricStorage),
	}
}

func (c *Controller) Router() http.Handler {
	return c.router
}

func (c *Controller) PeriodicallySaveBackUp(filename string, restore bool, interval time.Duration) {
	ctx := context.Background()
	ctx = c.metricStorage.Logger.WithContextFields(ctx,
		zap.String("app", "logging"),
		zap.String("service", "main"))

	ticker := time.NewTicker(interval)
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
}
