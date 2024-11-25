package storage

import "strconv"

type Metric struct {
	Gauge   float64
	Counter int64
	Name    string
}
type MemStorage struct {
	Metrics map[string]Metric
}

type IMetricStorage interface {
	SetGauge(key string, value float64)
	SetCounter(key string, value int64)
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Metrics: make(map[string]Metric),
	}
}

func (m *MemStorage) SetCounter(key string, counter int64) {
	if metric, ok := m.Metrics[key]; ok {
		m.Metrics[key] = Metric{Name: key, Counter: metric.Counter + counter}
	} else {
		m.Metrics[key] = Metric{Name: key, Counter: counter}
	}
}

func (m *MemStorage) SetGauge(key string, value float64) {
	m.Metrics[key] = Metric{Name: key, Gauge: value}
}

func StrToGauge(input string) (float64, error) {
	return strconv.ParseFloat(input, 64)
}
