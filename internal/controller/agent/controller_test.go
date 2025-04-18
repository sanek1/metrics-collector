package controller

import (
	"context"
	"net/http"
	"testing"

	"go.uber.org/zap"

	flags "github.com/sanek1/metrics-collector/internal/flags/agent"
	l "github.com/sanek1/metrics-collector/pkg/logging"
	"github.com/stretchr/testify/assert"
)

func TestNewController(t *testing.T) {
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	options := &flags.Options{
		FlagRunAddr:    ":8080",
		CryptoKey:      "test-key",
		ReportInterval: 10,
		PollInterval:   2,
		RateLimit:      2,
	}

	controller := NewController(options, logger)

	assert.NotNil(t, controller)
	assert.Equal(t, options, controller.opt)
	assert.Equal(t, logger, controller.l)
	assert.NotNil(t, controller.s)
}

func TestSendingCounterMetrics(t *testing.T) {
	ctx := context.Background()
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	client := &http.Client{}
	options := &flags.Options{
		FlagRunAddr:    ":8080",
		CryptoKey:      "test-key",
		ReportInterval: 10,
		PollInterval:   2,
		RateLimit:      2,
	}

	controller := NewController(options, logger)

	var pollCount int64 = 42
	controller.SendingCounterMetrics(ctx, &pollCount, client)
}

func TestSendingGaugeMetrics(t *testing.T) {
	ctx := context.Background()
	logger, _ := l.NewZapLogger(zap.InfoLevel)

	client := &http.Client{}
	opt := &flags.Options{FlagRunAddr: ":8080"}
	ctrl := NewController(opt, logger)

	metrics := map[string]float64{
		"gauge1": 10.56,
		"gauge2": 20.75,
	}

	ctrl.SendingGaugeMetrics(ctx, metrics, client)
}

func TestSendingGaugeMetricsWithRateLimit(t *testing.T) {
	ctx := context.Background()
	logger, _ := l.NewZapLogger(zap.InfoLevel)

	client := &http.Client{}
	opt := &flags.Options{
		FlagRunAddr: ":8080",
		RateLimit:   5,
	}

	ctrl := NewController(opt, logger)
	metrics := map[string]float64{
		"gauge1": 10.56,
		"gauge2": 20.75,
		"gauge3": 30.99,
	}
	ctrl.SendingGaugeMetrics(ctx, metrics, client)
}
