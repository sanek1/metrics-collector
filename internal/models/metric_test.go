package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMetricCounter(t *testing.T) {
	id := "test_counter"
	delta := int64(42)
	metric := NewMetricCounter(id, &delta)
	assert.NotNil(t, metric)
	assert.Equal(t, id, metric.ID)
	assert.Equal(t, TypeCounter, metric.MType)
	assert.Equal(t, &delta, metric.Delta)
	assert.Nil(t, metric.Value)
}

func TestNewMetricGauge(t *testing.T) {
	id := "test_gauge"
	value := float64(123.45)
	metric := NewMetricGauge(id, &value)

	assert.NotNil(t, metric)
	assert.Equal(t, id, metric.ID)
	assert.Equal(t, TypeGauge, metric.MType)
	assert.Equal(t, &value, metric.Value)
	assert.Nil(t, metric.Delta)
}

func TestNewArrMetricGauge(t *testing.T) {
	metrics := map[string]float64{
		"test_gauge1": 123.45,
		"test_gauge2": 678.90,
		"test_gauge3": 0.0,
	}

	result := NewArrMetricGauge(metrics)

	assert.NotNil(t, result)
	assert.Equal(t, len(metrics), len(result))

	resultMap := make(map[string]float64)
	for _, m := range result {
		assert.Equal(t, TypeGauge, m.MType)
		assert.NotNil(t, m.Value)
		assert.Nil(t, m.Delta)
		resultMap[m.ID] = *m.Value
	}

	for id, value := range metrics {
		resultValue, ok := resultMap[id]
		assert.True(t, ok, "metric %s is not found", id)
		assert.Equal(t, value, resultValue, "value metric %s is not equal", id)
	}
}

func TestMetricsConstants(t *testing.T) {
	assert.Equal(t, "gauge", TypeGauge)
	assert.Equal(t, "counter", TypeCounter)
}

func TestMetrics_Structure(t *testing.T) {
	t.Run("counter metric", func(t *testing.T) {
		delta := int64(42)
		metric := Metrics{
			ID:    "test_counter",
			MType: TypeCounter,
			Delta: &delta,
		}

		assert.Equal(t, "test_counter", metric.ID)
		assert.Equal(t, TypeCounter, metric.MType)
		assert.Equal(t, &delta, metric.Delta)
		assert.Nil(t, metric.Value)
	})

	t.Run("gauge metric", func(t *testing.T) {
		value := float64(123.45)
		metric := Metrics{
			ID:    "test_gauge",
			MType: TypeGauge,
			Value: &value,
		}

		assert.Equal(t, "test_gauge", metric.ID)
		assert.Equal(t, TypeGauge, metric.MType)
		assert.Equal(t, &value, metric.Value)
		assert.Nil(t, metric.Delta)
	})
}
