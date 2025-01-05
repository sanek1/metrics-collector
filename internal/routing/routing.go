package routing

import (
	"net/http"

	"github.com/go-chi/chi"
	h "github.com/sanek1/metrics-collector/internal/handlers"
	storage "github.com/sanek1/metrics-collector/internal/storage/server"
	v "github.com/sanek1/metrics-collector/internal/validation"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

type Controller struct {
	r         chi.Router
	l         *l.ZapLogger
	midleware *v.MiddlewareController
	storage   storage.Storage
	ms        *h.Storage
}

func New(s storage.Storage, logger *l.ZapLogger) *Controller {
	return &Controller{
		l:         logger,
		r:         chi.NewRouter(),
		midleware: v.New(logger),
		storage:   s,
		ms:        h.NewStorage(s, logger),
	}
}

func (c *Controller) InitRouting() http.Handler {
	r := chi.NewRouter()
	r.Use(c.midleware.Recover, v.GzipMiddleware)

	r.Route("/", func(r chi.Router) {
		// Get routes
		r.Get("/*", http.HandlerFunc(c.ms.MainPageHandler))
		r.Get("/{value}/{type}/*", http.HandlerFunc(c.ms.GetMetricsByNameHandler))

		// Post routes
		r.Post("/*", c.midleware.ValidationOld(http.HandlerFunc(h.NotImplementedHandler)))
		r.Post("/value/*", http.HandlerFunc(c.ms.GetMetricsByValueHandler))

		r.Route("/update", func(r chi.Router) {
			r.Post("/*", http.HandlerFunc(c.ms.GetMetricsHandler))
			r.Post("/gauge/*", c.midleware.ValidationOld(http.HandlerFunc(c.ms.GetMetricsHandler)))
			r.Post("/counter/*", c.midleware.ValidationOld(http.HandlerFunc(c.ms.GetMetricsHandler)))
		})
	})

	return r
}
