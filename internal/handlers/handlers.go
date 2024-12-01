package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	c "github.com/sanek1/metrics-collector/internal/config"
	"github.com/sanek1/metrics-collector/internal/storage"
)

type IMetricStorage interface {
	SetGauge(key string, value float64) string
	SetCounter(key string, value float64) string
	GetAllMetrics() []string
	GetMetrics(metricType, metricName, metricValue string) string
}

type MetricStorage struct {
	Storage IMetricStorage
}

func NotImplementedHandler(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}

func BadRequestHandler(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Bad Request Handler", http.StatusBadRequest)
}

func (ms MetricStorage) MainPageHandler(rw http.ResponseWriter, r *http.Request) {
	res := ms.Storage.GetAllMetrics()

	data := `<html lang="ru"><head><meta charset="UTF-8"><title>Метрики</title></head><body><h1>Метрики</h1>`
	for _, metric := range res {
		data += `<p>` + metric + `</p>`
	}
	data += `</body></html>`

	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(data))
}

func (ms MetricStorage) GetMetricsByNameHandler(rw http.ResponseWriter, r *http.Request) {

	//typeMetric := chi.URLParam(r, "type")
	value := chi.URLParam(r, "value")

	//nameMetric := ""
	//if typeMetric == "gauge" {
	//	nameMetric = chi.URLParam(r, c.Gauge)
	//} else {
	//	nameMetric = chi.URLParam(r, c.Counter)
	//}
	//res := ms.Storage.GetMetrics(typeMetric, nameMetric, value)

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(value))
	//rw.Header().Set("Content-Type", "application/json")
	//json.NewEncoder(rw).Encode(struct {
	//	Value string `json:"value"`
	//}{res})
}

func (ms MetricStorage) GaugeHandler(rw http.ResponseWriter, r *http.Request) {
	key, val, err := readingDataFromURL(r)
	if err != nil {
		// log error
		http.Error(rw, "The value does not match the expected type.", http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusOK)

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(struct {
		Value string `json:"value"`
	}{ms.Storage.SetGauge(key, val)})
}

func (ms MetricStorage) CounterHandler(rw http.ResponseWriter, r *http.Request) {
	key, val, err := readingDataFromURL(r)
	if err != nil {
		http.Error(rw, "The value does not match the expected type.", http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(struct {
		Value string `json:"value"`
	}{ms.Storage.SetCounter(key, val)})
}

func readingDataFromURL(r *http.Request) (key string, value float64, err error) {
	splitedPath := strings.Split(r.URL.Path, "/")
	metricKey := splitedPath[c.MetricName]
	metricValue, err := storage.StrToGauge(splitedPath[c.MetricVal])
	return metricKey, metricValue, err
}
