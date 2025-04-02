package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sanek1/metrics-collector/internal/models"
	"github.com/sanek1/metrics-collector/internal/storage/server/mocks"
)

func TestFilterBatchesBeforeSaving(t *testing.T) {
	mockStorage := mocks.NewStorageHelper(t)
	testMetrics := []models.Metrics{
		{ID: "test1", MType: "gauge"},
		{ID: "test2", MType: "counter"},
	}
	t.Run("filter metrics", func(t *testing.T) {
		mockStorage.On("FilterBatchesBeforeSaving", testMetrics).Return(testMetrics[:1]).Once()

		result := mockStorage.FilterBatchesBeforeSaving(testMetrics)
		assert.Len(t, result, 1)
		mockStorage.AssertExpectations(t)
	})
}

func Test_GetMetricsOnDBs(t *testing.T) {
	mockStorage := mocks.NewStorageHelper(t)
	ctx := context.Background()
	metrics := []models.Metrics{
		{ID: "test1", MType: "gauge"},
	}

	t.Run("success get metrics", func(t *testing.T) {
		expected := []*models.Metrics{{ID: "test1", MType: "gauge"}}
		mockStorage.On("GetMetricsOnDBs", ctx, mock.Anything).Return(expected, nil).Once()

		result, err := mockStorage.GetMetricsOnDBs(ctx, metrics...)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mockStorage.AssertExpectations(t)
	})

	t.Run("empty result", func(t *testing.T) {
		mockStorage.On("GetMetricsOnDBs", ctx, mock.Anything).Return([]*models.Metrics{}, nil).Once()

		result, err := mockStorage.GetMetricsOnDBs(ctx, metrics...)
		require.NoError(t, err)
		assert.Empty(t, result)
		mockStorage.AssertExpectations(t)
	})
}

func TestStorageHelper_InsertUpdateMetrics(t *testing.T) {
	mockStorage := mocks.NewStorageHelper(t)
	ctx := context.Background()
	metrics := []models.Metrics{
		{ID: "test1", MType: "gauge"},
	}

	t.Run("insert metrics", func(t *testing.T) {
		mockStorage.On("InsertMetric", ctx, metrics).Return(nil).Once()

		err := mockStorage.InsertMetric(ctx, metrics)
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("update metrics", func(t *testing.T) {
		mockStorage.On("UpdateMetrics", ctx, metrics).Return(nil).Once()

		err := mockStorage.UpdateMetrics(ctx, metrics)
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})
}

func TestStorageHelper_SetMetrics(t *testing.T) {
	mockStorage := mocks.NewStorageHelper(t)
	ctx := context.Background()
	input := []models.Metrics{
		{ID: "test1", MType: "counter", Delta: ptr(int64(10))},
		{ID: "test2", MType: "gauge", Value: ptr(123.45)},
	}

	t.Run("successful set metrics", func(t *testing.T) {
		expectedResult := []*models.Metrics{
			{ID: "test1", MType: "counter", Delta: ptr(int64(10))},
			{ID: "test2", MType: "gauge", Value: ptr(123.45)},
		}

		mockStorage.On("SetMetrics", ctx, input).Return(expectedResult, nil).Once()
		result, err := mockStorage.SetMetrics(ctx, input)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		mockStorage.AssertExpectations(t)
	})
}

func TestStorageHelper_SortingBatchData(t *testing.T) {
	mockStorage := mocks.NewStorageHelper(t)

	existing := []*models.Metrics{
		{ID: "test1", MType: "counter"},
	}
	input := []models.Metrics{
		{ID: "test1", MType: "counter"},
		{ID: "test2", MType: "gauge"},
	}

	t.Run("sort metrics", func(t *testing.T) {
		expectedUpdates := []models.Metrics{{ID: "test1", MType: "counter"}}
		expectedInserts := []models.Metrics{{ID: "test2", MType: "gauge"}}

		mockStorage.On("SortingBatchData", existing, input).Return(expectedUpdates, expectedInserts).Once()

		updates, inserts := mockStorage.SortingBatchData(existing, input)
		assert.Equal(t, expectedUpdates, updates)
		assert.Equal(t, expectedInserts, inserts)
		mockStorage.AssertExpectations(t)
	})
}

func TestStorageHelper_CollectorQuery(t *testing.T) {
	mockStorage := mocks.NewStorageHelper(t)
	ctx := context.Background()
	metrics := []models.Metrics{
		{ID: "test1", MType: "counter"},
	}

	t.Run("generate query", func(t *testing.T) {
		expectedQuery := "SELECT ..."
		expectedTypes := []string{"counter"}
		expectedArgs := []interface{}{"test1"}

		mockStorage.On("CollectorQuery", ctx, metrics).Return(expectedQuery, expectedTypes, expectedArgs).Once()

		q, types, args := mockStorage.CollectorQuery(ctx, metrics)
		assert.Equal(t, expectedQuery, q)
		assert.Equal(t, expectedTypes, types)
		assert.Equal(t, expectedArgs, args)
		mockStorage.AssertExpectations(t)
	})
}

func ptr[T any](v T) *T {
	return &v
}
