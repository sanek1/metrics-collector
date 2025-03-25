package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	m "github.com/sanek1/metrics-collector/internal/models"
	storage "github.com/sanek1/metrics-collector/internal/storage/server"
	l "github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/zap"
)

// Services предоставляет сервисы для обработки метрик.
// Реализует бизнес-логику для работы с различными типами метрик
// и обеспечивает взаимодействие между хранилищем и HTTP-обработчиками.
type Services struct {
	s      storage.Storage
	models *[]m.Metrics
	logger *l.ZapLogger
}

// HServices определяет интерфейс сервисов обработки метрик.
// Предоставляет методы для работы с различными типами метрик
// и проверки соединения с хранилищем.
type HServices interface {
	// PingService проверяет соединение с хранилищем метрик.
	PingService(c *gin.Context)

	// CounterService обрабатывает запросы для метрик типа counter.
	CounterService(c *gin.Context)

	// GaugeService обрабатывает запросы для метрик типа gauge.
	GaugeService(c *gin.Context)

	// HistogramService обрабатывает запросы для метрик типа histogram.
	HistogramService(c *gin.Context)

	// MetricsService обрабатывает запросы для метрик любого типа.
	MetricsService(c *gin.Context, models ...*m.Metrics)
}

// NewHandlerServices создает новый экземпляр сервисов обработки метрик.
// Параметры:
//   - st: хранилище метрик
//   - models: указатель на срез метрик (может быть nil)
//   - zl: логгер
//
// Возвращает:
//   - указатель на новый экземпляр Services
func NewHandlerServices(st storage.Storage, models *[]m.Metrics, zl *l.ZapLogger) *Services {
	return &Services{
		s:      st,
		models: models,
		logger: zl,
	}
}

