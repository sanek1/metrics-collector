package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sanek1/metrics-collector/internal/storage"
	v "github.com/sanek1/metrics-collector/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMetricsByBody(t *testing.T) {
	value1 := float64(123)
	value2 := int64(-123)

	tests := []struct {
		name           string
		model          v.Metrics
		expectedStatus int
	}{
		{
			name: "counter",
			model: v.Metrics{
				ID:    "test1",
				MType: "counter",
				Delta: &value2,
			},
			expectedStatus: http.StatusOK, //test counter type
		},
		{
			name: "gauge",
			model: v.Metrics{
				ID:    "test2",
				MType: "gauge",
				Value: &value1,
			},
			expectedStatus: http.StatusOK, //test gauge type
		},
		// {
		// 	name: "unknown type",
		// 	model: v.Metrics{
		// 		ID:    "test3",
		// 		MType: "unknown",
		// 	},
		// 	expectedStatus: http.StatusBadRequest, //test unknown type
		// },
	}

	logger, err := v.Initialize("test_level")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ms := MetricStorage{
				Storage: &storage.MemoryStorage{
					Metrics: make(map[string]v.Metrics),
					Logger:  logger,
				},
			}
			b, _ := json.Marshal(test.model)
			req, err := http.NewRequestWithContext(ctx, "POST", "/", bytes.NewBuffer(b))
			require.NoError(t, err)
			w := httptest.NewRecorder()
			ms.GetMetricsHandler(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)
		})
	}
}
