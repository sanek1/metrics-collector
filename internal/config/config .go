package config

import "time"

type Config struct {
}

const (
	PollInterval   = 2 * time.Second
	ReportInterval = 10 * time.Second
)
