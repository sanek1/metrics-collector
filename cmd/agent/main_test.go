package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sanek1/metrics-collector/internal/services"
	m "github.com/sanek1/metrics-collector/internal/validation"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_reportClient(t *testing.T) {
	value1 := float64(123)
	value2 := float64(-123)
	value3 := float64(0)
	value4 := float64(1.1)

	tests := []m.Metrics{
		{
			ID:    "Alloc",
			MType: "gauge",
			Value: &value1,
			//url:   "/update1/gauge/Alloc/123",
		},
		{
			ID:    "BuckHashSys",
			MType: "gauge",
			Delta: nil,
			Value: &value2,
			//url:   "/up2ate/gauge/BuckHashSys/-123",
		},
		{
			ID:    "Frees",
			MType: "gauge",
			Delta: nil,
			Value: &value3,
			//url:   "/update/gauge/Frees/0",
		},
		{
			ID:    "GCCPUFraction",
			MType: "gauge",
			Delta: nil,
			Value: &value4,
			//url:   "/update/gauge/GCCPUFraction/0",
		},
		{
			ID:    "wrong path",
			MType: "gauge",
			Delta: nil,
			Value: &value4,
			//url:   "/update111/gauge/GCCPUFraction/0",
		},
	}

	testServer := httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(rw, "")
			rw.Header().Set("Content-Type", "application/json")
			rw.Header().Set("content-encoding", "")
			rw.Header().Set("Accept-Encoding", "")
		}),
	)
	defer testServer.Close()
	//logger := log.New(io.Discard, "", log.LstdFlags)
	logger, _ := m.Initialize("info")
	logger.Info("agent started ", zap.String("time: ", time.DateTime))

	for _, tt := range tests {
		t.Run(tt.ID, func(t *testing.T) {
			metricURL := fmt.Sprint(testServer.URL, "/update/gauge/"+tt.ID+"/"+fmt.Sprintf("%f", *tt.Value))
			err := services.SendToServer(&http.Client{}, metricURL, tt, logger)
			assert.Equal(t, err, nil)
		})
	}
}
