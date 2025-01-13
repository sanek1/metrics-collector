package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/sanek1/metrics-collector/internal/config"
	m "github.com/sanek1/metrics-collector/internal/models"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

const (
	fileMode = 0600
)

type MetricsStorage struct {
	mtx     sync.RWMutex
	Metrics map[string]m.Metrics
	Logger  *l.ZapLogger
	Errors  []string
}

func NewMetricsStorage(logger *l.ZapLogger) *MetricsStorage {
	return &MetricsStorage{
		Metrics: make(map[string]m.Metrics),
		Logger:  logger,
	}
}

func (ms *MetricsStorage) GetAllMetrics() []string {
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

func (ms *MetricsStorage) GetMetrics(ctx context.Context, key, metricName string) (*m.Metrics, bool) {
	metric, ok := ms.Metrics[metricName]
	if !ok {
		return nil, false
	}
	return &metric, true
}

func (ms *MetricsStorage) SetCounter(ctx context.Context, models ...m.Metrics) ([]*m.Metrics, error) {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	results := make([]*m.Metrics, len(models))
	errors := make([]error, len(models))

	for i, model := range models {
		SetLog(ctx, ms, &model, "SetCounter")
		metric, exists := ms.Metrics[model.ID]
		if exists && metric.Delta != nil {
			*metric.Delta += *model.Delta
		} else {
			metric = m.Metrics{
				ID:    model.ID,
				MType: model.MType,
				Delta: model.Delta,
			}
		}
		ms.Metrics[model.ID] = metric
		results[i] = &metric
	}

	hasErrors := false
	for _, err := range errors {
		if err != nil {
			hasErrors = true
			break
		}
	}

	if hasErrors {
		return results, fmt.Errorf("errors: %v", errors)
	}

	return results, nil
}

func (ms *MetricsStorage) SetGauge(ctx context.Context, models ...m.Metrics) ([]*m.Metrics, error) {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	results := make([]*m.Metrics, len(models))
	errors := make([]error, len(models))

	for i, model := range models {
		SetLog(ctx, ms, &model, "SetGauge")
		ms.Metrics[model.ID] = m.Metrics{ID: model.ID, MType: model.MType, Value: model.Value}
		res := ms.Metrics[model.ID]
		results[i] = &res
		errors[i] = nil
	}

	hasErrors := false
	for _, err := range errors {
		if err != nil {
			hasErrors = true
			break
		}
	}

	if hasErrors {
		return results, fmt.Errorf("errors: %v", errors)
	}

	return results, nil

}

func (ms *MetricsStorage) SaveToFile(fname string) error {
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

func (ms *MetricsStorage) LoadFromFile(filename string) error {
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
