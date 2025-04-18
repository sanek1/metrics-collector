package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	m "github.com/sanek1/metrics-collector/internal/models"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

type MockDBStorage struct {
	*DBStorage
	PingError       error
	GetMetricsData  map[string]*m.Metrics
	GetMetricsFound bool
	AllMetrics      []string
}

func NewMockDBStorage(logger *l.ZapLogger) *MockDBStorage {
	return &MockDBStorage{
		DBStorage: &DBStorage{
			Logger: logger,
		},
		GetMetricsData: make(map[string]*m.Metrics),
		AllMetrics:     []string{},
	}
}

func (m *MockDBStorage) PingIsOk() bool {
	return m.PingError == nil
}

func (m *MockDBStorage) GetMetrics(ctx context.Context, mType, id string) (*m.Metrics, bool) {
	key := mType + ":" + id
	metric, ok := m.GetMetricsData[key]
	return metric, ok || m.GetMetricsFound
}

func (m *MockDBStorage) GetAllMetrics() []string {
	return m.AllMetrics
}

func TestDBStorage_PingIsOk(t *testing.T) {
	logger, _ := l.NewZapLogger(zap.InfoLevel)

	t.Run("Success", func(t *testing.T) {
		mockStorage := NewMockDBStorage(logger)
		mockStorage.PingError = nil

		result := mockStorage.PingIsOk()
		assert.True(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		mockStorage := NewMockDBStorage(logger)
		mockStorage.PingError = errors.New("connection error")

		result := mockStorage.PingIsOk()
		assert.False(t, result)
	})
}

func TestMockDBStorage_GetMetrics(t *testing.T) {
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	mockStorage := NewMockDBStorage(logger)

	gaugeValue := float64(42.5)
	metric := &m.Metrics{
		ID:    "test_gauge",
		MType: "gauge",
		Value: &gaugeValue,
	}
	mockStorage.GetMetricsData["gauge:test_gauge"] = metric

	ctx := context.Background()

	// Check the retrieval of an existing metric
	t.Run("ExistingMetric", func(t *testing.T) {
		result, found := mockStorage.GetMetrics(ctx, "gauge", "test_gauge")
		assert.True(t, found)
		assert.NotNil(t, result)
		assert.Equal(t, "test_gauge", result.ID)
		assert.Equal(t, "gauge", result.MType)
		assert.Equal(t, gaugeValue, *result.Value)
	})

	// Check the retrieval of a non-existent metric
	t.Run("NonExistingMetric", func(t *testing.T) {
		mockStorage.GetMetricsFound = false
		result, found := mockStorage.GetMetrics(ctx, "gauge", "non_existent")
		assert.False(t, found)
		assert.Nil(t, result)
	})

	// Check the case when GetMetricsFound = true
	t.Run("MetricFoundByFlag", func(t *testing.T) {
		mockStorage.GetMetricsFound = true
		_, found := mockStorage.GetMetrics(ctx, "gauge", "non_existent")
		assert.True(t, found)
	})
}

func TestMockDBStorage_GetAllMetrics(t *testing.T) {
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	mockStorage := NewMockDBStorage(logger)

	// Check the empty list of metrics
	t.Run("EmptyList", func(t *testing.T) {
		result := mockStorage.GetAllMetrics()
		assert.Empty(t, result)
	})

	// Check the non-empty list of metrics
	t.Run("NonEmptyList", func(t *testing.T) {
		mockStorage.AllMetrics = []string{"gauge:metric1", "counter:metric2"}
		result := mockStorage.GetAllMetrics()
		assert.Len(t, result, 2)
		assert.Contains(t, result, "gauge:metric1")
		assert.Contains(t, result, "counter:metric2")
	})
}
