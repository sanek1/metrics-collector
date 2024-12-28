package app

import (
	"context"
	"net/http"
	"time"

	"github.com/sanek1/metrics-collector/internal/controller"
	"github.com/sanek1/metrics-collector/internal/flags"
	"github.com/sanek1/metrics-collector/internal/storage"
	l "github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/zap"
)

const (
	countMetrics = 30
)

func Run(opt *flags.Options) error {
	ctx := context.Background()
	logger, _ := InitLogger()
	pollTick, reportTick, metrics := initDataAgent(opt)
	client := &http.Client{}
	var pollCount int64 = 0

	defer func() {
		pollTick.Stop()
		reportTick.Stop()
	}()

	for {
		select {
		case <-pollTick.C:
			pollCount++
			storage.InitPoolMetrics(metrics)
		case <-reportTick.C:
			controller.ReportMetrics(ctx, metrics, &pollCount, client, logger, opt)
		}
	}
}

func InitLogger() (*l.ZapLogger, error) {
	logger, err := l.NewZapLogger(zap.InfoLevel)
	if err != nil {
		return nil, err
	}
	logger.InfoCtx(context.Background(), "agent started ", zap.String("time: ", time.DateTime))
	return logger, nil
}

func initDataAgent(opt *flags.Options) (pollTick, reportTick *time.Ticker, metrics map[string]float64) {
	pollTick = time.NewTicker(time.Duration(opt.PollInterval) * time.Second)
	reportTick = time.NewTicker(time.Duration(opt.ReportInterval) * time.Second)
	metrics = make(map[string]float64, countMetrics)
	return pollTick, reportTick, metrics
}
