package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	c "github.com/sanek1/metrics-collector/internal/config"
	v "github.com/sanek1/metrics-collector/internal/validation"
	"go.uber.org/zap"
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
		resp, err := json.Marshal(m)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		sendResultStatusOK(rw, resp)
		return
	}
	http.Error(rw, "No such value exists", http.StatusNotFound)
}

func (ms MetricStorage) GetMetricsByValueHandler(rw http.ResponseWriter, r *http.Request) {
	model, err := parseModel(rw, r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	if m, ok := ms.Storage.GetMetrics(model.MType, model.ID); ok {
		resp, err := json.Marshal(m)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		sendResultStatusOK(rw, resp)
		return
	}
	http.Error(rw, "No such value exists", http.StatusNotFound)
}

func (ms MetricStorage) GaugeHandler(rw http.ResponseWriter, r *http.Request) {
	model, err := parseModel(rw, r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if ok := ms.Storage.SetGauge(model); !ok {
		http.Error(rw, "No such value exists", http.StatusNotFound)
		return
	}
	resp, err := json.Marshal(model)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	sendResultStatusOK(rw, resp)
}

func (ms MetricStorage) CounterHandler(rw http.ResponseWriter, r *http.Request) {
	model, err := parseModel(rw, r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	newmodel := ms.Storage.SetCounter(model)
	resp, err := json.Marshal(newmodel)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	sendResultStatusOK(rw, resp)
}

func parseModel(rw http.ResponseWriter, r *http.Request) (v.Metrics, error) {
	var model v.Metrics
	if r.ContentLength == 0 {
		key, name, val, err := readingDataFromURL(r)
		if err != nil {
			http.Error(rw, "The value does not match the expected type.", http.StatusBadRequest)
			return model, err
		}
		intVal := int64(*val)
		model = v.Metrics{
			ID:    name,
			MType: key,
			Delta: &intVal,
			Value: val,
		}
		resp, err := json.Marshal(model)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return model, err
		}
		r.Body = io.NopCloser(bytes.NewReader(resp))
	}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&model); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return model, err
	}

	if model.MType == v.TypeGauge || model.MType == v.TypeCounter {
		return model, nil
	} else {
		v.Loger.Error("unsupported request type", zap.String("type", model.MType))
		rw.WriteHeader(http.StatusInternalServerError)
		return model, fmt.Errorf("unsupported request type: %s", model.MType)
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

func sendResultStatusOK(rw http.ResponseWriter, resp []byte) {
	rw.WriteHeader(http.StatusOK)
	rw.Header().Set("Content-Type", "application/json")
	_, err := rw.Write(resp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func readingDataFromURL(r *http.Request) (key, name string, value *float64, err error) {
	splitedPath := strings.Split(r.URL.Path, "/")

	metrickType := splitedPath[c.TypeMetric]
	metrickName := splitedPath[c.MetricName]
	metricValue, err := convertStringToFloat64(splitedPath[c.MetricVal])

	return metrickType, metrickName, metricValue, err
}

func convertStringToFloat64(str string) (*float64, error) {
	floatVal, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return nil, err
	}
	return &floatVal, nil
}
