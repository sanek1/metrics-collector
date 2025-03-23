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

type Storage struct {
	Storage         storage.Storage
	Logger          *l.ZapLogger
	handlerServices *Services
}

func NewStorage(s storage.Storage, zl *l.ZapLogger) *Storage {
	hs := NewHandlerServices(s, nil, zl)

	return &Storage{Storage: s, Logger: zl, handlerServices: hs}
}

func (s Storage) MainPageHandler(c *gin.Context) {
	metrics := s.Storage.GetAllMetrics()
	htmlData := GenerateHTMLServices(metrics)
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, htmlData)
}

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

func (s Storage) GetMetricsByValueHandler(c *gin.Context) {
	s.Logger.InfoCtx(c.Request.Context(), "Handling GetMetricsByValueHandler request")
	s.handlerServices.GetMetricsByValueGin(c)
}

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

func (s Storage) PingDBHandler(c *gin.Context) {
	s.Logger.InfoCtx(c.Request.Context(), "handler PingDBHandler")
	c.Header("Content-Type", "application/json")
	s.handlerServices.models = nil
	s.handlerServices.PingService(c)
}

func NotImplementedHandler(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusNotFound)
}

func BadRequestHandler(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Bad Request Handler", http.StatusBadRequest)
}
