package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	m "github.com/sanek1/metrics-collector/internal/models"
	"github.com/sanek1/metrics-collector/internal/storage/server/mocks"
)

func TestStorage_GetAllMetrics(t *testing.T) {
	mockStorage := mocks.NewStorage(t)
	t.Run("returns all metrics", func(t *testing.T) {
		expected := []string{
			"counter:test1=42",
			"gauge:test2=123.45",
		}

		mockStorage.On("GetAllMetrics").Return(expected).Once()

		result := mockStorage.GetAllMetrics()
		assert.Equal(t, expected, result)
		mockStorage.AssertExpectations(t)
	})

	t.Run("returns empty list", func(t *testing.T) {
		mockStorage.On("GetAllMetrics").Return([]string{}).Once()

		result := mockStorage.GetAllMetrics()
		assert.Empty(t, result)
		mockStorage.AssertExpectations(t)
	})
}

func TestStorage_GetMetrics(t *testing.T) {
	mockStorage := mocks.NewStorage(t)
	ctx := context.Background()

	t.Run("existing counter metric", func(t *testing.T) {
		delta := int64(42)
		expected := &m.Metrics{
			ID:    "test1",
			MType: "counter",
			Delta: &delta,
		}

		mockStorage.On("GetMetrics", ctx, "counter", "test1").Return(expected, true).Once()

		metric, found := mockStorage.GetMetrics(ctx, "counter", "test1")
		require.True(t, found)
		assert.Equal(t, expected, metric)
		mockStorage.AssertExpectations(t)
	})

	t.Run("non-existing gauge metric", func(t *testing.T) {
		mockStorage.On("GetMetrics", ctx, "gauge", "unknown").Return((*m.Metrics)(nil), false).Once()

		_, found := mockStorage.GetMetrics(ctx, "gauge", "unknown")
		assert.False(t, found)
		mockStorage.AssertExpectations(t)
	})
}

func TestStorage_SetCounter(t *testing.T) {
	mockStorage := mocks.NewStorage(t)
	ctx := context.Background()
	delta := int64(10)
	metrics := []m.Metrics{
		{ID: "test1", MType: "counter", Delta: &delta},
	}

	t.Run("successful counter update", func(t *testing.T) {
		expected := []*m.Metrics{
			{ID: "test1", MType: "counter", Delta: &delta},
		}

		mockStorage.On("SetCounter", ctx, mock.Anything).Return(expected, nil).Once()

		result, err := mockStorage.SetCounter(ctx, metrics...)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, delta, *result[0].Delta)
		mockStorage.AssertExpectations(t)
	})

	t.Run("error on counter update", func(t *testing.T) {
		mockStorage.On("SetCounter", ctx, mock.Anything).Return(
			([]*m.Metrics)(nil), assert.AnError,
		).Once()

		_, err := mockStorage.SetCounter(ctx, metrics...)
		assert.Error(t, err)
		mockStorage.AssertExpectations(t)
	})
}

func TestStorage_SetGauge(t *testing.T) {
	mockStorage := mocks.NewStorage(t)
	ctx := context.Background()
	value := 123.45
	metrics := []m.Metrics{
		{ID: "test2", MType: "gauge", Value: &value},
	}

	t.Run("successful gauge update", func(t *testing.T) {
		expected := []*m.Metrics{
			{ID: "test2", MType: "gauge", Value: &value},
		}

		mockStorage.On("SetGauge", ctx, mock.Anything).Return(expected, nil).Once()

		result, err := mockStorage.SetGauge(ctx, metrics...)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, value, *result[0].Value)
		mockStorage.AssertExpectations(t)
	})

	t.Run("error on gauge update", func(t *testing.T) {
		mockStorage.On("SetGauge", ctx, mock.Anything).Return(
			([]*m.Metrics)(nil), assert.AnError,
		).Once()

		_, err := mockStorage.SetGauge(ctx, metrics...)
		assert.Error(t, err)
		mockStorage.AssertExpectations(t)
	})
}

