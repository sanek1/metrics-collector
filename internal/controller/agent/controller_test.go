package controller

import (
	"context"
	"net/http"
	"testing"

	"go.uber.org/zap"

	flags "github.com/sanek1/metrics-collector/internal/flags/agent"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

func TestSendingGaugeMetrics(t *testing.T) {
	ctx := context.Background()
	l, _ := l.NewZapLogger(zap.InfoLevel)

	client := &http.Client{}
	opt := &flags.Options{FlagRunAddr: ":8080"}

	ctrl := NewController(opt, l)

	metrics := map[string]float64{
		"gauge1": 10.56,
		"gauge2": 20.75,
	}

	ctrl.SendingGaugeMetrics(ctx, metrics, client)
}
