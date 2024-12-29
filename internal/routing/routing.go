package routing

import (
	"net/http"

	"github.com/go-chi/chi"
	h "github.com/sanek1/metrics-collector/internal/handlers"
	v "github.com/sanek1/metrics-collector/internal/validation"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

type Controller struct {
	r         chi.Router
	l         *l.ZapLogger
	midleware *v.MiddlewareController
}

func New(logger *l.ZapLogger) *Controller {
	return &Controller{
		l:         logger,
		r:         chi.NewRouter(),
		midleware: v.New(logger),
	}
}

func (c *Controller) InitRouting(ms h.MetricStorage) http.Handler {
	r := chi.NewRouter()
	//vv := v.New(c.l)
	r.Use(c.midleware.Recover, v.GzipMiddleware)

	r.Route("/", func(r chi.Router) {
		// Get routes
		r.Get("/*", http.HandlerFunc(ms.MainPageHandler))
		r.Get("/{value}/{type}/*", http.HandlerFunc(ms.GetMetricsByNameHandler))

		// Post routes
		r.Post("/*", c.midleware.ValidationOld(http.HandlerFunc(h.NotImplementedHandler)))
		r.Post("/value/*", http.HandlerFunc(ms.GetMetricsByValueHandler))

		r.Route("/update", func(r chi.Router) {
			r.Post("/*", http.HandlerFunc(ms.GetMetricsHandler))
			r.Post("/gauge/*", c.midleware.ValidationOld(http.HandlerFunc(ms.GetMetricsHandler)))
			r.Post("/counter/*", c.midleware.ValidationOld(http.HandlerFunc(ms.GetMetricsHandler)))
		})
	})

	return r
}
