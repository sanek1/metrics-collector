package main

import (
	"log"
	"net/http"

	h "github.com/sanek1/metrics-collector/internal/handlers"
	rc "github.com/sanek1/metrics-collector/internal/routingController"
	s "github.com/sanek1/metrics-collector/internal/storage"
)

func main() {
	ParseFlags()
	if err := RunServer(); err != nil {
		log.Fatal(err)
	}
}

func RunServer() error {

	memStorage := s.NewMemStorage()
	metricStorage := h.MetricStorage{
		Storage: memStorage,
	}

	r := rc.InitRouting(metricStorage)
	log.Println("Server start on ", Options.flagRunAddr)
	log.Fatal(http.ListenAndServe(Options.flagRunAddr, r))
	return nil
}
