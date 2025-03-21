package storage

import (
	"context"
	"testing"

	flags "github.com/sanek1/metrics-collector/internal/flags/server"
	m "github.com/sanek1/metrics-collector/internal/models"
	l "github.com/sanek1/metrics-collector/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
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
	// filter duplicates and sort before updating/inserting
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

	// Verify metrics are inserted correctly
	metricMap := make(map[string]m.Metrics)
	for _, metric := range models {
		metricMap[metric.ID] = metric
	}

	// Verify metrics are inserted correctly
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