// PingService проверяет соединение с хранилищем метрик.
// Проверяет доступность базы данных, если хранилище поддерживает интерфейс DatabaseStorage.
// Возвращает HTTP-ответ в зависимости от результата проверки.
func (s *Services) PingService(c *gin.Context) {
	if dbs, ok := s.s.(storage.DatabaseStorage); ok {
		if !dbs.PingIsOk() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Database connection failed"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{})
}

// CounterService обрабатывает запросы для метрик типа counter.
// Устанавливает значение метрики и возвращает результат обновления.
func (s *Services) CounterService(c *gin.Context) {
	models := *s.models
	updatedModels, err := s.s.SetCounter(c.Request.Context(), models...)
	if err != nil {
		s.logger.ErrorCtx(c.Request.Context(), "The metric counter was not saved", zap.Any("err", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.MetricsService(c, updatedModels...)
}

// GaugeService обрабатывает запросы для метрик типа gauge.
// Устанавливает значение метрики и возвращает результат обновления.
func (s *Services) GaugeService(c *gin.Context) {
	models := *s.models
	updatedModels, err := s.s.SetGauge(c.Request.Context(), models...)
	if err != nil {
		s.logger.ErrorCtx(c.Request.Context(), "One or more metrics were not saved", zap.Any("err", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "One or more metrics were not saved"})
		return
	}
	if len(updatedModels) == 1 {
		res := updatedModels[0]
		s.MetricsService(c, res)
		return
	}
	s.MetricsService(c, updatedModels...)
}

// MetricsService обрабатывает запросы для метрик любого типа.
// Обрабатывает список метрик и возвращает результат обновления.
func (s *Services) MetricsService(c *gin.Context, models ...*m.Metrics) {
	if len(models) == 1 {
		model := models[0]
		c.JSON(http.StatusOK, model)
		return
	}
	c.JSON(http.StatusOK, models)
}

// ParseMetricsServices разбирает метрики из тела HTTP-запроса.
// Поддерживает разбор как одиночной метрики, так и массива метрик в формате JSON.
// Также поддерживает сжатие gzip.
//
// Возвращает:
//   - срез разобранных метрик
//   - ошибку, если не удалось разобрать метрики
func (s *Services) ParseMetricsServices(c *gin.Context) ([]m.Metrics, error) {
	var models []m.Metrics
	r := c.Request

	if r.ContentLength == 0 {
		if err := s.buildJSONBody(c); err != nil {
			s.logger.ErrorCtx(r.Context(), "The metric was not parsed", zap.Any("err", err.Error()))
			return nil, fmt.Errorf("buildJSONBody: %w", err)
		}
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.ErrorCtx(r.Context(), "The metric was not read", zap.Any("err", err.Error()))
		return nil, fmt.Errorf("read body: %w", err)
	}
	if c.GetHeader("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(bytes.NewReader(bodyBytes))
		if err != nil {
			s.logger.ErrorCtx(r.Context(), "Failed to decompress gzip data", zap.Error(err))
			return nil, fmt.Errorf("gzip decompression: %w", err)
		}
		decompressed, err := io.ReadAll(reader)
		if err != nil {
			s.logger.ErrorCtx(r.Context(), "Failed to read decompressed data", zap.Error(err))
			return nil, fmt.Errorf("read decompressed: %w", err)
		}
		bodyBytes = decompressed
	}

	var model m.Metrics
	if err := json.Unmarshal(bodyBytes, &model); err != nil {
		if err := json.Unmarshal(bodyBytes, &models); err != nil {
			s.logger.ErrorCtx(r.Context(), "The metric was not parsed"+string(bodyBytes)+err.Error(), zap.Any("err", err.Error()))
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

// GetMetricsByValueGin обрабатывает запрос на получение значения метрики через JSON API.
// Разбирает метрику из тела запроса, получает её значение из хранилища
// и возвращает результат.
func (s *Services) GetMetricsByValueGin(c *gin.Context) {
	bodyBytes, err := c.GetRawData()
	if err != nil {
		s.logger.ErrorCtx(c.Request.Context(), "Failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read request body",
		})
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var requestMetric m.Metrics
	if err := json.Unmarshal(bodyBytes, &requestMetric); err != nil {
		s.logger.ErrorCtx(c.Request.Context(), "Failed to parse metric request",
			zap.Error(err),
			zap.String("body", string(bodyBytes)))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid metric request format",
		})
		return
	}

	if requestMetric.ID == "" || requestMetric.MType == "" {
		s.logger.WarnCtx(c.Request.Context(), "Missing metric ID or type in request",
			zap.String("id", requestMetric.ID),
			zap.String("type", requestMetric.MType))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Metric ID and type are required",
		})
		return
	}

	s.logger.InfoCtx(c.Request.Context(), "Getting metric value",
		zap.String("id", requestMetric.ID),
		zap.String("type", requestMetric.MType))

	metric, found := s.s.GetMetrics(c.Request.Context(), requestMetric.MType, requestMetric.ID)
	if !found {
		s.logger.WarnCtx(c.Request.Context(), "Metric not found",
			zap.String("id", requestMetric.ID),
			zap.String("type", requestMetric.MType))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Metric not found",
		})
		return
	}

	s.logger.InfoCtx(c.Request.Context(), "Metric found, returning value",
		zap.String("id", metric.ID),
		zap.String("type", metric.MType))

	c.JSON(http.StatusOK, metric)
}

// buildJSONBody формирует тело JSON-ответа для метрики.
// Используется для формирования ответа при обработке метрик.
//
// Возвращает:
//   - ошибку, если не удалось сформировать тело ответа
func (s *Services) buildJSONBody(c *gin.Context) (err error) {
	metricType := c.Param("metricType")
	metricName := c.Param("metricName")
	metricValue, err := strconv.ParseFloat(c.Param("metricValue"), 64)
	if err != nil {
		s.logger.ErrorCtx(c.Request.Context(), "Failed to parse metric value", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "The value does not match the expected type."})
		return err
	}

	intVal := int64(metricValue)
	model := m.Metrics{
		ID:    metricName,
		MType: metricType,
		Delta: &intVal,
		Value: &metricValue,
	}

	models := []m.Metrics{model}
	resp, err := json.Marshal(models)

	if err != nil {
		s.logger.ErrorCtx(c.Request.Context(), "Marshaling error", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(resp))
	return nil
}

// GenerateHTMLServices генерирует HTML-страницу со списком всех метрик.
// Параметры:
//   - metrics: срез строк с именами метрик
//
// Возвращает:
//   - строку с HTML-разметкой
func GenerateHTMLServices(metrics []string) string {
	data := `<html lang="ru"><head><meta charset="UTF-8"><title>Метрики</title></head><body><h1>Метрики</h1>`
	for _, metric := range metrics {
		data += `<p>` + metric + `</p>`
	}
	data += `</body></html>`
	return data
}

// SendResultStatusOK отправляет успешный ответ с данными.
// Параметры:
//   - rw: объект ResponseWriter для записи ответа
//   - resp: данные для отправки
func SendResultStatusOK(rw http.ResponseWriter, resp []byte) {
	rw.WriteHeader(http.StatusOK)
	rw.Header().Set("Content-Type", "application/json")
	_, err := rw.Write(resp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

// SendResultStatusNotOK отправляет ответ с ошибкой.
// Параметры:
//   - rw: объект ResponseWriter для записи ответа
//   - err: ошибка для отправки
func SendResultStatusNotOK(rw http.ResponseWriter, err error) {
	http.Error(rw, err.Error(), http.StatusBadRequest)
}
