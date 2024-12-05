package controller

import (
	"net/http"

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
	}

	return &Controller{
		metricStorage: metricStorage,
		router:        routing.InitRouting(metricStorage),
	}
}

func (c *Controller) Router() http.Handler {
	return c.router
}
