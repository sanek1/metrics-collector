package app

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sanek1/metrics-collector/internal/controller"
	"github.com/sanek1/metrics-collector/internal/flags"
	"github.com/sanek1/metrics-collector/internal/storage"
)

const (
	countMetrics = 30
)

func Run(opt *flags.Options) error {
	logger := log.New(os.Stdout, "Agent\t", log.Ldate|log.Ltime)
	logger.Println("started")

	pollTick := time.NewTicker(time.Duration(opt.PollInterval) * time.Second)
	reportTick := time.NewTicker(time.Duration(opt.ReportInterval) * time.Second)
	defer func() {
		pollTick.Stop()
		reportTick.Stop()
	}()
	client := &http.Client{}

	var pollCount int64 = 0
	metrics := make(map[string]float64, countMetrics)
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