func TestStorage_EdgeCases(t *testing.T) {
	mockStorage := mocks.NewStorage(t)
	ctx := context.Background()

	t.Run("empty metrics list", func(t *testing.T) {
		mockStorage.On("SetCounter", ctx).Return([]*m.Metrics{}, nil).Once()

		result, err := mockStorage.SetCounter(ctx)
		require.NoError(t, err)
		assert.Empty(t, result)
		mockStorage.AssertExpectations(t)
	})

	t.Run("nil values handling", func(t *testing.T) {
		invalidMetric := m.Metrics{
			ID:    "invalid",
			MType: "counter",
			Delta: nil,
		}

		mockStorage.On("SetCounter", ctx, invalidMetric).Return(
			([]*m.Metrics)(nil), assert.AnError,
		).Once()

		_, err := mockStorage.SetCounter(ctx, invalidMetric)
		assert.Error(t, err)
		mockStorage.AssertExpectations(t)
	})
}

func TestStorageWithMocks(t *testing.T) {
	mockStorage := mocks.NewStorage(t)
	ctx := context.Background()
	value := float64(123.45)
	gaugeMetric := m.Metrics{
		ID:    "TestGauge",
		MType: "gauge",
		Value: &value,
	}

	resultGaugeMetric := &m.Metrics{
		ID:    "TestGauge",
		MType: "gauge",
		Value: &value,
	}

	mockStorage.On("SetGauge", mock.Anything, mock.Anything).Return([]*m.Metrics{resultGaugeMetric}, nil)
	mockStorage.On("GetMetrics", mock.Anything, "gauge", "TestGauge").Return(resultGaugeMetric, true)

	delta := int64(42)
	counterMetric := m.Metrics{
		ID:    "TestCounter",
		MType: "counter",
		Delta: &delta,
	}

	resultCounterMetric := &m.Metrics{
		ID:    "TestCounter",
		MType: "counter",
		Delta: &delta,
	}

	mockStorage.On("SetCounter", mock.Anything, mock.Anything).Return([]*m.Metrics{resultCounterMetric}, nil)
	mockStorage.On("GetMetrics", mock.Anything, "counter", "TestCounter").Return(resultCounterMetric, true)

	mockStorage.On("GetAllMetrics").Return([]string{
		"gauge:TestGauge=123.45",
		"counter:TestCounter=42",
	})

	t.Run("SetGauge", func(t *testing.T) {
		updatedMetrics, err := mockStorage.SetGauge(ctx, gaugeMetric)
		require.NoError(t, err)
		require.Len(t, updatedMetrics, 1)
		assert.Equal(t, "TestGauge", updatedMetrics[0].ID)
		assert.Equal(t, "gauge", updatedMetrics[0].MType)
		assert.Equal(t, value, *updatedMetrics[0].Value)
	})

	t.Run("GetGaugeMetric", func(t *testing.T) {
		metric, found := mockStorage.GetMetrics(ctx, "gauge", "TestGauge")
		assert.True(t, found)
		assert.Equal(t, "TestGauge", metric.ID)
		assert.Equal(t, "gauge", metric.MType)
		assert.Equal(t, value, *metric.Value)
	})

	t.Run("SetCounter", func(t *testing.T) {
		updatedMetrics, err := mockStorage.SetCounter(ctx, counterMetric)
		require.NoError(t, err)
		require.Len(t, updatedMetrics, 1)
		assert.Equal(t, "TestCounter", updatedMetrics[0].ID)
		assert.Equal(t, "counter", updatedMetrics[0].MType)
		assert.Equal(t, delta, *updatedMetrics[0].Delta)
	})

	t.Run("GetCounterMetric", func(t *testing.T) {
		metric, found := mockStorage.GetMetrics(ctx, "counter", "TestCounter")
		assert.True(t, found)
		assert.Equal(t, "TestCounter", metric.ID)
		assert.Equal(t, "counter", metric.MType)
		assert.Equal(t, delta, *metric.Delta)
	})

	t.Run("GetAllMetrics", func(t *testing.T) {
		metrics := mockStorage.GetAllMetrics()
		assert.Len(t, metrics, 2)
		assert.Contains(t, metrics, "gauge:TestGauge=123.45")
		assert.Contains(t, metrics, "counter:TestCounter=42")
	})

	mockStorage.AssertExpectations(t)
}

func TestEnsureMetricsTableExists(t *testing.T) {
	mockStorage := mocks.NewStorageHelper(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockStorage.On("EnsureMetricsTableExists", ctx).Return(nil).Once()

		err := mockStorage.EnsureMetricsTableExists(ctx)
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockStorage.On("EnsureMetricsTableExists", ctx).Return(assert.AnError).Once()

		err := mockStorage.EnsureMetricsTableExists(ctx)
		assert.Error(t, err)
		mockStorage.AssertExpectations(t)
	})
}
