package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

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
		return model, fmt.Errorf("empty body")
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
