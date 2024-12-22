package controller

import (
	"net/http"
	"testing"

	"github.com/sanek1/metrics-collector/internal/flags"
	v "github.com/sanek1/metrics-collector/internal/validation"
	"go.uber.org/mock/gomock"
)

func TestSendingGaugeMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, _ := v.Initialize("info")

	client := &http.Client{}
	opt := &flags.Options{FlagRunAddr: ":8080"}

	metrics := map[string]float64{
		"gauge1": 10.56,
		"gauge2": 20.75,
	}

	//expectedURL1 := "http://localhost:8080/update/gauge/gauge1/10.500000"
	//expectedURL2 := "http://localhost:8080/update/gauge/gauge2/20.750000"

	SendingGaugeMetrics(metrics, client, logger, opt.FlagRunAddr)
}
