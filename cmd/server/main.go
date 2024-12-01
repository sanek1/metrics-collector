package main

import (
	"log"
	"net"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	c "github.com/sanek1/metrics-collector/internal/config"
	h "github.com/sanek1/metrics-collector/internal/handlers"
	s "github.com/sanek1/metrics-collector/internal/storage"
	v "github.com/sanek1/metrics-collector/internal/validation"
)

func main() {
	if err := RunServer(); err != nil {
		log.Fatal(err)
	}
}

func RunServer() error {

	memStorage := s.NewMemStorage()
	metricStorage := h.MetricStorage{
		Storage: memStorage,
	}

	r := InitRouting(metricStorage)
	log.Println("Server start on ", net.JoinHostPort(c.Address, c.Port))
	log.Fatal(http.ListenAndServe(net.JoinHostPort(c.Address, c.Port), r))
	return nil
}

func InitRouting(ms h.MetricStorage) http.Handler {

	r := chi.NewRouter()
	///update := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/*", http.HandlerFunc(ms.MainPageHandler))
	r.Route("/{value}/{type}", func(r chi.Router) {
		r.Get("/*", http.HandlerFunc(ms.GetMetricsByNameHandler))
		//r.Get("/{counter}", http.HandlerFunc(ms.GetMetricsByNameHandler))
	})

	r.Route("/update", func(r chi.Router) {
		r.Post("/*", h.BadRequestHandler)
		r.Post("/gauge/*", v.Validation(http.HandlerFunc(ms.GaugeHandler)))
		r.Post("/counter/*", v.Validation(http.HandlerFunc(ms.CounterHandler)))
	})
	r.Post("/*", h.NotImplementedHandler)
	r.Get("/*", h.NotImplementedHandler)
	return r
}
