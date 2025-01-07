package handlers

import (
	"context"
	"database/sql"
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
	Storage storage.Storage
	Logger  *l.ZapLogger
	db      *sql.DB
}

func NewStorage(s storage.Storage, db *sql.DB, zl *l.ZapLogger) *Storage {
	return &Storage{Storage: s, Logger: zl, db: db}
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

	if m, ok := s.Storage.GetMetrics(typeMetric, nameMetric); ok {
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
	model, err := ParseMetricServices(rw, r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	if m, ok := s.Storage.GetMetrics(model.MType, model.ID); ok {
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

func (s Storage) PingDBHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer cancel()
	services := NewHandlerServices(s.Storage, s.db, nil, s.Logger)
	services.PingService(r.Context(), rw)
}

func (s Storage) GetMetricsHandler(rw http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	model, err := ParseMetricServices(rw, r)
	ctx = s.Logger.WithContextFields(ctx,
		zap.String("type", model.MType))
	if err != nil {
		s.Logger.ErrorCtx(ctx, "The metric was not parsed", zap.Any("err", err.Error()))
		SendResultStatusNotOK(rw, []byte(`{"error": "failed to read body"}`))
		return
	}

	services := NewHandlerServices(s.Storage, s.db, &model, s.Logger)

	switch model.MType {
	case "counter":
		services.CounterService(ctx, rw)
	case "gauge":
		services.GaugeService(ctx, rw)
	default:
		http.Error(rw, "No such value exists", http.StatusNotFound)
		return
	}
}

func (s Storage) SaveToFile(fname string) error {
	// serialize to json
	data, err := json.MarshalIndent(s.Storage, "", "   ")
	if err != nil {
		return err
	}
	// save to file
	return os.WriteFile(fname, data, fileMode)
}

func NotImplementedHandler(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusBadRequest)
}

func BadRequestHandler(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Bad Request Handler", http.StatusBadRequest)
}
