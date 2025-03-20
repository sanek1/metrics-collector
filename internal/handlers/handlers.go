package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	storage "github.com/sanek1/metrics-collector/internal/storage/server"
	l "github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/zap"
)

const (
	fileMode = 0600
)

type Storage struct {
	Storage         storage.Storage
	Logger          *l.ZapLogger
	handlerServices *Services
}

func NewStorage(s storage.Storage, zl *l.ZapLogger) *Storage {
	hs := NewHandlerServices(s, nil, zl)

	return &Storage{Storage: s, Logger: zl, handlerServices: hs}
}

func (s Storage) MainPageHandler(rw http.ResponseWriter, r *http.Request) {
	metrics := s.Storage.GetAllMetrics()
	htmlData := GenerateHTMLServices(metrics)
	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(http.StatusOK)
	_, err := rw.Write([]byte(htmlData))
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func (s Storage) GetMetricsByNameHandler(rw http.ResponseWriter, r *http.Request) {
	typeMetric := chi.URLParam(r, "type")
	nameMetric := chi.URLParam(r, "*")
	s.Logger.InfoCtx(r.Context(),
		fmt.Sprintf("handler GetMetricsByNameHandler. GetMetricsByNameHandler typeMetric %s nameMetric %s", typeMetric, nameMetric))

	if m, ok := s.Storage.GetMetrics(r.Context(), typeMetric, nameMetric); ok {
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

func (s Storage) GetMetricsByValueHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	models, err := s.handlerServices.ParseMetricsServices(rw, r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	model := &models[0]
	if m, ok := s.Storage.GetMetrics(r.Context(), model.MType, model.ID); ok {
		resp, err := json.Marshal(m)
		if err != nil {
			s.Logger.ErrorCtx(r.Context(), "The metric not marshaled", zap.Any("err", err.Error()))
			return
		}
		SendResultStatusOK(rw, resp)
		return
	} else {
		http.Error(rw, "No such value exists", http.StatusNotFound)
		return
	}
}

func (s Storage) MetricHandler(rw http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	models, err := s.handlerServices.ParseMetricsServices(rw, r)
	if err != nil {
		s.Logger.ErrorCtx(ctx, "The metric was not parsed")
		SendResultStatusNotOK(rw, err )
		return
	}
	s.handlerServices.models = &models
	switch models[0].MType {
	case "counter":
		s.handlerServices.CounterService(ctx, rw)
	case "gauge":
		s.handlerServices.GaugeService(ctx, rw)
	default:
		http.Error(rw, "No such value exists3", http.StatusNotFound)
		return
	}
}

func (s Storage) SaveToFile(fname string) error {
	data, err := json.MarshalIndent(s.Storage, "", "   ")
	if err != nil {
		return err
	}
	go func() {
		if err := os.WriteFile(fname, data, fileMode); err != nil {
			s.Logger.ErrorCtx(context.Background(), "Async save failed", zap.Any("err", err.Error()))
		}
	}()
	return nil
}

func (s Storage) PingDBHandler(rw http.ResponseWriter, r *http.Request) {
	s.Logger.InfoCtx(r.Context(), "handler PingDBHandler")
	rw.Header().Set("Content-Type", "application/json")
	s.handlerServices.models = nil
	s.handlerServices.PingService(r.Context(), rw)
}

func NotImplementedHandler(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusBadRequest)
}

func BadRequestHandler(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Bad Request Handler", http.StatusBadRequest)
}

// func BuildMetricByChiParam(r *http.Request) (*store.Metric, error) {
// 	metric := &store.Metric{}
// 	metric.MType = chi.URLParam(r, "metricType")
// 	metric.ID = chi.URLParam(r, "metricName")
// 	v := chi.URLParam(r, "metricValue")
// 	if metric.MType == store.MTypeCounter {
// 		delta, err := strconv.ParseInt(v, 10, 64)
// 		if err != nil {
// 			return nil, err
// 		}
// 		metric.Delta = &delta
// 	} else if metric.MType == store.MTypeGauge {
// 		value, err := strconv.ParseFloat(v, 64)
// 		if err != nil {
// 			return nil, err
// 		}
// 		metric.Value = &value
// 	} else {
// 		return nil, errors.New("unknown type")
// 	}

// 	return metric, nil
// }
