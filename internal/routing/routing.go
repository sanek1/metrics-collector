package routing

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi"
	h "github.com/sanek1/metrics-collector/internal/handlers"
	storage "github.com/sanek1/metrics-collector/internal/storage/server"
	v "github.com/sanek1/metrics-collector/internal/validation"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

type Controller struct {
	r          chi.Router
	l          *l.ZapLogger
	middleware *v.MiddlewareController
	storage    storage.Storage
	s          *h.Storage
}

func New(s storage.Storage, db *sql.DB, logger *l.ZapLogger) *Controller {
	c := &Controller{
		l:       logger,
		r:       chi.NewRouter(),
		storage: s,
	}

	c.s = h.NewStorage(s, db, logger)
	c.middleware = v.New(c.s, logger)
	return c
}

func (c *Controller) InitRouting() http.Handler {
	r := chi.NewRouter()
	r.Use(c.middleware.Recover, v.GzipMiddleware)

	r.Route("/", func(r chi.Router) {
		// Get routes
		r.Get("/*", c.middleware.CheckForPingMiddleware(http.HandlerFunc(c.s.MainPageHandler)))
		r.Get("/ping/", http.HandlerFunc(c.s.PingDBHandler))
		r.Get("/{value}/{type}/*", http.HandlerFunc(c.s.GetMetricsByNameHandler))

		// Post routes
		r.Post("/*", c.middleware.ValidationOld(http.HandlerFunc(h.NotImplementedHandler)))
		r.Post("/value/*", http.HandlerFunc(c.s.GetMetricsByValueHandler))

		r.Route("/update", func(r chi.Router) {
			r.Post("/*", http.HandlerFunc(c.s.GetMetricsHandler))
			r.Post("/gauge/*", c.middleware.ValidationOld(http.HandlerFunc(c.s.GetMetricsHandler)))
			r.Post("/counter/*", c.middleware.ValidationOld(http.HandlerFunc(c.s.GetMetricsHandler)))
		})
	})

	return r
}
