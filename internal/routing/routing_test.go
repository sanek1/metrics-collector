package routing

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	sf "github.com/sanek1/metrics-collector/internal/flags/server"
	m "github.com/sanek1/metrics-collector/internal/models"
	"github.com/sanek1/metrics-collector/internal/storage/server/mocks"
	"github.com/sanek1/metrics-collector/pkg/logging"
)

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) InfoCtx(ctx context.Context, msg string, fields ...interface{}) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) ErrorCtx(ctx context.Context, msg string, fields ...interface{}) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) WarnCtx(ctx context.Context, msg string, fields ...interface{}) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) DebugCtx(ctx context.Context, msg string, fields ...interface{}) {
	m.Called(ctx, msg, fields)
}

func TestRouterWithMocks_InitRouting(t *testing.T) {
	mockStorage := new(mocks.Storage)
	mockLogger := new(MockLogger)

	mockLogger.On("InfoCtx", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("ErrorCtx", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("WarnCtx", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("DebugCtx", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	opts := &sf.ServerOptions{
		CryptoKey: "test_key",
	}

	gaugeValue := float64(123.45)
	mockGaugeMetric := &m.Metrics{
		ID:    "TestGauge",
		MType: "gauge",
		Value: &gaugeValue,
	}
	mockStorage.On("SetGauge", mock.Anything, mock.Anything).Return([]*m.Metrics{mockGaugeMetric}, nil).Maybe()
	mockStorage.On("GetMetrics", mock.Anything, "gauge", "TestGauge").Return(mockGaugeMetric, true).Maybe()
	mockStorage.On("GetMetrics", mock.Anything, mock.Anything, mock.Anything).Return(mockGaugeMetric, true).Maybe()

	counterValue := int64(42)
	mockCounterMetric := &m.Metrics{
		ID:    "TestCounter",
		MType: "counter",
		Delta: &counterValue,
	}
	mockStorage.On("SetCounter", mock.Anything, mock.Anything).Return([]*m.Metrics{mockCounterMetric}, nil).Maybe()
	mockStorage.On("GetMetrics", mock.Anything, "counter", "TestCounter").Return(mockCounterMetric, true).Maybe()
	mockStorage.On("GetAllMetrics").Return([]string{"gauge:TestGauge=123.45", "counter:TestCounter=42"}).Maybe()
	mockStorage.On("Ping", mock.Anything).Return(nil).Maybe()

	l, _ := logging.NewZapLogger(zap.InfoLevel)
	router := NewRouting(mockStorage, opts, l)
	handler := router.InitRouting()

	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		expectedBody   string
		validate       func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "POST update gauge metric via URL",
			method:         http.MethodPost,
			path:           "/update/gauge/TestGauge/123.45",
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var metric m.Metrics
				err := json.Unmarshal(resp.Body.Bytes(), &metric)
				require.NoError(t, err)
				assert.Equal(t, "gauge", metric.MType)
				assert.Equal(t, "TestGauge", metric.ID)
				assert.NotNil(t, metric.Value)
				assert.Equal(t, 123.45, *metric.Value)
			},
		},
		{
			name:           "POST update counter metric via URL",
			method:         http.MethodPost,
			path:           "/update/counter/TestCounter/42",
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var metric m.Metrics
				err := json.Unmarshal(resp.Body.Bytes(), &metric)
				require.NoError(t, err)
				assert.Equal(t, "counter", metric.MType)
				assert.Equal(t, "TestCounter", metric.ID)
				assert.NotNil(t, metric.Delta)
				assert.Equal(t, int64(42), *metric.Delta)
			},
		},
		{
			name:           "POST get gauge metric value",
			method:         http.MethodPost,
			path:           "/value/",
			body:           m.Metrics{ID: "TestGauge", MType: "gauge"},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var metric m.Metrics
				err := json.Unmarshal(resp.Body.Bytes(), &metric)
				require.NoError(t, err)
				assert.Equal(t, "gauge", metric.MType)
				assert.Equal(t, "TestGauge", metric.ID)
				assert.NotNil(t, metric.Value)
				assert.Equal(t, 123.45, *metric.Value)
			},
		},
		{
			name:           "GET ping",
			method:         http.MethodGet,
			path:           "/ping",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET main page",
			method:         http.MethodGet,
			path:           "/",
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Contains(t, resp.Header().Get("Content-Type"), "text/html")
				assert.Contains(t, resp.Body.String(), "<html")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request

			if tt.body != nil {
				bodyBytes, err := json.Marshal(tt.body)
				require.NoError(t, err)
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)
			assert.Equal(t, tt.expectedStatus, resp.Code)
			if tt.validate != nil {
				tt.validate(t, resp)
			}
		})
	}
}
