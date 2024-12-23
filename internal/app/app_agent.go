package app

import (
	"net/http"
	"time"

	"github.com/sanek1/metrics-collector/internal/controller"
	"github.com/sanek1/metrics-collector/internal/flags"
	"github.com/sanek1/metrics-collector/internal/storage"
	v "github.com/sanek1/metrics-collector/internal/validation"
	"go.uber.org/zap"
)

const (
	countMetrics = 30
)

func Run(opt *flags.Options) error {
	logger, _ := initLogger()
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
			controller.ReportMetrics(metrics, &pollCount, client, logger, opt)
		}
	}
}

func initLogger() (*zap.SugaredLogger, error) {
	logger, err := v.Initialize("info")
	if err != nil {
		return nil, err
	}
	logger.Logger.Info("agent started ", zap.String("time: ", time.DateTime))
	return logger.Logger, nil
}

func initDataAgent(opt *flags.Options) (pollTick, reportTick *time.Ticker, metrics map[string]float64) {
	pollTick = time.NewTicker(time.Duration(opt.PollInterval) * time.Second)
	reportTick = time.NewTicker(time.Duration(opt.ReportInterval) * time.Second)
	metrics = make(map[string]float64, countMetrics)
	return pollTick, reportTick, metrics
}
