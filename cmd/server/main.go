package main

import (
	"net"
	"net/http"

	h "github.com/sanek1/metrics-collector/internal/handlers"
	s "github.com/sanek1/metrics-collector/internal/storage"
	v "github.com/sanek1/metrics-collector/internal/validation"
)

func main() {
	handlersStorage := h.MetricStorage{
		Storage: s.NewMemStorage(),
	}

	mux := initRouting(handlersStorage)
	srv := &http.Server{
		Addr:    net.JoinHostPort(s.Address, s.Port),
		Handler: mux,
	}
	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}

func initRouting(ms h.MetricStorage) *http.ServeMux {
	mux := http.NewServeMux()
	mainPage := http.HandlerFunc(h.MainPage)
	gaugeHandler := http.HandlerFunc(ms.GaugeHandler)
	counterHandler := http.HandlerFunc(ms.CounterHandler)

	mux.HandleFunc(`/`, v.Validation(mainPage))
	mux.HandleFunc(`/update/gauge/`, v.Validation(gaugeHandler))
	mux.HandleFunc(`/update/counter/`, v.Validation(counterHandler))
	return mux
}
