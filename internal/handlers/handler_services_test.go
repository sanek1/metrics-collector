package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	m "github.com/sanek1/metrics-collector/internal/models"
	storage "github.com/sanek1/metrics-collector/internal/storage/server"
	"github.com/sanek1/metrics-collector/internal/storage/server/mocks"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

func TestNewHandlerServices(t *testing.T) {
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	s := storage.GetStorage(false, nil, logger)
	hashKey := "test-key"
	cryptoKey := ""
	services := NewHandlerServices(s, &hashKey, cryptoKey, logger)
	assert.NotNil(t, services)
	assert.Equal(t, s, services.s)
	assert.Equal(t, hashKey, *services.hashKey)
	assert.Equal(t, logger, services.logger)
}

func TestSetMetricsByBodyGin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	mockStorage := mocks.NewStorage(t)

	gaugeValue := float64(123.45)
	gaugeMetric := m.Metrics{
		ID:    "test_gauge",
		MType: "gauge",
		Value: &gaugeValue,
	}
	returnedGaugeMetric := &gaugeMetric
	mockStorage.On("SetGauge", mock.Anything, gaugeMetric).Return([]*m.Metrics{returnedGaugeMetric}, nil)

	counterValue := int64(42)
	counterMetric := m.Metrics{
		ID:    "test_counter",
		MType: "counter",
		Delta: &counterValue,
	}
	returnedCounterMetric := &counterMetric
	mockStorage.On("SetCounter", mock.Anything, counterMetric).Return([]*m.Metrics{returnedCounterMetric}, nil)

	services := NewHandlerServices(mockStorage, nil, "", logger)
	t.Run("gauge metric", func(t *testing.T) {
		jsonData, _ := json.Marshal(gaugeMetric)
		req, _ := http.NewRequest("POST", "/update", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		bodyBytes, _ := json.Marshal(gaugeMetric)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		services.SetMetricsByBodyGin(c)
		assert.Equal(t, http.StatusOK, w.Code)

		var respMetric m.Metrics
		err := json.Unmarshal(w.Body.Bytes(), &respMetric)
		require.NoError(t, err)

		assert.Equal(t, "test_gauge", respMetric.ID)
		assert.Equal(t, "gauge", respMetric.MType)
		assert.Equal(t, gaugeValue, *respMetric.Value)
	})

	t.Run("counter metric", func(t *testing.T) {
		jsonData, _ := json.Marshal(counterMetric)
		req, _ := http.NewRequest("POST", "/update", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		bodyBytes, _ := json.Marshal(counterMetric)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		services.SetMetricsByBodyGin(c)
		assert.Equal(t, http.StatusOK, w.Code)

		var respMetric m.Metrics
		err := json.Unmarshal(w.Body.Bytes(), &respMetric)
		require.NoError(t, err)

		assert.Equal(t, "test_counter", respMetric.ID)
		assert.Equal(t, "counter", respMetric.MType)
		assert.Equal(t, counterValue, *respMetric.Delta)
	})
}

func TestGetMetricsByValueGin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := l.NewZapLogger(zap.InfoLevel)

	mockStorage := mocks.NewStorage(t)
	gaugeValue := float64(123.45)
	gaugeMetric := m.Metrics{
		ID:    "test_gauge",
		MType: "gauge",
		Value: &gaugeValue,
	}

	mockStorage.On("GetMetrics", mock.Anything, "gauge", "test_gauge").Return(&gaugeMetric, true)
	services := NewHandlerServices(mockStorage, nil, "", logger)

	t.Run("get existing metric", func(t *testing.T) {
		request := m.Metrics{
			ID:    "test_gauge",
			MType: "gauge",
		}
		jsonData, _ := json.Marshal(request)
		req, _ := http.NewRequest("POST", "/value", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		bodyBytes, _ := json.Marshal(request)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		services.GetMetricsByValueGin(c)
		assert.Equal(t, http.StatusOK, w.Code)

		var respMetric m.Metrics
		err := json.Unmarshal(w.Body.Bytes(), &respMetric)
		require.NoError(t, err)

		assert.Equal(t, "test_gauge", respMetric.ID)
		assert.Equal(t, "gauge", respMetric.MType)
		assert.Equal(t, gaugeValue, *respMetric.Value)
	})

	t.Run("get non-existent metric", func(t *testing.T) {
		mockStorage := mocks.NewStorage(t)
		mockStorage.On("GetMetrics", mock.Anything, "gauge", "non_existent").Return(nil, false)

		services := NewHandlerServices(mockStorage, nil, "", logger)
		request := m.Metrics{
			ID:    "non_existent",
			MType: "gauge",
		}
		jsonData, _ := json.Marshal(request)
		req, _ := http.NewRequest("POST", "/value", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		bodyBytes, _ := json.Marshal(request)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		services.GetMetricsByValueGin(c)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestCheckValue(t *testing.T) {
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	s := storage.GetStorage(false, nil, logger)
	services := NewHandlerServices(s, nil, "", logger)

	t.Run("check counter", func(t *testing.T) {
		delta := int64(42)
		metric := m.Metrics{
			ID:    "test_counter",
			MType: "counter",
			Delta: &delta,
		}

		ok := services.CheckValue(&metric)
		assert.True(t, ok)
	})

	t.Run("check gauge", func(t *testing.T) {
		value := float64(123.45)
		metric := m.Metrics{
			ID:    "test_gauge",
			MType: "gauge",
			Value: &value,
		}

		ok := services.CheckValue(&metric)
		assert.True(t, ok)
	})

	t.Run("check invalid type", func(t *testing.T) {
		value := float64(123.45)
		metric := m.Metrics{
			ID:    "test_invalid",
			MType: "invalid",
			Value: &value,
		}

		ok := services.CheckValue(&metric)
		assert.False(t, ok)
	})

	t.Run("check counter without value", func(t *testing.T) {
		metric := m.Metrics{
			ID:    "test_counter",
			MType: "counter",
		}

		ok := services.CheckValue(&metric)
		assert.False(t, ok)
	})

	t.Run("check gauge without value", func(t *testing.T) {
		metric := m.Metrics{
			ID:    "test_gauge",
			MType: "gauge",
		}

		ok := services.CheckValue(&metric)
		assert.False(t, ok)
	})
}
