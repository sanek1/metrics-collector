package handlers

import (
	"bytes"
	con "context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	c "github.com/sanek1/metrics-collector/internal/config"
	m "github.com/sanek1/metrics-collector/internal/models"
	storage "github.com/sanek1/metrics-collector/internal/storage/server"
	l "github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/zap"
)

type Services struct {
	s      storage.Storage
	models *[]m.Metrics
	logger *l.ZapLogger
	db     *sql.DB
}

type SetGaugeInterface interface {
	SetGauge(con.Context, m.Metrics) (*m.Metrics, error)
}

type MetricsServiceInterface interface {
	MetricsService(con.Context, http.ResponseWriter, []*m.Metrics)
}

func NewHandlerServices(st storage.Storage,
	db *sql.DB,
	models *[]m.Metrics,
	//sss SetGaugeInterface,
	// MetricsServiceInterface,
	zl *l.ZapLogger) *Services {
	return &Services{
		s:      st,
		models: models,
		logger: zl,
		db:     db,
	}
}

func (s *Services) PingService(ctx con.Context, rw http.ResponseWriter) {
	s.logger.InfoCtx(ctx, "PingService start")
	if s.db == nil {
		s.logger.ErrorCtx(ctx, "Database is nil")
		SendResultStatusNotOK(rw, nil)
		return
	}
	err := s.db.PingContext(ctx)
	if err != nil {
		s.logger.ErrorCtx(ctx, "Database is not available", zap.Any("err", err.Error()))
		SendResultStatusNotOK(rw, nil)
		return
	}
	s.logger.InfoCtx(ctx, "PingService success")
	SendResultStatusOK(rw, nil)
}

func (s *Services) CounterService(ctx con.Context, rw http.ResponseWriter) {
	var models []m.Metrics
	models = *s.models
	updatedModels, err := s.s.SetCounter(ctx, models...)
	if err != nil {
		s.logger.ErrorCtx(ctx, "The metric counter was not saved", zap.Any("err", err.Error()))
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	s.MetricsService(ctx, rw, updatedModels...)
}

func (s *Services) GaugeService(ctx con.Context, rw http.ResponseWriter) {
	var models []m.Metrics
	models = *s.models
	updatedModels, err := s.s.SetGauge(ctx, models...)
	if err != nil {
		str := "One or more metrics were not saved: " + err.Error() + "\n"
		s.logger.ErrorCtx(ctx, str, zap.Any("err", "metrics not saved"))
		http.Error(rw, "One or more metrics were not saved", http.StatusInternalServerError)
		return
	}
	if len(updatedModels) == 1 {
		res := updatedModels[0]
		s.MetricsService(ctx, rw, res)
		return
	}

	s.MetricsService(ctx, rw, updatedModels...)
}

func (s *Services) MetricsService(ctx con.Context, rw http.ResponseWriter, models ...*m.Metrics) {
	var err error
	var resp []byte

	if len(models) == 1 {
		resp, err = json.Marshal(models[0])
	} else {
		resp, err = json.Marshal(models)
	}

	if err != nil {
		s.logger.ErrorCtx(ctx, "The metric was not marshaled", zap.Any("err", err.Error()))
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	SendResultStatusOK(rw, resp)
}

func (s *Services) ParseMetricsServices(rw http.ResponseWriter, r *http.Request) ([]m.Metrics, error) {
	var models []m.Metrics
	if r.ContentLength == 0 {
		if err := s.buildJSONBody(rw, r); err != nil {
			s.logger.ErrorCtx(r.Context(), "The metric was not parsed", zap.Any("err", err.Error()))
			return nil, fmt.Errorf("buildJSONBody: %w", err)
		}
	}

	//todo: check r.Body array or single json string
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.ErrorCtx(r.Context(), "The metric was not read", zap.Any("err", err.Error()))
		return nil, fmt.Errorf("read body: %w", err)
	}

	var model m.Metrics
	if err := json.Unmarshal(bodyBytes, &model); err != nil {
		s.logger.InfoCtx(r.Context(), "The metric was not parsed", zap.Any("err", err.Error()))
		if err := json.Unmarshal(bodyBytes, &models); err != nil {
			s.logger.ErrorCtx(r.Context(), "The metric was not parsed", zap.Any("err", err.Error()))
			return nil, fmt.Errorf("unmarshal: %w", err)
		}
	}
	defer r.Body.Close()

	if model != (m.Metrics{}) {
		models = append(models, model)
	}

	for _, model := range models {
		if model.MType != m.TypeGauge && model.MType != m.TypeCounter {
			s.logger.ErrorCtx(r.Context(), "The metric has unsupported type", zap.Any("err", "unsupported request type"))
			return nil, fmt.Errorf("unsupported request type: %s", model.MType)
		}
	}
	return models, nil
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

func (s *Services) buildJSONBody(rw http.ResponseWriter, r *http.Request) (err error) {
	key, name, val, err := readingDataFromURL(r)
	if err != nil {
		fmt.Println("url parsing")
		http.Error(rw, "The value does not match the expected type.", http.StatusBadRequest)
		return err
	}
	intVal := int64(*val)
	model := m.Metrics{
		ID:    name,
		MType: key,
		Delta: &intVal,
		Value: val,
	}
	models := []m.Metrics{model}
	resp, err := json.Marshal(models)
	if err != nil {
		fmt.Println("marshaling error")
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return err
	}
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
