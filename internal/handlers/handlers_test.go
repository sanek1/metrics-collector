package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	m "github.com/sanek1/metrics-collector/internal/models"
	storage "github.com/sanek1/metrics-collector/internal/storage/server"
	l "github.com/sanek1/metrics-collector/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGetMetricsByBody(t *testing.T) {
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
				ID:    "test1",
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := Storage{
				Storage: &storage.MetricsStorage{
					Metrics: make(map[string]m.Metrics),
					Logger:  l,
				},
			}
			b, _ := json.Marshal(test.model)
			req, err := http.NewRequestWithContext(ctx, "POST", "/", bytes.NewBuffer(b))
			require.NoError(t, err)
			w := httptest.NewRecorder()
			s.GetMetricsHandler(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)
		})
	}
}
