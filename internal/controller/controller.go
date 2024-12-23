package controller

import (
	"net/http"
	"time"

	"github.com/sanek1/metrics-collector/internal/handlers"
	"github.com/sanek1/metrics-collector/internal/routing"
	"github.com/sanek1/metrics-collector/internal/storage"
)

type Controller struct {
	metricStorage handlers.MetricStorage
	router        http.Handler
}

func New() *Controller {
	storageImpl := storage.NewMemoryStorage()
	metricStorage := handlers.MetricStorage{
		Storage: storageImpl,
		Logger:  storageImpl.Logger,
	}

	return &Controller{
		metricStorage: metricStorage,
		router:        routing.InitRouting(metricStorage),
	}
}

func (c *Controller) Router() http.Handler {
	return c.router
}

func (c *Controller) PeriodicallySaveBackUp(filename string, restore bool, interval time.Duration) {
	ticker := time.NewTicker(interval)
	if restore {
		err := c.metricStorage.Storage.LoadFromFile(filename)
		if err != nil {
			c.metricStorage.Logger.Infoln("Error loading metrics from file")
		}
	}

	for range ticker.C {
		c.metricStorage.Logger.Infoln("PeriodicallySaveBackUp")
		err := c.metricStorage.Storage.SaveToFile(filename)
		if err != nil {
			c.metricStorage.Logger.Infoln("Error saving metrics to file")
		}
	}
}
