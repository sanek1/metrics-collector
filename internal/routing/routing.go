package routing

import (
	"net/http"

	"github.com/go-chi/chi"
	sf "github.com/sanek1/metrics-collector/internal/flags/server"
	h "github.com/sanek1/metrics-collector/internal/handlers"
	ss "github.com/sanek1/metrics-collector/internal/storage/server"
	v "github.com/sanek1/metrics-collector/internal/validation"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

type Controller struct {
	r              chi.Router
	l              *l.ZapLogger
	middleware     *v.MiddlewareController
	middlewareHash *v.Secret
	storage        ss.Storage
	s              *h.Storage
	opt            *sf.ServerOptions
}

func NewRouting(s ss.Storage, opt *sf.ServerOptions, logger *l.ZapLogger) *Controller {
	c := &Controller{
		l:       logger,
		r:       chi.NewRouter(),
		storage: s,
		opt:     opt,
	}

	c.s = h.NewStorage(s, logger)
	c.middleware = v.NewValidation(c.s, logger)
	c.middlewareHash = v.NewHash(opt.CryptoKey)
	return c
}

func (c *Controller) InitRouting() http.Handler {
	r := chi.NewRouter()
	r.Use(c.middleware.Recover, v.GzipMiddleware)
	if c.opt.CryptoKey != "" {
		r.Use(c.middlewareHash.HashMiddleware)
	}

	r.Route("/", func(r chi.Router) {
		// Get routes
		r.Get("/*", c.middleware.CheckForPingMiddleware(http.HandlerFunc(c.s.MainPageHandler)))
		r.Get("/ping/", http.HandlerFunc(c.s.PingDBHandler))
		r.Get("/{value}/{type}/*", http.HandlerFunc(c.s.GetMetricsByNameHandler))

		// Post routes
		r.Post("/*", c.middleware.ValidationOld(http.HandlerFunc(h.NotImplementedHandler)))
		r.Post("/value/*", http.HandlerFunc(c.s.GetMetricsByValueHandler))
		r.Post("/updates/", http.HandlerFunc(c.s.MetricHandler))

		r.Route("/update", func(r chi.Router) {
			r.Post("/*", http.HandlerFunc(c.s.MetricHandler))
			r.Post("/gauge/*", c.middleware.ValidationOld(http.HandlerFunc(c.s.MetricHandler)))
			r.Post("/counter/*", c.middleware.ValidationOld(http.HandlerFunc(c.s.MetricHandler)))
		})
	})

	return r
}
