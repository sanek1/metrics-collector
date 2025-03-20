package controller

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	flags "github.com/sanek1/metrics-collector/internal/flags/agent"
	m "github.com/sanek1/metrics-collector/internal/models"
	services "github.com/sanek1/metrics-collector/internal/services/agent"
	l "github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/zap"
)

type Controller struct {
	opt *flags.Options
	l   *l.ZapLogger
	s   *services.Services
}

func NewController(options *flags.Options, logger *l.ZapLogger) *Controller {
	return &Controller{
		opt: options,
		l:   logger,
		s:   services.NewServices(options, logger),
	}
}

func (c *Controller) SendingCounterMetrics(ctx context.Context, pollCount *int64, client *http.Client) {
	metricCounter := m.NewMetricCounter("PollCount", pollCount)
	pollMetricsURL := &url.URL{
		Scheme: "http",
		Host:   c.opt.FlagRunAddr,
		Path:   "/update/counter/PollCount/" + fmt.Sprintf("%d", *pollCount),
	}
	if err := c.s.SendToServerMetric(ctx, client, pollMetricsURL.String(), *metricCounter); err != nil {
		c.l.ErrorCtx(ctx, "message"+err.Error(), zap.String("sendingCounterMetrics", fmt.Sprintf("Error reporting metrics%v", err)))
	}
}

func (c *Controller) SendingGaugeMetrics(ctx context.Context, metrics map[string]float64, client *http.Client) {
	updatesURL := &url.URL{
		Scheme: "http",
		Host:   c.opt.FlagRunAddr,
		Path:   "/updates/",
	}
	metricGauges := m.NewArrMetricGauge(metrics)

	if c.opt.RateLimit != 0 {
		if err := c.s.SendToServerAsync(ctx, client, updatesURL.String(), metricGauges); err != nil {
			c.l.InfoCtx(ctx, "message", zap.String("SendingGaugeMetrics", fmt.Sprintf("Error reporting metrics%v", err)))
		}
		return
	}

	// for _, gauge := range metricGauges {
	// 	metricURL := &url.URL{
	// 		Scheme: "http",
	// 		Host:   c.opt.FlagRunAddr,
	// 		Path:   fmt.Sprintf("/update/gauge/%s/%f", gauge.ID, *gauge.Value),
	// 	}
	// 	if err := c.s.SendToServerMetric(ctx, client, metricURL.String(), gauge); err != nil {
	// 		c.l.InfoCtx(ctx, "message", zap.String("SendingGaugeMetrics", fmt.Sprintf("Error reporting metrics%v", err)))
	// 	}
	// }

	var wg sync.WaitGroup
	concurrency := 10 // Настраиваемое значение
	sem := make(chan struct{}, concurrency)

	for _, gauge := range metricGauges {
		wg.Add(1)
		sem <- struct{}{}

		go func(g m.Metrics) {
			defer func() {
				<-sem
				wg.Done()
			}()

			metricURL := &url.URL{
				Scheme: "http",
				Host:   c.opt.FlagRunAddr,
				Path:   fmt.Sprintf("/update/gauge/%s/%f", gauge.ID, *gauge.Value),
			}
			// metricURL.Path = path.Join(metricURL.Path,
			// 	fmt.Sprintf("/update/gauge/%s/%f", g.ID, *g.Value))

			if err := c.s.SendToServerMetric(
				ctx,
				client,
				metricURL.String(),
				g,
			); err != nil {
				c.l.InfoCtx(ctx, "gauge metric send failed",
					zap.String("metric", g.ID),
					zap.Error(err),
				)
			}
		}(gauge)
	}

	wg.Wait()
}
