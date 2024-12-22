package controller

import (
	"fmt"
	"net/http"
	"os"

	"github.com/sanek1/metrics-collector/internal/flags"
	"github.com/sanek1/metrics-collector/internal/services"
	"github.com/sanek1/metrics-collector/internal/validation"
	"go.uber.org/zap"
)

func ReportMetrics(metrics map[string]float64, pollCount *int64, client *http.Client, logger *zap.SugaredLogger, opt *flags.Options) {
	addr := opt.FlagRunAddr
	fmt.Fprintf(os.Stdout, "--------- start response ---------\n\n")
	sendingCounterMetrics(pollCount, client, logger, addr)
	SendingGaugeMetrics(metrics, client, logger, addr)
	fmt.Fprintf(os.Stdout, "--------- end response ---------\n\n")
	fmt.Fprintf(os.Stdout, "--------- NEW ITERATION %d ---------> \n\n", *pollCount)
}

func sendingCounterMetrics(pollCount *int64, client *http.Client, logger *zap.SugaredLogger, addr string) {
	metricURL := fmt.Sprintf("http://%s/update/counter/PollCount/%d", addr, *pollCount)
	metricCounter := validation.NewMetricCounter("PollCount", pollCount)
	if err := services.SendToServer(client, metricURL, *metricCounter, logger); err != nil {
		logger.Infoln("Error reporting metrics: %v", err)
	}
}

func SendingGaugeMetrics(metrics map[string]float64, client *http.Client, logger *zap.SugaredLogger, addr string) {
	for name, v := range metrics {
		metricURL := fmt.Sprintf("http://%s/update/gauge/%s/%f", addr, name, v)
		metricGauge := validation.NewMetricGauge(name, &v)
		if err := services.SendToServer(client, metricURL, *metricGauge, logger); err != nil {
			logger.Infoln("Error reporting metrics: %v", err)
		}
	}
}
