// Package config представляет собой конфигурацию метрик
package config

const (
	TypeMethod = iota + 1
	TypeMetric
	MetricName
	MetricVal
	MinPathLen
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)
