package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/sanek1/metrics-collector/internal/config"
	m "github.com/sanek1/metrics-collector/internal/models"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

func TestNewMetricsStorage(t *testing.T) {
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	storage := NewMetricsStorage(logger)
	assert.NotNil(t, storage)
	assert.Empty(t, storage.Metrics)
	assert.Empty(t, storage.Errors)
	assert.Equal(t, logger, storage.Logger)
}

func TestSetGauge(t *testing.T) {
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	storage := NewMetricsStorage(logger)

	t.Run("single gauge", func(t *testing.T) {
		val := 123.45
		models := []m.Metrics{
			{ID: "test1", MType: config.Gauge, Value: &val},
		}

		results, err := storage.SetGauge(context.Background(), models...)
		require.NoError(t, err)
		require.Len(t, results, 1)

		assert.Equal(t, "test1", results[0].ID)
		assert.Equal(t, 123.45, *results[0].Value)
	})

	t.Run("multiple gauges", func(t *testing.T) {
		vals := []float64{10.1, 20.2, 30.3}
		models := make([]m.Metrics, 3)
		for i := range models {
			models[i] = m.Metrics{
				ID:    fmt.Sprintf("metric%d", i+1),
				MType: config.Gauge,
				Value: &vals[i],
			}
		}

		results, err := storage.SetGauge(context.Background(), models...)
		require.NoError(t, err)
		require.Len(t, results, 3)

		for i := range results {
			assert.Equal(t, fmt.Sprintf("metric%d", i+1), results[i].ID)
			assert.Equal(t, vals[i], *results[i].Value)
		}
	})
}

func TestSetCounter(t *testing.T) {
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	storage := NewMetricsStorage(logger)

	t.Run("new counter", func(t *testing.T) {
		delta := int64(5)
		models := []m.Metrics{
			{ID: "counter1", MType: config.Counter, Delta: &delta},
		}

		results, err := storage.SetCounter(context.Background(), models...)
		require.NoError(t, err)
		require.Len(t, results, 1)

		assert.Equal(t, int64(5), *results[0].Delta)
	})

	t.Run("increment existing counter", func(t *testing.T) {
		delta := int64(3)
		models := []m.Metrics{
			{ID: "counter1", MType: config.Counter, Delta: &delta},
		}

		results, err := storage.SetCounter(context.Background(), models...)
		require.NoError(t, err)
		require.Len(t, results, 1)

		assert.Equal(t, int64(8), *results[0].Delta)
	})
}

func TestGetAllMetrics(t *testing.T) {
	storage := NewMetricsStorage(nil)
	gaugeVal := 123.45
	counterVal := int64(42)
	storage.Metrics = map[string]m.Metrics{
		"gauge1":   {ID: "gauge1", MType: config.Gauge, Value: &gaugeVal},
		"counter1": {ID: "counter1", MType: config.Counter, Delta: &counterVal},
	}

	result := storage.GetAllMetrics()
	expected := []string{
		"counter1: 42",
		"gauge1: 123.45",
	}

	assert.ElementsMatch(t, expected, result)
}

func TestGetMetrics(t *testing.T) {
	storage := NewMetricsStorage(nil)
	gaugeVal := 99.9
	storage.Metrics["test"] = m.Metrics{ID: "test", MType: config.Gauge, Value: &gaugeVal}

	t.Run("existing metric", func(t *testing.T) {
		metric, ok := storage.GetMetrics(context.Background(), "", "test")
		require.True(t, ok)
		assert.Equal(t, 99.9, *metric.Value)
	})

	t.Run("non-existent metric", func(t *testing.T) {
		_, ok := storage.GetMetrics(context.Background(), "", "unknown")
		assert.False(t, ok)
	})
}
