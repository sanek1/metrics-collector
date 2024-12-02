package config

const (
	TypeMethod = 1
	TypeMetric = 2
	MetricName = 3
	MetricVal  = 4
	MinPathLen = 5

	Gauge   = "gauge"
	Counter = "counter"
)

type Config struct {
	ADDRESS         string `env:"ADDRESS"`
	REPORT_INTERVAL int64  `env:"TASK_DURATION"`
	POLL_INTERVAL   int64  `env:"POLL_INTERVAL"`
}
