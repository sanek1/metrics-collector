package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	m "github.com/sanek1/metrics-collector/internal/models"
	storage "github.com/sanek1/metrics-collector/internal/storage/server"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

const (
	fileMode = 0600
)

// Storage представляет собой структуру для обработки HTTP-запросов к метрикам.
// Предоставляет методы для взаимодействия с хранилищем метрик через HTTP API.
type Storage struct {
	Storage         storage.Storage
	Logger          *l.ZapLogger
	handlerServices *Services
}

// NewStorage создает новый экземпляр обработчика метрик.
// Параметры:
//   - s: хранилище метрик
//   - zl: логгер
//
// Возвращает:
//   - указатель на новый экземпляр Storage
func NewStorage(s storage.Storage, zl *l.ZapLogger) *Storage {
	hs := NewHandlerServices(s, nil, zl)

	return &Storage{Storage: s, Logger: zl, handlerServices: hs}
}

// MainPageHandler обрабатывает запрос к главной странице, отображая все доступные метрики.
// Возвращает HTML-страницу со списком всех метрик.
func (s Storage) MainPageHandler(c *gin.Context) {
	metrics := s.Storage.GetAllMetrics()
	htmlData := GenerateHTMLServices(metrics)
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, htmlData)
}

// GetMetricsByNameHandler обрабатывает запрос на получение метрики по имени и типу.
// URL-параметры:
//   - metricName: имя метрики
//   - metricType: тип метрики (gauge или counter)
//
// Возвращает значение метрики или ошибку, если метрика не найдена.
// @Summary Получить значение метрики
// @Description Возвращает значение метрики по имени и типу
// @Tags Metrics
// @Produce  json
// @Param metricType path string true "Тип метрики"
// @Param metricName path string true "Имя метрики"
// @Success 200 {object} m.Metrics
// @Router /value/{metricType}/{metricName} [get]
func (s Storage) GetMetricsByNameHandler(c *gin.Context) {
	nameMetric := c.Param("metricName")
	typeMetric := c.Param("metricType")
	s.Logger.InfoCtx(c.Request.Context(),
		fmt.Sprintf("handler GetMetricsByNameHandler. GetMetricsByNameHandler typeMetric %s nameMetric %s", typeMetric, nameMetric))

	if metric, ok := s.Storage.GetMetrics(c.Request.Context(), typeMetric, nameMetric); ok {
		c.Header("Content-Type", "text/plain; charset=utf-8")
		answer := ""
		if metric.MType == m.TypeCounter {
			answer = fmt.Sprint(*metric.Delta)
		} else if metric.MType == m.TypeGauge {
			answer = fmt.Sprint(*metric.Value)
		}

		c.String(http.StatusOK, answer)
		return
	}
	c.String(http.StatusNotFound, "GetMetricsByNameHandler No such value exists")
}

// GetMetricsByValueHandler обрабатывает запрос на получение метрики через JSON-формат.
// Делегирует обработку запроса сервису handlerServices.
// @Success 200 {object} m.Metrics
// @Failure 400 {object} m.Metrics
// @Failure 500 {object} m.Metrics
// @Router /metrics [get]
// @Success 200 {object} m.Metrics
// @Failure 400 {object} m.Metrics
func (s Storage) GetMetricsByValueHandler(c *gin.Context) {
	s.Logger.InfoCtx(c.Request.Context(), "Handling GetMetricsByValueHandler request")
	s.handlerServices.GetMetricsByValueGin(c)
}

