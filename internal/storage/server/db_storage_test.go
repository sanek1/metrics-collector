package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	flags "github.com/sanek1/metrics-collector/internal/flags/server"
	m "github.com/sanek1/metrics-collector/internal/models"
	"github.com/sanek1/metrics-collector/internal/storage/server/mocks"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

func TestInsertUpdateMultipleMetricsInBatch(t *testing.T) {
	value1 := float64(123)
	value2 := int64(-123)
	value3 := int64(1)

	ctx := context.Background()
	opt := flags.ParseServerFlags()
	if opt.DBPath == "" {
		return
	}
	logger, _ := l.NewZapLogger(zap.InfoLevel)

	storage := NewDBStorage(opt, logger)

	// Prepare test data
	testMetrics := []m.Metrics{
		{ID: "metric1", MType: "counter", Delta: &value2, Value: nil},
		{ID: "metric2", MType: "gauge", Delta: nil, Value: &value1},
		{ID: "metric3", MType: "counter", Delta: &value2, Value: &value1},
		{ID: "metric1", MType: "counter", Delta: &value3, Value: nil},
	}
	// check EnsureMetricsTableExists
	if err := storage.EnsureMetricsTableExists(ctx); err != nil {
		t.Error("failed to ensure metrics table exists: %w", err)
	}

	models := FilterBatchesBeforeSaving(testMetrics)

	existingMetrics, _ := storage.getMetricsOnDBs(ctx, models...)
	updatingBatch, insertingBatch := SortingBatchData(existingMetrics, models)

	if len(updatingBatch) != 0 {
		if err := storage.updateMetrics(ctx, updatingBatch); err != nil {
			storage.Logger.ErrorCtx(ctx, "failed to update metric", zap.Error(err))
			t.Error("failed to update metric: %w", err)
		}
	}
	if len(insertingBatch) != 0 {
		if err := storage.insertMetric(ctx, insertingBatch); err != nil {
			storage.Logger.ErrorCtx(ctx, "failed to insert metric", zap.Error(err))
			t.Error("failed to insert metric: %w", err)
		}
	}

	metricMap := make(map[string]m.Metrics)
	for _, metric := range models {
		metricMap[metric.ID] = metric
	}

	metricMapBefore := make(map[string]m.Metrics)
	for _, metric := range existingMetrics {
		metricMapBefore[metric.ID] = *metric
	}

	retrievedMetrics, err := storage.getMetricsOnDBs(ctx, testMetrics...)
	require.NoError(t, err)

	for _, retrievedMetric := range retrievedMetrics {
		expectedMetric, exists := metricMap[retrievedMetric.ID]
		require.True(t, exists)
		assert.Equal(t, expectedMetric.ID, retrievedMetric.ID)
		assert.Equal(t, expectedMetric.MType, retrievedMetric.MType)

		if retrievedMetric.MType == "counter" {
			if metricMapBefore[expectedMetric.ID].Delta == nil {
				assert.Equal(t, *expectedMetric.Delta, *retrievedMetric.Delta)
			} else {
				oldDelta := *metricMapBefore[expectedMetric.ID].Delta
				assert.Equal(t, *expectedMetric.Delta+oldDelta, *retrievedMetric.Delta)
			}
		} else {
			assert.Equal(t, expectedMetric.Delta, retrievedMetric.Delta)
		}
		assert.Equal(t, expectedMetric.Value, retrievedMetric.Value)
	}
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
