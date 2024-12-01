package config

import "time"

type Config struct {
}

const (
	PollInterval   = 2 * time.Second
	ReportInterval = 10 * time.Second

	Address = "localhost"
	Port    = "8080"

	TypeMethod = 1
	TypeMetric = 2
	MetricName = 3
	MetricVal  = 4
	MinPathLen = 5

	Gauge   = "gauge"
	Counter = "counter"
)
