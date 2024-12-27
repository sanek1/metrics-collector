package controller

import (
	"context"
	"net/http"
	"testing"

	"github.com/sanek1/metrics-collector/internal/flags"
	l "github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestSendingGaugeMetrics(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := l.NewZapLogger(zap.InfoLevel)

	client := &http.Client{}
	opt := &flags.Options{FlagRunAddr: ":8080"}

	metrics := map[string]float64{
		"gauge1": 10.56,
		"gauge2": 20.75,
	}

	//expectedURL1 := "http://localhost:8080/update/gauge/gauge1/10.500000"
	//expectedURL2 := "http://localhost:8080/update/gauge/gauge2/20.750000"

	SendingGaugeMetrics(ctx, metrics, client, logger, opt.FlagRunAddr)
}
