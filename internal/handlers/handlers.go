package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

func (ms MetricStorage) MainPageHandler(rw http.ResponseWriter, r *http.Request) {
	metrics := ms.Storage.GetAllMetrics()
	htmlData := GenerateHTMLServices(metrics)
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

	ms.Logger.Infoln(
		"uri", r.RequestURI,
		"method", r.Method,
		"Content-Type", r.Header.Get("Content-Type"),
		"hander", "GetMetricsByNameHandler",
		"typeMetric", typeMetric,
		"nameMetric", nameMetric,
	)

	if m, ok := ms.Storage.GetMetrics(typeMetric, nameMetric); ok {
		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		answer := ""
		if m.MType == "counter" {
			answer = fmt.Sprint(*m.Delta)
		} else if m.MType == "gauge" {
			answer = fmt.Sprint(*m.Value)
		}

		if _, err := rw.Write([]byte(answer)); err != nil {
			log.Printf("Error writing response: %v", err)
		}
		return
	}
	http.Error(rw, "No such value exists", http.StatusNotFound)
}

func (ms MetricStorage) GetMetricsByValueHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	model, err := ParseMetricServices(rw, r)
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
		SendResultStatusOK(rw, resp)
		return
	}
	http.Error(rw, "No such value exists", http.StatusNotFound)
}

func (ms MetricStorage) GetMetricsHandler(rw http.ResponseWriter, r *http.Request) {
	model, err := ParseMetricServices(rw, r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		SendResultStatusNotOK(rw, []byte(err.Error()))
		return
	}
	switch model.MType {
	case "counter":
		CounterService(rw, &model, &ms)
	case "gauge":
		GaugeService(rw, &model, &ms)
	default:
		http.Error(rw, "No such value exists", http.StatusNotFound)
		SendResultStatusNotOK(rw, []byte(err.Error()))
	}
}

func NotImplementedHandler(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusBadRequest)
}

func BadRequestHandler(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Bad Request Handler", http.StatusBadRequest)
}
