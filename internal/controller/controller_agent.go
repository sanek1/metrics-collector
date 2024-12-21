package controller

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sanek1/metrics-collector/internal/flags"
	"github.com/sanek1/metrics-collector/internal/services"
	"github.com/sanek1/metrics-collector/internal/validation"
)

func ReportMetrics(metrics map[string]float64, pollCount *int64, client *http.Client, logger *log.Logger, opt *flags.Options) {
	fmt.Fprintf(os.Stdout, "--------- start response ---------\n\n")
	addr := opt.FlagRunAddr
	metricURL := fmt.Sprint("http://", addr, "/update/counter/PollCount/", *pollCount)
	metrics2 := validation.Metrics{
		ID:    "PollCount",
		MType: "counter",
		Delta: pollCount,
	}
	if err := services.SendToServer(client, metricURL, metrics2, logger); err != nil {
		logger.Printf("Error reporting metrics: %v", err)
	}

	for name, v := range metrics {
		metricURL = fmt.Sprint("http://", addr, "/update/gauge/", name, "/", v)

		metrics3 := validation.Metrics{
			ID:    name,
			MType: "gauge",
			Value: &v,
		}

		if err := services.SendToServer(client, metricURL, metrics3, logger); err != nil {
			logger.Printf("Error reporting metrics: %v", err)
		}
	}
	fmt.Fprintf(os.Stdout, "--------- end response ---------\n\n")
	fmt.Fprintf(os.Stdout, "--------- NEW ITERATION %d ---------> \n\n", *pollCount)
}
