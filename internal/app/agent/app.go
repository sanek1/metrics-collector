package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	ac "github.com/sanek1/metrics-collector/internal/controller/agent"
	flags "github.com/sanek1/metrics-collector/internal/flags/agent"
	storage "github.com/sanek1/metrics-collector/internal/storage/agent"
	l "github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/zap"
)

const (
	countMetrics = 30
)

type App struct {
	controller *ac.Controller
	opt        *flags.Options
	logger     *l.ZapLogger
}

func New(opt *flags.Options) *App {
	logger := startLogger()
	ctrl := ac.New(opt, logger)

	return &App{
		controller: ctrl,
		opt:        opt,
		logger:     logger,
	}
}
func (a *App) Run() error {
	ctx := context.Background()

	pollTick, reportTick, metrics := initDataAgent(a.opt)
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
			fmt.Fprintf(os.Stdout, "--------- start response ---------\n\n")
			a.controller.SendingCounterMetrics(ctx, &pollCount, client)
			a.controller.SendingGaugeMetrics(ctx, metrics, client)
			fmt.Fprintf(os.Stdout, "--------- end response ---------\n\n")
			fmt.Fprintf(os.Stdout, "--------- NEW ITERATION %d ---------> \n\n", pollCount)
		}
	}
}

func startLogger() *l.ZapLogger {
	ctx := context.Background()
	logger, err := l.NewZapLogger(zap.InfoLevel)

	if err != nil {
		log.Panic(err)
	}

	_ = logger.WithContextFields(ctx,
		zap.String("agent started", ""))

	defer logger.Sync()
	return logger
}

func initDataAgent(opt *flags.Options) (pollTick, reportTick *time.Ticker, metrics map[string]float64) {
	pollTick = time.NewTicker(time.Duration(opt.PollInterval) * time.Second)
	reportTick = time.NewTicker(time.Duration(opt.ReportInterval) * time.Second)
	metrics = make(map[string]float64, countMetrics)
	return pollTick, reportTick, metrics
}