// UpdateMetricFromURLHandler обрабатывает запрос на обновление метрики через URL-параметры.
// URL-параметры:
//   - metricType: тип метрики (gauge или counter)
//   - metricName: имя метрики
//   - metricValue: значение метрики
//
// Обновляет метрику и возвращает результат операции.
// @author Firstname Lastname
func (s Storage) UpdateMetricFromURLHandler(c *gin.Context) {
	metricType := c.Param("metricType")
	metricName := c.Param("metricName")
	metricValue := c.Param("metricValue")

	s.Logger.InfoCtx(c.Request.Context(), "Processing URL-based metric update",
		zap.String("type", metricType),
		zap.String("name", metricName),
		zap.String("value", metricValue))

	var metric m.Metrics
	metric.ID = metricName
	metric.MType = metricType

	switch metricType {
	case m.TypeCounter:
		delta, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			s.Logger.WarnCtx(c.Request.Context(), "Invalid counter value",
				zap.String("value", metricValue),
				zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid counter value",
				"details": err.Error(),
			})
			return
		}
		metric.Delta = &delta

	case m.TypeGauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			s.Logger.WarnCtx(c.Request.Context(), "Invalid gauge value",
				zap.String("value", metricValue),
				zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid gauge value",
				"details": err.Error(),
			})
			return
		}
		metric.Value = &value

	default:
		s.Logger.WarnCtx(c.Request.Context(), "Unsupported metric type",
			zap.String("type", metricType))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unsupported metric type",
			"type":  metricType,
		})
		return
	}

	var updatedMetrics []*m.Metrics
	var err error

	if metricType == m.TypeCounter {
		updatedMetrics, err = s.handlerServices.s.SetCounter(c.Request.Context(), metric)
	} else {
		updatedMetrics, err = s.handlerServices.s.SetGauge(c.Request.Context(), metric)
	}

	if err != nil {
		s.Logger.ErrorCtx(c.Request.Context(), "Failed to update metric",
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update metric",
			"details": err.Error(),
		})
		return
	}

	if len(updatedMetrics) == 0 {
		s.Logger.WarnCtx(c.Request.Context(), "No metrics were updated")
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "No metrics were updated",
		})
		return
	}

	s.Logger.InfoCtx(c.Request.Context(), "Successfully updated metric",
		zap.String("id", updatedMetrics[0].ID),
		zap.String("type", updatedMetrics[0].MType))

	c.JSON(http.StatusOK, updatedMetrics[0])
}

// MetricHandler обрабатывает запрос на обновление метрики через JSON в теле запроса.
// Принимает JSON-представление метрики и делегирует обработку соответствующему сервису
// в зависимости от типа метрики.
// @Summary Обновление метрики через URL
// @Description Обновляет значение метрики указанного типа
// @Tags metrics
// @Produce text/plain
// @Param type path string true "Тип метрики (gauge/counter)"
// @Param name path string true "Имя метрики"
// @Param value path string true "Значение метрики"
// @Success 200 {string} string "Метрика обновлена"
// @Failure 400 {object} models.ErrorResponse
// @Router /update/{type}/{name}/{value} [post]
func (s Storage) MetricHandler(c *gin.Context) {
	ctx := context.Background()
	models, err := s.handlerServices.ParseMetricsServices(c)
	if err != nil {
		s.Logger.ErrorCtx(ctx, "MetricHandler The metric was not parsed")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.handlerServices.models = &models
	switch models[0].MType {
	case m.TypeCounter:
		s.handlerServices.CounterService(c)
	case m.TypeGauge:
		s.handlerServices.GaugeService(c)
	default:
		c.JSON(http.StatusNotFound, gin.H{"error": "No such value exists"})
		return
	}
}

// SaveToFile сохраняет состояние хранилища метрик в файл.
// Параметры:
//   - fname: имя файла для сохранения
//
// Возвращает:
//   - ошибку, если не удалось сохранить файл
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

// PingDBHandler обрабатывает запрос на проверку соединения с базой данных.
// Делегирует обработку запроса сервису handlerServices.
func (s Storage) PingDBHandler(c *gin.Context) {
	s.Logger.InfoCtx(c.Request.Context(), "handler PingDBHandler")
	c.Header("Content-Type", "application/json")
	s.handlerServices.models = nil
	s.handlerServices.PingService(c)
}

// NotImplementedHandler является обработчиком для еще не реализованных эндпоинтов.
// Возвращает статус 404 Not Found.
func NotImplementedHandler(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusNotFound)
}

// BadRequestHandler является обработчиком для некорректных запросов.
// Возвращает статус 400 Bad Request.
func BadRequestHandler(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Bad Request Handler", http.StatusBadRequest)
}
