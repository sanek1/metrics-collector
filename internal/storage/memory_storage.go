package storage

import (
	"fmt"
	"strconv"

	"github.com/sanek1/metrics-collector/internal/config"
	m "github.com/sanek1/metrics-collector/internal/validation"
	"go.uber.org/zap"
)

// testcounter [ {"id": "counter1", "type": "counter", "delta": 1, "value": 123.4}]
// testSetGet32 [ {"id": "testSetGet33", "type": "gauge", "delta": 1, "value": 123.4}
type MemoryStorage struct {
	Metrics map[string]m.Metrics
	Logger  *zap.SugaredLogger
}

// NewMemoryStorage returns a new MemoryStorage instance.
// It creates an empty map of Metrics and a new Logger writing to os.Stdout.
func NewMemoryStorage() *MemoryStorage {
	logger, err := m.Initialize("test_level")
	if err != nil {
		panic(err)
	}

	return &MemoryStorage{
		Metrics: make(map[string]m.Metrics),
		Logger:  logger,
	}
}

// metrics.Value = Gauge
// metrics.Delta = Counter
func (ms *MemoryStorage) GetAllMetrics() []string {
	result := make([]string, 0, len(ms.Metrics))
	for _, metric := range ms.Metrics {
		if *metric.Value != 0 || *metric.Delta != 0 {
			var value string
			if metric.MType == config.Counter {
				value = strconv.FormatInt(*metric.Delta, 10)
			} else {
				value = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
			}
			result = append(result, fmt.Sprintf("%s: %s", metric.ID, value))
		}
	}
	return result
}

func (ms *MemoryStorage) GetMetrics(key, metricName string) (*m.Metrics, bool) {
	metric, ok := ms.Metrics[metricName]
	if !ok {
		return nil, false
	}
	return &metric, true
}

func (ms *MemoryStorage) SetCounter(model m.Metrics) m.Metrics {
	if metric, ok := ms.Metrics[model.ID]; ok {
		*metric.Delta += *model.Delta
		ms.Metrics[model.ID] = metric
	} else {
		ms.Metrics[model.ID] = m.Metrics{
			ID:    model.ID,
			MType: model.MType,
			Delta: model.Delta,
		}
	}
	ms.Logger.Infoln(
		"hander", "SetCounter",
		"id", model.ID,
		"MType", model.MType,
		"Delta_before", *ms.Metrics[model.ID].Delta,
		"Delta_after", *model.Delta,
	)

	return ms.Metrics[model.ID]
}

func (ms *MemoryStorage) SetGauge(model m.Metrics) bool {
	ms.Metrics[model.ID] = m.Metrics{ID: model.ID, MType: model.MType, Value: model.Value}
	ms.Logger.Infoln(
		"hander", "SetGauge",
		"id", model.ID,
		"MType", model.MType,
		"Value_before", *ms.Metrics[model.ID].Value,
		"Value_after", *model.Value,
	)
	return true
}
