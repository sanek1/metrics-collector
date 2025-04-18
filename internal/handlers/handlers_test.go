package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	m "github.com/sanek1/metrics-collector/internal/models"
	storage "github.com/sanek1/metrics-collector/internal/storage/server"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

func TestMainPageHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	s := storage.GetStorage(false, nil, logger)
	memStorage := NewStorage(s, logger)

	value1 := float64(123.45)
	value2 := int64(42)

	ctx := context.Background()
	_, _ = s.SetGauge(ctx, m.Metrics{ID: "gauge1", MType: "gauge", Value: &value1})
	_, _ = s.SetCounter(ctx, m.Metrics{ID: "counter1", MType: "counter", Delta: &value2})

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	memStorage.MainPageHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Метрики")
	assert.Contains(t, w.Body.String(), "gauge1")
	assert.Contains(t, w.Body.String(), "counter1")
}

func TestGetMetricsByNameHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	s := storage.GetStorage(false, nil, logger)
	memStorage := NewStorage(s, logger)

	gaugeValue := float64(123.45)
	counterValue := int64(42)
	ctx := context.Background()
	_, _ = s.SetGauge(ctx, m.Metrics{ID: "test_gauge", MType: "gauge", Value: &gaugeValue})
	_, _ = s.SetCounter(ctx, m.Metrics{ID: "test_counter", MType: "counter", Delta: &counterValue})

	tests := []struct {
		name           string
		metricType     string
		metricName     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Get existing gauge metric",
			metricType:     "gauge",
			metricName:     "test_gauge",
			expectedStatus: http.StatusOK,
			expectedBody:   strconv.FormatFloat(gaugeValue, 'f', -1, 64),
		},
		{
			name:           "Get existing counter metric",
			metricType:     "counter",
			metricName:     "test_counter",
			expectedStatus: http.StatusOK,
			expectedBody:   strconv.FormatInt(counterValue, 10),
		},
		{
			name:           "Get non-existent metric",
			metricType:     "gauge",
			metricName:     "non_existent",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "GetMetricsByNameHandler No such value exists",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("/value/%s/%s", tc.metricType, tc.metricName), nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = []gin.Param{
				{Key: "metricType", Value: tc.metricType},
				{Key: "metricName", Value: tc.metricName},
			}

			memStorage.GetMetricsByNameHandler(c)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Equal(t, tc.expectedBody, w.Body.String())
		})
	}
}

func TestUpdateMetricFromURLHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	s := storage.GetStorage(false, nil, logger)
	memStorage := NewStorage(s, logger)

	tests := []struct {
		name           string
		metricType     string
		metricName     string
		metricValue    string
		expectedStatus int
		validateBody   func(t *testing.T, body string)
	}{
		{
			name:           "Update gauge metric",
			metricType:     "gauge",
			metricName:     "test_gauge",
			metricValue:    "123.45",
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				var metric m.Metrics
				err := json.Unmarshal([]byte(body), &metric)
				require.NoError(t, err)
				assert.Equal(t, "test_gauge", metric.ID)
				assert.Equal(t, "gauge", metric.MType)
				require.NotNil(t, metric.Value)
				assert.Equal(t, 123.45, *metric.Value)
			},
		},
		{
			name:           "Update counter metric",
			metricType:     "counter",
			metricName:     "test_counter",
			metricValue:    "42",
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				var metric m.Metrics
				err := json.Unmarshal([]byte(body), &metric)
				require.NoError(t, err)
				assert.Equal(t, "test_counter", metric.ID)
				assert.Equal(t, "counter", metric.MType)
				require.NotNil(t, metric.Delta)
				assert.Equal(t, int64(42), *metric.Delta)
			},
		},
		{
			name:           "Invalid gauge value",
			metricType:     "gauge",
			metricName:     "invalid_gauge",
			metricValue:    "not_a_number",
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "Invalid gauge value")
			},
		},
		{
			name:           "Invalid counter value",
			metricType:     "counter",
			metricName:     "invalid_counter",
			metricValue:    "not_a_number",
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "Invalid counter value")
			},
		},
		{
			name:           "Unsupported metric type",
			metricType:     "unknown",
			metricName:     "test",
			metricValue:    "123",
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "Unsupported metric type")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", fmt.Sprintf("/update/%s/%s/%s", tc.metricType, tc.metricName, tc.metricValue), nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = []gin.Param{
				{Key: "metricType", Value: tc.metricType},
				{Key: "metricName", Value: tc.metricName},
				{Key: "metricValue", Value: tc.metricValue},
			}

			memStorage.UpdateMetricFromURLHandler(c)

			assert.Equal(t, tc.expectedStatus, w.Code)
			tc.validateBody(t, w.Body.String())
		})
	}
}

func TestGetMetricsByValueHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	s := storage.GetStorage(false, nil, logger)
	memStorage := NewStorage(s, logger)

	gaugeValue := float64(123.45)
	ctx := context.Background()
	_, _ = s.SetGauge(ctx, m.Metrics{ID: "test_gauge", MType: "gauge", Value: &gaugeValue})

	metricRequest := m.Metrics{
		ID:    "test_gauge",
		MType: "gauge",
	}
	jsonData, _ := json.Marshal(metricRequest)
	req, _ := http.NewRequest("POST", "/metrics", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	memStorage.GetMetricsByValueHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPingDBHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	t.Run("Memory storage", func(t *testing.T) {
		s := storage.GetStorage(false, nil, logger)
		memStorage := NewStorage(s, logger)

		req, _ := http.NewRequest("GET", "/ping", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		memStorage.PingDBHandler(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestSaveToFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	s := storage.GetStorage(false, nil, logger)
	memStorage := NewStorage(s, logger)

	err := os.MkdirAll("./testdata", 0755)
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll("./testdata")
	}()

	err = memStorage.SaveToFile("./testdata/test.json")
	assert.NoError(t, err)

	err = memStorage.SaveToFile("/non-existent-dir/test.json")
	assert.NoError(t, err)
}

func TestGetMetricsByBody_MetricHandler(t *testing.T) {
	value1 := float64(123)
	value2 := int64(-123)

	tests := []struct {
		name           string
		model          m.Metrics
		expectedStatus int
	}{
		{
			name: "counter",
			model: m.Metrics{
				ID:    "test3",
				MType: "counter",
				Delta: &value2,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "gauge",
			model: m.Metrics{
				ID:    "test2",
				MType: "gauge",
				Value: &value1,
			},
			expectedStatus: http.StatusOK,
		},
	}

	ctx := context.Background()
	l, err := l.NewZapLogger(zap.InfoLevel)
	if err != nil {
		log.Panic(err)
	}
	s := storage.GetStorage(false, nil, l)
	memStorage := NewStorage(s, l)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b, _ := json.Marshal(test.model)
			req, err := http.NewRequestWithContext(ctx, "POST", "/", bytes.NewBuffer(b))
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			memStorage.MetricHandler(c)

			assert.Equal(t, test.expectedStatus, w.Code)
		})
	}
}

func TestBatchMetricsByBody(t *testing.T) {
	value1 := float64(123)
	value2 := int64(-123)

	tests := []struct {
		name           string
		model          []m.Metrics
		expectedStatus int
	}{
		{
			name: "counter",
			model: []m.Metrics{
				{
					ID:    "test1",
					MType: "counter",
					Value: &value1,
				},
				{
					ID:    "test2",
					MType: "gauge",
					Delta: &value2,
				},
			},
			expectedStatus: http.StatusOK,
		},
	}

	ctx := context.Background()
	l, err := l.NewZapLogger(zap.InfoLevel)
	if err != nil {
		log.Panic(err)
	}
	s := storage.GetStorage(false, nil, l)
	memStorage := NewStorage(s, l)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b, _ := json.Marshal(test.model)
			req, err := http.NewRequestWithContext(ctx, "POST", "/", bytes.NewBuffer(b))
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			memStorage.MetricHandler(c)

			assert.Equal(t, test.expectedStatus, w.Code)
		})
	}
}
