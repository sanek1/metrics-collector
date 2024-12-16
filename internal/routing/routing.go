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
	r.Post("/value/*", http.HandlerFunc(ms.GetMetricsByValueHandler))
	// post
	r.Post("/*", v.ValidationOld(http.HandlerFunc(h.NotImplementedHandler)))
	r.Route("/update", func(r chi.Router) {
		//r.Post("/*", v.ValidationOld(http.HandlerFunc(h.BadRequestHandler)))
		r.Post("/*", http.HandlerFunc(ms.GetMetricsByValueHandler2))
		//r.Post("/*", http.HandlerFunc(ms.GetMetricsByValueHandler2))
		r.Post("/gauge/*", v.ValidationOld(http.HandlerFunc(ms.GaugeHandler)))
		r.Post("/counter/*", v.ValidationOld(http.HandlerFunc(ms.CounterHandler)))
	})

	return r
}
