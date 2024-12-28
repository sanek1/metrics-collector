package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/sanek1/metrics-collector/internal/config"
	m "github.com/sanek1/metrics-collector/internal/models"
	l "github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/zap"
)

const (
	fileMode = 0600
)

// testcounter [ {"id": "counter1", "type": "counter", "delta": 1, "value": 123.4}]
// testSetGet32 [ {"id": "testSetGet33", "type": "gauge", "delta": 1, "value": 123.4}
type MemoryStorage struct {
	Metrics map[string]m.Metrics
	Logger  *l.ZapLogger
}

// NewMemoryStorage returns a new MemoryStorage instance.
// It creates an empty map of Metrics and a new Logger writing to os.Stdout.
func NewMemoryStorage(logger *l.ZapLogger) *MemoryStorage {
	return &MemoryStorage{
		Metrics: make(map[string]m.Metrics),
		Logger:  logger,
	}
}

func (ms *MemoryStorage) GetAllMetrics() []string {
	result := make([]string, 0, len(ms.Metrics))
	for id, metric := range ms.Metrics {
		var value string
		if metric.MType == config.Counter && metric.Delta != nil {
			if *metric.Delta != 0 {
				value = strconv.FormatInt(*metric.Delta, 10)
			}
		} else if metric.MType == config.Gauge && metric.Value != nil {
			if *metric.Value != 0 {
				value = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
			}
		}
		if value != "" {
			result = append(result, fmt.Sprintf("%s: %s", id, value))
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

func (ms *MemoryStorage) SetCounter(ctx context.Context, model m.Metrics) m.Metrics {
	setLog(ctx, ms, &model, "SetCounter")
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
	return ms.Metrics[model.ID]
}

func (ms *MemoryStorage) SetGauge(ctx context.Context, model m.Metrics) bool {
	setLog(ctx, ms, &model, "SetGauge")
	ms.Metrics[model.ID] = m.Metrics{ID: model.ID, MType: model.MType, Value: model.Value}
	return true
}

func setLog(ctx context.Context, ms *MemoryStorage, model *m.Metrics, name string) {
	ms.Logger.InfoCtx(ctx, name, zap.String(name, fmt.Sprintf("model%s", formatMetric(*model))))
}

func (ms *MemoryStorage) SaveToFile(fname string) error {
	// serialize to json
	data, err := json.MarshalIndent(ms.Metrics, "", "   ")
	if err != nil {
		return err
	}
	// save to file
	err = os.WriteFile(fname, data, fileMode)
	if err != nil {
		return err
	}
	fmt.Printf("Data saved to file: %s\n", fname)
	return nil
}

func (ms *MemoryStorage) LoadFromFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Data file not found. Let's start with empty values.")
			return nil
		}
		return fmt.Errorf("file read error: %v", err)
	}

	err = json.Unmarshal(content, &ms.Metrics)
	if err != nil {
		return fmt.Errorf("data unmarshalling error: %v", err)
	}

	fmt.Println("Previous metric values have been loaded.")
	return nil
}

func formatMetric(model m.Metrics) string {
	data, _ := json.Marshal(model) //
	return string(data)
}
