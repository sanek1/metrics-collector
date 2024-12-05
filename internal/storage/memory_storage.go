package storage

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/sanek1/metrics-collector/internal/config"
)

type Metric struct {
	Gauge   float64
	Counter int64
	Key     string
}
type MemoryStorage struct {
	Metrics map[string]Metric
	Logger  *log.Logger
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		Metrics: make(map[string]Metric),
		Logger:  log.New(os.Stdout, "Server\t", log.Ldate|log.Ltime),
	}
}

func (ms *MemoryStorage) GetAllMetrics() []string {
	var result []string

	for _, metric := range ms.Metrics {
		if metric.Gauge != 0 || metric.Counter != 0 {
			var value string
			if metric.Key != config.Gauge {
				value = metric.CounterStr()
			} else {
				value = metric.GaugeStr()
			}
			result = append(result, metric.Key+": "+value)
		}
	}
	return result
}

func (ms *MemoryStorage) GetMetrics(key, metricName string) (string, bool) {
	metric, ok := ms.Metrics[metricName]
	if !ok {
		return "", false
	}
	switch key {
	case "gauge":
		return metric.GaugeStr(), true
	case "counter":
		return metric.CounterStr(), true
	}
	return "", false
}

func (ms *MemoryStorage) SetCounter(key string, value float64) string {
	metric, ok := ms.Metrics[key]
	if !ok {
		metric = Metric{Key: key}
	}
	metric.Counter += int64(value)
	ms.Metrics[key] = metric
	ms.Logger.Printf("Update Metric: %s, Counter: %f", key, value)

	return writResult(key, value)
}

func (ms *MemoryStorage) SetGauge(key string, value float64) string {
	ms.Metrics[key] = Metric{Key: key, Gauge: value}
	ms.Logger.Printf("Update Metric: %s, Gauge: %.2f", key, value)

	return writResult(key, value)
}

func (m Metric) GaugeStr() string {
	return fmt.Sprint(m.Gauge)
}

func (m Metric) CounterStr() string {
	return fmt.Sprint(m.Counter)
}

func StrToGauge(input string) (float64, error) {
	return strconv.ParseFloat(input, 64)
}

func writResult(key string, value float64) string {
	parts := []string{"Metric Name: ", key, ", Metric Value:", fmt.Sprint(value)}
	return strings.Join(parts, "")
}
