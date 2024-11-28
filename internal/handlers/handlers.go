package handlers

import (
	"net/http"
	"strings"

	"github.com/sanek1/metrics-collector/internal/storage"
)

const (
	metricName = 3
	metricVal  = 4
	minPathLen = 5
)

type IMetricStorage interface {
	SetGauge(key string, value float64)
	SetCounter(key string, value int64)
}

type MetricStorage struct {
	Storage IMetricStorage
}

func MainPage(rw http.ResponseWriter, req *http.Request) {
	data := []byte(" ---------   Main Page ---------")
	rw.Write(data)

}
func (ms MetricStorage) GaugeHandler(rw http.ResponseWriter, r *http.Request) {
	if metricKey, metricValue, err := readingDataFromURL(rw, r); err == nil {
		ms.Storage.SetGauge(metricKey, metricValue)
		rw.WriteHeader(http.StatusOK)
	} else {
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}
}

func (ms MetricStorage) CounterHandler(rw http.ResponseWriter, r *http.Request) {
	key, val, err := readingDataFromURL(rw, r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	ms.Storage.SetCounter(key, int64(val))
	rw.WriteHeader(http.StatusOK)
}

func readingDataFromURL(rw http.ResponseWriter, r *http.Request) (key string, value float64, err error) {
	splitedPath := strings.Split(r.URL.Path, "/")
	metricKey := splitedPath[metricName]
	metricValue, err := storage.StrToGauge(splitedPath[metricVal])
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	return metricKey, metricValue, nil
}
