package controller

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/sanek1/metrics-collector/internal/flags"
	m "github.com/sanek1/metrics-collector/internal/models"
	"github.com/sanek1/metrics-collector/internal/services"
	"github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/zap"
)

func ReportMetrics(ctx context.Context,
	metrics map[string]float64,
	pollCount *int64,
	client *http.Client,
	l *logging.ZapLogger,
	opt *flags.Options) {
	addr := opt.FlagRunAddr
	fmt.Fprintf(os.Stdout, "--------- start response ---------\n\n")
	sendingCounterMetrics(ctx, pollCount, client, l, addr)
	SendingGaugeMetrics(ctx, metrics, client, l, addr)
	fmt.Fprintf(os.Stdout, "--------- end response ---------\n\n")
	fmt.Fprintf(os.Stdout, "--------- NEW ITERATION %d ---------> \n\n", *pollCount)
}

func sendingCounterMetrics(ctx context.Context, pollCount *int64, client *http.Client, l *logging.ZapLogger, addr string) {
	metricURL := fmt.Sprintf("http://%s/update/counter/PollCount/%d", addr, *pollCount)
	metricCounter := m.NewMetricCounter("PollCount", pollCount)
	if err := services.SendToServer(client, metricURL, *metricCounter, l); err != nil {
		l.InfoCtx(ctx, "message", zap.String("sendingCounterMetrics", fmt.Sprintf("Error reporting metrics%v", err)))
	}
}

func SendingGaugeMetrics(ctx context.Context, metrics map[string]float64, client *http.Client, l *logging.ZapLogger, addr string) {
	for name, v := range metrics {
		metricURL := fmt.Sprintf("http://%s/update/gauge/%s/%f", addr, name, v)
		metricGauge := m.NewMetricGauge(name, &v)
		if err := services.SendToServer(client, metricURL, *metricGauge, l); err != nil {
			l.InfoCtx(ctx, "message", zap.String("SendingGaugeMetrics", fmt.Sprintf("Error reporting metrics%v", err)))
		}
	}
}
