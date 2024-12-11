package storage

type Storage interface {
	SetGauge(key string, value float64) string
	SetCounter(key string, value float64) string
	GetAllMetrics() []string
	GetMetrics(metricType, metricName string) (string, bool)
}
