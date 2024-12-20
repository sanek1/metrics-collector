package routing

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	h "github.com/sanek1/metrics-collector/internal/handlers"
	v "github.com/sanek1/metrics-collector/internal/validation"
)

func InitRouting(ms h.MetricStorage) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// get
	r.Get("/*", http.HandlerFunc(ms.MainPageHandler))
	r.Get("/{value}/{type}/*", http.HandlerFunc(ms.GetMetricsByNameHandler))

	// post
	r.Post("/*", v.Validation(http.HandlerFunc(h.NotImplementedHandler)))
	r.Route("/update", func(r chi.Router) {
		r.Post("/*", v.Validation(http.HandlerFunc(h.BadRequestHandler)))
		r.Post("/gauge/*", v.Validation(http.HandlerFunc(ms.GaugeHandler)))
		r.Post("/counter/*", v.Validation(http.HandlerFunc(ms.CounterHandler)))
	})

	return r
}
