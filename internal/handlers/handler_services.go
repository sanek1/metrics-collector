package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	c "github.com/sanek1/metrics-collector/internal/config"
	v "github.com/sanek1/metrics-collector/internal/validation"
	"go.uber.org/zap"
)

func CounterService(rw http.ResponseWriter, model *v.Metrics, ms *MetricStorage) {
	newmodel := ms.Storage.SetCounter(*model)
	resp, err := json.Marshal(newmodel)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	SendResultStatusOK(rw, resp)
}

func GaugeService(rw http.ResponseWriter, model *v.Metrics, ms *MetricStorage) {
	if ok := ms.Storage.SetGauge(*model); !ok {
		http.Error(rw, "No such value exists", http.StatusNotFound)
		return
	}
	resp, err := json.Marshal(model)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	SendResultStatusOK(rw, resp)
}

func ParseMetricServices(rw http.ResponseWriter, r *http.Request) (v.Metrics, error) {
	var model v.Metrics
	if r.ContentLength == 0 {
		if err := buildJSONBody(rw, r); err != nil {
			return model, err
		}
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
		rw.WriteHeader(http.StatusBadRequest)
		return model, fmt.Errorf("unsupported request type")
	}
}

func GenerateHTMLServices(metrics []string) string {
	data := `<html lang="ru"><head><meta charset="UTF-8"><title>Метрики</title></head><body><h1>Метрики</h1>`
	for _, metric := range metrics {
		data += `<p>` + metric + `</p>`
	}
	data += `</body></html>`
	return data
}

func SendResultStatusOK(rw http.ResponseWriter, resp []byte) {
	rw.WriteHeader(http.StatusOK)
	rw.Header().Set("Content-Type", "application/json")
	_, err := rw.Write(resp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func SendResultStatusNotOK(rw http.ResponseWriter, resp []byte) {
	rw.WriteHeader(http.StatusBadRequest)
	rw.Header().Set("Content-Type", "application/json")
	_, err := rw.Write(resp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func buildJSONBody(rw http.ResponseWriter, r *http.Request) (err error) {
	key, name, val, err := readingDataFromURL(r)
	if err != nil {
		http.Error(rw, "The value does not match the expected type.", http.StatusBadRequest)
		return err
	}
	intVal := int64(*val)
	model := v.Metrics{
		ID:    name,
		MType: key,
		Delta: &intVal,
		Value: val,
	}
	resp, err := json.Marshal(model)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return err
	}
	// set body
	r.Body = io.NopCloser(bytes.NewReader(resp))
	return nil
}

func readingDataFromURL(r *http.Request) (key, name string, value *float64, err error) {
	splitedPath := strings.Split(r.URL.Path, "/")
	metrickType := splitedPath[c.TypeMetric]
	metrickName := splitedPath[c.MetricName]
	metricValue, err := strconv.ParseFloat(splitedPath[c.MetricVal], 64)
	return metrickType, metrickName, &metricValue, err
}