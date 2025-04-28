// Package app
package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	ac "github.com/sanek1/metrics-collector/internal/controller/agent"
	af "github.com/sanek1/metrics-collector/internal/flags/agent"
	as "github.com/sanek1/metrics-collector/internal/storage/agent"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

const (
	countMetrics = 30
)

type App struct {
	controller *ac.Controller
	opt        *af.Options
	logger     *l.ZapLogger
}

func New(opt *af.Options) *App {
	logger := startLogger()
	ctrl := ac.NewController(opt, logger)
	return &App{
		controller: ctrl,
		opt:        opt,
	}
}
func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	if err := a.startAgent(ctx); err != nil {
		a.logger.ErrorCtx(ctx, "Error starting agent", zap.Error(err))
		return err
	}
	return nil
}

func (a *App) startAgent(ctx context.Context) error {
	pollTick, reportTick, metrics, gpMetrics := initDataAgent(a.opt)
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: 10 * time.Second,
	}
	var pollCount int64 = 0

	defer func() {
		pollTick.Stop()
		reportTick.Stop()
	}()

	for {
		select {
		case <-pollTick.C:
			pollCount++
			as.InitPoolMetrics(metrics)
			as.GetGopsuiteMetrics(gpMetrics)
		case <-reportTick.C:
			_, _ = fmt.Fprintf(os.Stdout, "--------- start response ---------\n\n")
			a.controller.SendingCounterMetrics(ctx, &pollCount, client)
			a.controller.SendingGaugeMetrics(ctx, metrics, client)
			a.controller.SendingGaugeMetrics(ctx, gpMetrics, client)
			_, _ = fmt.Fprintf(os.Stdout, "--------- end response ---------\n\n")
			_, _ = fmt.Fprintf(os.Stdout, "--------- NEW ITERATION %d ---------> \n\n", pollCount)

		case <-ctx.Done():
			//a.logger.InfoCtx(ctx, "Agent stopped.", zap.String("signal", ctx.Err().Error()))
			if a.controller != nil && gpMetrics != nil {
				a.controller.SendingGaugeMetrics(ctx, gpMetrics, client)
			} else {
				if a.controller == nil {
					//a.logger.WarnCtx(ctx, "controller is nil, cannot send final metrics")
				}
				if gpMetrics == nil {
					//a.logger.WarnCtx(ctx, "gpMetrics is nil, nothing to send")
				}
			}
			//a.logger.InfoCtx(ctx, "Agent stopped.")
			return nil
		}
	}
}

func startLogger() *l.ZapLogger {
	ctx := context.Background()
	logger, err := l.NewZapLogger(zap.DebugLevel)

	if err != nil {
		log.Panic(err)
	}

	_ = logger.WithContextFields(ctx,
		zap.String("agent started", ""))

	defer logger.Sync()
	return logger
}

func initDataAgent(opt *af.Options) (pollTick, reportTick *time.Ticker, metrics, gpMetrics map[string]float64) {
	pollTick = time.NewTicker(time.Duration(opt.PollInterval) * time.Second)
	reportTick = time.NewTicker(time.Duration(opt.ReportInterval) * time.Second)
	metrics = make(map[string]float64, countMetrics)
	gpMetrics = make(map[string]float64, countMetrics)
	return pollTick, reportTick, metrics, gpMetrics
}
