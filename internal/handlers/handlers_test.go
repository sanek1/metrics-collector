package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	m "github.com/sanek1/metrics-collector/internal/models"
	storage "github.com/sanek1/metrics-collector/internal/storage/server"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

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
