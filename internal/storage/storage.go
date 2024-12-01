package storage

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	c "github.com/sanek1/metrics-collector/internal/config"
)

type Metric struct {
	Gauge   float64
	Counter int64
	Key     string
}
type MemStorage struct {
	Metrics map[string]Metric
	Logger  *log.Logger
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Metrics: make(map[string]Metric),
		Logger:  log.New(os.Stdout, "Server\t", log.Ldate|log.Ltime),
	}
}

func (ms *MemStorage) SetCounter(key string, value float64) string {
	metric, ok := ms.Metrics[key]
	if !ok {
		metric = Metric{Key: key}
	}
	metric.Counter += int64(value)
	ms.Metrics[key] = metric
	ms.Logger.Printf("Update Metric: %s, Counter: %f", key, value)

	return writResult(key, value)
}

func (ms *MemStorage) SetGauge(key string, value float64) string {
	ms.Metrics[key] = Metric{Key: key, Gauge: value}
	ms.Logger.Printf("Update Metric: %s, Gauge: %.2f", key, value)

	return writResult(key, value)
}

func (ms *MemStorage) GetAllMetrics() []string {
	var result []string

	for _, metric := range ms.Metrics {
		if metric.Gauge != 0 || metric.Counter != 0 {
			var value string
			if metric.Key != c.Gauge {
				value = metric.CounterStr()
			} else {
				value = metric.GaugeStr()
			}
			result = append(result, metric.Key+": "+value)
		}
	}
	return result
}

func (ms *MemStorage) GetMetrics(key, metricName, metricValue string) string {
	metric, ok := ms.Metrics[key]
	if !ok {
		return "Not Exists"
	}

	switch metricName {
	case c.Gauge:
		if metric.GaugeStr() == metricValue {
			return "Gauge: " + metric.GaugeStr()
		}
	case c.Counter:
		if metric.CounterStr() == metricValue {
			return "Counter: " + metric.CounterStr()
		}
	}
	return "Not Found"
}

func (m Metric) GaugeStr() string {
	return fmt.Sprintf("%.2f", m.Gauge)
}

func (m Metric) CounterStr() string {
	return fmt.Sprint(m.Counter)
}

func writResult(key string, value float64) string {
	parts := []string{"Metric Name: ", key, ", Metric Value:", fmt.Sprint(value)}
	return strings.Join(parts, "")
}

func StrToGauge(input string) (float64, error) {
	return strconv.ParseFloat(input, 64)
}
