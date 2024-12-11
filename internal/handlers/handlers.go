package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	c "github.com/sanek1/metrics-collector/internal/config"
	"github.com/sanek1/metrics-collector/internal/storage"
)

func (ms MetricStorage) MainPageHandler(rw http.ResponseWriter, r *http.Request) {
	metrics := ms.Storage.GetAllMetrics()
	htmlData := generateHTML(metrics)
	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(http.StatusOK)
	_, err := rw.Write([]byte(htmlData))
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func (ms MetricStorage) GetMetricsByNameHandler(rw http.ResponseWriter, r *http.Request) {
	typeMetric := chi.URLParam(r, "type")
	nameMetric := chi.URLParam(r, "*")

	if m, ok := ms.Storage.GetMetrics(typeMetric, nameMetric); ok {
		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if _, err := rw.Write([]byte(m)); err != nil {
			log.Printf("Error writing response: %v", err)
		}
		return
	}
	http.Error(rw, "No such value exists", http.StatusNotFound)
}

func (ms MetricStorage) GaugeHandler(rw http.ResponseWriter, r *http.Request) {
	key, val, err := readingDataFromURL(r)
	if err != nil {
		http.Error(rw, "The value does not match the expected type.", http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(rw).Encode(struct {
		Value string `json:"value"`
	}{ms.Storage.SetGauge(key, val)})
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func (ms MetricStorage) CounterHandler(rw http.ResponseWriter, r *http.Request) {
	key, val, err := readingDataFromURL(r)
	if err != nil {
		http.Error(rw, "The value does not match the expected type.", http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(rw).Encode(struct {
		Value string `json:"value"`
	}{ms.Storage.SetCounter(key, val)})
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func NotImplementedHandler(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusBadRequest)
}

func BadRequestHandler(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Bad Request Handler", http.StatusBadRequest)
}

func generateHTML(metrics []string) string {
	data := `<html lang="ru"><head><meta charset="UTF-8"><title>Метрики</title></head><body><h1>Метрики</h1>`
	for _, metric := range metrics {
		data += `<p>` + metric + `</p>`
	}
	data += `</body></html>`
	return data
}

func readingDataFromURL(r *http.Request) (key string, value float64, err error) {
	splitedPath := strings.Split(r.URL.Path, "/")
	metricKey := splitedPath[c.MetricName]
	metricValue, err := storage.StrToGauge(splitedPath[c.MetricVal])
	return metricKey, metricValue, err
}
