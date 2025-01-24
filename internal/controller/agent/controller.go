package controller

import (
	"context"
	"fmt"
	"net/http"

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

func New(options *flags.Options, logger *l.ZapLogger) *Controller {
	return &Controller{
		opt: options,
		l:   logger,
		s:   services.NewServices(logger),
	}
}

func (c *Controller) SendingCounterMetrics(ctx context.Context, pollCount *int64, client *http.Client) {
	metricURL := fmt.Sprintf("http://%s/update/counter/PollCount/%d", c.opt.FlagRunAddr, *pollCount)
	metricCounter := m.NewMetricCounter("PollCount", pollCount)
	if err := c.s.SendToServer(ctx, client, metricURL, *metricCounter); err != nil {
		c.l.InfoCtx(ctx, "message", zap.String("sendingCounterMetrics", fmt.Sprintf("Error reporting metrics%v", err)))
	}
}

func (c *Controller) SendingGaugeMetrics(ctx context.Context, metrics map[string]float64, client *http.Client) {
	for name, v := range metrics {
		metricURL := fmt.Sprintf("http://%s/update/gauge/%s/%f", c.opt.FlagRunAddr, name, v)
		metricGauge := m.NewMetricGauge(name, &v)
		if err := c.s.SendToServer(ctx, client, metricURL, *metricGauge); err != nil {
			c.l.InfoCtx(ctx, "message", zap.String("SendingGaugeMetrics", fmt.Sprintf("Error reporting metrics%v", err)))
		}
	}
}
