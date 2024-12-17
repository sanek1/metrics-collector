package routing

import (
	"net/http"

	"github.com/go-chi/chi"
	h "github.com/sanek1/metrics-collector/internal/handlers"
	v "github.com/sanek1/metrics-collector/internal/validation"
)

func InitRouting(ms h.MetricStorage) http.Handler {
	r := chi.NewRouter()
	r.Use(v.WithLogging)

	r.Get("/*", http.HandlerFunc(ms.MainPageHandler))
	r.Get("/{value}/{type}/*", http.HandlerFunc(ms.GetMetricsByNameHandler))

	// post
	r.Post("/*", v.ValidationOld(http.HandlerFunc(h.NotImplementedHandler)))
	r.Post("/value/*", http.HandlerFunc(ms.GetMetricsByValueHandler))
	r.Route("/update", func(r chi.Router) {
		r.Post("/*", http.HandlerFunc(ms.GetMetricsHandler))
		r.Post("/gauge/*", v.ValidationOld(http.HandlerFunc(ms.GetMetricsHandler)))
		r.Post("/counter/*", v.ValidationOld(http.HandlerFunc(ms.GetMetricsHandler)))
	})

	return r
}
