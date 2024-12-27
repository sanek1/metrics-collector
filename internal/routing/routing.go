package routing

import (
	"net/http"

	"github.com/go-chi/chi"
	h "github.com/sanek1/metrics-collector/internal/handlers"
	v "github.com/sanek1/metrics-collector/internal/validation"
	l "github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/zap"
)

type Controller struct {
	r chi.Router
	l *l.ZapLogger
}

func New(logger *l.ZapLogger) *Controller {
	return &Controller{
		l: logger,
		r: chi.NewRouter(),
	}
}

func (c *Controller) recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				c.l.PanicCtx(r.Context(), "recovered from panic", zap.Any("panic", rec))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (c *Controller) InitRouting(ms h.MetricStorage) http.Handler {
	r := chi.NewRouter()
	r.Use(c.recover, v.GzipMiddleware)

	r.Route("/", func(r chi.Router) {
		// Get routes
		r.Get("/*", http.HandlerFunc(ms.MainPageHandler))
		r.Get("/{value}/{type}/*", http.HandlerFunc(ms.GetMetricsByNameHandler))

		// Post routes
		r.Post("/*", v.ValidationOld(http.HandlerFunc(h.NotImplementedHandler)))
		r.Post("/value/*", http.HandlerFunc(ms.GetMetricsByValueHandler))

		r.Route("/update", func(r chi.Router) {
			r.Post("/*", http.HandlerFunc(ms.GetMetricsHandler))
			r.Post("/gauge/*", v.ValidationOld(http.HandlerFunc(ms.GetMetricsHandler)))
			r.Post("/counter/*", v.ValidationOld(http.HandlerFunc(ms.GetMetricsHandler)))
		})
	})

	return r
}
