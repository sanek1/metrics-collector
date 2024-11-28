package storage

import (
	"log"
	"os"
	"strconv"
)

const (
	Address = "localhost"
	Port    = "8080"
)

type Metric struct {
	Gauge   float64
	Counter int64
	Name    string
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

func (ms *MemStorage) SetCounter(key string, counter int64) {
	metric, ok := ms.Metrics[key]
	if !ok {
		metric = Metric{Name: key}
	}
	metric.Counter += counter
	ms.Metrics[key] = metric
	ms.Logger.Printf("Update Metric: %s, Gauge: %d", key, counter)
}

func (ms *MemStorage) SetGauge(key string, value float64) {
	ms.Metrics[key] = Metric{Name: key, Gauge: value}
	ms.Logger.Printf("Update Metric: %s, Gauge: %.2f", key, value)
}

func StrToGauge(input string) (float64, error) {
	return strconv.ParseFloat(input, 64)
}
